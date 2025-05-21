package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/lei/yaml-helm-pipeline/internal/extractor"
	"github.com/lei/yaml-helm-pipeline/internal/git"
	"github.com/lei/yaml-helm-pipeline/internal/github"
	"github.com/lei/yaml-helm-pipeline/internal/helm"
)

// getValueFilesPaths returns a list of value file paths based on environment variable or default path
func getValueFilesPaths(localRepoPath string) []string {
	valueFilesPaths := os.Getenv("VALUE_FILES_PATHS")
	if valueFilesPaths == "" {
		// Default path
		defaultPath := filepath.Join(localRepoPath, "values", "values.yaml")
		log.Printf("Using default values file path: %s", defaultPath)
		return []string{defaultPath}
	}

	// Split by comma
	paths := strings.Split(valueFilesPaths, ",")
	var result []string

	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}

		// Convert to absolute path if it's relative
		if !filepath.IsAbs(path) {
			path = filepath.Join(localRepoPath, path)
		}

		log.Printf("Adding values file path: %s", path)
		result = append(result, path)
	}

	return result
}

// Handler handles API requests
type Handler struct {
	githubService    *github.Service
	helmService      *helm.Service
	gitService       *git.Service
	extractorService *extractor.Service
}

// NewHandler creates a new API handler
func NewHandler(githubService *github.Service, helmService *helm.Service, gitService *git.Service, extractorService *extractor.Service) *Handler {
	return &Handler{
		githubService:    githubService,
		helmService:      helmService,
		gitService:       gitService,
		extractorService: extractorService,
	}
}

// SetupRoutes sets up the API routes
func SetupRoutes(router chi.Router, githubService *github.Service, helmService *helm.Service, gitService *git.Service) {
	extractorService := extractor.NewService()

	handler := NewHandler(githubService, helmService, gitService, extractorService)

	router.Route("/api", func(r chi.Router) {
		r.Get("/branches", handler.ListBranches)
		r.Post("/preview", handler.PreviewChanges)
		r.Post("/commit", handler.CommitChanges)
		r.Get("/health", handler.HealthCheck)
	})
}

// ListBranches lists the branches in the repository
func (h *Handler) ListBranches(w http.ResponseWriter, r *http.Request) {
	branches, err := h.githubService.ListBranches(context.Background())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var branchNames []string
	for _, branch := range branches {
		branchNames = append(branchNames, *branch.Name)
	}

	render.JSON(w, r, map[string]interface{}{
		"branches": branchNames,
	})
}

// PreviewRequest represents a request to preview changes
type PreviewRequest struct {
	Branch string `json:"branch"`
}

// PreviewChanges previews the changes that will be made
func (h *Handler) PreviewChanges(w http.ResponseWriter, r *http.Request) {
	var req PreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Branch == "" {
		http.Error(w, "Branch is required", http.StatusBadRequest)
		return
	}

	// Get repository information
	repo, err := h.githubService.GetRepository(context.Background())
	if err != nil {
		http.Error(w, "Failed to get repository information: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Clone the repository
	repoURL := *repo.CloneURL
	repoOwner := *repo.Owner.Login
	repoName := *repo.Name

	localRepoPath := h.gitService.GetLocalRepoPath(repoOwner, repoName, req.Branch)

	if err := h.gitService.CloneRepository(repoURL, localRepoPath, req.Branch); err != nil {
		http.Error(w, "Failed to clone repository: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Use repository root as chart directory
	chartPath := localRepoPath

	// Check if Chart.yaml exists
	if _, err := os.Stat(filepath.Join(chartPath, "Chart.yaml")); os.IsNotExist(err) {
		http.Error(w, "Chart.yaml not found in repository root", http.StatusInternalServerError)
		return
	}

	// Check if templates directory exists
	if _, err := os.Stat(filepath.Join(chartPath, "templates")); os.IsNotExist(err) {
		http.Error(w, "Templates directory not found", http.StatusInternalServerError)
		return
	}

	// Get value files paths
	valuesPaths := getValueFilesPaths(localRepoPath)
	if len(valuesPaths) == 0 {
		http.Error(w, "No values files found", http.StatusInternalServerError)
		return
	}

	// Generate the YAML using Helm
	yamlOutput, err := h.helmService.TemplateChart(chartPath, valuesPaths)
	if err != nil {
		http.Error(w, "Failed to template chart: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Extract keys from the YAML
	keys, err := h.extractorService.ExtractKeys(yamlOutput)
	if err != nil {
		http.Error(w, "Failed to extract keys: "+err.Error(), http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, map[string]interface{}{
		"keys":   keys,
		"branch": req.Branch,
	})
}

// CommitRequest represents a request to commit changes
type CommitRequest struct {
	Branch  string `json:"branch"`
	Message string `json:"message"`
}

// CommitChanges commits the changes to the repository
func (h *Handler) CommitChanges(w http.ResponseWriter, r *http.Request) {
	var req CommitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Branch == "" {
		http.Error(w, "Branch is required", http.StatusBadRequest)
		return
	}

	if req.Message == "" {
		http.Error(w, "Message is required", http.StatusBadRequest)
		return
	}

	// Get repository information
	repo, err := h.githubService.GetRepository(context.Background())
	if err != nil {
		http.Error(w, "Failed to get repository information: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Clone the repository
	repoURL := *repo.CloneURL
	repoOwner := *repo.Owner.Login
	repoName := *repo.Name

	localRepoPath := h.gitService.GetLocalRepoPath(repoOwner, repoName, req.Branch)

	if err := h.gitService.CloneRepository(repoURL, localRepoPath, req.Branch); err != nil {
		http.Error(w, "Failed to clone repository: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Use repository root as chart directory
	chartPath := localRepoPath

	// Check if Chart.yaml exists
	if _, err := os.Stat(filepath.Join(chartPath, "Chart.yaml")); os.IsNotExist(err) {
		http.Error(w, "Chart.yaml not found in repository root", http.StatusInternalServerError)
		return
	}

	// Check if templates directory exists
	if _, err := os.Stat(filepath.Join(chartPath, "templates")); os.IsNotExist(err) {
		http.Error(w, "Templates directory not found", http.StatusInternalServerError)
		return
	}

	// Get value files paths
	valuesPaths := getValueFilesPaths(localRepoPath)
	if len(valuesPaths) == 0 {
		http.Error(w, "No values files found", http.StatusInternalServerError)
		return
	}

	// Generate the YAML using Helm
	yamlOutput, err := h.helmService.TemplateChart(chartPath, valuesPaths)
	if err != nil {
		http.Error(w, "Failed to template chart: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Determine output repository and path
	outputRepoPath, _, err := h.gitService.GetOutputRepoPath(localRepoPath)
	if err != nil {
		http.Error(w, "Failed to prepare output repository: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get output directory within the repository
	outputDirEnv := os.Getenv("OUTPUT_DIR")
	outputDir := outputRepoPath
	if outputDirEnv != "" {
		outputDir = filepath.Join(outputRepoPath, outputDirEnv)
		// Create directory if it doesn't exist
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			http.Error(w, "Failed to create output directory: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Get output filename
	outputFilename := os.Getenv("OUTPUT_FILENAME")
	if outputFilename == "" {
		outputFilename = "generated.yaml"
	}

	// Check if the file already exists and compare content
	existingContent, err := os.ReadFile(filepath.Join(outputDir, outputFilename))
	fileExists := err == nil
	contentChanged := true

	if fileExists {
		contentChanged = !bytes.Equal(existingContent, yamlOutput)
	}

	// Write the YAML to the output file
	outputPath := filepath.Join(outputDir, outputFilename)
	if err := os.WriteFile(outputPath, yamlOutput, 0644); err != nil {
		http.Error(w, "Failed to write YAML file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare commit message
	commitMessage := req.Message
	if os.Getenv("OUTPUT_REPO_URL") != "" {
		commitMessage = fmt.Sprintf("%s (generated from %s/%s branch: %s)",
			req.Message, *repo.Owner.Login, *repo.Name, req.Branch)
	}

	// Commit and push the changes
	if err := h.gitService.CommitAndPush(outputRepoPath, commitMessage); err != nil {
		http.Error(w, "Failed to commit and push changes: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare response message
	responseMessage := "Changes committed and pushed successfully"
	if !contentChanged && fileExists {
		responseMessage = "No changes detected. The generated content is identical to the existing file."
	}

	render.JSON(w, r, map[string]interface{}{
		"success": true,
		"message": responseMessage,
		"branch":  req.Branch,
	})
}

// HealthCheck checks the health of the API
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Check GitHub authentication
	isAuthenticated := h.githubService.IsAuthenticated(context.Background())

	// Check if Helm is installed
	helmInstalled := true
	cmd := "helm version"
	_, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		helmInstalled = false
	}

	// Check value files configuration
	valueFilesConfig := os.Getenv("VALUE_FILES_PATHS")
	valueFilesConfigured := valueFilesConfig != ""

	// Check output configuration
	outputRepoURL := os.Getenv("OUTPUT_REPO_URL")
	outputRepoConfigured := outputRepoURL != ""
	outputDir := os.Getenv("OUTPUT_DIR")
	outputFilename := os.Getenv("OUTPUT_FILENAME")
	if outputFilename == "" {
		outputFilename = "generated.yaml"
	}

	render.JSON(w, r, map[string]interface{}{
		"status":                 "ok",
		"github_authenticated":   isAuthenticated,
		"helm_installed":         helmInstalled,
		"value_files_configured": valueFilesConfigured,
		"value_files_paths":      valueFilesConfig,
		"output_configured":      outputRepoConfigured,
		"output_repo_url":        outputRepoURL,
		"output_dir":             outputDir,
		"output_filename":        outputFilename,
	})
}

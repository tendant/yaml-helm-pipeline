package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/lei/yaml-helm-pipeline/internal/config"
	"github.com/lei/yaml-helm-pipeline/internal/extractor"
	"github.com/lei/yaml-helm-pipeline/internal/git"
	"github.com/lei/yaml-helm-pipeline/internal/github"
	"github.com/lei/yaml-helm-pipeline/internal/helm"
)

// Helper functions for configuration groups

// findConfigGroup finds a configuration group by name
func (h *Handler) findConfigGroup(name string) (*config.ConfigGroup, error) {
	for _, group := range h.config.Groups {
		if group.Name == name {
			return &group, nil
		}
	}
	return nil, fmt.Errorf("configuration group not found: %s", name)
}

// cloneValuesRepositories clones the values repositories for a configuration group
func (h *Handler) cloneValuesRepositories(group *config.ConfigGroup) ([]string, error) {
	var valuesPaths []string

	for _, valuesRepo := range group.ValuesRepos {
		// Construct the repository URL
		repoURL := config.GetRepoURL(valuesRepo.Owner, valuesRepo.Repo)

		// Create a unique path for this values repository
		valuesRepoPath := filepath.Join(
			os.TempDir(),
			fmt.Sprintf("values-%s-%s-%s", valuesRepo.Owner, valuesRepo.Repo, valuesRepo.Branch),
		)

		// Clone the repository
		if err := h.gitService.CloneRepository(repoURL, valuesRepoPath, valuesRepo.Branch); err != nil {
			return nil, fmt.Errorf("failed to clone values repository %s/%s: %w",
				valuesRepo.Owner, valuesRepo.Repo, err)
		}

		// Add the values file path
		valuesPath := filepath.Join(valuesRepoPath, valuesRepo.Path)
		valuesPaths = append(valuesPaths, valuesPath)
	}

	return valuesPaths, nil
}

// cloneOutputRepository clones the output repository for a configuration group
func (h *Handler) cloneOutputRepository(group *config.ConfigGroup) (string, error) {
	outputRepo := group.OutputRepo

	// Construct the repository URL
	repoURL := config.GetRepoURL(outputRepo.Owner, outputRepo.Repo)

	// Create a unique path for this output repository
	outputRepoPath := filepath.Join(
		os.TempDir(),
		fmt.Sprintf("output-%s-%s-%s", outputRepo.Owner, outputRepo.Repo, outputRepo.Branch),
	)

	// Clone the repository
	if err := h.gitService.CloneRepository(repoURL, outputRepoPath, outputRepo.Branch); err != nil {
		return "", fmt.Errorf("failed to clone output repository %s/%s: %w",
			outputRepo.Owner, outputRepo.Repo, err)
	}

	return outputRepoPath, nil
}

// processConfigGroup processes a configuration group
func (h *Handler) processConfigGroup(
	ctx context.Context,
	groupName string,
	templateRepoBranch string,
	commitMessage string,
	previewOnly bool,
) (map[string]interface{}, error) {
	// Find the configuration group
	group, err := h.findConfigGroup(groupName)
	if err != nil {
		return nil, err
	}

	// Get template repository information
	repo, err := h.githubService.GetRepository(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository information: %w", err)
	}

	// Clone the template repository
	repoURL := *repo.CloneURL
	repoOwner := *repo.Owner.Login
	repoName := *repo.Name

	templateRepoPath := h.gitService.GetLocalRepoPath(repoOwner, repoName, templateRepoBranch)
	if err := h.gitService.CloneRepository(repoURL, templateRepoPath, templateRepoBranch); err != nil {
		return nil, fmt.Errorf("failed to clone template repository: %w", err)
	}

	// Use repository root as chart directory
	chartPath := templateRepoPath

	// Check if Chart.yaml exists
	if _, err := os.Stat(filepath.Join(chartPath, "Chart.yaml")); os.IsNotExist(err) {
		return nil, fmt.Errorf("Chart.yaml not found in repository root")
	}

	// Check if templates directory exists
	if _, err := os.Stat(filepath.Join(chartPath, "templates")); os.IsNotExist(err) {
		return nil, fmt.Errorf("templates directory not found")
	}

	// Clone values repositories and get values files
	valuesPaths, err := h.cloneValuesRepositories(group)
	if err != nil {
		return nil, err
	}

	if len(valuesPaths) == 0 {
		return nil, fmt.Errorf("no values files found for group %s", groupName)
	}

	// Generate the YAML using Helm
	yamlOutput, err := h.helmService.TemplateChart(chartPath, valuesPaths)
	if err != nil {
		return nil, fmt.Errorf("failed to template chart: %w", err)
	}

	// If preview only, compare with existing content
	if previewOnly {
		// Get output filename and path for comparison
		outputFilename := group.OutputRepo.Filename
		if outputFilename == "" {
			outputFilename = "generated.yaml"
		}

		// Clone output repository to get existing content
		outputRepoPath, err := h.cloneOutputRepository(group)
		if err != nil {
			return nil, err
		}

		// Create output directory path
		outputDir := outputRepoPath
		if group.OutputRepo.Path != "" {
			outputDir = filepath.Join(outputRepoPath, group.OutputRepo.Path)
		}

		outputPath := filepath.Join(outputDir, outputFilename)
		existingContent, err := os.ReadFile(outputPath)
		
		var changes map[string]interface{}
		if err != nil {
			// File doesn't exist, extract keys from new content only
			keys, err := h.extractorService.ExtractKeys(yamlOutput)
			if err != nil {
				return nil, fmt.Errorf("failed to extract keys: %w", err)
			}
			changes = map[string]interface{}{
				"all_new": true,
				"keys":    keys,
			}
		} else {
			// File exists, compare old vs new
			changes, err = h.extractorService.CompareYAML(existingContent, yamlOutput)
			if err != nil {
				return nil, fmt.Errorf("failed to compare YAML: %w", err)
			}
		}

		result := map[string]interface{}{
			"changes": changes,
		}
		return result, nil
	}

	// Extract keys from the YAML
	keys, err := h.extractorService.ExtractKeys(yamlOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to extract keys: %w", err)
	}

	result := map[string]interface{}{
		"keys": keys,
	}

	// Clone output repository
	outputRepoPath, err := h.cloneOutputRepository(group)
	if err != nil {
		return nil, err
	}

	// Create output directory if it doesn't exist
	outputDir := outputRepoPath
	if group.OutputRepo.Path != "" {
		outputDir = filepath.Join(outputRepoPath, group.OutputRepo.Path)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Get output filename
	outputFilename := group.OutputRepo.Filename
	if outputFilename == "" {
		outputFilename = "generated.yaml"
	}

	// Check if the file already exists and compare content
	outputPath := filepath.Join(outputDir, outputFilename)
	existingContent, err := os.ReadFile(outputPath)
	fileExists := err == nil
	contentChanged := true

	if fileExists {
		contentChanged = !bytes.Equal(existingContent, yamlOutput)
	}

	// Write the YAML to the output file
	if err := os.WriteFile(outputPath, yamlOutput, 0644); err != nil {
		return nil, fmt.Errorf("failed to write YAML file: %w", err)
	}

	// Prepare commit message
	finalCommitMessage := commitMessage
	if commitMessage != "" {
		finalCommitMessage = fmt.Sprintf("%s (generated from %s/%s branch: %s, group: %s)",
			commitMessage, repoOwner, repoName, templateRepoBranch, groupName)
	}

	// Commit and push the changes
	if err := h.gitService.CommitAndPush(outputRepoPath, finalCommitMessage); err != nil {
		return nil, fmt.Errorf("failed to commit and push changes: %w", err)
	}

	// Prepare response message
	responseMessage := "Changes committed and pushed successfully"
	if !contentChanged && fileExists {
		responseMessage = "No changes detected. The generated content is identical to the existing file."
	}

	result["message"] = responseMessage
	result["content_changed"] = contentChanged

	return result, nil
}

// Handler handles API requests
type Handler struct {
	githubService    *github.Service
	helmService      *helm.Service
	gitService       *git.Service
	extractorService *extractor.Service
	config           *config.Config
}

// NewHandler creates a new API handler
func NewHandler(githubService *github.Service, helmService *helm.Service, gitService *git.Service, extractorService *extractor.Service, config *config.Config) *Handler {
	return &Handler{
		githubService:    githubService,
		helmService:      helmService,
		gitService:       gitService,
		extractorService: extractorService,
		config:           config,
	}
}

// SetupRoutes sets up the API routes
func SetupRoutes(router chi.Router, githubService *github.Service, helmService *helm.Service, gitService *git.Service, config *config.Config) {
	extractorService := extractor.NewService()

	handler := NewHandler(githubService, helmService, gitService, extractorService, config)

	router.Route("/api", func(r chi.Router) {
		r.Get("/branches", handler.ListBranches)
		r.Get("/groups", handler.ListConfigGroups)
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

// ListConfigGroups lists available configuration groups
func (h *Handler) ListConfigGroups(w http.ResponseWriter, r *http.Request) {
	var groupNames []string
	for _, group := range h.config.Groups {
		groupNames = append(groupNames, group.Name)
	}

	render.JSON(w, r, map[string]interface{}{
		"groups": groupNames,
	})
}

// PreviewRequest represents a request to preview changes
type PreviewRequest struct {
	Branch string   `json:"branch"`
	Groups []string `json:"groups"`
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

	// If no groups specified, use all groups
	selectedGroups := req.Groups
	if len(selectedGroups) == 0 {
		for _, group := range h.config.Groups {
			selectedGroups = append(selectedGroups, group.Name)
		}
	}

	if len(selectedGroups) == 0 {
		http.Error(w, "No configuration groups available", http.StatusInternalServerError)
		return
	}

	// Process each selected group
	results := make(map[string]interface{})

	for _, groupName := range selectedGroups {
		result, err := h.processConfigGroup(r.Context(), groupName, req.Branch, "", true)
		if err != nil {
			results[groupName] = map[string]interface{}{
				"error": err.Error(),
			}
		} else {
			results[groupName] = result
		}
	}

	render.JSON(w, r, map[string]interface{}{
		"results": results,
		"branch":  req.Branch,
	})
}

// CommitRequest represents a request to commit changes
type CommitRequest struct {
	Branch  string   `json:"branch"`
	Message string   `json:"message"`
	Groups  []string `json:"groups"`
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

	// If no groups specified, use all groups
	selectedGroups := req.Groups
	if len(selectedGroups) == 0 {
		for _, group := range h.config.Groups {
			selectedGroups = append(selectedGroups, group.Name)
		}
	}

	if len(selectedGroups) == 0 {
		http.Error(w, "No configuration groups available", http.StatusInternalServerError)
		return
	}

	// Process each selected group
	results := make(map[string]interface{})

	for _, groupName := range selectedGroups {
		result, err := h.processConfigGroup(r.Context(), groupName, req.Branch, req.Message, false)
		if err != nil {
			results[groupName] = map[string]interface{}{
				"error": err.Error(),
			}
		} else {
			results[groupName] = result
		}
	}

	render.JSON(w, r, map[string]interface{}{
		"results": results,
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

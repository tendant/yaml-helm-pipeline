package api

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/lei/yaml-helm-pipeline/internal/extractor"
	"github.com/lei/yaml-helm-pipeline/internal/git"
	"github.com/lei/yaml-helm-pipeline/internal/github"
	"github.com/lei/yaml-helm-pipeline/internal/helm"
)

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
func SetupRoutes(router *gin.Engine, githubService *github.Service, helmService *helm.Service, gitService *git.Service) {
	extractorService := extractor.NewService()

	handler := NewHandler(githubService, helmService, gitService, extractorService)

	api := router.Group("/api")
	{
		api.GET("/branches", handler.ListBranches)
		api.POST("/preview", handler.PreviewChanges)
		api.POST("/commit", handler.CommitChanges)
		api.GET("/health", handler.HealthCheck)
	}
}

// ListBranches lists the branches in the repository
func (h *Handler) ListBranches(c *gin.Context) {
	branches, err := h.githubService.ListBranches(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var branchNames []string
	for _, branch := range branches {
		branchNames = append(branchNames, *branch.Name)
	}

	c.JSON(http.StatusOK, gin.H{"branches": branchNames})
}

// PreviewRequest represents a request to preview changes
type PreviewRequest struct {
	Branch string `json:"branch" binding:"required"`
}

// PreviewChanges previews the changes that will be made
func (h *Handler) PreviewChanges(c *gin.Context) {
	var req PreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get repository information
	repo, err := h.githubService.GetRepository(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository information: " + err.Error()})
		return
	}

	// Clone the repository
	repoURL := *repo.CloneURL
	repoOwner := *repo.Owner.Login
	repoName := *repo.Name

	localRepoPath := h.gitService.GetLocalRepoPath(repoOwner, repoName, req.Branch)

	if err := h.gitService.CloneRepository(repoURL, localRepoPath, req.Branch); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clone repository: " + err.Error()})
		return
	}

	// Find the Helm chart and values
	chartPath := filepath.Join(localRepoPath, "chart")
	valuesPath := filepath.Join(localRepoPath, "values", "values.yaml")

	// Check if the files exist
	if _, err := os.Stat(chartPath); os.IsNotExist(err) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Chart directory not found"})
		return
	}

	if _, err := os.Stat(valuesPath); os.IsNotExist(err) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Values file not found"})
		return
	}

	// Generate the YAML using Helm
	yamlOutput, err := h.helmService.TemplateChart(chartPath, valuesPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to template chart: " + err.Error()})
		return
	}

	// Extract keys from the YAML
	keys, err := h.extractorService.ExtractKeys(yamlOutput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to extract keys: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"keys":   keys,
		"branch": req.Branch,
	})
}

// CommitRequest represents a request to commit changes
type CommitRequest struct {
	Branch  string `json:"branch" binding:"required"`
	Message string `json:"message" binding:"required"`
}

// CommitChanges commits the changes to the repository
func (h *Handler) CommitChanges(c *gin.Context) {
	var req CommitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get repository information
	repo, err := h.githubService.GetRepository(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository information: " + err.Error()})
		return
	}

	// Clone the repository
	repoURL := *repo.CloneURL
	repoOwner := *repo.Owner.Login
	repoName := *repo.Name

	localRepoPath := h.gitService.GetLocalRepoPath(repoOwner, repoName, req.Branch)

	if err := h.gitService.CloneRepository(repoURL, localRepoPath, req.Branch); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clone repository: " + err.Error()})
		return
	}

	// Find the Helm chart and values
	chartPath := filepath.Join(localRepoPath, "chart")
	valuesPath := filepath.Join(localRepoPath, "values", "values.yaml")

	// Check if the files exist
	if _, err := os.Stat(chartPath); os.IsNotExist(err) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Chart directory not found"})
		return
	}

	if _, err := os.Stat(valuesPath); os.IsNotExist(err) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Values file not found"})
		return
	}

	// Generate the YAML using Helm
	yamlOutput, err := h.helmService.TemplateChart(chartPath, valuesPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to template chart: " + err.Error()})
		return
	}

	// Write the YAML to a file
	outputPath := filepath.Join(localRepoPath, "generated-secrets.yaml")
	if err := os.WriteFile(outputPath, yamlOutput, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write YAML file: " + err.Error()})
		return
	}

	// Commit and push the changes
	if err := h.gitService.CommitAndPush(localRepoPath, req.Message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit and push changes: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Changes committed and pushed successfully",
		"branch":  req.Branch,
	})
}

// HealthCheck checks the health of the API
func (h *Handler) HealthCheck(c *gin.Context) {
	// Check GitHub authentication
	isAuthenticated := h.githubService.IsAuthenticated(context.Background())

	// Check if Helm is installed
	helmInstalled := true
	cmd := "helm version"
	_, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		helmInstalled = false
	}

	c.JSON(http.StatusOK, gin.H{
		"status":               "ok",
		"github_authenticated": isAuthenticated,
		"helm_installed":       helmInstalled,
	})
}

package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/lei/yaml-helm-pipeline/internal/api"
	"github.com/lei/yaml-helm-pipeline/internal/git"
	"github.com/lei/yaml-helm-pipeline/internal/github"
	"github.com/lei/yaml-helm-pipeline/internal/helm"
)

func main() {
	// Check for required environment variables
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		log.Fatal("GITHUB_TOKEN environment variable is required")
	}

	repoOwner := os.Getenv("REPO_OWNER")
	if repoOwner == "" {
		log.Fatal("REPO_OWNER environment variable is required")
	}

	repoName := os.Getenv("REPO_NAME")
	if repoName == "" {
		log.Fatal("REPO_NAME environment variable is required")
	}

	// Initialize services
	githubService := github.NewService(githubToken, repoOwner, repoName)
	helmService := helm.NewService()
	gitService := git.NewService(githubToken)

	// Initialize router
	router := gin.Default()

	// Setup CORS
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Serve static files for frontend
	router.Static("/", "./frontend/dist")

	// Setup API routes
	api.SetupRoutes(router, githubService, helmService, gitService)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("Server starting on port %s...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

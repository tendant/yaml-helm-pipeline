package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/lei/yaml-helm-pipeline/internal/api"
	"github.com/lei/yaml-helm-pipeline/internal/git"
	"github.com/lei/yaml-helm-pipeline/internal/github"
	"github.com/lei/yaml-helm-pipeline/internal/helm"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

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
	router := chi.NewRouter()

	// Setup middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))

	// Setup CORS
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Serve static files for frontend
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "frontend/dist"))
	FileServer(router, "/", filesDir)

	// Setup API routes
	api.SetupRoutes(router, githubService, helmService, gitService)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.HasSuffix(path, "/") {
		r.Get(path+"*", func(w http.ResponseWriter, r *http.Request) {
			rctx := chi.RouteContext(r.Context())
			pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
			fs := http.StripPrefix(pathPrefix, http.FileServer(root))
			fs.ServeHTTP(w, r)
		})
	} else {
		r.Get(path, func(w http.ResponseWriter, r *http.Request) {
			http.FileServer(root).ServeHTTP(w, r)
		})
	}
}

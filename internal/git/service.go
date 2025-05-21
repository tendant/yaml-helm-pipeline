package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

// Service handles Git operations
type Service struct {
	token string
}

// NewService creates a new Git service
func NewService(token string) *Service {
	return &Service{
		token: token,
	}
}

// CloneRepository clones a repository to a local directory
func (s *Service) CloneRepository(url, directory, branch string) error {
	// Remove directory if it exists
	if _, err := os.Stat(directory); err == nil {
		if err := os.RemoveAll(directory); err != nil {
			return fmt.Errorf("failed to remove existing directory: %w", err)
		}
	}

	// Create directory
	if err := os.MkdirAll(directory, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Clone the repository
	_, err := git.PlainClone(directory, false, &git.CloneOptions{
		URL:           url,
		Progress:      os.Stdout,
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch)),
		SingleBranch:  true,
		Auth: &http.BasicAuth{
			Username: "git", // This can be anything except an empty string
			Password: s.token,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	return nil
}

// CommitAndPush commits changes to a repository and pushes them
func (s *Service) CommitAndPush(repoPath, message string) error {
	// Open the repository
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Get the worktree
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Add all changes
	if err := worktree.AddGlob("."); err != nil {
		return fmt.Errorf("failed to add changes: %w", err)
	}

	// Commit changes
	_, err = worktree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Helm Pipeline",
			Email: "helm-pipeline@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	// Push changes
	err = repo.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: "git", // This can be anything except an empty string
			Password: s.token,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to push changes: %w", err)
	}

	return nil
}

// GetLocalRepoPath returns the path to the local repository
func (s *Service) GetLocalRepoPath(owner, repo, branch string) string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("%s-%s-%s", owner, repo, branch))
}

// CloneOutputRepository clones the output repository if specified
func (s *Service) CloneOutputRepository(outputRepoURL, outputBranch string) (string, error) {
	if outputRepoURL == "" {
		return "", fmt.Errorf("output repository URL is empty")
	}

	// Create a unique directory for the output repository
	outputRepoPath := filepath.Join(os.TempDir(), fmt.Sprintf("output-repo-%s", time.Now().Format("20060102150405")))

	// Clone the repository
	err := s.CloneRepository(outputRepoURL, outputRepoPath, outputBranch)
	if err != nil {
		return "", fmt.Errorf("failed to clone output repository: %w", err)
	}

	return outputRepoPath, nil
}

// GetOutputRepoPath determines the path to use for output files
func (s *Service) GetOutputRepoPath(sourceRepoPath string) (string, string, error) {
	outputRepoURL := os.Getenv("OUTPUT_REPO_URL")
	outputBranch := os.Getenv("OUTPUT_REPO_BRANCH")

	// If no output repo specified, use the source repo
	if outputRepoURL == "" || strings.TrimSpace(outputRepoURL) == "" {
		return sourceRepoPath, "", nil
	}

	// Default to main branch if not specified
	if outputBranch == "" {
		outputBranch = "main"
	}

	// Clone the output repository
	outputRepoPath, err := s.CloneOutputRepository(outputRepoURL, outputBranch)
	if err != nil {
		return "", "", err
	}

	return outputRepoPath, outputBranch, nil
}

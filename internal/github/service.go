package github

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
)

// Service handles GitHub API operations
type Service struct {
	client    *github.Client
	repoOwner string
	repoName  string
}

// NewService creates a new GitHub service
func NewService(token, repoOwner, repoName string) *Service {
	// Create an OAuth2 client with the token
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)

	// Create a GitHub client
	client := github.NewClient(tc)

	return &Service{
		client:    client,
		repoOwner: repoOwner,
		repoName:  repoName,
	}
}

// ListBranches returns a list of branches in the repository
func (s *Service) ListBranches(ctx context.Context) ([]*github.Branch, error) {
	branches, _, err := s.client.Repositories.ListBranches(ctx, s.repoOwner, s.repoName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}
	return branches, nil
}

// GetContents retrieves the contents of a file from the repository
func (s *Service) GetContents(ctx context.Context, path, branch string) ([]byte, error) {
	opts := &github.RepositoryContentGetOptions{
		Ref: branch,
	}

	fileContent, _, _, err := s.client.Repositories.GetContents(
		ctx,
		s.repoOwner,
		s.repoName,
		path,
		opts,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get file contents: %w", err)
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return nil, fmt.Errorf("failed to decode content: %w", err)
	}

	return []byte(content), nil
}

// CreateOrUpdateFile creates or updates a file in the repository
func (s *Service) CreateOrUpdateFile(ctx context.Context, path, branch, message string, content []byte) error {
	// Get the current file to check if it exists and get its SHA
	var sha *string
	fileContent, _, _, err := s.client.Repositories.GetContents(
		ctx,
		s.repoOwner,
		s.repoName,
		path,
		&github.RepositoryContentGetOptions{Ref: branch},
	)
	if err == nil && fileContent != nil {
		sha = fileContent.SHA
	}

	// Create or update the file
	_, _, err = s.client.Repositories.CreateFile(
		ctx,
		s.repoOwner,
		s.repoName,
		path,
		&github.RepositoryContentFileOptions{
			Message: &message,
			Content: content,
			Branch:  &branch,
			SHA:     sha,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create/update file: %w", err)
	}

	return nil
}

// GetRepository returns the repository information
func (s *Service) GetRepository(ctx context.Context) (*github.Repository, error) {
	repo, _, err := s.client.Repositories.Get(ctx, s.repoOwner, s.repoName)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}
	return repo, nil
}

// IsAuthenticated checks if the GitHub token is valid
func (s *Service) IsAuthenticated(ctx context.Context) bool {
	_, resp, err := s.client.Users.Get(ctx, "")
	if err != nil || resp.StatusCode != http.StatusOK {
		return false
	}
	return true
}

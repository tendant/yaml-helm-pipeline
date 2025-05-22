package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Groups []ConfigGroup `yaml:"groups" json:"groups"`
}

// ConfigGroup represents a group of values files and their output destination
type ConfigGroup struct {
	Name        string       `yaml:"name" json:"name"`
	ValuesRepos []ValuesRepo `yaml:"values_repos" json:"values_repos"`
	OutputRepo  OutputRepo   `yaml:"output_repo" json:"output_repo"`
}

// ValuesRepo represents a repository containing values files
type ValuesRepo struct {
	Owner  string `yaml:"owner" json:"owner"`
	Repo   string `yaml:"repo" json:"repo"`
	Path   string `yaml:"path" json:"path"`
	Branch string `yaml:"branch,omitempty" json:"branch,omitempty"` // Optional, defaults to "main"
}

// OutputRepo represents a repository for output files
type OutputRepo struct {
	Owner    string `yaml:"owner" json:"owner"`
	Repo     string `yaml:"repo" json:"repo"`
	Path     string `yaml:"path" json:"path"`         // Path within the repository
	Filename string `yaml:"filename" json:"filename"` // Output filename
	Branch   string `yaml:"branch" json:"branch"`     // Branch to commit to
}

// LoadConfig loads the configuration from a file or environment variables
func LoadConfig(configPath string) (*Config, error) {
	// Try to load from file first
	config, err := loadConfigFromFile(configPath)
	if err == nil {
		return config, nil
	}

	// If file loading failed, try environment variables
	fmt.Printf("Failed to load config from file: %v\n", err)
	fmt.Println("Attempting to load config from environment variables...")

	return loadConfigFromEnv()
}

// loadConfigFromFile loads configuration from a YAML file
func loadConfigFromFile(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = "config.yaml" // Default path
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// loadConfigFromEnv loads configuration from environment variables
func loadConfigFromEnv() (*Config, error) {
	// Try to load from CONFIG_GROUPS environment variable first
	configGroupsEnv := os.Getenv("CONFIG_GROUPS")
	if configGroupsEnv != "" {
		return parseConfigGroupsJSON(configGroupsEnv)
	}

	// Try to load from prefixed environment variables
	config, err := parseConfigGroupsFromPrefixedEnv()
	if err == nil && len(config.Groups) > 0 {
		return config, nil
	}

	// If we get here, no valid configuration was found
	return nil, fmt.Errorf("no valid configuration found in environment variables")
}

// parseConfigGroupsJSON parses the CONFIG_GROUPS JSON environment variable
func parseConfigGroupsJSON(configGroupsEnv string) (*Config, error) {
	var groups []ConfigGroup
	if err := json.Unmarshal([]byte(configGroupsEnv), &groups); err != nil {
		return nil, fmt.Errorf("failed to parse CONFIG_GROUPS JSON: %w", err)
	}

	config := &Config{Groups: groups}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

// parseConfigGroupsFromPrefixedEnv parses configuration from prefixed environment variables
func parseConfigGroupsFromPrefixedEnv() (*Config, error) {
	var config Config

	// Find all environment variables with the CONFIG_GROUP prefix
	for i := 1; ; i++ {
		groupPrefix := fmt.Sprintf("CONFIG_GROUP_%d_", i)
		groupName := os.Getenv(groupPrefix + "NAME")

		// If no name is found for this group, we've reached the end
		if groupName == "" {
			break
		}

		group := ConfigGroup{
			Name: groupName,
		}

		// Parse values repositories
		for j := 1; ; j++ {
			valuesRepoEnv := os.Getenv(fmt.Sprintf("%sVALUES_REPO_%d", groupPrefix, j))
			if valuesRepoEnv == "" {
				break
			}

			valuesRepo, err := parseValuesRepoString(valuesRepoEnv)
			if err != nil {
				return nil, err
			}

			group.ValuesRepos = append(group.ValuesRepos, valuesRepo)
		}

		// Parse output repository
		outputRepoEnv := os.Getenv(groupPrefix + "OUTPUT_REPO")
		if outputRepoEnv != "" {
			outputRepo, err := parseOutputRepoString(outputRepoEnv)
			if err != nil {
				return nil, err
			}
			group.OutputRepo = outputRepo
		}

		config.Groups = append(config.Groups, group)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// parseValuesRepoString parses a string in the format "owner/repo:path:branch"
func parseValuesRepoString(s string) (ValuesRepo, error) {
	parts := strings.Split(s, ":")
	if len(parts) < 2 {
		return ValuesRepo{}, fmt.Errorf("invalid values repo format: %s", s)
	}

	repoPath := strings.Split(parts[0], "/")
	if len(repoPath) != 2 {
		return ValuesRepo{}, fmt.Errorf("invalid repository format: %s", parts[0])
	}

	repo := ValuesRepo{
		Owner: repoPath[0],
		Repo:  repoPath[1],
		Path:  parts[1],
	}

	if len(parts) > 2 {
		repo.Branch = parts[2]
	} else {
		repo.Branch = "main" // Default branch
	}

	return repo, nil
}

// parseOutputRepoString parses a string in the format "owner/repo:path/filename:branch"
func parseOutputRepoString(s string) (OutputRepo, error) {
	parts := strings.Split(s, ":")
	if len(parts) < 2 {
		return OutputRepo{}, fmt.Errorf("invalid output repo format: %s", s)
	}

	repoPath := strings.Split(parts[0], "/")
	if len(repoPath) != 2 {
		return OutputRepo{}, fmt.Errorf("invalid repository format: %s", parts[0])
	}

	// Split the second part into path and filename
	pathParts := strings.Split(parts[1], "/")
	filename := pathParts[len(pathParts)-1]
	path := strings.Join(pathParts[:len(pathParts)-1], "/")

	repo := OutputRepo{
		Owner:    repoPath[0],
		Repo:     repoPath[1],
		Path:     path,
		Filename: filename,
	}

	if len(parts) > 2 {
		repo.Branch = parts[2]
	} else {
		repo.Branch = "main" // Default branch
	}

	return repo, nil
}

// validateConfig validates the configuration
func validateConfig(config *Config) error {
	if len(config.Groups) == 0 {
		return fmt.Errorf("no configuration groups defined")
	}

	for i, group := range config.Groups {
		if group.Name == "" {
			return fmt.Errorf("group %d has no name", i+1)
		}

		if len(group.ValuesRepos) == 0 {
			return fmt.Errorf("group %s has no values repositories", group.Name)
		}

		for j, repo := range group.ValuesRepos {
			if repo.Owner == "" || repo.Repo == "" || repo.Path == "" {
				return fmt.Errorf("group %s, values repo %d has missing fields", group.Name, j+1)
			}

			// Set default branch if not specified
			if repo.Branch == "" {
				config.Groups[i].ValuesRepos[j].Branch = "main"
			}
		}

		// Validate output repo
		if group.OutputRepo.Owner == "" || group.OutputRepo.Repo == "" {
			return fmt.Errorf("group %s has invalid output repository", group.Name)
		}

		// Set default filename if not specified
		if group.OutputRepo.Filename == "" {
			config.Groups[i].OutputRepo.Filename = "generated.yaml"
		}

		// Set default branch if not specified
		if group.OutputRepo.Branch == "" {
			config.Groups[i].OutputRepo.Branch = "main"
		}
	}

	return nil
}

// GetRepoURL returns the GitHub URL for a repository
func GetRepoURL(owner, repo string) string {
	return fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
}

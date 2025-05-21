package helm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Service handles Helm operations
type Service struct{}

// NewService creates a new Helm service
func NewService() *Service {
	return &Service{}
}

// TemplateChart renders a Helm chart with the given values
func (s *Service) TemplateChart(chartPath string, valuesPaths []string) ([]byte, error) {
	// Create a temporary directory for the output
	tempDir, err := ioutil.TempDir("", "helm-template")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Build the helm template command
	args := []string{"template", chartPath, "--output-dir", tempDir}

	// Add each values file
	for _, valuesPath := range valuesPaths {
		// Check if the file exists
		if _, err := os.Stat(valuesPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("values file not found: %s", valuesPath)
		}
		args = append(args, "-f", valuesPath)
	}

	// Run helm template command
	cmd := exec.Command("helm", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("helm template failed: %w, stderr: %s", err, stderr.String())
	}

	// Read all generated files and combine them
	var combinedOutput bytes.Buffer
	err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml" {
			content, err := ioutil.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read generated file: %w", err)
			}
			combinedOutput.WriteString("---\n")
			combinedOutput.Write(content)
			combinedOutput.WriteString("\n")
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to read generated files: %w", err)
	}

	return combinedOutput.Bytes(), nil
}

// ExtractKeys extracts keys from YAML content without their values
func (s *Service) ExtractKeys(yamlContent []byte) (map[string]interface{}, error) {
	var data map[string]interface{}
	if err := yaml.Unmarshal(yamlContent, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	// Extract keys recursively
	result := make(map[string]interface{})
	extractKeysRecursive(data, result)

	return result, nil
}

// CompareYAML compares two YAML contents and returns the keys that will change
func (s *Service) CompareYAML(oldYAML, newYAML []byte) (map[string]interface{}, error) {
	var oldData, newData map[string]interface{}

	if err := yaml.Unmarshal(oldYAML, &oldData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal old YAML: %w", err)
	}

	if err := yaml.Unmarshal(newYAML, &newData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal new YAML: %w", err)
	}

	// Find differences
	diff := make(map[string]interface{})
	findDifferences(oldData, newData, diff, "")

	return diff, nil
}

// extractKeysRecursive extracts keys from a nested map without their values
func extractKeysRecursive(data, result map[string]interface{}) {
	for k, v := range data {
		switch val := v.(type) {
		case map[string]interface{}:
			nestedResult := make(map[string]interface{})
			extractKeysRecursive(val, nestedResult)
			result[k] = nestedResult
		case []interface{}:
			// For arrays, we just indicate they exist but don't show values
			result[k] = "[...]"
		default:
			// For other types, we just indicate they exist but don't show values
			result[k] = "..."
		}
	}
}

// findDifferences finds differences between old and new data
func findDifferences(oldData, newData, diff map[string]interface{}, prefix string) {
	for k, newVal := range newData {
		path := k
		if prefix != "" {
			path = prefix + "." + k
		}

		oldVal, exists := oldData[k]
		if !exists {
			// Key exists in new but not in old
			diff[path] = "added"
			continue
		}

		// Check if values are different
		switch newTyped := newVal.(type) {
		case map[string]interface{}:
			if oldTyped, ok := oldVal.(map[string]interface{}); ok {
				// Recursively check nested maps
				findDifferences(oldTyped, newTyped, diff, path)
			} else {
				// Types are different
				diff[path] = "changed"
			}
		case []interface{}:
			// For arrays, we just indicate they changed
			if _, ok := oldVal.([]interface{}); !ok || !compareArrays(oldVal.([]interface{}), newTyped) {
				diff[path] = "changed"
			}
		default:
			// For other types, check equality
			oldJSON, _ := json.Marshal(oldVal)
			newJSON, _ := json.Marshal(newVal)
			if string(oldJSON) != string(newJSON) {
				diff[path] = "changed"
			}
		}
	}

	// Check for keys in old that don't exist in new
	for k := range oldData {
		path := k
		if prefix != "" {
			path = prefix + "." + k
		}

		if _, exists := newData[k]; !exists {
			diff[path] = "removed"
		}
	}
}

// compareArrays compares two arrays for equality
func compareArrays(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		aJSON, _ := json.Marshal(a[i])
		bJSON, _ := json.Marshal(b[i])
		if string(aJSON) != string(bJSON) {
			return false
		}
	}

	return true
}

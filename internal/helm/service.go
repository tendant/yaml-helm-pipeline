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
func (s *Service) TemplateChart(templatesPath string, valuesPaths []string) ([]byte, error) {
	// Create a temporary directory for the chart
	chartDir, err := ioutil.TempDir("", "helm-chart")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp chart directory: %w", err)
	}
	defer os.RemoveAll(chartDir)

	// Create a templates directory in the chart directory
	chartTemplatesDir := filepath.Join(chartDir, "templates")
	if err := os.Mkdir(chartTemplatesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create templates directory: %w", err)
	}

	// Copy all template files to the chart templates directory
	err = filepath.Walk(templatesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Read the template file
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read template file: %w", err)
		}

		// Get the relative path from the templates directory
		relPath, err := filepath.Rel(templatesPath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Write the template file to the chart templates directory
		destPath := filepath.Join(chartTemplatesDir, relPath)
		destDir := filepath.Dir(destPath)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return fmt.Errorf("failed to create destination directory: %w", err)
		}
		if err := ioutil.WriteFile(destPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write template file: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to copy template files: %w", err)
	}

	// Create a Chart.yaml file
	chartYaml := []byte(`apiVersion: v2
name: generated-chart
description: A Helm chart generated from templates
type: application
version: 0.1.0
appVersion: "1.0.0"
`)
	if err := ioutil.WriteFile(filepath.Join(chartDir, "Chart.yaml"), chartYaml, 0644); err != nil {
		return nil, fmt.Errorf("failed to create Chart.yaml: %w", err)
	}

	// Create a temporary directory for the output
	outputDir, err := ioutil.TempDir("", "helm-output")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp output directory: %w", err)
	}
	defer os.RemoveAll(outputDir)

	// Build the helm template command
	args := []string{"template", chartDir, "--output-dir", outputDir}

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
	isFirstFile := true
	err = filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
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

			// Skip empty files
			if len(bytes.TrimSpace(content)) == 0 {
				return nil
			}

			// Only add separator if it's not the first file
			if !isFirstFile {
				combinedOutput.WriteString("---\n")
			} else {
				isFirstFile = false
			}

			combinedOutput.Write(content)
			combinedOutput.WriteString("\n")
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to read generated files: %w", err)
	}

	// Clean up the output to remove any duplicate separators
	output := combinedOutput.Bytes()
	output = bytes.ReplaceAll(output, []byte("---\n---\n"), []byte("---\n"))

	return output, nil
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

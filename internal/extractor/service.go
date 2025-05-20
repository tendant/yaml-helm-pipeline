package extractor

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// Service handles key extraction from YAML files
type Service struct{}

// NewService creates a new extractor service
func NewService() *Service {
	return &Service{}
}

// ExtractKeys extracts keys from YAML content without their values
func (s *Service) ExtractKeys(yamlContent []byte) (map[string]interface{}, error) {
	var data map[string]interface{}
	if err := yaml.Unmarshal(yamlContent, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	// Extract keys recursively
	result := make(map[string]interface{})
	s.extractKeysRecursive(data, result)

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
	s.findDifferences(oldData, newData, diff, "")

	return diff, nil
}

// extractKeysRecursive extracts keys from a nested map without their values
func (s *Service) extractKeysRecursive(data, result map[string]interface{}) {
	for k, v := range data {
		switch val := v.(type) {
		case map[string]interface{}:
			nestedResult := make(map[string]interface{})
			s.extractKeysRecursive(val, nestedResult)
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
func (s *Service) findDifferences(oldData, newData, diff map[string]interface{}, prefix string) {
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
				s.findDifferences(oldTyped, newTyped, diff, path)
			} else {
				// Types are different
				diff[path] = "changed"
			}
		case []interface{}:
			// For arrays, we just indicate they changed
			if _, ok := oldVal.([]interface{}); !ok || !s.compareArrays(oldVal.([]interface{}), newTyped) {
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
func (s *Service) compareArrays(a, b []interface{}) bool {
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

// FormatChanges formats the changes for display
func (s *Service) FormatChanges(changes map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for path, changeType := range changes {
		result[path] = map[string]interface{}{
			"type": changeType,
		}
	}

	return result
}

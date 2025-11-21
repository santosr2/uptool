// Copyright (c) 2024 santosr2
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package integrations provides metadata about available integrations.
package integrations

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Metadata contains information about an integration.
type Metadata struct {
	DisplayName  string   `yaml:"displayName"`
	Description  string   `yaml:"description"`
	URL          string   `yaml:"url"`
	Category     string   `yaml:"category"`
	FilePatterns []string `yaml:"filePatterns"`
	Datasources  []string `yaml:"datasources"`
	Experimental bool     `yaml:"experimental"`
	Disabled     bool     `yaml:"disabled"`
}

// DatasourceMetadata contains information about a datasource.
type DatasourceMetadata struct {
	Name        string `yaml:"name"`
	URL         string `yaml:"url"`
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
}

// CategoryMetadata contains information about an integration category.
type CategoryMetadata struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// RegistryMetadata represents the full integrations.yaml structure.
type RegistryMetadata struct {
	Integrations map[string]Metadata           `yaml:"integrations"`
	Datasources  map[string]DatasourceMetadata `yaml:"datasources"`
	Categories   map[string]CategoryMetadata   `yaml:"categories"`
	Version      string                        `yaml:"version"`
}

var cachedMetadata *RegistryMetadata

// findRegistryFile searches for integrations.yaml in the current directory and parent directories.
func findRegistryFile() (string, error) {
	// Try current directory first
	if _, err := os.Stat("integrations.yaml"); err == nil {
		return "integrations.yaml", nil
	}

	// Walk up to find repository root (contains go.mod)
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		// Check for integrations.yaml
		registryPath := filepath.Join(dir, "integrations.yaml")
		if _, err := os.Stat(registryPath); err == nil {
			return registryPath, nil
		}

		// Check if we're at repository root (go.mod exists)
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			// At repo root, use this location
			return filepath.Join(dir, "integrations.yaml"), nil
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("integrations.yaml not found")
}

// LoadMetadata loads integration metadata from the integrations.yaml file.
func LoadMetadata() (*RegistryMetadata, error) {
	if cachedMetadata != nil {
		return cachedMetadata, nil
	}

	registryPath, err := findRegistryFile()
	if err != nil {
		return nil, err
	}

	// Validate path for security
	err = ValidateFilePath(registryPath)
	if err != nil {
		return nil, fmt.Errorf("invalid registry path: %w", err)
	}

	data, err := os.ReadFile(registryPath) // #nosec G304 - path is validated above
	if err != nil {
		return nil, fmt.Errorf("reading integrations.yaml: %w", err)
	}

	var metadata RegistryMetadata
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("parsing integrations.yaml: %w", err)
	}

	cachedMetadata = &metadata
	return &metadata, nil
}

// GetMetadata returns metadata for a specific integration.
func GetMetadata(name string) (*Metadata, error) {
	metadata, err := LoadMetadata()
	if err != nil {
		return nil, err
	}

	meta, ok := metadata.Integrations[name]
	if !ok {
		return nil, fmt.Errorf("integration %q not found in registry", name)
	}

	return &meta, nil
}

// ListIntegrations returns a list of all integration names with their metadata.
func ListIntegrations() (map[string]Metadata, error) {
	metadata, err := LoadMetadata()
	if err != nil {
		return nil, err
	}

	return metadata.Integrations, nil
}

// ListByCategory returns integrations grouped by category.
func ListByCategory(category string) (map[string]Metadata, error) {
	metadata, err := LoadMetadata()
	if err != nil {
		return nil, err
	}

	result := make(map[string]Metadata)
	for name, meta := range metadata.Integrations {
		if meta.Category == category {
			result[name] = meta
		}
	}

	return result, nil
}

// IsDisabled checks if an integration is disabled in the registry.
func IsDisabled(name string) bool {
	meta, err := GetMetadata(name)
	if err != nil {
		return false
	}
	return meta.Disabled
}

// IsExperimental checks if an integration is marked as experimental.
func IsExperimental(name string) bool {
	meta, err := GetMetadata(name)
	if err != nil {
		return false
	}
	return meta.Experimental
}

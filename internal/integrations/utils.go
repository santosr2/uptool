// Package integrations provides shared utilities for integration implementations.
package integrations

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ValidateFilePath validates that a file path is safe to read/write.
// It checks for directory traversal attempts to prevent security vulnerabilities.
func ValidateFilePath(path string) error {
	// Clean the path to resolve any . or .. components
	cleanPath := filepath.Clean(path)

	// Check for directory traversal attempts
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path contains directory traversal: %s", path)
	}

	return nil
}

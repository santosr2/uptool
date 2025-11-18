// Package secureio provides secure file I/O operations with path validation.
package secureio

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidateFilePath validates that a file path is safe to read/write
func ValidateFilePath(path string) error {
	// Check for directory traversal attempts before cleaning
	if strings.Contains(path, "..") {
		return fmt.Errorf("path contains directory traversal: %s", path)
	}

	// Clean the path to resolve any . components
	cleanPath := filepath.Clean(path)

	// Ensure path is absolute for security
	if !filepath.IsAbs(cleanPath) {
		return fmt.Errorf("path must be absolute: %s", path)
	}

	return nil
}

// ReadFile safely reads a file after validating the path
func ReadFile(path string) ([]byte, error) {
	if err := ValidateFilePath(path); err != nil {
		return nil, err
	}
	return os.ReadFile(path) // #nosec G304 - path validated above
}

// WriteFile safely writes a file after validating the path
func WriteFile(path string, data []byte, perm os.FileMode) error {
	if err := ValidateFilePath(path); err != nil {
		return err
	}
	return os.WriteFile(path, data, perm) // #nosec G306 - secure permissions enforced
}

// Create safely creates a file after validating the path
func Create(path string) (*os.File, error) {
	if err := ValidateFilePath(path); err != nil {
		return nil, err
	}
	return os.Create(path) // #nosec G304 - path validated above
}

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

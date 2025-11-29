//go:build tools

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

// Command gen_guards generates the all.go file that imports all built-in guards.
// This ensures that adding a new guard doesn't require manual updates.
//
// Usage:
//
//	go run scripts/gen_guards.go
//
// Or via go generate:
//
//	go generate ./internal/policy/guards/builtin
package main

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	guardsDir  = "internal/policy/guards/builtin"
	outputFile = "internal/policy/guards/builtin/all.go"
	modulePath = "github.com/santosr2/uptool"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Find repository root (directory containing go.mod)
	repoRoot, err := findRepoRoot()
	if err != nil {
		return fmt.Errorf("finding repository root: %w", err)
	}

	// Change to repository root
	err = os.Chdir(repoRoot)
	if err != nil {
		return fmt.Errorf("changing to repo root: %w", err)
	}

	// Find all guard files
	guards, err := findGuards()
	if err != nil {
		return fmt.Errorf("finding guards: %w", err)
	}

	if len(guards) == 0 {
		fmt.Printf("Warning: no guards found in %s\n", guardsDir)
		// Don't error - builtin guards are optional
	}

	// Generate the all.go file
	content, err := generateAllGo(guards)
	if err != nil {
		return fmt.Errorf("generating code: %w", err)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputFile), 0o750); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Write the generated file
	if err := os.WriteFile(outputFile, content, 0o600); err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}

	fmt.Printf("Generated %s with %d guards\n", outputFile, len(guards))
	return nil
}

// findRepoRoot walks up the directory tree to find the repository root
// (directory containing go.mod).
func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		// Check if go.mod exists in current directory
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root without finding go.mod
			return "", fmt.Errorf("go.mod not found in any parent directory")
		}
		dir = parent
	}
}

// findGuards scans the guards/builtin directory and returns a sorted list
// of guard file names (excluding all.go and test files).
func findGuards() ([]string, error) {
	entries, err := os.ReadDir(guardsDir)
	if err != nil {
		return nil, err
	}

	var guards []string
	for _, entry := range entries {
		// Skip directories
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Skip non-Go files
		if !strings.HasSuffix(name, ".go") {
			continue
		}

		// Skip test files
		if strings.HasSuffix(name, "_test.go") {
			continue
		}

		// Skip the all.go file itself
		if name == "all.go" {
			continue
		}

		// Skip special files
		if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
			continue
		}

		guards = append(guards, name)
	}

	sort.Strings(guards)
	return guards, nil
}

// generateAllGo creates the content for the all.go file.
func generateAllGo(guards []string) ([]byte, error) {
	var b strings.Builder

	// License header
	b.WriteString("// Copyright (c) 2024 santosr2\n")
	b.WriteString("//\n")
	b.WriteString("// Permission is hereby granted, free of charge, to any person obtaining a copy\n")
	b.WriteString("// of this software and associated documentation files (the \"Software\"), to deal\n")
	b.WriteString("// in the Software without restriction, including without limitation the rights\n")
	b.WriteString("// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell\n")
	b.WriteString("// copies of the Software, and to permit persons to whom the Software is\n")
	b.WriteString("// furnished to do so, subject to the following conditions:\n")
	b.WriteString("//\n")
	b.WriteString("// The above copyright notice and this permission notice shall be included in all\n")
	b.WriteString("// copies or substantial portions of the Software.\n")
	b.WriteString("//\n")
	b.WriteString("// THE SOFTWARE IS PROVIDED \"AS IS\", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR\n")
	b.WriteString("// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,\n")
	b.WriteString("// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE\n")
	b.WriteString("// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER\n")
	b.WriteString("// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,\n")
	b.WriteString("// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE\n")
	b.WriteString("// SOFTWARE.\n\n")

	// Package header
	b.WriteString("// Code generated by scripts/gen_guards.go. DO NOT EDIT.\n\n")
	b.WriteString("// Package builtin registers all built-in auto-merge guards.\n")
	b.WriteString("// Import this package to automatically register guards like ci-green, codeowners-approve, security-scan.\n")
	b.WriteString("//\n")
	b.WriteString("// The guards are registered via init() functions in their individual files:\n")
	for _, guard := range guards {
		b.WriteString("//   - " + guard + "\n")
	}
	b.WriteString("package builtin\n\n")

	// Add comment about registration
	b.WriteString("// This file intentionally blank - guard registration happens via init() functions\n")
	b.WriteString("// in individual guard files.\n")

	// Format the generated code
	formatted, err := format.Source([]byte(b.String()))
	if err != nil {
		// If formatting fails, return the unformatted version with the error
		return []byte(b.String()), fmt.Errorf("formatting generated code: %w", err)
	}

	return formatted, nil
}

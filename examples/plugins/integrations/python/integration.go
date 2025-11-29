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

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/santosr2/uptool/internal/engine"
)

const integrationName = "python"

// Integration implements the engine.Integration interface for Python requirements.txt.
type Integration struct {
	client *PyPIClient
}

// New creates a new Python integration instance.
func New() engine.Integration {
	return &Integration{
		client: NewPyPIClient(),
	}
}

// Name returns the integration identifier.
func (i *Integration) Name() string {
	return integrationName
}

// Detect finds requirements.txt files in the repository.
func (i *Integration) Detect(ctx context.Context, repoRoot string) ([]*engine.Manifest, error) {
	var manifests []*engine.Manifest

	// Walk the repository looking for requirements.txt files
	err := filepath.Walk(repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			// Skip hidden directories and common non-code directories
			name := filepath.Base(path)
			if name != "." && name != ".." && strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			if name == "node_modules" || name == "vendor" || name == "__pycache__" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if this is a requirements.txt file
		basename := filepath.Base(path)
		if basename != "requirements.txt" && !strings.HasPrefix(basename, "requirements-") {
			return nil
		}
		if !strings.HasSuffix(basename, ".txt") {
			return nil
		}

		// Read and parse the requirements file
		content, err := os.ReadFile(path) // #nosec G304 -- path is from filepath.Walk, scoped to repoRoot
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		// Parse dependencies
		deps, err := ParseRequirements(string(content))
		if err != nil {
			return fmt.Errorf("parsing %s: %w", path, err)
		}

		// Skip files with no dependencies
		if len(deps) == 0 {
			return nil
		}

		// Convert pointer slice to value slice
		valueDeps := make([]engine.Dependency, len(deps))
		for i, dep := range deps {
			valueDeps[i] = *dep
		}

		// Create manifest
		manifests = append(manifests, &engine.Manifest{
			Path:         path,
			Type:         integrationName,
			Dependencies: valueDeps,
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("scanning for requirements.txt: %w", err)
	}

	return manifests, nil
}

// Plan generates an update plan for a requirements.txt file.
func (i *Integration) Plan(ctx context.Context, manifest *engine.Manifest, planCtx *engine.PlanContext) (*engine.UpdatePlan, error) {
	var updates []engine.Update

	for _, dep := range manifest.Dependencies {
		// Query PyPI for latest version
		latestVersion, err := i.client.GetLatestVersion(ctx, dep.Name)
		if err != nil {
			// Log warning but continue with other packages
			fmt.Fprintf(os.Stderr, "Warning: failed to get latest version for %s: %v\n", dep.Name, err)
			continue
		}

		// Check if update is needed
		if dep.CurrentVersion != latestVersion {
			// TODO: Calculate semantic version impact
			updates = append(updates, engine.Update{
				Dependency:    dep,
				TargetVersion: latestVersion,
				Impact:        string(engine.ImpactMinor), // Simplified for example
				PolicySource:  planCtx.GetPolicySource(),
			})
		}
	}

	return &engine.UpdatePlan{
		Manifest: manifest,
		Updates:  updates,
	}, nil
}

// Apply executes the update plan by rewriting requirements.txt.
func (i *Integration) Apply(ctx context.Context, plan *engine.UpdatePlan) (*engine.ApplyResult, error) {
	// Read current file content
	content, err := os.ReadFile(plan.Manifest.Path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", plan.Manifest.Path, err)
	}

	// Apply updates to content
	updated := string(content)
	for _, update := range plan.Updates {
		// Replace version in requirements.txt
		// This is a simplified implementation - a production version would be more robust
		oldSpec := fmt.Sprintf("%s==%s", update.Dependency.Name, update.Dependency.CurrentVersion)
		newSpec := fmt.Sprintf("%s==%s", update.Dependency.Name, update.TargetVersion)
		updated = strings.ReplaceAll(updated, oldSpec, newSpec)
	}

	// Write updated content
	if err := os.WriteFile(plan.Manifest.Path, []byte(updated), 0600); err != nil {
		return nil, fmt.Errorf("writing %s: %w", plan.Manifest.Path, err)
	}

	return &engine.ApplyResult{
		Manifest: plan.Manifest,
		Applied:  len(plan.Updates),
	}, nil
}

// Validate checks if a requirements.txt file is valid.
func (i *Integration) Validate(ctx context.Context, manifest *engine.Manifest) error {
	// Read file
	content, err := os.ReadFile(manifest.Path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", manifest.Path, err)
	}

	// Try to parse
	_, err = ParseRequirements(string(content))
	if err != nil {
		return fmt.Errorf("invalid requirements.txt: %w", err)
	}

	return nil
}

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

// Package asdf provides integration for asdf tool version manager.
// It detects and updates .tool-versions files.
//
// Status: EXPERIMENTAL - Version resolution not yet implemented
package asdf

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/santosr2/uptool/internal/engine"
	"github.com/santosr2/uptool/internal/integrations"
)

func init() {
	integrations.Register("asdf", func() engine.Integration {
		return New()
	})
}

const integrationName = "asdf"

// Integration implements the engine.Integration interface for asdf.
type Integration struct{}

// New creates a new asdf integration.
func New() *Integration {
	return &Integration{}
}

// Name returns the integration identifier.
func (i *Integration) Name() string {
	return integrationName
}

// Detect scans for .tool-versions files.
func (i *Integration) Detect(ctx context.Context, repoRoot string) ([]*engine.Manifest, error) {
	var manifests []*engine.Manifest

	err := filepath.Walk(repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories except root
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && path != repoRoot {
			return filepath.SkipDir
		}

		// Skip node_modules
		if info.IsDir() && info.Name() == "node_modules" {
			return filepath.SkipDir
		}

		// Check for .tool-versions
		if !info.IsDir() && info.Name() == ".tool-versions" {
			manifest, err := i.parseManifest(path)
			if err != nil {
				return fmt.Errorf("parse %s: %w", path, err)
			}
			if manifest != nil {
				manifests = append(manifests, manifest)
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk repository: %w", err)
	}

	return manifests, nil
}

// parseManifest parses .tool-versions files.
func (i *Integration) parseManifest(path string) (*engine.Manifest, error) {
	// Validate path for security
	if err := integrations.ValidateFilePath(path); err != nil {
		return nil, err
	}

	content, err := os.ReadFile(path) // #nosec G304 - path is validated above
	if err != nil {
		return nil, err
	}

	manifest := &engine.Manifest{
		Path:         path,
		Type:         "asdf",
		Dependencies: []engine.Dependency{},
		Content:      content,
		Metadata:     map[string]interface{}{},
	}

	return i.parseToolVersions(manifest, content)
}

// parseToolVersions parses .tool-versions format.
func (i *Integration) parseToolVersions(manifest *engine.Manifest, content []byte) (*engine.Manifest, error) {
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue // Skip invalid lines
		}

		tool := parts[0]
		version := parts[1]

		manifest.Dependencies = append(manifest.Dependencies, engine.Dependency{
			Name:           tool,
			CurrentVersion: version,
			Type:           "runtime",
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan file: %w", err)
	}

	return manifest, nil
}

// Plan generates an update plan for asdf tools.
//
// Note: asdf integration is experimental. Version resolution is not implemented because
// each tool (.tool-versions can contain nodejs, python, ruby, terraform, etc.) has its own
// registry and update mechanism. This would require datasources for every possible runtime.
//
// Recommended approach: Use native asdf commands:
//   - asdf plugin update --all     # Update plugin versions
//   - asdf latest --all            # Show latest versions
//   - asdf install <tool> latest   # Install latest version
//
// Future enhancement: Could implement version checking via tool-specific datasources
// (npm registry, python.org, ruby gems, etc.) or by calling asdf native commands.
//
// The planCtx parameter is accepted for interface compatibility but not currently used.
func (i *Integration) Plan(ctx context.Context, manifest *engine.Manifest, planCtx *engine.PlanContext) (*engine.UpdatePlan, error) {
	// Note: planCtx is not currently used as asdf version resolution is not implemented.
	_ = planCtx

	return &engine.UpdatePlan{
		Manifest: manifest,
		Updates:  []engine.Update{},
		Strategy: "native_command", // asdf has native update commands
	}, nil
}

// Apply applies updates to asdf manifest files.
//
// Note: Apply is not implemented for asdf. Use native asdf commands instead:
//   - asdf plugin update --all     # Update all plugins
//   - asdf install <tool> latest   # Install latest version of a tool
//
// To manually update versions in .tool-versions:
//  1. Check available versions: asdf list all <tool>
//  2. Edit .tool-versions with desired versions
//  3. Install: asdf install
func (i *Integration) Apply(ctx context.Context, plan *engine.UpdatePlan) (*engine.ApplyResult, error) {
	if len(plan.Updates) == 0 {
		return &engine.ApplyResult{
			Manifest: plan.Manifest,
			Applied:  0,
			Failed:   0,
		}, nil
	}

	return &engine.ApplyResult{
		Manifest: plan.Manifest,
		Applied:  0,
		Failed:   len(plan.Updates),
		Errors: []string{
			"asdf integration is experimental - automatic apply not supported",
			"Use native asdf commands: 'asdf plugin update --all' and 'asdf install <tool> latest'",
		},
	}, nil
}

// Validate validates an asdf manifest.
func (i *Integration) Validate(ctx context.Context, manifest *engine.Manifest) error {
	// Validation would require asdf to be installed
	// Skip for now
	return nil
}

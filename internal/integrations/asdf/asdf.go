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

// Integration implements the engine.Integration interface for asdf.
type Integration struct{}

// New creates a new asdf integration.
func New() *Integration {
	return &Integration{}
}

// Name returns the integration identifier.
func (i *Integration) Name() string {
	return "asdf"
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
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
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
// TODO: Implement version resolution by querying tool-specific registries
func (i *Integration) Plan(ctx context.Context, manifest *engine.Manifest) (*engine.UpdatePlan, error) {
	// Version resolution not yet implemented
	// Would need to query each tool's registry (nodejs, python, ruby, etc.)
	return &engine.UpdatePlan{
		Manifest: manifest,
		Updates:  []engine.Update{},
		Strategy: "custom_rewrite",
	}, nil
}

// Apply applies updates to asdf manifest files.
func (i *Integration) Apply(ctx context.Context, plan *engine.UpdatePlan) (*engine.ApplyResult, error) {
	if len(plan.Updates) == 0 {
		return &engine.ApplyResult{
			Manifest: plan.Manifest,
			Applied:  0,
			Failed:   0,
		}, nil
	}

	// TODO: Implement file rewriting for .tool-versions
	return &engine.ApplyResult{
		Manifest: plan.Manifest,
		Applied:  0,
		Failed:   len(plan.Updates),
		Errors:   []string{"asdf integration is experimental - apply not yet implemented"},
	}, nil
}

// Validate validates an asdf manifest.
func (i *Integration) Validate(ctx context.Context, manifest *engine.Manifest) error {
	// Validation would require asdf to be installed
	// Skip for now
	return nil
}

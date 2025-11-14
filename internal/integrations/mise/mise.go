// Package mise provides integration for mise tool version manager.
// It detects and updates mise.toml, .mise.toml, and optionally .tool-versions files.
// mise is backward-compatible with asdf's .tool-versions format.
//
// Status: EXPERIMENTAL - Version resolution not yet implemented
package mise

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/santosr2/uptool/internal/engine"
	"github.com/santosr2/uptool/internal/integrations"
)

func init() {
	integrations.Register("mise", func() engine.Integration {
		return New()
	})
}

// Integration implements the engine.Integration interface for mise.
type Integration struct{}

// New creates a new mise integration.
func New() *Integration {
	return &Integration{}
}

// Name returns the integration identifier.
func (i *Integration) Name() string {
	return "mise"
}

// Detect scans for mise.toml and .mise.toml files.
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

		// Check for mise.toml or .mise.toml
		if !info.IsDir() && (info.Name() == "mise.toml" || info.Name() == ".mise.toml") {
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

// parseManifest parses mise.toml or .mise.toml files.
func (i *Integration) parseManifest(path string) (*engine.Manifest, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	manifest := &engine.Manifest{
		Path:         path,
		Type:         "mise",
		Dependencies: []engine.Dependency{},
		Content:      content,
		Metadata:     map[string]interface{}{},
	}

	return i.parseMiseToml(manifest, content)
}

// MiseConfig represents the structure of a mise.toml file.
type MiseConfig struct {
	Tools map[string]interface{} `toml:"tools"`
}

// parseMiseToml parses mise.toml format.
func (i *Integration) parseMiseToml(manifest *engine.Manifest, content []byte) (*engine.Manifest, error) {
	var config MiseConfig
	if err := toml.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("parse toml: %w", err)
	}

	// Extract tools section
	for tool, version := range config.Tools {
		// Version can be a string or a map with additional config
		var versionStr string
		switch v := version.(type) {
		case string:
			versionStr = v
		case map[string]interface{}:
			// If it's a map, look for "version" key
			if ver, ok := v["version"].(string); ok {
				versionStr = ver
			}
		}

		if versionStr != "" {
			manifest.Dependencies = append(manifest.Dependencies, engine.Dependency{
				Name:           tool,
				CurrentVersion: versionStr,
				Type:           "runtime",
			})
		}
	}

	return manifest, nil
}

// Plan generates an update plan for mise tools.
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

// Apply applies updates to mise manifest files.
func (i *Integration) Apply(ctx context.Context, plan *engine.UpdatePlan) (*engine.ApplyResult, error) {
	if len(plan.Updates) == 0 {
		return &engine.ApplyResult{
			Manifest: plan.Manifest,
			Applied:  0,
			Failed:   0,
		}, nil
	}

	// TODO: Implement file rewriting for mise.toml and .mise.toml
	return &engine.ApplyResult{
		Manifest: plan.Manifest,
		Applied:  0,
		Failed:   len(plan.Updates),
		Errors:   []string{"mise integration is experimental - apply not yet implemented"},
	}, nil
}

// Validate validates a mise manifest.
func (i *Integration) Validate(ctx context.Context, manifest *engine.Manifest) error {
	// Validation would require mise to be installed
	// Skip for now
	return nil
}

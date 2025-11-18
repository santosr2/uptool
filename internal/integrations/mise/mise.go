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

const integrationName = "mise"

// Integration implements the engine.Integration interface for mise.
type Integration struct{}

// New creates a new mise integration.
func New() *Integration {
	return &Integration{}
}

// Name returns the integration identifier.
func (i *Integration) Name() string {
	return integrationName
}

// validateFilePath validates that a file path is safe to read/write
func validateFilePath(path string) error {
	// Clean the path to resolve any . or .. components
	cleanPath := filepath.Clean(path)

	// Check for directory traversal attempts
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path contains directory traversal: %s", path)
	}

	return nil
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
	// Validate path for security
	if err := validateFilePath(path); err != nil {
		return nil, err
	}

	content, err := os.ReadFile(path) //nolint:gosec // path validated above
	if err != nil {
		return nil, err
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

// Config represents the structure of a mise.toml file.
type Config struct {
	Tools map[string]interface{} `toml:"tools"`
}

// parseMiseToml parses mise.toml format.
func (i *Integration) parseMiseToml(manifest *engine.Manifest, content []byte) (*engine.Manifest, error) {
	var config Config
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
//
// Note: mise integration is experimental. Version resolution is not implemented because
// each tool (mise.toml can contain nodejs, python, ruby, terraform, etc.) has its own
// registry and update mechanism. This would require datasources for every possible runtime.
//
// Recommended approach: Use native mise commands:
//   - mise upgrade                 # Upgrade all tools to latest versions
//   - mise outdated                # Show outdated tools
//   - mise use <tool>@latest       # Pin to latest version
//
// Future enhancement: Could implement version checking via tool-specific datasources
// (npm registry, python.org, ruby gems, etc.) or by calling mise native commands.
func (i *Integration) Plan(ctx context.Context, manifest *engine.Manifest) (*engine.UpdatePlan, error) {
	return &engine.UpdatePlan{
		Manifest: manifest,
		Updates:  []engine.Update{},
		Strategy: "native_command", // mise has native update commands
	}, nil
}

// Apply applies updates to mise manifest files.
//
// Note: Apply is not implemented for mise. Use native mise commands instead:
//   - mise upgrade                 # Upgrade all tools to latest
//   - mise use <tool>@latest       # Pin specific tool to latest
//
// To manually update versions in mise.toml or .mise.toml:
//  1. Check available versions: mise ls-remote <tool>
//  2. Edit mise.toml with desired versions
//  3. Install: mise install
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
			"mise integration is experimental - automatic apply not supported",
			"Use native mise commands: 'mise upgrade' or 'mise use <tool>@latest'",
		},
	}, nil
}

// Validate validates a mise manifest.
func (i *Integration) Validate(ctx context.Context, manifest *engine.Manifest) error {
	// Validation would require mise to be installed
	// Skip for now
	return nil
}

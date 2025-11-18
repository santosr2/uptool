// Package tflint implements the tflint integration for updating plugin versions in .tflint.hcl files.
// It detects tflint configuration files, parses HCL to extract plugin versions, queries GitHub Releases
// for plugin updates, and rewrites versions while preserving HCL formatting.
package tflint

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"

	"github.com/santosr2/uptool/internal/datasource"
	"github.com/santosr2/uptool/internal/engine"
	"github.com/santosr2/uptool/internal/integrations"
)

func init() {
	integrations.Register("tflint", func() engine.Integration {
		return New()
	})
}

const integrationName = "tflint"

// Integration implements tflint configuration updates.
type Integration struct {
	ds datasource.Datasource
}

// New creates a new tflint integration.
func New() *Integration {
	ds, err := datasource.Get("github-releases")
	if err != nil {
		// Fallback to creating a new instance if not registered
		ds = datasource.NewGitHubDatasource()
	}
	return &Integration{
		ds: ds,
	}
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

// Name returns the integration identifier.
func (i *Integration) Name() string {
	return integrationName
}

// Config represents .tflint.hcl structure.
type Config struct {
	Remain  hcl.Body `hcl:",remain"`
	Plugins []Plugin `hcl:"plugin,block"`
	Rules   []Rule   `hcl:"rule,block"`
}

// Plugin represents a tflint plugin block.
type Plugin struct {
	Remain  hcl.Body `hcl:",remain"`
	Name    string   `hcl:"name,label"`
	Version string   `hcl:"version,optional"`
	Source  string   `hcl:"source,optional"`
	Enabled bool     `hcl:"enabled,optional"`
}

// Rule represents a tflint rule block.
type Rule struct {
	Remain  hcl.Body `hcl:",remain"`
	Name    string   `hcl:"name,label"`
	Enabled bool     `hcl:"enabled,optional"`
}

// Detect finds .tflint.hcl files in the repository.
func (i *Integration) Detect(ctx context.Context, repoRoot string) ([]*engine.Manifest, error) {
	var manifests []*engine.Manifest

	err := filepath.Walk(repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && path != repoRoot {
			return filepath.SkipDir
		}

		if info.Name() == ".tflint.hcl" {
			relPath, err := filepath.Rel(repoRoot, path)
			if err != nil {
				return err
			}

			// Validate path for security
			err = validateFilePath(path)
			if err != nil {
				return err
			}

			content, err := os.ReadFile(path) // #nosec G304 - path is validated above
			if err != nil {
				return err
			}

			var config Config
			if err := hclsimple.Decode(path, content, nil, &config); err != nil {
				return err
			}

			deps := i.extractDependencies(&config)

			manifest := &engine.Manifest{
				Path:         relPath,
				Type:         integrationName,
				Dependencies: deps,
				Content:      content,
				Metadata: map[string]any{
					"plugins_count": len(config.Plugins),
					"rules_count":   len(config.Rules),
				},
			}

			manifests = append(manifests, manifest)
		}

		return nil
	})

	return manifests, err
}

// extractDependencies extracts plugins as dependencies.
func (i *Integration) extractDependencies(config *Config) []engine.Dependency {
	deps := make([]engine.Dependency, 0, len(config.Plugins))

	for _, plugin := range config.Plugins {
		if plugin.Source == "" {
			continue
		}

		deps = append(deps, engine.Dependency{
			Name:           plugin.Source,
			CurrentVersion: plugin.Version,
			Type:           "direct",
			Registry:       "github",
		})
	}

	return deps
}

// Plan determines available updates for tflint plugins.
func (i *Integration) Plan(ctx context.Context, manifest *engine.Manifest) (*engine.UpdatePlan, error) {
	var config Config
	if err := hclsimple.Decode(manifest.Path, manifest.Content, nil, &config); err != nil {
		return nil, fmt.Errorf("parse HCL: %w", err)
	}

	var updates []engine.Update

	for _, plugin := range config.Plugins {
		if plugin.Source == "" || plugin.Version == "" {
			continue
		}

		// Parse GitHub URL to extract owner/repo
		source := plugin.Source
		source = strings.TrimPrefix(source, "https://")
		source = strings.TrimPrefix(source, "http://")
		source = strings.TrimPrefix(source, "github.com/")
		source = strings.TrimSuffix(source, ".git")

		// Validate format (should be owner/repo)
		parts := strings.Split(source, "/")
		if len(parts) < 2 {
			continue
		}
		pkg := parts[0] + "/" + parts[1]

		// Get latest version using datasource
		latest, err := i.ds.GetLatestVersion(ctx, pkg)
		if err != nil {
			continue
		}

		// Compare versions
		if latest != plugin.Version && !strings.HasPrefix(latest, "v"+plugin.Version) {
			updates = append(updates, engine.Update{
				Dependency: engine.Dependency{
					Name:           plugin.Source,
					CurrentVersion: plugin.Version,
					Type:           "direct",
					Registry:       "github",
				},
				TargetVersion: latest,
				Impact:        determineImpact(plugin.Version, latest),
			})
		}
	}

	return &engine.UpdatePlan{
		Manifest: manifest,
		Updates:  updates,
		Strategy: "hcl_rewrite",
	}, nil
}

// Apply executes the update by rewriting the HCL file.
func (i *Integration) Apply(ctx context.Context, plan *engine.UpdatePlan) (*engine.ApplyResult, error) {
	if len(plan.Updates) == 0 {
		return &engine.ApplyResult{
			Manifest: plan.Manifest,
			Applied:  0,
			Failed:   0,
		}, nil
	}

	// Read old content for diff
	// Validate path for security
	if err := validateFilePath(plan.Manifest.Path); err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	oldContent, err := os.ReadFile(plan.Manifest.Path) // #nosec G304 - path is validated above
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	// Parse HCL for writing
	file, diags := hclwrite.ParseConfig(oldContent, plan.Manifest.Path, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, fmt.Errorf("parse HCL for writing: %s", diags.Error())
	}

	// Create update map for quick lookup
	updateMap := make(map[string]string)
	for i := range plan.Updates {
		update := &plan.Updates[i]
		updateMap[update.Dependency.Name] = update.TargetVersion
	}

	applied := 0

	// Parse config to get source values
	var config Config
	if err := hclsimple.Decode(plan.Manifest.Path, oldContent, nil, &config); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// Create source-to-label map
	sourceToLabel := make(map[string]string)
	for _, plugin := range config.Plugins {
		if plugin.Source != "" {
			sourceToLabel[plugin.Source] = plugin.Name
		}
	}

	// Update plugin blocks by label
	for _, block := range file.Body().Blocks() {
		if block.Type() != "plugin" {
			continue
		}

		labels := block.Labels()
		if len(labels) == 0 {
			continue
		}

		pluginLabel := labels[0]

		// Find the source for this plugin label
		var source string
		for _, plugin := range config.Plugins {
			if plugin.Name == pluginLabel {
				source = plugin.Source
				break
			}
		}

		if source == "" {
			continue
		}

		// Check if this plugin needs updating
		if newVersion, ok := updateMap[source]; ok {
			versionAttr := block.Body().GetAttribute("version")
			if versionAttr != nil {
				// Update version
				block.Body().SetAttributeValue("version", cty.StringVal(newVersion))
				applied++
			}
		}
	}

	// Write updated content
	newContent := file.Bytes()
	if err := os.WriteFile(plan.Manifest.Path, newContent, 0o600); err != nil {
		return nil, fmt.Errorf("write config: %w", err)
	}

	// Generate diff
	diff := generateDiff(string(oldContent), string(newContent))

	return &engine.ApplyResult{
		Manifest:     plan.Manifest,
		Applied:      applied,
		Failed:       0,
		ManifestDiff: diff,
	}, nil
}

// Validate checks if the HCL file is valid.
func (i *Integration) Validate(ctx context.Context, manifest *engine.Manifest) error {
	var config Config
	if err := hclsimple.Decode(manifest.Path, manifest.Content, nil, &config); err != nil {
		return fmt.Errorf("invalid HCL: %w", err)
	}
	return nil
}

// determineImpact tries to determine the impact of an update.
func determineImpact(old, newVer string) string {
	// Simple heuristic: if major version changes (v1 -> v2), it's major
	oldParts := strings.Split(strings.TrimPrefix(old, "v"), ".")
	newParts := strings.Split(strings.TrimPrefix(newVer, "v"), ".")

	if len(oldParts) > 0 && len(newParts) > 0 && oldParts[0] != newParts[0] {
		return "major"
	}

	if len(oldParts) > 1 && len(newParts) > 1 && oldParts[1] != newParts[1] {
		return "minor"
	}

	return "patch"
}

// generateDiff creates a simple diff between old and new content.
func generateDiff(old, newContent string) string {
	if old == newContent {
		return ""
	}

	oldLines := strings.Split(old, "\n")
	newLines := strings.Split(newContent, "\n")

	var diff strings.Builder
	diff.WriteString("--- .tflint.hcl\n")
	diff.WriteString("+++ .tflint.hcl\n")

	// Find version lines that changed
	versionRE := regexp.MustCompile(`^\s*version\s*=\s*"([^"]+)"`)

	for i := 0; i < len(oldLines) || i < len(newLines); i++ {
		var oldLine, newLine string
		if i < len(oldLines) {
			oldLine = oldLines[i]
		}
		if i < len(newLines) {
			newLine = newLines[i]
		}

		if oldLine != newLine {
			// Check if it's a version line
			oldMatch := versionRE.MatchString(oldLine)
			newMatch := versionRE.MatchString(newLine)

			if oldMatch || newMatch {
				if oldLine != "" {
					diff.WriteString("- " + oldLine + "\n")
				}
				if newLine != "" {
					diff.WriteString("+ " + newLine + "\n")
				}
			}
		}
	}

	return diff.String()
}

// Package helm implements the Helm chart integration for updating Chart.yaml dependencies.
// It detects Chart.yaml files, queries Helm chart repositories for version updates,
// and rewrites chart dependency versions while preserving YAML structure.
package helm

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/santosr2/uptool/internal/datasource"
	"github.com/santosr2/uptool/internal/engine"
	"github.com/santosr2/uptool/internal/integrations"
	"gopkg.in/yaml.v3"
)

func init() {
	integrations.Register("helm", func() engine.Integration {
		return New()
	})
}

const integrationName = "helm"

// Integration implements helm chart updates.
type Integration struct {
	ds datasource.Datasource
}

// New creates a new helm integration.
func New() *Integration {
	ds, err := datasource.Get("helm")
	if err != nil {
		// Fallback to creating a new instance if not registered
		ds = datasource.NewHelmDatasource()
	}
	return &Integration{
		ds: ds,
	}
}

// Name returns the integration identifier.
func (i *Integration) Name() string {
	return integrationName
}

// Chart represents the structure of Chart.yaml.
type Chart struct {
	APIVersion   string         `yaml:"apiVersion"`
	Name         string         `yaml:"name"`
	Description  string         `yaml:"description"`
	Type         string         `yaml:"type"`
	Version      string         `yaml:"version"`
	AppVersion   string         `yaml:"appVersion"`
	Dependencies []Dependency   `yaml:"dependencies,omitempty"`
	Raw          map[string]any `yaml:",inline"`
}

// Dependency represents a chart dependency.
type Dependency struct {
	Name       string `yaml:"name"`
	Version    string `yaml:"version"`
	Repository string `yaml:"repository"`
	Condition  string `yaml:"condition,omitempty"`
	Tags       string `yaml:"tags,omitempty"`
	Enabled    bool   `yaml:"enabled,omitempty"`
	Alias      string `yaml:"alias,omitempty"`
}

// Detect finds Chart.yaml files in the repository.
func (i *Integration) Detect(ctx context.Context, repoRoot string) ([]*engine.Manifest, error) {
	var manifests []*engine.Manifest

	err := filepath.Walk(repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip hidden directories
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && path != repoRoot {
			return filepath.SkipDir
		}

		if info.Name() == "Chart.yaml" {
			relPath, err := filepath.Rel(repoRoot, path)
			if err != nil {
				return nil
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			var chart Chart
			if err := yaml.Unmarshal(content, &chart); err != nil {
				return nil // Skip invalid YAML
			}

			deps := i.extractDependencies(&chart)

			manifest := &engine.Manifest{
				Path:         relPath,
				Type:         integrationName,
				Dependencies: deps,
				Content:      content,
				Metadata: map[string]any{
					"chart_name":    chart.Name,
					"chart_version": chart.Version,
					"deps_count":    len(chart.Dependencies),
				},
			}

			manifests = append(manifests, manifest)
		}

		return nil
	})

	return manifests, err
}

// extractDependencies extracts chart dependencies.
func (i *Integration) extractDependencies(chart *Chart) []engine.Dependency {
	var deps []engine.Dependency

	for _, dep := range chart.Dependencies {
		// Skip OCI repositories for now
		if strings.HasPrefix(dep.Repository, "oci://") {
			continue
		}

		// Skip local dependencies
		if strings.HasPrefix(dep.Repository, "file://") {
			continue
		}

		deps = append(deps, engine.Dependency{
			Name:           dep.Name,
			CurrentVersion: dep.Version,
			Type:           "chart",
			Registry:       dep.Repository,
		})
	}

	return deps
}

// Plan determines available updates for helm chart dependencies.
func (i *Integration) Plan(ctx context.Context, manifest *engine.Manifest) (*engine.UpdatePlan, error) {
	var updates []engine.Update

	for _, dep := range manifest.Dependencies {
		// Get latest version from chart repository
		// Datasource expects format: "repository_url|chart_name"
		pkg := fmt.Sprintf("%s|%s", dep.Registry, dep.Name)
		latest, err := i.ds.GetLatestVersion(ctx, pkg)
		if err != nil {
			// Skip charts we can't query
			continue
		}

		// Compare versions
		if latest != dep.CurrentVersion {
			updates = append(updates, engine.Update{
				Dependency:    dep,
				TargetVersion: latest,
				Impact:        determineImpact(dep.CurrentVersion, latest),
			})
		}
	}

	return &engine.UpdatePlan{
		Manifest: manifest,
		Updates:  updates,
		Strategy: "yaml_rewrite",
	}, nil
}

// Apply executes the update by rewriting Chart.yaml.
func (i *Integration) Apply(ctx context.Context, plan *engine.UpdatePlan) (*engine.ApplyResult, error) {
	if len(plan.Updates) == 0 {
		return &engine.ApplyResult{
			Manifest: plan.Manifest,
			Applied:  0,
			Failed:   0,
		}, nil
	}

	// Read old content for diff
	oldContent, err := os.ReadFile(plan.Manifest.Path)
	if err != nil {
		return nil, fmt.Errorf("read Chart.yaml: %w", err)
	}

	// Parse Chart.yaml
	var chart Chart
	if err := yaml.Unmarshal(oldContent, &chart); err != nil {
		return nil, fmt.Errorf("parse Chart.yaml: %w", err)
	}

	// Create update map for quick lookup
	updateMap := make(map[string]string)
	for _, update := range plan.Updates {
		updateMap[update.Dependency.Name] = update.TargetVersion
	}

	applied := 0

	// Update dependency versions
	for i := range chart.Dependencies {
		if newVersion, ok := updateMap[chart.Dependencies[i].Name]; ok {
			chart.Dependencies[i].Version = newVersion
			applied++
		}
	}

	// Marshal updated chart
	newContent, err := yaml.Marshal(&chart)
	if err != nil {
		return nil, fmt.Errorf("marshal Chart.yaml: %w", err)
	}

	// Write updated content
	if err := os.WriteFile(plan.Manifest.Path, newContent, 0644); err != nil {
		return nil, fmt.Errorf("write Chart.yaml: %w", err)
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

// Validate checks if the Chart.yaml is valid.
func (i *Integration) Validate(ctx context.Context, manifest *engine.Manifest) error {
	var chart Chart
	if err := yaml.Unmarshal(manifest.Content, &chart); err != nil {
		return fmt.Errorf("invalid Chart.yaml: %w", err)
	}

	if chart.APIVersion == "" {
		return fmt.Errorf("Chart.yaml missing apiVersion")
	}

	if chart.Name == "" {
		return fmt.Errorf("Chart.yaml missing name")
	}

	if chart.Version == "" {
		return fmt.Errorf("Chart.yaml missing version")
	}

	return nil
}

// determineImpact tries to determine the impact of an update.
func determineImpact(old, new string) string {
	// Simple heuristic: if major version changes (v1 -> v2), it's major
	oldParts := strings.Split(strings.TrimPrefix(old, "v"), ".")
	newParts := strings.Split(strings.TrimPrefix(new, "v"), ".")

	if len(oldParts) > 0 && len(newParts) > 0 && oldParts[0] != newParts[0] {
		return "major"
	}

	if len(oldParts) > 1 && len(newParts) > 1 && oldParts[1] != newParts[1] {
		return "minor"
	}

	return "patch"
}

// generateDiff creates a simple diff between old and new content.
func generateDiff(old, new string) string {
	if old == new {
		return ""
	}

	oldLines := strings.Split(old, "\n")
	newLines := strings.Split(new, "\n")

	var diff strings.Builder
	diff.WriteString("--- Chart.yaml\n")
	diff.WriteString("+++ Chart.yaml\n")

	maxLines := len(oldLines)
	if len(newLines) > maxLines {
		maxLines = len(newLines)
	}

	for i := 0; i < maxLines; i++ {
		var oldLine, newLine string
		if i < len(oldLines) {
			oldLine = oldLines[i]
		}
		if i < len(newLines) {
			newLine = newLines[i]
		}

		if oldLine != newLine {
			// Check if it's a version line
			if strings.Contains(oldLine, "version:") || strings.Contains(newLine, "version:") {
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

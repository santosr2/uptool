// Package npm implements the npm integration for updating package.json dependencies.
// It detects package.json files, queries the npm registry for version updates,
// and rewrites dependency versions while preserving constraint prefixes (^, ~, >=).
package npm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/santosr2/uptool/internal/engine"
	"github.com/santosr2/uptool/internal/integrations"
	"github.com/santosr2/uptool/internal/registry"
)

func init() {
	integrations.Register("npm", func() engine.Integration {
		return New()
	})
}

// Integration implements npm package.json updates.
type Integration struct {
	client *registry.NPMClient
}

// New creates a new npm integration.
func New() *Integration {
	return &Integration{
		client: registry.NewNPMClient(),
	}
}

// Name returns the integration identifier.
func (i *Integration) Name() string {
	return "npm"
}

// PackageJSON represents the structure of package.json.
type PackageJSON struct {
	Name                 string            `json:"name,omitempty"`
	Version              string            `json:"version,omitempty"`
	Dependencies         map[string]string `json:"dependencies,omitempty"`
	DevDependencies      map[string]string `json:"devDependencies,omitempty"`
	PeerDependencies     map[string]string `json:"peerDependencies,omitempty"`
	OptionalDependencies map[string]string `json:"optionalDependencies,omitempty"`
}

// Detect finds package.json files in the repository.
func (i *Integration) Detect(ctx context.Context, repoRoot string) ([]*engine.Manifest, error) {
	var manifests []*engine.Manifest

	err := filepath.Walk(repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip node_modules directories
		if info.IsDir() && info.Name() == "node_modules" {
			return filepath.SkipDir
		}

		// Skip hidden directories
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
			return filepath.SkipDir
		}

		if info.Name() == "package.json" {
			relPath, err := filepath.Rel(repoRoot, path)
			if err != nil {
				return nil
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			var pkg PackageJSON
			if err := json.Unmarshal(content, &pkg); err != nil {
				return nil // Skip invalid JSON
			}

			deps := i.extractDependencies(&pkg)

			manifest := &engine.Manifest{
				Path:         relPath,
				Type:         "npm",
				Dependencies: deps,
				Content:      content,
				Metadata: map[string]interface{}{
					"package_name": pkg.Name,
				},
			}

			manifests = append(manifests, manifest)
		}

		return nil
	})

	return manifests, err
}

// extractDependencies extracts all dependencies from package.json.
func (i *Integration) extractDependencies(pkg *PackageJSON) []engine.Dependency {
	var deps []engine.Dependency

	for name, version := range pkg.Dependencies {
		deps = append(deps, engine.Dependency{
			Name:           name,
			CurrentVersion: version,
			Constraint:     version,
			Type:           "direct",
			Registry:       "npm",
		})
	}

	for name, version := range pkg.DevDependencies {
		deps = append(deps, engine.Dependency{
			Name:           name,
			CurrentVersion: version,
			Constraint:     version,
			Type:           "dev",
			Registry:       "npm",
		})
	}

	for name, version := range pkg.PeerDependencies {
		deps = append(deps, engine.Dependency{
			Name:           name,
			CurrentVersion: version,
			Constraint:     version,
			Type:           "peer",
			Registry:       "npm",
		})
	}

	for name, version := range pkg.OptionalDependencies {
		deps = append(deps, engine.Dependency{
			Name:           name,
			CurrentVersion: version,
			Constraint:     version,
			Type:           "optional",
			Registry:       "npm",
		})
	}

	return deps
}

// Plan determines available updates for npm dependencies.
func (i *Integration) Plan(ctx context.Context, manifest *engine.Manifest) (*engine.UpdatePlan, error) {
	var updates []engine.Update

	for _, dep := range manifest.Dependencies {
		// Skip file: and link: dependencies
		if strings.HasPrefix(dep.Constraint, "file:") || strings.HasPrefix(dep.Constraint, "link:") {
			continue
		}

		// Skip git URLs
		if strings.Contains(dep.Constraint, "git") || strings.HasPrefix(dep.Constraint, "http") {
			continue
		}

		// Get the latest version
		latest, err := i.client.GetLatestVersion(ctx, dep.Name)
		if err != nil {
			// Skip packages that can't be resolved
			continue
		}

		// Check if update is needed
		if i.needsUpdate(dep.CurrentVersion, latest) {
			impact := i.determineImpact(dep.CurrentVersion, latest)

			updates = append(updates, engine.Update{
				Dependency:    dep,
				TargetVersion: latest,
				Impact:        impact,
				ChangelogURL:  fmt.Sprintf("https://www.npmjs.com/package/%s", dep.Name),
			})
		}
	}

	return &engine.UpdatePlan{
		Manifest: manifest,
		Updates:  updates,
		Strategy: "custom_rewrite", // We rewrite package.json directly
	}, nil
}

// needsUpdate checks if an update is needed.
func (i *Integration) needsUpdate(current, latest string) bool {
	// Remove npm constraint prefixes
	currentClean := strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(current, "^"), "~"), ">=")
	currentClean = strings.TrimSpace(currentClean)

	// Try to parse as semver
	currentVer, err1 := semver.NewVersion(currentClean)
	latestVer, err2 := semver.NewVersion(latest)

	if err1 != nil || err2 != nil {
		// If parsing fails, just compare strings
		return current != latest
	}

	return latestVer.GreaterThan(currentVer)
}

// determineImpact calculates the update impact.
func (i *Integration) determineImpact(current, target string) string {
	currentClean := strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(current, "^"), "~"), ">=")
	currentClean = strings.TrimSpace(currentClean)

	currentVer, err1 := semver.NewVersion(currentClean)
	targetVer, err2 := semver.NewVersion(target)

	if err1 != nil || err2 != nil {
		return "unknown"
	}

	if targetVer.Major() > currentVer.Major() {
		return "major"
	}
	if targetVer.Minor() > currentVer.Minor() {
		return "minor"
	}
	return "patch"
}

// Apply executes the update plan by rewriting package.json.
func (i *Integration) Apply(ctx context.Context, plan *engine.UpdatePlan) (*engine.ApplyResult, error) {
	if len(plan.Updates) == 0 {
		return &engine.ApplyResult{
			Manifest: plan.Manifest,
			Applied:  0,
			Failed:   0,
		}, nil
	}

	// Read the current package.json
	fullPath := plan.Manifest.Path
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("read package.json: %w", err)
	}

	var pkg PackageJSON
	if err := json.Unmarshal(content, &pkg); err != nil {
		return nil, fmt.Errorf("parse package.json: %w", err)
	}

	oldContent := string(content)
	applied := 0

	// Apply updates
	for _, update := range plan.Updates {
		if i.updateDependency(&pkg, update) {
			applied++
		}
	}

	// Write back to package.json with formatting
	newContent, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal package.json: %w", err)
	}

	// Add trailing newline
	newContent = append(newContent, '\n')

	if err := os.WriteFile(fullPath, newContent, 0644); err != nil {
		return nil, fmt.Errorf("write package.json: %w", err)
	}

	// Generate diff
	diff := generateDiff(oldContent, string(newContent))

	return &engine.ApplyResult{
		Manifest:     plan.Manifest,
		Applied:      applied,
		Failed:       len(plan.Updates) - applied,
		ManifestDiff: diff,
	}, nil
}

// updateDependency updates a dependency in the package.json structure.
func (i *Integration) updateDependency(pkg *PackageJSON, update engine.Update) bool {
	name := update.Dependency.Name
	newVersion := update.TargetVersion

	// Preserve constraint prefix (^, ~, >=)
	prefix := ""
	oldVersion := update.Dependency.CurrentVersion
	if strings.HasPrefix(oldVersion, "^") {
		prefix = "^"
	} else if strings.HasPrefix(oldVersion, "~") {
		prefix = "~"
	} else if strings.HasPrefix(oldVersion, ">=") {
		prefix = ">="
	}

	newVersionWithPrefix := prefix + newVersion

	// Update in the appropriate section
	switch update.Dependency.Type {
	case "direct":
		if pkg.Dependencies != nil {
			if _, ok := pkg.Dependencies[name]; ok {
				pkg.Dependencies[name] = newVersionWithPrefix
				return true
			}
		}
	case "dev":
		if pkg.DevDependencies != nil {
			if _, ok := pkg.DevDependencies[name]; ok {
				pkg.DevDependencies[name] = newVersionWithPrefix
				return true
			}
		}
	case "peer":
		if pkg.PeerDependencies != nil {
			if _, ok := pkg.PeerDependencies[name]; ok {
				pkg.PeerDependencies[name] = newVersionWithPrefix
				return true
			}
		}
	case "optional":
		if pkg.OptionalDependencies != nil {
			if _, ok := pkg.OptionalDependencies[name]; ok {
				pkg.OptionalDependencies[name] = newVersionWithPrefix
				return true
			}
		}
	}

	return false
}

// Validate runs npm validation (optional).
func (i *Integration) Validate(ctx context.Context, manifest *engine.Manifest) error {
	// Could run `npm install --package-lock-only` to validate
	// For now, just check if package.json is valid JSON
	var pkg PackageJSON
	return json.Unmarshal(manifest.Content, &pkg)
}

// generateDiff creates a simple diff between old and new content.
func generateDiff(old, new string) string {
	if old == new {
		return ""
	}

	oldLines := strings.Split(old, "\n")
	newLines := strings.Split(new, "\n")

	var diff strings.Builder
	diff.WriteString("--- package.json\n")
	diff.WriteString("+++ package.json\n")

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
			if oldLine != "" {
				diff.WriteString("- " + oldLine + "\n")
			}
			if newLine != "" {
				diff.WriteString("+ " + newLine + "\n")
			}
		}
	}

	return diff.String()
}

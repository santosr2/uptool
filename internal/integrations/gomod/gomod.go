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

// Package gomod implements the Go modules integration for updating go.mod dependencies.
// It detects go.mod files, queries the Go module proxy for version updates,
// and rewrites dependency versions while preserving the go.mod format.
package gomod

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/santosr2/uptool/internal/datasource"
	"github.com/santosr2/uptool/internal/engine"
	"github.com/santosr2/uptool/internal/integrations"
	"github.com/santosr2/uptool/internal/resolve"
)

func init() {
	integrations.Register("gomod", func() engine.Integration {
		return New()
	})
}

// Integration implements Go modules go.mod updates.
type Integration struct {
	ds datasource.Datasource
}

// New creates a new gomod integration.
func New() *Integration {
	ds, err := datasource.Get("go")
	if err != nil {
		// Fallback to creating a new instance if not registered
		ds = datasource.NewGoDatasource()
	}
	return &Integration{
		ds: ds,
	}
}

// Name returns the integration identifier.
func (i *Integration) Name() string {
	return "gomod"
}

// Regex patterns for parsing go.mod files.
var (
	modulePattern  = regexp.MustCompile(`^module\s+(.+)$`)
	goVersionPat   = regexp.MustCompile(`^go\s+(\d+\.\d+(?:\.\d+)?)$`)
	requirePattern = regexp.MustCompile(`^\s*(\S+)\s+(v\S+)(\s*//\s*indirect)?$`)
	replacePattern = regexp.MustCompile(`^\s*(\S+)\s+=>\s+`)
)

// Detect finds go.mod files in the repository.
func (i *Integration) Detect(ctx context.Context, repoRoot string) ([]*engine.Manifest, error) {
	var manifests []*engine.Manifest

	err := filepath.Walk(repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor directories
		if info.IsDir() && info.Name() == "vendor" {
			return filepath.SkipDir
		}

		// Skip hidden directories
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
			return filepath.SkipDir
		}

		// Skip testdata directories
		if info.IsDir() && info.Name() == "testdata" {
			return filepath.SkipDir
		}

		if info.Name() == "go.mod" {
			relPath, err := filepath.Rel(repoRoot, path)
			if err != nil {
				return err
			}

			// Validate path for security
			err = integrations.ValidateFilePath(path)
			if err != nil {
				return err
			}

			content, err := os.ReadFile(path) // #nosec G304 - path is validated above
			if err != nil {
				return err
			}

			deps, metadata := i.parseGoMod(content)

			manifest := &engine.Manifest{
				Path:         relPath,
				Type:         "gomod",
				Dependencies: deps,
				Content:      content,
				Metadata:     metadata,
			}

			manifests = append(manifests, manifest)
		}

		return nil
	})

	return manifests, err
}

// parseGoMod extracts dependencies and metadata from go.mod content.
func (i *Integration) parseGoMod(content []byte) ([]engine.Dependency, map[string]interface{}) {
	deps := make([]engine.Dependency, 0)
	metadata := make(map[string]interface{})

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	inRequireBlock := false
	inReplaceBlock := false
	replacements := make(map[string]bool)

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Track module name
		if matches := modulePattern.FindStringSubmatch(trimmedLine); len(matches) > 1 {
			metadata["module_name"] = matches[1]
			continue
		}

		// Track Go version
		if matches := goVersionPat.FindStringSubmatch(trimmedLine); len(matches) > 1 {
			metadata["go_version"] = matches[1]
			continue
		}

		// Handle require block start
		if trimmedLine == "require (" {
			inRequireBlock = true
			continue
		}

		// Handle replace block start
		if trimmedLine == "replace (" {
			inReplaceBlock = true
			continue
		}

		// Handle block end
		if trimmedLine == ")" {
			inRequireBlock = false
			inReplaceBlock = false
			continue
		}

		// Track replaced modules
		if inReplaceBlock {
			if matches := replacePattern.FindStringSubmatch(trimmedLine); len(matches) > 1 {
				replacements[matches[1]] = true
			}
			continue
		}

		// Handle single-line require
		if strings.HasPrefix(trimmedLine, "require ") && !strings.HasSuffix(trimmedLine, "(") {
			requireLine := strings.TrimPrefix(trimmedLine, "require ")
			if dep := i.parseDependencyLine(requireLine); dep != nil {
				deps = append(deps, *dep)
			}
			continue
		}

		// Parse dependencies in require block
		if inRequireBlock {
			if dep := i.parseDependencyLine(trimmedLine); dep != nil {
				deps = append(deps, *dep)
			}
		}
	}

	// Store replacements in metadata for Plan phase to use
	metadata["replacements"] = replacements

	return deps, metadata
}

// parseDependencyLine parses a single dependency line from go.mod.
func (i *Integration) parseDependencyLine(line string) *engine.Dependency {
	matches := requirePattern.FindStringSubmatch(line)
	if len(matches) < 3 {
		return nil
	}

	modulePath := matches[1]
	version := matches[2]
	isIndirect := len(matches) > 3 && matches[3] != ""

	depType := "direct"
	if isIndirect {
		depType = "indirect"
	}

	return &engine.Dependency{
		Name:           modulePath,
		CurrentVersion: version,
		Constraint:     version, // Go modules use exact versions in go.mod
		Type:           depType,
		Registry:       "go",
	}
}

// Plan determines available updates for Go module dependencies.
// It applies policy precedence: CLI flags > uptool.yaml > manifest constraints.
func (i *Integration) Plan(ctx context.Context, manifest *engine.Manifest, planCtx *engine.PlanContext) (*engine.UpdatePlan, error) {
	updates := make([]engine.Update, 0, len(manifest.Dependencies))

	// Get replacements from manifest metadata
	replacements := make(map[string]bool)
	if manifest.Metadata != nil {
		if repl, ok := manifest.Metadata["replacements"].(map[string]bool); ok {
			replacements = repl
		}
	}

	for _, dep := range manifest.Dependencies {
		// Skip indirect dependencies by default (they're managed by go mod tidy)
		if dep.Type == "indirect" {
			continue
		}

		// Skip replaced modules
		if replacements[dep.Name] {
			continue
		}

		// Skip local paths and git references
		if strings.HasPrefix(dep.CurrentVersion, "v0.0.0-") {
			// This is likely a pseudo-version from a git commit
			continue
		}

		// Get all available versions
		availableVersions, err := i.ds.GetVersions(ctx, dep.Name)
		if err != nil {
			// Fallback: try to get just the latest version
			latest, latestErr := i.ds.GetLatestVersion(ctx, dep.Name)
			if latestErr != nil {
				// Skip packages that can't be resolved
				continue
			}
			availableVersions = []string{latest}
		}

		// Use policy-aware version selection
		targetVersion, impact, err := resolve.SelectVersionWithContext(
			dep.CurrentVersion,
			dep.Constraint,
			availableVersions,
			planCtx,
		)
		if err != nil || targetVersion == "" {
			continue
		}

		updates = append(updates, engine.Update{
			Dependency:    dep,
			TargetVersion: targetVersion,
			Impact:        string(impact),
			ChangelogURL:  fmt.Sprintf("https://pkg.go.dev/%s?tab=versions", dep.Name),
			PolicySource:  planCtx.GetPolicySource(),
		})
	}

	return &engine.UpdatePlan{
		Manifest: manifest,
		Updates:  updates,
		Strategy: "custom_rewrite", // We rewrite go.mod directly
	}, nil
}

// Apply executes the update plan by rewriting go.mod.
func (i *Integration) Apply(ctx context.Context, plan *engine.UpdatePlan) (*engine.ApplyResult, error) {
	if len(plan.Updates) == 0 {
		return &engine.ApplyResult{
			Manifest: plan.Manifest,
			Applied:  0,
			Failed:   0,
		}, nil
	}

	// Read the current go.mod
	fullPath := plan.Manifest.Path

	// Validate path for security
	if err := integrations.ValidateFilePath(fullPath); err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	content, err := os.ReadFile(fullPath) // #nosec G304 - path is validated above
	if err != nil {
		return nil, fmt.Errorf("read go.mod: %w", err)
	}

	oldContent := string(content)
	newContent := oldContent
	applied := 0

	// Apply updates by replacing version strings
	for idx := range plan.Updates {
		update := &plan.Updates[idx]
		oldVersion := update.Dependency.CurrentVersion
		newVersion := update.TargetVersion

		// Build the pattern to find and replace
		// Match: "module/path vX.Y.Z" in require statements
		oldPattern := regexp.QuoteMeta(update.Dependency.Name) + `\s+` + regexp.QuoteMeta(oldVersion)
		newReplacement := update.Dependency.Name + " " + newVersion

		re, err := regexp.Compile(oldPattern)
		if err != nil {
			continue
		}

		if re.MatchString(newContent) {
			newContent = re.ReplaceAllString(newContent, newReplacement)
			applied++
		}
	}

	if applied == 0 {
		return &engine.ApplyResult{
			Manifest: plan.Manifest,
			Applied:  0,
			Failed:   len(plan.Updates),
		}, nil
	}

	// Write back to go.mod
	if err := os.WriteFile(fullPath, []byte(newContent), 0o600); err != nil {
		return nil, fmt.Errorf("write go.mod: %w", err)
	}

	// Generate diff
	diff := generateDiff(oldContent, newContent)

	return &engine.ApplyResult{
		Manifest:     plan.Manifest,
		Applied:      applied,
		Failed:       len(plan.Updates) - applied,
		ManifestDiff: diff,
	}, nil
}

// Validate checks if go.mod is valid.
func (i *Integration) Validate(ctx context.Context, manifest *engine.Manifest) error {
	// Basic validation: check that we can parse the go.mod
	deps, metadata := i.parseGoMod(manifest.Content)

	if metadata["module_name"] == nil {
		return fmt.Errorf("go.mod missing module directive")
	}

	// Check for at least one dependency or empty is valid
	_ = deps

	return nil
}

// generateDiff creates a simple diff between old and new content.
func generateDiff(old, newContent string) string {
	if old == newContent {
		return ""
	}

	oldLines := strings.Split(old, "\n")
	newLines := strings.Split(newContent, "\n")

	var diff strings.Builder
	diff.WriteString("--- go.mod\n")
	diff.WriteString("+++ go.mod\n")

	maxLines := len(oldLines)
	if len(newLines) > maxLines {
		maxLines = len(newLines)
	}

	for idx := 0; idx < maxLines; idx++ {
		var oldLine, newLine string
		if idx < len(oldLines) {
			oldLine = oldLines[idx]
		}
		if idx < len(newLines) {
			newLine = newLines[idx]
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

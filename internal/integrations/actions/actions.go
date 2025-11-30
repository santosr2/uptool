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

// Package actions implements the GitHub Actions integration for updating workflow files.
// It detects .github/workflows/*.yml files, parses action references (uses: owner/repo@ref),
// queries GitHub Releases for version updates, and rewrites workflow files while preserving
// YAML structure and comments.
//
//nolint:govet // YAML struct field order is intentional for readability
package actions

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/santosr2/uptool/internal/datasource"
	"github.com/santosr2/uptool/internal/engine"
	"github.com/santosr2/uptool/internal/integrations"
	"github.com/santosr2/uptool/internal/resolve"
)

func init() {
	integrations.Register("actions", func() engine.Integration {
		return New()
	})
}

const integrationName = "actions"

// actionRefPattern matches GitHub Action references like:
// uses: actions/checkout@v4
// uses: actions/checkout@v4.2.2
// uses: actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683
var actionRefPattern = regexp.MustCompile(`^([a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+)@(.+)$`)

// Integration implements GitHub Actions workflow updates.
type Integration struct {
	ds datasource.Datasource
}

// New creates a new GitHub Actions integration.
func New() *Integration {
	ds, err := datasource.Get("github-releases")
	if err != nil {
		ds = datasource.NewGitHubDatasource()
	}
	return &Integration{
		ds: ds,
	}
}

// Name returns the integration identifier.
func (i *Integration) Name() string {
	return integrationName
}

// Workflow represents the structure of a GitHub Actions workflow file.
type Workflow struct {
	Name string                 `yaml:"name,omitempty"`
	On   interface{}            `yaml:"on,omitempty"`
	Jobs map[string]Job         `yaml:"jobs,omitempty"`
	Raw  map[string]interface{} `yaml:",inline"`
}

// Job represents a job in a workflow.
type Job struct {
	Name        string                 `yaml:"name,omitempty"`
	RunsOn      interface{}            `yaml:"runs-on,omitempty"`
	Steps       []Step                 `yaml:"steps,omitempty"`
	Strategy    interface{}            `yaml:"strategy,omitempty"`
	Permissions interface{}            `yaml:"permissions,omitempty"`
	Env         map[string]string      `yaml:"env,omitempty"`
	Needs       interface{}            `yaml:"needs,omitempty"`
	If          string                 `yaml:"if,omitempty"`
	Raw         map[string]interface{} `yaml:",inline"`
}

// Step represents a step in a job.
type Step struct {
	Name            string                 `yaml:"name,omitempty"`
	Uses            string                 `yaml:"uses,omitempty"`
	Run             string                 `yaml:"run,omitempty"`
	With            map[string]interface{} `yaml:"with,omitempty"`
	Env             map[string]string      `yaml:"env,omitempty"`
	If              string                 `yaml:"if,omitempty"`
	ID              string                 `yaml:"id,omitempty"`
	ContinueOnError interface{}            `yaml:"continue-on-error,omitempty"`
	Raw             map[string]interface{} `yaml:",inline"`
}

// Detect finds GitHub Actions workflow files in the repository.
func (i *Integration) Detect(ctx context.Context, repoRoot string) ([]*engine.Manifest, error) {
	var manifests []*engine.Manifest

	workflowsDir := filepath.Join(repoRoot, ".github", "workflows")
	if _, err := os.Stat(workflowsDir); os.IsNotExist(err) {
		return manifests, nil
	}

	err := filepath.Walk(workflowsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(info.Name()))
		if ext != ".yml" && ext != ".yaml" {
			return nil
		}

		if pathErr := integrations.ValidateFilePath(path); pathErr != nil {
			return pathErr
		}

		content, err := os.ReadFile(path) // #nosec G304 - path is validated above
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}

		deps, workflowName := i.extractDependencies(content)
		if len(deps) == 0 {
			return nil
		}

		manifest := &engine.Manifest{
			Path:         relPath,
			Type:         integrationName,
			Dependencies: deps,
			Content:      content,
			Metadata: map[string]interface{}{
				"workflow_name": workflowName,
				"action_count":  len(deps),
			},
		}

		manifests = append(manifests, manifest)
		return nil
	})

	return manifests, err
}

// extractDependencies parses workflow content and extracts action references.
func (i *Integration) extractDependencies(content []byte) ([]engine.Dependency, string) {
	var workflow Workflow
	if err := yaml.Unmarshal(content, &workflow); err != nil {
		return nil, ""
	}

	deps := make([]engine.Dependency, 0)
	seen := make(map[string]bool)

	for jobName := range workflow.Jobs {
		job := workflow.Jobs[jobName]
		for _, step := range job.Steps {
			if step.Uses == "" {
				continue
			}

			// Skip local actions (e.g., uses: ./.github/actions/my-action)
			if strings.HasPrefix(step.Uses, "./") || strings.HasPrefix(step.Uses, ".\\") {
				continue
			}

			// Skip Docker Hub references (e.g., uses: docker://alpine:3.8)
			if strings.HasPrefix(step.Uses, "docker://") {
				continue
			}

			matches := actionRefPattern.FindStringSubmatch(step.Uses)
			if matches == nil {
				continue
			}

			repo := matches[1]
			version := matches[2]

			// Create unique key to avoid duplicates
			key := fmt.Sprintf("%s@%s", repo, version)
			if seen[key] {
				continue
			}
			seen[key] = true

			// Determine version type
			depType := determineVersionType(version)

			deps = append(deps, engine.Dependency{
				Name:           repo,
				CurrentVersion: version,
				Constraint:     version,
				Type:           depType,
				Registry:       "github",
			})
		}
	}

	return deps, workflow.Name
}

// determineVersionType determines if a version is a tag, branch, or commit SHA.
func determineVersionType(version string) string {
	// Check if it's a commit SHA (40 hex characters)
	if len(version) == 40 && isHex(version) {
		return "sha"
	}

	// Check if it looks like a semver tag (v1, v1.2, v1.2.3, etc.)
	if strings.HasPrefix(version, "v") {
		return "tag"
	}

	// Assume it's a branch or other reference
	return "ref"
}

// isHex checks if a string contains only hexadecimal characters.
func isHex(s string) bool {
	for _, c := range s {
		isDigit := c >= '0' && c <= '9'
		isLowerHex := c >= 'a' && c <= 'f'
		isUpperHex := c >= 'A' && c <= 'F'
		if !isDigit && !isLowerHex && !isUpperHex {
			return false
		}
	}
	return true
}

// Plan determines available updates for GitHub Actions.
func (i *Integration) Plan(ctx context.Context, manifest *engine.Manifest, planCtx *engine.PlanContext) (*engine.UpdatePlan, error) {
	updates := make([]engine.Update, 0, len(manifest.Dependencies))

	for _, dep := range manifest.Dependencies {
		// Skip SHA pinned actions by default (they're usually pinned for security)
		if dep.Type == "sha" {
			continue
		}

		// Query GitHub releases for this action
		availableVersions, err := i.ds.GetVersions(ctx, dep.Name)
		if err != nil {
			continue
		}

		if len(availableVersions) == 0 {
			continue
		}

		// Extract current version number (strip 'v' prefix if present)
		currentVersion := strings.TrimPrefix(dep.CurrentVersion, "v")

		// Use policy-aware version selection
		targetVersion, impact, err := resolve.SelectVersionWithContext(
			currentVersion,
			dep.Constraint,
			availableVersions,
			planCtx,
		)
		if err != nil || targetVersion == "" {
			continue
		}

		// Add 'v' prefix back for GitHub Actions
		targetVersionWithPrefix := "v" + targetVersion

		// Skip if no update needed
		if targetVersionWithPrefix == dep.CurrentVersion {
			continue
		}

		updates = append(updates, engine.Update{
			Dependency:    dep,
			TargetVersion: targetVersionWithPrefix,
			Impact:        string(impact),
			PolicySource:  planCtx.GetPolicySource(),
		})
	}

	return &engine.UpdatePlan{
		Manifest: manifest,
		Updates:  updates,
		Strategy: "yaml_rewrite",
	}, nil
}

// Apply executes the update by rewriting workflow files.
func (i *Integration) Apply(ctx context.Context, plan *engine.UpdatePlan) (*engine.ApplyResult, error) {
	if len(plan.Updates) == 0 {
		return &engine.ApplyResult{
			Manifest: plan.Manifest,
			Applied:  0,
			Failed:   0,
		}, nil
	}

	if err := integrations.ValidateFilePath(plan.Manifest.Path); err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	oldContent, err := os.ReadFile(plan.Manifest.Path) // #nosec G304 - path is validated above
	if err != nil {
		return nil, fmt.Errorf("read workflow: %w", err)
	}

	// Create update map: old uses -> new uses
	updateMap := make(map[string]string)
	for idx := range plan.Updates {
		update := &plan.Updates[idx]
		oldRef := fmt.Sprintf("%s@%s", update.Dependency.Name, update.Dependency.CurrentVersion)
		newRef := fmt.Sprintf("%s@%s", update.Dependency.Name, update.TargetVersion)
		updateMap[oldRef] = newRef
	}

	// Replace action references in content
	newContent := string(oldContent)
	applied := 0

	for oldRef, newRef := range updateMap {
		if strings.Contains(newContent, oldRef) {
			newContent = strings.ReplaceAll(newContent, oldRef, newRef)
			applied++
		}
	}

	// Write updated content
	if err := os.WriteFile(plan.Manifest.Path, []byte(newContent), 0o600); err != nil {
		return nil, fmt.Errorf("write workflow: %w", err)
	}

	// Generate diff
	diff := generateDiff(plan.Manifest.Path, string(oldContent), newContent)

	return &engine.ApplyResult{
		Manifest:     plan.Manifest,
		Applied:      applied,
		Failed:       0,
		ManifestDiff: diff,
	}, nil
}

// Validate checks if the workflow file is valid YAML.
func (i *Integration) Validate(ctx context.Context, manifest *engine.Manifest) error {
	var workflow Workflow
	if err := yaml.Unmarshal(manifest.Content, &workflow); err != nil {
		return fmt.Errorf("invalid workflow YAML: %w", err)
	}

	if len(workflow.Jobs) == 0 {
		return fmt.Errorf("workflow has no jobs defined")
	}

	return nil
}

// generateDiff creates a simple diff between old and new content.
func generateDiff(path, old, newContent string) string {
	if old == newContent {
		return ""
	}

	oldLines := strings.Split(old, "\n")
	newLines := strings.Split(newContent, "\n")

	var diff strings.Builder
	diff.WriteString(fmt.Sprintf("--- %s\n", path))
	diff.WriteString(fmt.Sprintf("+++ %s\n", path))

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
			if strings.Contains(oldLine, "uses:") || strings.Contains(newLine, "uses:") {
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

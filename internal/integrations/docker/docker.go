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

// Package docker implements the Docker integration for updating Dockerfile and docker-compose files.
// It detects Dockerfile and docker-compose.yml files, parses image references (FROM image:tag),
// queries Docker Hub for version updates, and rewrites files while preserving structure.
//
//nolint:govet // YAML struct field order is intentional for readability
package docker

import (
	"bufio"
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
	integrations.Register("docker", func() engine.Integration {
		return New()
	})
}

const integrationName = "docker"

// fromPattern matches Dockerfile FROM instructions:
// FROM image:tag
// FROM image:tag AS builder
// FROM --platform=linux/amd64 image:tag
var fromPattern = regexp.MustCompile(`^FROM\s+(?:--platform=[^\s]+\s+)?([^:\s]+)(?::([^\s@]+))?(?:@sha256:[a-f0-9]+)?(?:\s+AS\s+\S+)?`)

const defaultTag = "latest"

// Integration implements Docker file updates.
type Integration struct {
	ds datasource.Datasource
}

// New creates a new Docker integration.
func New() *Integration {
	ds, err := datasource.Get("docker-hub")
	if err != nil {
		ds = datasource.NewDockerHubDatasource()
	}
	return &Integration{
		ds: ds,
	}
}

// Name returns the integration identifier.
func (i *Integration) Name() string {
	return integrationName
}

// ComposeFile represents the structure of a docker-compose.yml file.
type ComposeFile struct {
	Version  string                 `yaml:"version,omitempty"`
	Services map[string]Service     `yaml:"services,omitempty"`
	Raw      map[string]interface{} `yaml:",inline"`
}

// Service represents a service in docker-compose.
type Service struct {
	Image       string                 `yaml:"image,omitempty"`
	Build       interface{}            `yaml:"build,omitempty"`
	Environment interface{}            `yaml:"environment,omitempty"`
	Ports       []string               `yaml:"ports,omitempty"`
	Volumes     []string               `yaml:"volumes,omitempty"`
	DependsOn   interface{}            `yaml:"depends_on,omitempty"`
	Raw         map[string]interface{} `yaml:",inline"`
}

// Detect finds Dockerfile and docker-compose files in the repository.
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

		// Skip vendor and node_modules directories
		if info.IsDir() && (info.Name() == "vendor" || info.Name() == "node_modules") {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		name := info.Name()
		isDockerfile := name == "Dockerfile" || strings.HasPrefix(name, "Dockerfile.")
		isCompose := name == "docker-compose.yml" || name == "docker-compose.yaml" ||
			name == "compose.yml" || name == "compose.yaml"

		if !isDockerfile && !isCompose {
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

		var deps []engine.Dependency
		var metadata map[string]interface{}

		if isDockerfile {
			deps = i.extractDockerfileDeps(content)
			metadata = map[string]interface{}{
				"file_type":   "dockerfile",
				"image_count": len(deps),
			}
		} else {
			deps = i.extractComposeDeps(content)
			metadata = map[string]interface{}{
				"file_type":   "compose",
				"image_count": len(deps),
			}
		}

		if len(deps) == 0 {
			return nil
		}

		manifest := &engine.Manifest{
			Path:         relPath,
			Type:         integrationName,
			Dependencies: deps,
			Content:      content,
			Metadata:     metadata,
		}

		manifests = append(manifests, manifest)
		return nil
	})

	return manifests, err
}

// extractDockerfileDeps parses Dockerfile content and extracts image references.
func (i *Integration) extractDockerfileDeps(content []byte) []engine.Dependency {
	deps := make([]engine.Dependency, 0)
	seen := make(map[string]bool)
	scanner := bufio.NewScanner(strings.NewReader(string(content)))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for FROM instruction
		if strings.HasPrefix(strings.ToUpper(line), "FROM") {
			matches := fromPattern.FindStringSubmatch(line)
			if matches == nil {
				continue
			}

			image := matches[1]
			tag := matches[2]

			// Skip scratch images
			if image == "scratch" {
				continue
			}

			// Skip build args (e.g., FROM ${BASE_IMAGE})
			if strings.Contains(image, "$") || strings.Contains(image, "{") {
				continue
			}

			// Default tag is "latest"
			if tag == "" {
				tag = defaultTag
			}

			// Create unique key
			key := fmt.Sprintf("%s:%s", image, tag)
			if seen[key] {
				continue
			}
			seen[key] = true

			deps = append(deps, engine.Dependency{
				Name:           image,
				CurrentVersion: tag,
				Constraint:     tag,
				Type:           "image",
				Registry:       "docker-hub",
			})
		}
	}

	return deps
}

// extractComposeDeps parses docker-compose content and extracts image references.
func (i *Integration) extractComposeDeps(content []byte) []engine.Dependency {
	var compose ComposeFile
	if err := yaml.Unmarshal(content, &compose); err != nil {
		return nil
	}

	deps := make([]engine.Dependency, 0)
	seen := make(map[string]bool)

	for _, service := range compose.Services {
		if service.Image == "" {
			continue
		}

		image, tag := parseImageReference(service.Image)
		if image == "" {
			continue
		}

		// Create unique key
		key := fmt.Sprintf("%s:%s", image, tag)
		if seen[key] {
			continue
		}
		seen[key] = true

		deps = append(deps, engine.Dependency{
			Name:           image,
			CurrentVersion: tag,
			Constraint:     tag,
			Type:           "image",
			Registry:       "docker-hub",
		})
	}

	return deps
}

// parseImageReference parses an image reference into image name and tag.
func parseImageReference(ref string) (string, string) {
	// Handle digest references (image@sha256:...)
	if strings.Contains(ref, "@sha256:") {
		parts := strings.SplitN(ref, "@", 2)
		return parts[0], "sha256"
	}

	// Handle normal references (image:tag)
	parts := strings.SplitN(ref, ":", 2)
	image := parts[0]
	tag := defaultTag

	if len(parts) == 2 {
		tag = parts[1]
	}

	// Skip variable references
	if strings.Contains(image, "$") || strings.Contains(image, "{") {
		return "", ""
	}

	return image, tag
}

// Plan determines available updates for Docker images.
func (i *Integration) Plan(ctx context.Context, manifest *engine.Manifest, planCtx *engine.PlanContext) (*engine.UpdatePlan, error) {
	updates := make([]engine.Update, 0, len(manifest.Dependencies))

	for _, dep := range manifest.Dependencies {
		// Skip latest tag (no specific version to update from)
		if dep.CurrentVersion == defaultTag || dep.CurrentVersion == "sha256" {
			continue
		}

		// Query Docker Hub for available tags
		availableVersions, err := i.ds.GetVersions(ctx, dep.Name)
		if err != nil {
			continue
		}

		if len(availableVersions) == 0 {
			continue
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

		// Skip if no update needed
		if targetVersion == dep.CurrentVersion {
			continue
		}

		updates = append(updates, engine.Update{
			Dependency:    dep,
			TargetVersion: targetVersion,
			Impact:        string(impact),
			PolicySource:  planCtx.GetPolicySource(),
		})
	}

	return &engine.UpdatePlan{
		Manifest: manifest,
		Updates:  updates,
		Strategy: "text_rewrite",
	}, nil
}

// Apply executes the update by rewriting Docker files.
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
		return nil, fmt.Errorf("read docker file: %w", err)
	}

	// Create update map: old reference -> new reference
	updateMap := make(map[string]string)
	for idx := range plan.Updates {
		update := &plan.Updates[idx]
		oldRef := fmt.Sprintf("%s:%s", update.Dependency.Name, update.Dependency.CurrentVersion)
		newRef := fmt.Sprintf("%s:%s", update.Dependency.Name, update.TargetVersion)
		updateMap[oldRef] = newRef
	}

	// Replace image references in content
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
		return nil, fmt.Errorf("write docker file: %w", err)
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

// Validate checks if the Docker file is valid.
func (i *Integration) Validate(ctx context.Context, manifest *engine.Manifest) error {
	fileType, ok := manifest.Metadata["file_type"].(string)
	if !ok {
		return fmt.Errorf("unknown docker file type")
	}

	if fileType == "compose" {
		var compose ComposeFile
		if err := yaml.Unmarshal(manifest.Content, &compose); err != nil {
			return fmt.Errorf("invalid docker-compose YAML: %w", err)
		}
		if len(compose.Services) == 0 {
			return fmt.Errorf("docker-compose has no services defined")
		}
		return nil
	}

	// Validate Dockerfile by checking for at least one FROM instruction
	scanner := bufio.NewScanner(strings.NewReader(string(manifest.Content)))
	hasFrom := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(strings.ToUpper(line), "FROM") {
			hasFrom = true
			break
		}
	}

	if !hasFrom {
		return fmt.Errorf("dockerfile has no FROM instruction")
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
			// Show changes to FROM lines or image: lines
			if strings.Contains(strings.ToUpper(oldLine), "FROM") ||
				strings.Contains(strings.ToUpper(newLine), "FROM") ||
				strings.Contains(oldLine, "image:") ||
				strings.Contains(newLine, "image:") {
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

// Package terraform implements the Terraform integration for updating module versions in .tf files.
// It detects Terraform configuration files, parses HCL to extract module and provider versions,
// queries the Terraform Registry for updates, and rewrites versions while preserving HCL formatting.
package terraform

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
	"github.com/santosr2/uptool/internal/datasource"
	"github.com/santosr2/uptool/internal/engine"
	"github.com/santosr2/uptool/internal/integrations"
	"github.com/zclconf/go-cty/cty"
)

func init() {
	integrations.Register("terraform", func() engine.Integration {
		return New()
	})
}

const integrationName = "terraform"

// Integration implements terraform configuration updates.
type Integration struct {
	ds datasource.Datasource
}

// New creates a new terraform integration.
func New() *Integration {
	ds, err := datasource.Get("terraform")
	if err != nil {
		// Fallback to creating a new instance if not registered
		ds = datasource.NewTerraformDatasource()
	}
	return &Integration{
		ds: ds,
	}
}

// Name returns the integration identifier.
func (i *Integration) Name() string {
	return integrationName
}

// Config represents terraform configuration structure.
type Config struct {
	Terraform []TerraformBlock `hcl:"terraform,block"`
	Modules   []ModuleBlock    `hcl:"module,block"`
	Providers []ProviderBlock  `hcl:"provider,block"`
	Remain    hcl.Body         `hcl:",remain"`
}

// TerraformBlock represents a terraform configuration block.
type TerraformBlock struct {
	RequiredVersion   string                  `hcl:"required_version,optional"`
	RequiredProviders *RequiredProvidersBlock `hcl:"required_providers,block"`
	Remain            hcl.Body                `hcl:",remain"`
}

// RequiredProvidersBlock represents the required_providers block.
type RequiredProvidersBlock struct {
	Body hcl.Body `hcl:",remain"`
}

// ModuleBlock represents a module block.
type ModuleBlock struct {
	Name    string   `hcl:"name,label"`
	Source  string   `hcl:"source,optional"`
	Version string   `hcl:"version,optional"`
	Remain  hcl.Body `hcl:",remain"`
}

// ProviderBlock represents a provider block.
type ProviderBlock struct {
	Name   string   `hcl:"name,label"`
	Remain hcl.Body `hcl:",remain"`
}

// Detect finds .tf files in the repository.
func (i *Integration) Detect(ctx context.Context, repoRoot string) ([]*engine.Manifest, error) {
	var manifests []*engine.Manifest
	manifestMap := make(map[string]*engine.Manifest)

	err := filepath.Walk(repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip hidden directories
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && path != repoRoot {
			return filepath.SkipDir
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".tf") {
			dir := filepath.Dir(path)
			relDir, err := filepath.Rel(repoRoot, dir)
			if err != nil {
				return nil
			}

			// Group all .tf files in the same directory
			if _, exists := manifestMap[relDir]; !exists {
				manifestMap[relDir] = &engine.Manifest{
					Path:         relDir,
					Type:         integrationName,
					Dependencies: []engine.Dependency{},
					Metadata: map[string]any{
						"files": []string{},
					},
				}
			}

			// Parse the file
			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			// Use hclwrite for more flexible parsing
			file, diags := hclwrite.ParseConfig(content, path, hcl.Pos{Line: 1, Column: 1})
			if diags.HasErrors() {
				return nil // Skip invalid HCL
			}

			manifest := manifestMap[relDir]

			// Extract module dependencies
			for _, block := range file.Body().Blocks() {
				if block.Type() == "module" {
					labels := block.Labels()
					if len(labels) == 0 {
						continue
					}

					sourceAttr := block.Body().GetAttribute("source")
					versionAttr := block.Body().GetAttribute("version")

					if sourceAttr != nil && versionAttr != nil {
						sourceTokens := sourceAttr.Expr().BuildTokens(nil)
						source := strings.Trim(string(sourceTokens.Bytes()), ` "`)

						versionTokens := versionAttr.Expr().BuildTokens(nil)
						version := strings.Trim(string(versionTokens.Bytes()), ` "`)

						// Only track registry modules
						if !strings.HasPrefix(source, ".") && !strings.Contains(source, "git::") {
							manifest.Dependencies = append(manifest.Dependencies, engine.Dependency{
								Name:           source,
								CurrentVersion: version,
								Type:           "module",
								Registry:       "terraform",
							})
						}
					}
				}
			}

			// Track files in this directory
			files := manifest.Metadata["files"].([]string)
			files = append(files, filepath.Base(path))
			manifest.Metadata["files"] = files
		}

		return nil
	})

	// Convert map to slice
	for _, manifest := range manifestMap {
		if len(manifest.Dependencies) > 0 {
			manifests = append(manifests, manifest)
		}
	}

	return manifests, err
}

// Plan determines available updates for terraform providers and modules.
func (i *Integration) Plan(ctx context.Context, manifest *engine.Manifest) (*engine.UpdatePlan, error) {
	var updates []engine.Update

	for _, dep := range manifest.Dependencies {
		if dep.Type == "provider" {
			// Get latest provider version
			latest, err := i.ds.GetLatestVersion(ctx, dep.Name)
			if err != nil {
				continue
			}

			// Strip constraint prefixes for comparison
			currentClean := strings.TrimPrefix(dep.CurrentVersion, "~> ")
			currentClean = strings.TrimPrefix(currentClean, ">= ")
			currentClean = strings.TrimPrefix(currentClean, "= ")
			currentClean = strings.TrimSpace(currentClean)

			latestClean := strings.TrimPrefix(latest, "v")
			currentClean = strings.TrimPrefix(currentClean, "v")

			// Compare versions - only add update if they're actually different
			if latestClean != currentClean && latest != "" && currentClean != "" {
				updates = append(updates, engine.Update{
					Dependency:    dep,
					TargetVersion: latest,
					Impact:        determineImpact(currentClean, latestClean),
				})
			}
		} else if dep.Type == "module" {
			// Get latest module version
			latest, err := i.ds.GetLatestVersion(ctx, dep.Name)
			if err != nil {
				continue
			}

			// Strip constraint prefixes for comparison
			currentClean := strings.TrimPrefix(dep.CurrentVersion, "~> ")
			currentClean = strings.TrimPrefix(currentClean, ">= ")
			currentClean = strings.TrimPrefix(currentClean, "= ")
			currentClean = strings.TrimSpace(currentClean)

			latestClean := strings.TrimPrefix(latest, "v")
			currentClean = strings.TrimPrefix(currentClean, "v")

			// Compare versions - only add update if they're actually different
			if latestClean != currentClean && latest != "" && currentClean != "" {
				updates = append(updates, engine.Update{
					Dependency:    dep,
					TargetVersion: latest,
					Impact:        determineImpact(currentClean, latestClean),
				})
			}
		}
	}

	return &engine.UpdatePlan{
		Manifest: manifest,
		Updates:  updates,
		Strategy: "hcl_rewrite",
	}, nil
}

// Apply executes the update by rewriting terraform files.
func (i *Integration) Apply(ctx context.Context, plan *engine.UpdatePlan) (*engine.ApplyResult, error) {
	if len(plan.Updates) == 0 {
		return &engine.ApplyResult{
			Manifest: plan.Manifest,
			Applied:  0,
			Failed:   0,
		}, nil
	}

	// Create update maps for quick lookup
	providerUpdates := make(map[string]string)
	moduleUpdates := make(map[string]string)

	for _, update := range plan.Updates {
		if update.Dependency.Type == "provider" {
			providerUpdates[update.Dependency.Name] = update.TargetVersion
		} else if update.Dependency.Type == "module" {
			moduleUpdates[update.Dependency.Name] = update.TargetVersion
		}
	}

	applied := 0
	var allDiffs strings.Builder

	// Get list of files to update
	files := plan.Manifest.Metadata["files"].([]string)

	for _, filename := range files {
		filePath := filepath.Join(plan.Manifest.Path, filename)

		// Read old content
		oldContent, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		// Parse HCL for writing
		file, diags := hclwrite.ParseConfig(oldContent, filePath, hcl.Pos{Line: 1, Column: 1})
		if diags.HasErrors() {
			continue
		}

		fileUpdated := false

		// Update terraform blocks (providers)
		for _, block := range file.Body().Blocks() {
			if block.Type() == "terraform" {
				for _, innerBlock := range block.Body().Blocks() {
					if innerBlock.Type() == "required_providers" {
						// Update each provider in required_providers
						for name := range providerUpdates {
							// Extract provider name from source (e.g., "hashicorp/aws" -> "aws")
							providerName := name
							if strings.Contains(name, "/") {
								parts := strings.Split(name, "/")
								providerName = parts[len(parts)-1]
							}

							providerAttr := innerBlock.Body().GetAttribute(providerName)
							if providerAttr != nil {
								// This is a complex attribute, need to update the version within it
								// For now, we'll use string replacement as HCL doesn't provide easy nested updates
								fileUpdated = true
								applied++
							}
						}
					}
				}
			}

			// Update module blocks
			if block.Type() == "module" {
				labels := block.Labels()
				if len(labels) == 0 {
					continue
				}

				sourceAttr := block.Body().GetAttribute("source")
				if sourceAttr == nil {
					continue
				}

				// Get source value by parsing the tokens
				sourceTokens := sourceAttr.Expr().BuildTokens(nil)
				source := strings.Trim(string(sourceTokens.Bytes()), ` "`)

				if newVersion, ok := moduleUpdates[source]; ok {
					versionAttr := block.Body().GetAttribute("version")
					if versionAttr != nil {
						block.Body().SetAttributeValue("version", cty.StringVal(newVersion))
						fileUpdated = true
						applied++
					}
				}
			}
		}

		if fileUpdated {
			// Write updated content
			newContent := file.Bytes()

			// For provider versions, use regex replacement since HCL doesn't support nested updates easily
			for providerSource, newVersion := range providerUpdates {
				providerName := providerSource
				if strings.Contains(providerSource, "/") {
					parts := strings.Split(providerSource, "/")
					providerName = parts[len(parts)-1]
				}

				// Match: provider_name = { ... version = "old_version" ... }
				re := regexp.MustCompile(fmt.Sprintf(`(%s\s*=\s*\{[^}]*version\s*=\s*)"([^"]*)"`, providerName))
				newContent = re.ReplaceAll(newContent, []byte(fmt.Sprintf(`${1}"%s"`, newVersion)))
			}

			if err := os.WriteFile(filePath, newContent, 0644); err != nil {
				continue
			}

			// Generate diff for this file
			diff := generateDiff(filename, string(oldContent), string(newContent))
			allDiffs.WriteString(diff)
		}
	}

	return &engine.ApplyResult{
		Manifest:     plan.Manifest,
		Applied:      applied,
		Failed:       0,
		ManifestDiff: allDiffs.String(),
	}, nil
}

// Validate checks if the terraform configuration is valid.
func (i *Integration) Validate(ctx context.Context, manifest *engine.Manifest) error {
	// Basic HCL validation
	files := manifest.Metadata["files"].([]string)
	for _, filename := range files {
		filePath := filepath.Join(manifest.Path, filename)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read file %s: %w", filename, err)
		}

		var config Config
		if err := hclsimple.Decode(filePath, content, nil, &config); err != nil {
			return fmt.Errorf("invalid HCL in %s: %w", filename, err)
		}
	}
	return nil
}

// determineImpact tries to determine the impact of an update.
func determineImpact(old, new string) string {
	// Strip constraint prefixes
	old = strings.TrimPrefix(old, "~> ")
	old = strings.TrimPrefix(old, ">= ")
	old = strings.TrimPrefix(old, "= ")

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
func generateDiff(filename, old, new string) string {
	if old == new {
		return ""
	}

	oldLines := strings.Split(old, "\n")
	newLines := strings.Split(new, "\n")

	var diff strings.Builder
	diff.WriteString(fmt.Sprintf("--- %s\n", filename))
	diff.WriteString(fmt.Sprintf("+++ %s\n", filename))

	// Find version lines that changed
	versionRE := regexp.MustCompile(`version\s*=\s*"([^"]+)"`)

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

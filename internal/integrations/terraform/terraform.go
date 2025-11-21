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
	"github.com/zclconf/go-cty/cty"

	"github.com/santosr2/uptool/internal/datasource"
	"github.com/santosr2/uptool/internal/engine"
	"github.com/santosr2/uptool/internal/integrations"
)

func init() {
	integrations.Register("terraform", func() engine.Integration {
		return New()
	})
}

const (
	integrationName = "terraform"
	blockTypeModule = "module"
)

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
	Remain    hcl.Body        `hcl:",remain"`
	Terraform []Block         `hcl:"terraform,block"`
	Modules   []ModuleBlock   `hcl:"module,block"`
	Providers []ProviderBlock `hcl:"provider,block"`
}

// Block represents a terraform configuration block.
type Block struct {
	Remain            hcl.Body                `hcl:",remain"`
	RequiredProviders *RequiredProvidersBlock `hcl:"required_providers,block"`
	RequiredVersion   string                  `hcl:"required_version,optional"`
}

// RequiredProvidersBlock represents the required_providers block.
type RequiredProvidersBlock struct {
	Body hcl.Body `hcl:",remain"`
}

// ModuleBlock represents a module block.
type ModuleBlock struct {
	Remain  hcl.Body `hcl:",remain"`
	Name    string   `hcl:"name,label"`
	Source  string   `hcl:"source,optional"`
	Version string   `hcl:"version,optional"`
}

// ProviderBlock represents a provider block.
type ProviderBlock struct {
	Remain hcl.Body `hcl:",remain"`
	Name   string   `hcl:"name,label"`
}

// Detect finds .tf files in the repository.
func (i *Integration) Detect(ctx context.Context, repoRoot string) ([]*engine.Manifest, error) {
	var manifests []*engine.Manifest
	manifestMap := make(map[string]*engine.Manifest)

	err := filepath.Walk(repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && path != repoRoot {
			return filepath.SkipDir
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".tf") {
			dir := filepath.Dir(path)
			relDir, err := filepath.Rel(repoRoot, dir)
			if err != nil {
				return err
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
			// Validate path for security
			err = integrations.ValidateFilePath(path)
			if err != nil {
				return err
			}

			content, err := os.ReadFile(path) // #nosec G304 - path is validated above
			if err != nil {
				return err
			}

			// Use hclwrite for more flexible parsing
			file, diags := hclwrite.ParseConfig(content, path, hcl.Pos{Line: 1, Column: 1})
			if diags.HasErrors() {
				return diags
			}

			manifest := manifestMap[relDir]

			// Extract module dependencies
			for _, block := range file.Body().Blocks() {
				if block.Type() != blockTypeModule {
					continue
				}

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

			// Track files in this directory
			files := manifest.Metadata["files"].([]string) //nolint:errcheck // metadata set by us
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

// processDependencyUpdate fetches and compares versions for a dependency
func (i *Integration) processDependencyUpdate(ctx context.Context, dep *engine.Dependency) (engine.Update, bool) {
	latest, err := i.ds.GetLatestVersion(ctx, dep.Name)
	if err != nil {
		return engine.Update{}, false
	}

	// Strip constraint prefixes for comparison
	currentClean := strings.TrimPrefix(dep.CurrentVersion, "~> ")
	currentClean = strings.TrimPrefix(currentClean, ">= ")
	currentClean = strings.TrimPrefix(currentClean, "= ")
	currentClean = strings.TrimSpace(currentClean)

	latestClean := strings.TrimPrefix(latest, "v")
	currentClean = strings.TrimPrefix(currentClean, "v")

	// Compare versions - only return update if they're actually different
	if latestClean != currentClean && latest != "" && currentClean != "" {
		return engine.Update{
			Dependency:    *dep,
			TargetVersion: latest,
			Impact:        determineImpact(currentClean, latestClean),
		}, true
	}

	return engine.Update{}, false
}

// Plan determines available updates for terraform providers and modules.
func (i *Integration) Plan(ctx context.Context, manifest *engine.Manifest) (*engine.UpdatePlan, error) {
	var updates []engine.Update

	for _, dep := range manifest.Dependencies {
		if dep.Type == "provider" || dep.Type == blockTypeModule {
			if update, ok := i.processDependencyUpdate(ctx, &dep); ok {
				updates = append(updates, update)
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

	for i := range plan.Updates {
		update := &plan.Updates[i]
		switch update.Dependency.Type {
		case "provider":
			providerUpdates[update.Dependency.Name] = update.TargetVersion
		case blockTypeModule:
			moduleUpdates[update.Dependency.Name] = update.TargetVersion
		}
	}

	applied := 0
	var allDiffs strings.Builder

	// Get list of files to update
	files := plan.Manifest.Metadata["files"].([]string) //nolint:errcheck // metadata set by us

	for _, filename := range files {
		filePath := filepath.Join(plan.Manifest.Path, filename)

		// Read old content
		// Validate path for security
		if err := integrations.ValidateFilePath(filePath); err != nil {
			continue
		}

		oldContent, err := os.ReadFile(filePath) // #nosec G304 - path is validated above
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
			if block.Type() == blockTypeModule {
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
				newContent = re.ReplaceAll(newContent, []byte(fmt.Sprintf(`${1}%q`, newVersion)))
			}

			if err := os.WriteFile(filePath, newContent, 0o600); err != nil {
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
	files := manifest.Metadata["files"].([]string) //nolint:errcheck // metadata set by us
	for _, filename := range files {
		filePath := filepath.Join(manifest.Path, filename)
		// Validate path for security
		if err := integrations.ValidateFilePath(filePath); err != nil {
			continue
		}

		content, err := os.ReadFile(filePath) // #nosec G304 - path is validated above
		if err != nil {
			continue
		}

		var config Config
		if err := hclsimple.Decode(filePath, content, nil, &config); err != nil {
			return fmt.Errorf("invalid HCL in %s: %w", filename, err)
		}
	}
	return nil
}

// determineImpact tries to determine the impact of an update.
func determineImpact(old, newVer string) string {
	// Strip constraint prefixes
	old = strings.TrimPrefix(old, "~> ")
	old = strings.TrimPrefix(old, ">= ")
	old = strings.TrimPrefix(old, "= ")

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
func generateDiff(filename, old, newContent string) string {
	if old == newContent {
		return ""
	}

	oldLines := strings.Split(old, "\n")
	newLines := strings.Split(newContent, "\n")

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

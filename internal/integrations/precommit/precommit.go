// Package precommit implements the pre-commit integration using the native autoupdate command.
// It detects .pre-commit-config.yaml files, runs 'pre-commit autoupdate' to update hook versions,
// and parses the output to report changes. This follows the manifest-first philosophy by using
// the native tool that directly updates the configuration file.
package precommit

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/santosr2/uptool/internal/engine"
	"github.com/santosr2/uptool/internal/integrations"
	"github.com/santosr2/uptool/internal/secureio"
)

func init() {
	integrations.Register("precommit", func() engine.Integration {
		return New()
	})
}

const integrationName = "precommit"

// Integration implements pre-commit hook updates using native autoupdate command.
type Integration struct{}

// New creates a new pre-commit integration.
func New() *Integration {
	return &Integration{}
}

// Name returns the integration identifier.
func (i *Integration) Name() string {
	return integrationName
}

// Config represents the structure of .pre-commit-config.yaml.
type Config struct {
	Repos []Repo `yaml:"repos"`
}

// Repo represents a pre-commit repository.
type Repo struct {
	Repo  string `yaml:"repo"`
	Rev   string `yaml:"rev"`
	Hooks []Hook `yaml:"hooks,omitempty"`
}

// Hook represents a pre-commit hook.
type Hook struct {
	ID string `yaml:"id"`
}

// Detect finds .pre-commit-config.yaml files in the repository.
func (i *Integration) Detect(ctx context.Context, repoRoot string) ([]*engine.Manifest, error) {
	var manifests []*engine.Manifest

	err := filepath.Walk(repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories except .pre-commit-config.yaml in root
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && path != repoRoot {
			return filepath.SkipDir
		}

		if info.Name() == ".pre-commit-config.yaml" {
			relPath, err := filepath.Rel(repoRoot, path)
			if err != nil {
				return err
			}

			content, err := secureio.ReadFile(path)
			if err != nil {
				return err
			}

			var config Config
			if err := yaml.Unmarshal(content, &config); err != nil {
				return err
			}

			deps := i.extractDependencies(&config)

			manifest := &engine.Manifest{
				Path:         relPath,
				Type:         integrationName,
				Dependencies: deps,
				Content:      content,
				Metadata: map[string]interface{}{
					"repos_count": len(config.Repos),
				},
			}

			manifests = append(manifests, manifest)
		}

		return nil
	})

	return manifests, err
}

// extractDependencies extracts hook repositories as dependencies.
func (i *Integration) extractDependencies(config *Config) []engine.Dependency {
	deps := make([]engine.Dependency, 0, len(config.Repos))

	for _, repo := range config.Repos {
		if repo.Repo == "" || repo.Repo == "local" || repo.Repo == "meta" {
			continue
		}

		deps = append(deps, engine.Dependency{
			Name:           repo.Repo,
			CurrentVersion: repo.Rev,
			Type:           "direct",
			Registry:       "git",
		})
	}

	return deps
}

// Plan determines available updates for pre-commit hooks.
// For pre-commit, we use the native autoupdate command in dry-run mode.
func (i *Integration) Plan(ctx context.Context, manifest *engine.Manifest) (*engine.UpdatePlan, error) {
	// Check if pre-commit is available
	if !i.isPreCommitAvailable() {
		return &engine.UpdatePlan{
			Manifest: manifest,
			Updates:  nil,
			Strategy: "native_command",
		}, nil
	}

	// Run pre-commit autoupdate in dry-run mode by checking output
	updates, err := i.detectUpdates(ctx, manifest.Path)
	if err != nil {
		return nil, err
	}

	return &engine.UpdatePlan{
		Manifest: manifest,
		Updates:  updates,
		Strategy: "native_command",
	}, nil
}

// detectUpdates runs pre-commit autoupdate and parses the output to detect changes.
func (i *Integration) detectUpdates(ctx context.Context, manifestPath string) ([]engine.Update, error) {
	// Create a temporary copy to test updates
	tmpDir, err := os.MkdirTemp("", "precommit-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Copy the config file
	content, err := secureio.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	tmpConfig := filepath.Join(tmpDir, ".pre-commit-config.yaml")
	if writeErr := os.WriteFile(tmpConfig, content, 0o600); writeErr != nil {
		return nil, fmt.Errorf("write temp config: %w", writeErr)
	}

	// Run autoupdate
	// Validate tmpConfig path to prevent command injection
	if !filepath.IsAbs(tmpConfig) || strings.Contains(tmpConfig, "..") {
		return nil, fmt.Errorf("invalid temp config path: %s", tmpConfig)
	}
	cmd := exec.CommandContext(ctx, "pre-commit", "autoupdate", "--config", tmpConfig) // #nosec G204 - tmpConfig is a validated temporary file path
	output, err := cmd.CombinedOutput()
	// Note: We continue even if there's an error, as the output might still be useful
	_ = err // Explicitly ignore error - we parse output regardless

	// Parse the output to find updates
	updates := i.parseAutoupdateOutput(string(output))

	return updates, nil
}

// parseAutoupdateOutput parses pre-commit autoupdate output.
// Format: "[<repo>] updating <old> -> <new>"
func (i *Integration) parseAutoupdateOutput(output string) []engine.Update {
	var updates []engine.Update

	// Regex to match update lines
	// Example: "[https://github.com/pre-commit/pre-commit-hooks] updating v4.3.0 -> v6.0.0"
	re := regexp.MustCompile(`\[(https://[^\]]+)\]\s+updating\s+(\S+)\s+->\s+(\S+)`)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if len(matches) == 4 {
			repo := matches[1]
			oldVer := matches[2]
			newVer := matches[3]

			updates = append(updates, engine.Update{
				Dependency: engine.Dependency{
					Name:           repo,
					CurrentVersion: oldVer,
					Type:           "direct",
				},
				TargetVersion: newVer,
				Impact:        i.determineImpact(oldVer, newVer),
			})
		}
	}

	return updates
}

// determineImpact tries to determine the impact of an update.
func (i *Integration) determineImpact(old, new string) string {
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

// Apply executes the update using native pre-commit autoupdate command.
func (i *Integration) Apply(ctx context.Context, plan *engine.UpdatePlan) (*engine.ApplyResult, error) {
	if len(plan.Updates) == 0 {
		return &engine.ApplyResult{
			Manifest: plan.Manifest,
			Applied:  0,
			Failed:   0,
		}, nil
	}

	// Check if pre-commit is available
	if !i.isPreCommitAvailable() {
		return nil, fmt.Errorf("pre-commit command not found")
	}

	// Read old content for diff
	oldContent, err := secureio.ReadFile(plan.Manifest.Path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	// Run pre-commit autoupdate
	// Validate manifest path to prevent command injection
	if !filepath.IsAbs(plan.Manifest.Path) || strings.Contains(plan.Manifest.Path, "..") {
		return &engine.ApplyResult{
			Manifest: plan.Manifest,
			Applied:  0,
			Failed:   len(plan.Updates),
			Errors:   []string{fmt.Sprintf("invalid manifest path: %s", plan.Manifest.Path)},
		}, nil
	}
	cmd := exec.CommandContext(ctx, "pre-commit", "autoupdate", "--config", plan.Manifest.Path) // #nosec G204 - manifest path is validated above
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &engine.ApplyResult{
			Manifest: plan.Manifest,
			Applied:  0,
			Failed:   len(plan.Updates),
			Errors:   []string{fmt.Sprintf("autoupdate failed: %v\n%s", err, output)},
		}, nil
	}

	// Read new content for diff
	newContent, err := secureio.ReadFile(plan.Manifest.Path)
	if err != nil {
		return nil, fmt.Errorf("read updated config: %w", err)
	}

	// Generate diff
	diff := generateDiff(string(oldContent), string(newContent))

	// Count actual updates from output
	applied := len(i.parseAutoupdateOutput(string(output)))

	return &engine.ApplyResult{
		Manifest:     plan.Manifest,
		Applied:      applied,
		Failed:       0,
		ManifestDiff: diff,
	}, nil
}

// Validate runs pre-commit validate-config.
func (i *Integration) Validate(ctx context.Context, manifest *engine.Manifest) error {
	if !i.isPreCommitAvailable() {
		return nil // Skip validation if pre-commit not available
	}

	// Validate manifest path to prevent command injection
	if !filepath.IsAbs(manifest.Path) || strings.Contains(manifest.Path, "..") {
		return fmt.Errorf("invalid manifest path: %s", manifest.Path)
	}
	cmd := exec.CommandContext(ctx, "pre-commit", "validate-config", manifest.Path) // #nosec G204 - manifest path is validated above
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("validation failed: %v\n%s", err, output)
	}

	return nil
}

// isPreCommitAvailable checks if pre-commit command is available.
func (i *Integration) isPreCommitAvailable() bool {
	_, err := exec.LookPath("pre-commit")
	return err == nil
}

// generateDiff creates a simple diff between old and new content.
func generateDiff(old, new string) string {
	if old == new {
		return ""
	}

	oldLines := strings.Split(old, "\n")
	newLines := strings.Split(new, "\n")

	var diff strings.Builder
	diff.WriteString("--- .pre-commit-config.yaml\n")
	diff.WriteString("+++ .pre-commit-config.yaml\n")

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

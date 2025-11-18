package precommit

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santosr2/uptool/internal/engine"
)

func TestNew(t *testing.T) {
	integ := New()
	if integ == nil {
		t.Fatal("New() returned nil")
	}
}

func TestName(t *testing.T) {
	integ := New()
	if integ.Name() != "precommit" {
		t.Errorf("Name() = %q, want %q", integ.Name(), "precommit")
	}
}

func TestDetect(t *testing.T) {
	ctx := context.Background()

	t.Run("finds .pre-commit-config.yaml in root", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".pre-commit-config.yaml")

		yaml := `repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.3.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
`
		if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
			t.Fatal(err)
		}

		integ := New()
		manifests, err := integ.Detect(ctx, tmpDir)

		if err != nil {
			t.Fatalf("Detect() error = %v", err)
		}
		if len(manifests) != 1 {
			t.Fatalf("Detect() found %d manifests, want 1", len(manifests))
		}

		m := manifests[0]
		if m.Path != ".pre-commit-config.yaml" {
			t.Errorf("Detect() path = %q, want %q", m.Path, ".pre-commit-config.yaml")
		}
		if m.Type != "precommit" {
			t.Errorf("Detect() type = %q, want %q", m.Type, "precommit")
		}
		if len(m.Dependencies) != 1 {
			t.Errorf("Detect() dependencies = %d, want 1", len(m.Dependencies))
		}
	})

	t.Run("finds multiple .pre-commit-config.yaml files", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Root config
		rootConfig := filepath.Join(tmpDir, ".pre-commit-config.yaml")
		if err := os.WriteFile(rootConfig, []byte("repos:\n  - repo: https://example.com\n    rev: v1.0.0"), 0644); err != nil {
			t.Fatal(err)
		}

		// Nested config
		nestedDir := filepath.Join(tmpDir, "subproject")
		if err := os.MkdirAll(nestedDir, 0755); err != nil {
			t.Fatal(err)
		}
		nestedConfig := filepath.Join(nestedDir, ".pre-commit-config.yaml")
		if err := os.WriteFile(nestedConfig, []byte("repos:\n  - repo: https://example2.com\n    rev: v2.0.0"), 0644); err != nil {
			t.Fatal(err)
		}

		integ := New()
		manifests, err := integ.Detect(ctx, tmpDir)

		if err != nil {
			t.Fatalf("Detect() error = %v", err)
		}
		if len(manifests) != 2 {
			t.Fatalf("Detect() found %d manifests, want 2", len(manifests))
		}
	})

	t.Run("skips hidden directories", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Root config
		rootConfig := filepath.Join(tmpDir, ".pre-commit-config.yaml")
		if err := os.WriteFile(rootConfig, []byte("repos:\n  - repo: https://example.com\n    rev: v1.0.0"), 0644); err != nil {
			t.Fatal(err)
		}

		// Hidden directory config (should be skipped)
		hiddenDir := filepath.Join(tmpDir, ".git", "hooks")
		if err := os.MkdirAll(hiddenDir, 0755); err != nil {
			t.Fatal(err)
		}
		hiddenConfig := filepath.Join(hiddenDir, ".pre-commit-config.yaml")
		if err := os.WriteFile(hiddenConfig, []byte("repos:\n  - repo: https://hidden.com\n    rev: v1.0.0"), 0644); err != nil {
			t.Fatal(err)
		}

		integ := New()
		manifests, err := integ.Detect(ctx, tmpDir)

		if err != nil {
			t.Fatalf("Detect() error = %v", err)
		}
		if len(manifests) != 1 {
			t.Fatalf("Detect() found %d manifests, want 1 (hidden dirs should be skipped)", len(manifests))
		}
	})

	t.Run("skips invalid YAML", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".pre-commit-config.yaml")

		// Invalid YAML
		if err := os.WriteFile(configPath, []byte("repos: invalid yaml:"), 0644); err != nil {
			t.Fatal(err)
		}

		integ := New()
		manifests, err := integ.Detect(ctx, tmpDir)

		if err != nil {
			t.Fatalf("Detect() error = %v", err)
		}
		if len(manifests) != 0 {
			t.Fatalf("Detect() found %d manifests, want 0 (invalid YAML should be skipped)", len(manifests))
		}
	})

	t.Run("sets metadata fields", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".pre-commit-config.yaml")

		yaml := `repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.3.0
    hooks:
      - id: trailing-whitespace
  - repo: https://github.com/psf/black
    rev: 23.1.0
    hooks:
      - id: black
`
		if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
			t.Fatal(err)
		}

		integ := New()
		manifests, err := integ.Detect(ctx, tmpDir)

		if err != nil || len(manifests) == 0 {
			t.Fatalf("Detect() error = %v", err)
		}

		m := manifests[0]
		if m.Metadata["repos_count"] != 2 {
			t.Errorf("Metadata[repos_count] = %v, want 2", m.Metadata["repos_count"])
		}
	})
}

func TestExtractDependencies(t *testing.T) {
	integ := New()

	t.Run("extracts repository dependencies", func(t *testing.T) {
		config := &Config{
			Repos: []Repo{
				{
					Repo: "https://github.com/pre-commit/pre-commit-hooks",
					Rev:  "v4.3.0",
					Hooks: []Hook{
						{ID: "trailing-whitespace"},
					},
				},
				{
					Repo: "https://github.com/psf/black",
					Rev:  "23.1.0",
					Hooks: []Hook{
						{ID: "black"},
					},
				},
			},
		}

		deps := integ.extractDependencies(config)

		if len(deps) != 2 {
			t.Fatalf("extractDependencies() count = %d, want 2", len(deps))
		}

		for _, dep := range deps {
			if dep.Type != "direct" {
				t.Errorf("dep %q type = %q, want %q", dep.Name, dep.Type, "direct")
			}
			if dep.Registry != "git" {
				t.Errorf("dep %q registry = %q, want %q", dep.Name, dep.Registry, "git")
			}
			if !strings.HasPrefix(dep.Name, "https://") {
				t.Errorf("dep name = %q, want https:// URL", dep.Name)
			}
		}
	})

	t.Run("skips local repos", func(t *testing.T) {
		config := &Config{
			Repos: []Repo{
				{Repo: "local", Rev: ""},
				{Repo: "https://github.com/example/repo", Rev: "v1.0.0"},
			},
		}

		deps := integ.extractDependencies(config)

		if len(deps) != 1 {
			t.Fatalf("extractDependencies() count = %d, want 1 (local should be skipped)", len(deps))
		}
		if deps[0].Name != "https://github.com/example/repo" {
			t.Errorf("extractDependencies() name = %q, want github URL", deps[0].Name)
		}
	})

	t.Run("skips meta repos", func(t *testing.T) {
		config := &Config{
			Repos: []Repo{
				{Repo: "meta", Rev: ""},
				{Repo: "https://github.com/example/repo", Rev: "v1.0.0"},
			},
		}

		deps := integ.extractDependencies(config)

		if len(deps) != 1 {
			t.Fatalf("extractDependencies() count = %d, want 1 (meta should be skipped)", len(deps))
		}
	})

	t.Run("skips empty repos", func(t *testing.T) {
		config := &Config{
			Repos: []Repo{
				{Repo: "", Rev: "v1.0.0"},
				{Repo: "https://github.com/example/repo", Rev: "v1.0.0"},
			},
		}

		deps := integ.extractDependencies(config)

		if len(deps) != 1 {
			t.Fatalf("extractDependencies() count = %d, want 1 (empty should be skipped)", len(deps))
		}
	})

	t.Run("handles empty config", func(t *testing.T) {
		config := &Config{}
		deps := integ.extractDependencies(config)

		if len(deps) != 0 {
			t.Errorf("extractDependencies() count = %d, want 0", len(deps))
		}
	})
}

func TestPlan(t *testing.T) {
	ctx := context.Background()
	integ := New()

	t.Run("returns strategy as native_command", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".pre-commit-config.yaml")

		yaml := `repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.3.0
    hooks:
      - id: trailing-whitespace
`
		if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
			t.Fatal(err)
		}

		manifest := &engine.Manifest{
			Path:         configPath,
			Type:         "precommit",
			Dependencies: []engine.Dependency{},
		}

		plan, err := integ.Plan(ctx, manifest)
		if err != nil {
			t.Fatalf("Plan() error = %v", err)
		}
		if plan.Strategy != "native_command" {
			t.Errorf("Plan() strategy = %q, want %q", plan.Strategy, "native_command")
		}
		// Updates may or may not be empty depending on whether pre-commit is installed
	})
}

func TestParseAutoupdateOutput(t *testing.T) {
	integ := New()

	t.Run("parses update lines", func(t *testing.T) {
		output := `[https://github.com/pre-commit/pre-commit-hooks] updating v4.3.0 -> v6.0.0
[https://github.com/psf/black] updating 23.1.0 -> 24.0.0`

		updates := integ.parseAutoupdateOutput(output)

		if len(updates) != 2 {
			t.Fatalf("parseAutoupdateOutput() count = %d, want 2", len(updates))
		}

		// Check first update
		if updates[0].Dependency.Name != "https://github.com/pre-commit/pre-commit-hooks" {
			t.Errorf("updates[0] name = %q, want pre-commit-hooks URL", updates[0].Dependency.Name)
		}
		if updates[0].Dependency.CurrentVersion != "v4.3.0" {
			t.Errorf("updates[0] current = %q, want %q", updates[0].Dependency.CurrentVersion, "v4.3.0")
		}
		if updates[0].TargetVersion != "v6.0.0" {
			t.Errorf("updates[0] target = %q, want %q", updates[0].TargetVersion, "v6.0.0")
		}

		// Check second update
		if updates[1].Dependency.Name != "https://github.com/psf/black" {
			t.Errorf("updates[1] name = %q, want black URL", updates[1].Dependency.Name)
		}
	})

	t.Run("handles empty output", func(t *testing.T) {
		updates := integ.parseAutoupdateOutput("")

		if len(updates) != 0 {
			t.Errorf("parseAutoupdateOutput() count = %d, want 0", len(updates))
		}
	})

	t.Run("handles output without updates", func(t *testing.T) {
		output := `Checking out pre-commit hooks...
All hooks are up to date.`

		updates := integ.parseAutoupdateOutput(output)

		if len(updates) != 0 {
			t.Errorf("parseAutoupdateOutput() count = %d, want 0", len(updates))
		}
	})

	t.Run("handles malformed lines", func(t *testing.T) {
		output := `[https://github.com/example/repo] invalid format
Some other text
[https://github.com/valid/repo] updating v1.0.0 -> v2.0.0`

		updates := integ.parseAutoupdateOutput(output)

		if len(updates) != 1 {
			t.Fatalf("parseAutoupdateOutput() count = %d, want 1", len(updates))
		}
		if updates[0].Dependency.Name != "https://github.com/valid/repo" {
			t.Errorf("parseAutoupdateOutput() name = %q, want valid repo", updates[0].Dependency.Name)
		}
	})
}

func TestDetermineImpact(t *testing.T) {
	integ := New()

	tests := []struct {
		name string
		old  string
		new  string
		want string
	}{
		{"major version change", "v1.0.0", "v2.0.0", "major"},
		{"major without v", "1.0.0", "2.0.0", "major"},
		{"minor version change", "v1.0.0", "v1.1.0", "minor"},
		{"minor without v", "1.0.0", "1.1.0", "minor"},
		{"patch version change", "v1.2.3", "v1.2.4", "patch"},
		{"patch without v", "1.2.3", "1.2.4", "patch"},
		{"same major and minor", "v1.0.0", "v1.0.1", "patch"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := integ.determineImpact(tt.old, tt.new)
			if got != tt.want {
				t.Errorf("determineImpact(%q, %q) = %q, want %q", tt.old, tt.new, got, tt.want)
			}
		})
	}
}

func TestApply(t *testing.T) {
	ctx := context.Background()
	integ := New()

	t.Run("returns early for no updates", func(t *testing.T) {
		manifest := &engine.Manifest{
			Path: ".pre-commit-config.yaml",
		}

		plan := &engine.UpdatePlan{
			Manifest: manifest,
			Updates:  []engine.Update{},
		}

		result, err := integ.Apply(ctx, plan)
		if err != nil {
			t.Fatalf("Apply() error = %v", err)
		}
		if result.Applied != 0 {
			t.Errorf("Apply() applied = %d, want 0", result.Applied)
		}
	})

	t.Run("returns error when pre-commit not available", func(t *testing.T) {
		// Skip if pre-commit is actually installed
		if integ.isPreCommitAvailable() {
			t.Skip("pre-commit is installed, skipping unavailability test")
		}

		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".pre-commit-config.yaml")
		if err := os.WriteFile(configPath, []byte("repos: []"), 0644); err != nil {
			t.Fatal(err)
		}

		manifest := &engine.Manifest{Path: configPath}
		update := engine.Update{
			Dependency:    engine.Dependency{Name: "https://example.com"},
			TargetVersion: "v2.0.0",
		}

		plan := &engine.UpdatePlan{
			Manifest: manifest,
			Updates:  []engine.Update{update},
		}

		result, err := integ.Apply(ctx, plan)
		if err == nil {
			t.Error("Apply() error = nil, want error when pre-commit not available")
		}
		if result != nil {
			t.Errorf("Apply() result = %v, want nil on error", result)
		}
	})
}

func TestIsPreCommitAvailable(t *testing.T) {
	integ := New()

	// Test that the function returns a boolean
	available := integ.isPreCommitAvailable()

	// Just verify it's a valid boolean value
	if available {
		t.Log("pre-commit is available")
	} else {
		t.Log("pre-commit is not available")
	}
}

func TestGenerateDiff(t *testing.T) {
	t.Run("returns empty string for identical content", func(t *testing.T) {
		diff := generateDiff("test", "test")
		if diff != "" {
			t.Errorf("generateDiff() = %q, want empty string", diff)
		}
	})

	t.Run("generates diff for different content", func(t *testing.T) {
		old := `repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.3.0
    hooks:
      - id: trailing-whitespace
`
		new := `repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v6.0.0
    hooks:
      - id: trailing-whitespace
`
		diff := generateDiff(old, new)
		if diff == "" {
			t.Error("generateDiff() returned empty string, want diff")
		}
		if !strings.Contains(diff, "--- .pre-commit-config.yaml") {
			t.Error("generateDiff() missing header")
		}
		if !strings.Contains(diff, "+++ .pre-commit-config.yaml") {
			t.Error("generateDiff() missing header")
		}
		if !strings.Contains(diff, "- ") && !strings.Contains(diff, "+ ") {
			t.Error("generateDiff() missing change markers")
		}
	})

	t.Run("handles different line counts", func(t *testing.T) {
		old := "line1\nline2"
		new := "line1\nline2\nline3"

		diff := generateDiff(old, new)
		if diff == "" {
			t.Error("generateDiff() returned empty string, want diff")
		}
		if !strings.Contains(diff, "+ line3") {
			t.Error("generateDiff() missing added line")
		}
	})
}

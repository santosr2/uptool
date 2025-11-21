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

package tflint

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santosr2/uptool/internal/engine"
)

const testAWSPlugin = `plugin "aws" {
  enabled = true
  version = "0.1.0"
  source  = "github.com/terraform-linters/tflint-ruleset-aws"
}
`

func TestNew(t *testing.T) {
	integ := New()
	if integ == nil {
		t.Fatal("New() returned nil")
	}
	if integ.ds == nil {
		t.Error("New() datasource is nil")
	}
}

func TestName(t *testing.T) {
	integ := New()
	if integ.Name() != integrationName {
		t.Errorf("Name() = %q, want %q", integ.Name(), integrationName)
	}
}

func TestDetect(t *testing.T) {
	ctx := context.Background()

	t.Run("finds .tflint.hcl in root", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".tflint.hcl")

		if err := os.WriteFile(configPath, []byte(testAWSPlugin), 0o644); err != nil {
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
		if m.Path != ".tflint.hcl" {
			t.Errorf("Detect() path = %q, want %q", m.Path, ".tflint.hcl")
		}
		if m.Type != integrationName {
			t.Errorf("Detect() type = %q, want %q", m.Type, integrationName)
		}
		if len(m.Dependencies) != 1 {
			t.Errorf("Detect() dependencies = %d, want 1", len(m.Dependencies))
		}
	})

	t.Run("finds multiple .tflint.hcl files", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Root config
		rootConfig := filepath.Join(tmpDir, ".tflint.hcl")
		if err := os.WriteFile(rootConfig, []byte(`plugin "aws" { enabled = true }`), 0o644); err != nil {
			t.Fatal(err)
		}

		// Nested config
		nestedDir := filepath.Join(tmpDir, "modules", "vpc")
		if err := os.MkdirAll(nestedDir, 0o755); err != nil {
			t.Fatal(err)
		}
		nestedConfig := filepath.Join(nestedDir, ".tflint.hcl")
		if err := os.WriteFile(nestedConfig, []byte(`plugin "azurerm" { enabled = true }`), 0o644); err != nil {
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
		rootConfig := filepath.Join(tmpDir, ".tflint.hcl")
		if err := os.WriteFile(rootConfig, []byte(`plugin "aws" { enabled = true }`), 0o644); err != nil {
			t.Fatal(err)
		}

		// Hidden directory config (should be skipped)
		hiddenDir := filepath.Join(tmpDir, ".terraform")
		if err := os.MkdirAll(hiddenDir, 0o755); err != nil {
			t.Fatal(err)
		}
		hiddenConfig := filepath.Join(hiddenDir, ".tflint.hcl")
		if err := os.WriteFile(hiddenConfig, []byte(`plugin "hidden" { enabled = true }`), 0o644); err != nil {
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

	t.Run("returns error for invalid HCL", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".tflint.hcl")

		// Invalid HCL
		if err := os.WriteFile(configPath, []byte(`plugin "aws" { invalid syntax`), 0o644); err != nil {
			t.Fatal(err)
		}

		integ := New()
		_, err := integ.Detect(ctx, tmpDir)

		if err == nil {
			t.Fatal("Detect() expected error for invalid HCL, got nil")
		}
	})

	t.Run("sets metadata fields", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".tflint.hcl")

		hcl := testAWSPlugin + `
plugin "azurerm" {
  enabled = true
  version = "0.2.0"
  source  = "github.com/terraform-linters/tflint-ruleset-azurerm"
}

rule "terraform_naming_convention" {
  enabled = true
}
`
		if err := os.WriteFile(configPath, []byte(hcl), 0o644); err != nil {
			t.Fatal(err)
		}

		integ := New()
		manifests, err := integ.Detect(ctx, tmpDir)

		if err != nil || len(manifests) == 0 {
			t.Fatalf("Detect() error = %v", err)
		}

		m := manifests[0]
		if m.Metadata["plugins_count"] != 2 {
			t.Errorf("Metadata[plugins_count] = %v, want 2", m.Metadata["plugins_count"])
		}
		if m.Metadata["rules_count"] != 1 {
			t.Errorf("Metadata[rules_count] = %v, want 1", m.Metadata["rules_count"])
		}
	})
}

func TestExtractDependencies(t *testing.T) {
	integ := New()

	t.Run("extracts plugin dependencies", func(t *testing.T) {
		config := &Config{
			Plugins: []Plugin{
				{
					Name:    "aws",
					Version: "0.1.0",
					Source:  "github.com/terraform-linters/tflint-ruleset-aws",
					Enabled: true,
				},
				{
					Name:    "azurerm",
					Version: "0.2.0",
					Source:  "github.com/terraform-linters/tflint-ruleset-azurerm",
					Enabled: true,
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
			if dep.Registry != "github" {
				t.Errorf("dep %q registry = %q, want %q", dep.Name, dep.Registry, "github")
			}
			if !strings.Contains(dep.Name, "github.com") {
				t.Errorf("dep name = %q, want github.com URL", dep.Name)
			}
		}
	})

	t.Run("skips plugins without source", func(t *testing.T) {
		config := &Config{
			Plugins: []Plugin{
				{
					Name:    "aws",
					Version: "0.1.0",
					Enabled: true,
					// No source
				},
			},
		}

		deps := integ.extractDependencies(config)

		if len(deps) != 0 {
			t.Fatalf("extractDependencies() count = %d, want 0 (no source should be skipped)", len(deps))
		}
	})

	t.Run("handles empty plugins", func(t *testing.T) {
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

	t.Run("returns empty plan for no plugins", func(t *testing.T) {
		hcl := `rule "terraform_naming_convention" {
  enabled = true
}
`
		manifest := &engine.Manifest{
			Path:         ".tflint.hcl",
			Type:         "tflint",
			Dependencies: []engine.Dependency{},
			Content:      []byte(hcl),
		}

		plan, err := integ.Plan(ctx, manifest)
		if err != nil {
			t.Fatalf("Plan() error = %v", err)
		}
		if len(plan.Updates) != 0 {
			t.Errorf("Plan() updates = %d, want 0", len(plan.Updates))
		}
		if plan.Strategy != "hcl_rewrite" {
			t.Errorf("Plan() strategy = %q, want %q", plan.Strategy, "hcl_rewrite")
		}
	})

	t.Run("skips plugins without source", func(t *testing.T) {
		hcl := `plugin "aws" {
  enabled = true
  version = "0.1.0"
}
`
		manifest := &engine.Manifest{
			Path:    ".tflint.hcl",
			Type:    "tflint",
			Content: []byte(hcl),
		}

		plan, err := integ.Plan(ctx, manifest)
		if err != nil {
			t.Fatalf("Plan() error = %v", err)
		}
		// Should skip because no source
		if len(plan.Updates) != 0 {
			t.Errorf("Plan() updates = %d, want 0 (no source should be skipped)", len(plan.Updates))
		}
	})

	t.Run("skips plugins without version", func(t *testing.T) {
		hcl := `plugin "aws" {
  enabled = true
  source  = "github.com/terraform-linters/tflint-ruleset-aws"
}
`
		manifest := &engine.Manifest{
			Path:    ".tflint.hcl",
			Type:    "tflint",
			Content: []byte(hcl),
		}

		plan, err := integ.Plan(ctx, manifest)
		if err != nil {
			t.Fatalf("Plan() error = %v", err)
		}
		// Should skip because no version
		if len(plan.Updates) != 0 {
			t.Errorf("Plan() updates = %d, want 0 (no version should be skipped)", len(plan.Updates))
		}
	})

	t.Run("parses various source formats", func(t *testing.T) {
		tests := []struct {
			name   string
			source string
		}{
			{"with https", "https://github.com/owner/repo"},
			{"with http", "http://github.com/owner/repo"},
			{"without protocol", "github.com/owner/repo"},
			{"with .git suffix", "github.com/owner/repo.git"},
			{"with extra path", "github.com/owner/repo/extra"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				hcl := `plugin "test" {
  enabled = true
  version = "1.0.0"
  source  = "` + tt.source + `"
}
`
				manifest := &engine.Manifest{
					Path:    ".tflint.hcl",
					Type:    "tflint",
					Content: []byte(hcl),
				}

				_, err := integ.Plan(ctx, manifest)
				if err != nil {
					t.Fatalf("Plan() error = %v", err)
				}
				// Just verify it doesn't crash on parsing
			})
		}
	})
}

func TestApply(t *testing.T) {
	ctx := context.Background()
	integ := New()

	t.Run("returns early for no updates", func(t *testing.T) {
		manifest := &engine.Manifest{
			Path: ".tflint.hcl",
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

	t.Run("applies updates to plugin versions", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".tflint.hcl")

		if err := os.WriteFile(configPath, []byte(testAWSPlugin), 0o644); err != nil {
			t.Fatal(err)
		}

		manifest := &engine.Manifest{
			Path: configPath,
		}

		update := engine.Update{
			Dependency: engine.Dependency{
				Name:           "github.com/terraform-linters/tflint-ruleset-aws",
				CurrentVersion: "0.1.0",
				Type:           "direct",
			},
			TargetVersion: "0.2.0",
		}

		plan := &engine.UpdatePlan{
			Manifest: manifest,
			Updates:  []engine.Update{update},
		}

		result, err := integ.Apply(ctx, plan)
		if err != nil {
			t.Fatalf("Apply() error = %v", err)
		}
		if result.Applied != 1 {
			t.Errorf("Apply() applied = %d, want 1", result.Applied)
		}

		// Verify file was updated
		content, _ := os.ReadFile(configPath) //nolint:errcheck // test data
		contentStr := string(content)

		if !strings.Contains(contentStr, `version = "0.2.0"`) {
			t.Errorf("Apply() version not updated in file, content:\n%s", contentStr)
		}

		if result.ManifestDiff == "" {
			t.Error("Apply() diff should not be empty")
		}
	})

	t.Run("updates multiple plugins", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".tflint.hcl")

		hcl := `plugin "aws" {
  enabled = true
  version = "0.1.0"
  source  = "github.com/terraform-linters/tflint-ruleset-aws"
}

plugin "azurerm" {
  enabled = true
  version = "0.2.0"
  source  = "github.com/terraform-linters/tflint-ruleset-azurerm"
}
`
		if err := os.WriteFile(configPath, []byte(hcl), 0o644); err != nil {
			t.Fatal(err)
		}

		manifest := &engine.Manifest{Path: configPath}
		updates := []engine.Update{
			{
				Dependency: engine.Dependency{
					Name: "github.com/terraform-linters/tflint-ruleset-aws",
				},
				TargetVersion: "0.1.5",
			},
			{
				Dependency: engine.Dependency{
					Name: "github.com/terraform-linters/tflint-ruleset-azurerm",
				},
				TargetVersion: "0.3.0",
			},
		}

		plan := &engine.UpdatePlan{
			Manifest: manifest,
			Updates:  updates,
		}

		result, err := integ.Apply(ctx, plan)
		if err != nil {
			t.Fatalf("Apply() error = %v", err)
		}
		if result.Applied != 2 {
			t.Errorf("Apply() applied = %d, want 2", result.Applied)
		}

		content, _ := os.ReadFile(configPath) //nolint:errcheck // test data
		contentStr := string(content)

		if !strings.Contains(contentStr, `version = "0.1.5"`) {
			t.Error("Apply() aws version not updated")
		}
		if !strings.Contains(contentStr, `version = "0.3.0"`) {
			t.Error("Apply() azurerm version not updated")
		}
	})
}

func TestValidate(t *testing.T) {
	ctx := context.Background()
	integ := New()

	t.Run("validates correct HCL", func(t *testing.T) {
		hcl := `plugin "aws" {
  enabled = true
  version = "0.1.0"
  source  = "github.com/terraform-linters/tflint-ruleset-aws"
}
`
		manifest := &engine.Manifest{
			Path:    ".tflint.hcl",
			Content: []byte(hcl),
		}

		err := integ.Validate(ctx, manifest)
		if err != nil {
			t.Errorf("Validate() error = %v, want nil", err)
		}
	})

	t.Run("fails for invalid HCL", func(t *testing.T) {
		manifest := &engine.Manifest{
			Path:    ".tflint.hcl",
			Content: []byte(`plugin "aws" { invalid syntax`),
		}

		err := integ.Validate(ctx, manifest)
		if err == nil {
			t.Error("Validate() error = nil, want error for invalid HCL")
		}
	})
}

func TestDetermineImpact(t *testing.T) {
	tests := []struct {
		name string
		old  string
		new  string
		want string
	}{
		{"major version change", "1.0.0", "2.0.0", "major"},
		{"major with v prefix", "v1.0.0", "v2.0.0", "major"},
		{"minor version change", "1.0.0", "1.1.0", "minor"},
		{"minor with v prefix", "v1.2.0", "v1.3.0", "minor"},
		{"patch version change", "1.2.3", "1.2.4", "patch"},
		{"patch with v prefix", "v1.2.3", "v1.2.4", "patch"},
		{"same major and minor", "1.0.0", "1.0.1", "patch"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineImpact(tt.old, tt.new)
			if got != tt.want {
				t.Errorf("determineImpact(%q, %q) = %q, want %q", tt.old, tt.new, got, tt.want)
			}
		})
	}
}

func TestGenerateDiff(t *testing.T) {
	t.Run("returns empty string for identical content", func(t *testing.T) {
		diff := generateDiff("test", "test")
		if diff != "" {
			t.Errorf("generateDiff() = %q, want empty string", diff)
		}
	})

	t.Run("generates diff for version changes", func(t *testing.T) {
		old := `plugin "aws" {
  enabled = true
  version = "0.1.0"
  source  = "github.com/terraform-linters/tflint-ruleset-aws"
}
`
		updated := `plugin "aws" {
  enabled = true
  version = "0.2.0"
  source  = "github.com/terraform-linters/tflint-ruleset-aws"
}
`
		diff := generateDiff(old, updated)
		if diff == "" {
			t.Error("generateDiff() returned empty string, want diff")
		}
		if !strings.Contains(diff, "--- .tflint.hcl") {
			t.Error("generateDiff() missing header")
		}
		if !strings.Contains(diff, "+++ .tflint.hcl") {
			t.Error("generateDiff() missing header")
		}
		if !strings.Contains(diff, `version = "0.1.0"`) {
			t.Error("generateDiff() missing old version line")
		}
		if !strings.Contains(diff, `version = "0.2.0"`) {
			t.Error("generateDiff() missing new version line")
		}
	})

	t.Run("only includes version lines", func(t *testing.T) {
		old := `plugin "aws" {
  enabled = true
  version = "0.1.0"
}
`
		updated := `plugin "aws" {
  enabled = false
  version = "0.2.0"
}
`
		diff := generateDiff(old, updated)
		if !strings.Contains(diff, "version") {
			t.Error("generateDiff() missing version line in diff")
		}
		// Should not include enabled change
		if strings.Contains(diff, "enabled") {
			t.Error("generateDiff() should only include version changes")
		}
	})
}

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

//nolint:dupl,goconst,govet // Test files use similar table-driven patterns
package actions

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santosr2/uptool/internal/datasource"
	"github.com/santosr2/uptool/internal/engine"
)

func TestNew(t *testing.T) {
	integration := New()
	if integration == nil {
		t.Fatal("New() returned nil")
	}
	if integration.ds == nil {
		t.Error("New() created integration with nil datasource")
	}
}

func TestIntegration_Name(t *testing.T) {
	integration := New()
	if got := integration.Name(); got != "actions" {
		t.Errorf("Name() = %q, want %q", got, "actions")
	}
}

func TestDetermineVersionType(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{"semver tag v4", "v4", "tag"},
		{"semver tag v4.2.2", "v4.2.2", "tag"},
		{"semver tag v1.0.0", "v1.0.0", "tag"},
		{"sha commit", "11bd71901bbe5b1630ceea73d27597364c9af683", "sha"},
		{"branch main", "main", "ref"},
		{"branch master", "master", "ref"},
		{"branch develop", "develop", "ref"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineVersionType(tt.version)
			if got != tt.expected {
				t.Errorf("determineVersionType(%q) = %q, want %q", tt.version, got, tt.expected)
			}
		})
	}
}

func TestIsHex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid hex lowercase", "11bd71901bbe5b1630ceea73d27597364c9af683", true},
		{"valid hex uppercase", "11BD71901BBE5B1630CEEA73D27597364C9AF683", true},
		{"valid hex mixed", "11Bd71901BbE5B1630CeEa73D27597364c9AF683", true},
		{"short hex", "abc123", true},
		{"contains non-hex", "11bd71901bbe5b1630ceea73d27597364c9af68g", false},
		{"contains dash", "11bd-7190", false},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isHex(tt.input)
			if got != tt.expected {
				t.Errorf("isHex(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIntegration_ExtractDependencies(t *testing.T) {
	integration := New()

	tests := []struct {
		name            string
		content         string
		expectedCount   int
		expectedActions []string
	}{
		{
			name: "single action",
			content: `
name: CI
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
`,
			expectedCount:   1,
			expectedActions: []string{"actions/checkout"},
		},
		{
			name: "multiple actions",
			content: `
name: CI
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
      - uses: actions/cache@v3
`,
			expectedCount:   3,
			expectedActions: []string{"actions/checkout", "actions/setup-node", "actions/cache"},
		},
		{
			name: "with SHA pinned action",
			content: `
name: CI
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683
`,
			expectedCount:   1,
			expectedActions: []string{"actions/checkout"},
		},
		{
			name: "skip local actions",
			content: `
name: CI
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: ./.github/actions/my-action
      - uses: actions/checkout@v4
`,
			expectedCount:   1,
			expectedActions: []string{"actions/checkout"},
		},
		{
			name: "skip docker actions",
			content: `
name: CI
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: docker://alpine:3.8
      - uses: actions/checkout@v4
`,
			expectedCount:   1,
			expectedActions: []string{"actions/checkout"},
		},
		{
			name: "deduplicate same action",
			content: `
name: CI
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/checkout@v4
`,
			expectedCount:   1,
			expectedActions: []string{"actions/checkout"},
		},
		{
			name: "multiple jobs",
			content: `
name: CI
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
`,
			expectedCount:   2,
			expectedActions: []string{"actions/checkout", "actions/setup-go"},
		},
		{
			name:            "no jobs",
			content:         `name: Empty`,
			expectedCount:   0,
			expectedActions: []string{},
		},
		{
			name:            "invalid yaml",
			content:         `invalid: [yaml`,
			expectedCount:   0,
			expectedActions: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, _ := integration.extractDependencies([]byte(tt.content))

			if len(deps) != tt.expectedCount {
				t.Errorf("extractDependencies() returned %d deps, want %d", len(deps), tt.expectedCount)
			}

			for i, expectedAction := range tt.expectedActions {
				if i < len(deps) && deps[i].Name != expectedAction {
					t.Errorf("deps[%d].Name = %q, want %q", i, deps[i].Name, expectedAction)
				}
			}
		})
	}
}

func TestIntegration_Detect(t *testing.T) {
	integration := New()
	ctx := context.Background()

	t.Run("detects workflow files", func(t *testing.T) {
		tmpDir := t.TempDir()
		workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
		if err := os.MkdirAll(workflowsDir, 0o755); err != nil {
			t.Fatal(err)
		}

		workflowContent := `
name: CI
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
`
		if err := os.WriteFile(filepath.Join(workflowsDir, "ci.yml"), []byte(workflowContent), 0o644); err != nil {
			t.Fatal(err)
		}

		manifests, err := integration.Detect(ctx, tmpDir)
		if err != nil {
			t.Fatalf("Detect() error = %v", err)
		}

		if len(manifests) != 1 {
			t.Errorf("Detect() returned %d manifests, want 1", len(manifests))
		}

		if len(manifests) > 0 {
			if manifests[0].Type != "actions" {
				t.Errorf("manifest.Type = %q, want %q", manifests[0].Type, "actions")
			}
			if len(manifests[0].Dependencies) != 1 {
				t.Errorf("manifest has %d deps, want 1", len(manifests[0].Dependencies))
			}
		}
	})

	t.Run("handles multiple workflow files", func(t *testing.T) {
		tmpDir := t.TempDir()
		workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
		if err := os.MkdirAll(workflowsDir, 0o755); err != nil {
			t.Fatal(err)
		}

		workflow1 := `
name: CI
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
`
		workflow2 := `
name: Release
on: release
jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: goreleaser/goreleaser-action@v5
`
		if err := os.WriteFile(filepath.Join(workflowsDir, "ci.yml"), []byte(workflow1), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(workflowsDir, "release.yaml"), []byte(workflow2), 0o644); err != nil {
			t.Fatal(err)
		}

		manifests, err := integration.Detect(ctx, tmpDir)
		if err != nil {
			t.Fatalf("Detect() error = %v", err)
		}

		if len(manifests) != 2 {
			t.Errorf("Detect() returned %d manifests, want 2", len(manifests))
		}
	})

	t.Run("skips non-yaml files", func(t *testing.T) {
		tmpDir := t.TempDir()
		workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
		if err := os.MkdirAll(workflowsDir, 0o755); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(filepath.Join(workflowsDir, "README.md"), []byte("# Workflows"), 0o644); err != nil {
			t.Fatal(err)
		}

		manifests, err := integration.Detect(ctx, tmpDir)
		if err != nil {
			t.Fatalf("Detect() error = %v", err)
		}

		if len(manifests) != 0 {
			t.Errorf("Detect() returned %d manifests, want 0", len(manifests))
		}
	})

	t.Run("returns empty for missing workflows dir", func(t *testing.T) {
		tmpDir := t.TempDir()

		manifests, err := integration.Detect(ctx, tmpDir)
		if err != nil {
			t.Fatalf("Detect() error = %v", err)
		}

		if len(manifests) != 0 {
			t.Errorf("Detect() returned %d manifests, want 0", len(manifests))
		}
	})
}

func TestIntegration_Validate(t *testing.T) {
	integration := New()
	ctx := context.Background()

	tests := []struct {
		name      string
		content   string
		expectErr bool
	}{
		{
			name: "valid workflow",
			content: `
name: CI
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
`,
			expectErr: false,
		},
		{
			name:      "invalid yaml",
			content:   `invalid: [yaml`,
			expectErr: true,
		},
		{
			name: "no jobs",
			content: `
name: Empty
on: push
`,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := &engine.Manifest{
				Content: []byte(tt.content),
			}

			err := integration.Validate(ctx, manifest)
			if (err != nil) != tt.expectErr {
				t.Errorf("Validate() error = %v, expectErr = %v", err, tt.expectErr)
			}
		})
	}
}

func TestIntegration_Plan(t *testing.T) {
	ctx := context.Background()

	t.Run("plans updates for tag versions", func(t *testing.T) {
		mockDS := &mockDatasource{
			versions: []string{"4.2.2", "4.2.1", "4.2.0", "4.1.0", "4.0.0"},
		}
		integration := &Integration{ds: mockDS}

		manifest := &engine.Manifest{
			Path: ".github/workflows/ci.yml",
			Type: "actions",
			Dependencies: []engine.Dependency{
				{
					Name:           "actions/checkout",
					CurrentVersion: "v4.0.0",
					Constraint:     "", // Actions don't have semver constraints like npm
					Type:           "tag",
					Registry:       "github",
				},
			},
		}

		planCtx := engine.NewPlanContext()
		plan, err := integration.Plan(ctx, manifest, planCtx)
		if err != nil {
			t.Fatalf("Plan() error = %v", err)
		}

		if len(plan.Updates) != 1 {
			t.Errorf("Plan() returned %d updates, want 1", len(plan.Updates))
		}

		if len(plan.Updates) > 0 && plan.Updates[0].TargetVersion != "v4.2.2" {
			t.Errorf("Plan() target = %q, want %q", plan.Updates[0].TargetVersion, "v4.2.2")
		}
	})

	t.Run("skips SHA pinned actions", func(t *testing.T) {
		mockDS := &mockDatasource{
			versions: []string{"4.2.2", "4.2.1"},
		}
		integration := &Integration{ds: mockDS}

		manifest := &engine.Manifest{
			Path: ".github/workflows/ci.yml",
			Type: "actions",
			Dependencies: []engine.Dependency{
				{
					Name:           "actions/checkout",
					CurrentVersion: "11bd71901bbe5b1630ceea73d27597364c9af683",
					Constraint:     "11bd71901bbe5b1630ceea73d27597364c9af683",
					Type:           "sha",
					Registry:       "github",
				},
			},
		}

		planCtx := engine.NewPlanContext()
		plan, err := integration.Plan(ctx, manifest, planCtx)
		if err != nil {
			t.Fatalf("Plan() error = %v", err)
		}

		if len(plan.Updates) != 0 {
			t.Errorf("Plan() returned %d updates, want 0 (SHA should be skipped)", len(plan.Updates))
		}
	})

	t.Run("handles no available versions", func(t *testing.T) {
		mockDS := &mockDatasource{
			versions: []string{},
		}
		integration := &Integration{ds: mockDS}

		manifest := &engine.Manifest{
			Path: ".github/workflows/ci.yml",
			Type: "actions",
			Dependencies: []engine.Dependency{
				{
					Name:           "unknown/action",
					CurrentVersion: "v1.0.0",
					Type:           "tag",
				},
			},
		}

		planCtx := engine.NewPlanContext()
		plan, err := integration.Plan(ctx, manifest, planCtx)
		if err != nil {
			t.Fatalf("Plan() error = %v", err)
		}

		if len(plan.Updates) != 0 {
			t.Errorf("Plan() returned %d updates, want 0", len(plan.Updates))
		}
	})
}

func TestIntegration_Apply(t *testing.T) {
	integration := New()
	ctx := context.Background()

	t.Run("applies updates", func(t *testing.T) {
		tmpDir := t.TempDir()
		workflowPath := filepath.Join(tmpDir, "ci.yml")
		originalContent := `name: CI
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4.0.0
`
		if err := os.WriteFile(workflowPath, []byte(originalContent), 0o644); err != nil {
			t.Fatal(err)
		}

		plan := &engine.UpdatePlan{
			Manifest: &engine.Manifest{
				Path: workflowPath,
			},
			Updates: []engine.Update{
				{
					Dependency: engine.Dependency{
						Name:           "actions/checkout",
						CurrentVersion: "v4.0.0",
					},
					TargetVersion: "v4.2.2",
				},
			},
		}

		result, err := integration.Apply(ctx, plan)
		if err != nil {
			t.Fatalf("Apply() error = %v", err)
		}

		if result.Applied != 1 {
			t.Errorf("Apply() applied = %d, want 1", result.Applied)
		}

		updatedContent, _ := os.ReadFile(workflowPath)
		if !strings.Contains(string(updatedContent), "actions/checkout@v4.2.2") {
			t.Errorf("Apply() did not update action reference")
		}

		if result.ManifestDiff == "" {
			t.Error("Apply() did not generate diff")
		}
	})

	t.Run("handles empty updates", func(t *testing.T) {
		plan := &engine.UpdatePlan{
			Manifest: &engine.Manifest{
				Path: "test.yml",
			},
			Updates: []engine.Update{},
		}

		result, err := integration.Apply(ctx, plan)
		if err != nil {
			t.Fatalf("Apply() error = %v", err)
		}

		if result.Applied != 0 {
			t.Errorf("Apply() applied = %d, want 0", result.Applied)
		}
	})

	t.Run("handles invalid path", func(t *testing.T) {
		plan := &engine.UpdatePlan{
			Manifest: &engine.Manifest{
				Path: "../../../etc/passwd",
			},
			Updates: []engine.Update{
				{
					Dependency: engine.Dependency{
						Name:           "test/action",
						CurrentVersion: "v1.0.0",
					},
					TargetVersion: "v2.0.0",
				},
			},
		}

		_, err := integration.Apply(ctx, plan)
		if err == nil {
			t.Error("Apply() expected error for path traversal")
		}
	})
}

func TestGenerateDiff(t *testing.T) {
	t.Run("generates diff for changed lines", func(t *testing.T) {
		old := "      - uses: actions/checkout@v4.0.0"
		newContent := "      - uses: actions/checkout@v4.2.2"

		diff := generateDiff("ci.yml", old, newContent)

		if diff == "" {
			t.Error("generateDiff() returned empty diff")
		}

		if !strings.Contains(diff, "- ") || !strings.Contains(diff, "+ ") {
			t.Error("generateDiff() missing diff markers")
		}
	})

	t.Run("returns empty for identical content", func(t *testing.T) {
		content := "uses: actions/checkout@v4"
		diff := generateDiff("ci.yml", content, content)

		if diff != "" {
			t.Errorf("generateDiff() = %q, want empty string", diff)
		}
	})
}

// mockDatasource is a test double for datasource.Datasource
type mockDatasource struct {
	versions []string
	err      error
}

func (m *mockDatasource) Name() string {
	return "mock"
}

func (m *mockDatasource) GetLatestVersion(ctx context.Context, pkg string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	if len(m.versions) > 0 {
		return m.versions[0], nil
	}
	return "", nil
}

func (m *mockDatasource) GetVersions(ctx context.Context, pkg string) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.versions, nil
}

func (m *mockDatasource) GetPackageInfo(ctx context.Context, pkg string) (*datasource.PackageInfo, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &datasource.PackageInfo{
		Name:     pkg,
		Versions: []datasource.VersionInfo{},
	}, nil
}

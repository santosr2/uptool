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

package asdf

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/santosr2/uptool/internal/engine"
)

func TestNew(t *testing.T) {
	integration := New()
	if integration == nil {
		t.Fatal("New() returned nil")
	}
}

func TestName(t *testing.T) {
	integration := New()
	if got := integration.Name(); got != integrationName {
		t.Errorf("Name() = %q, want %q", got, integrationName)
	}
}

func TestDetect(t *testing.T) {
	tests := []struct {
		setup     func(t *testing.T, dir string)
		name      string
		wantCount int
		wantErr   bool
	}{
		{
			name: "finds .tool-versions in root",
			setup: func(t *testing.T, dir string) {
				content := []byte("nodejs 18.16.0\npython 3.11.0\n")
				if err := os.WriteFile(filepath.Join(dir, ".tool-versions"), content, 0o644); err != nil {
					t.Fatal(err)
				}
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "finds multiple .tool-versions files",
			setup: func(t *testing.T, dir string) {
				content1 := []byte("nodejs 18.16.0\n")
				content2 := []byte("python 3.11.0\n")

				if err := os.WriteFile(filepath.Join(dir, ".tool-versions"), content1, 0o644); err != nil {
					t.Fatal(err)
				}

				subdir := filepath.Join(dir, "subproject")
				if err := os.Mkdir(subdir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(subdir, ".tool-versions"), content2, 0o644); err != nil {
					t.Fatal(err)
				}
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "skips hidden directories",
			setup: func(t *testing.T, dir string) {
				hiddenDir := filepath.Join(dir, ".hidden")
				if err := os.Mkdir(hiddenDir, 0o755); err != nil {
					t.Fatal(err)
				}
				content := []byte("nodejs 18.16.0\n")
				if err := os.WriteFile(filepath.Join(hiddenDir, ".tool-versions"), content, 0o644); err != nil {
					t.Fatal(err)
				}
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "skips node_modules",
			setup: func(t *testing.T, dir string) {
				nmDir := filepath.Join(dir, "node_modules")
				if err := os.Mkdir(nmDir, 0o755); err != nil {
					t.Fatal(err)
				}
				content := []byte("nodejs 18.16.0\n")
				if err := os.WriteFile(filepath.Join(nmDir, ".tool-versions"), content, 0o644); err != nil {
					t.Fatal(err)
				}
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "empty directory",
			setup: func(t *testing.T, dir string) {
				// No files
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "empty .tool-versions file",
			setup: func(t *testing.T, dir string) {
				content := []byte("")
				if err := os.WriteFile(filepath.Join(dir, ".tool-versions"), content, 0o644); err != nil {
					t.Fatal(err)
				}
			},
			wantCount: 1, // File is found, but has zero dependencies
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "asdf-test-*")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			tt.setup(t, tmpDir)

			integration := New()
			manifests, err := integration.Detect(context.Background(), tmpDir)

			if (err != nil) != tt.wantErr {
				t.Errorf("Detect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got := len(manifests); got != tt.wantCount {
				t.Errorf("Detect() found %d manifests, want %d", got, tt.wantCount)
			}

			// Verify all manifests have correct type
			for _, m := range manifests {
				if m.Type != "asdf" {
					t.Errorf("Manifest type = %q, want %q", m.Type, "asdf")
				}
			}
		})
	}
}

func TestParseToolVersions(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantDeps    []engine.Dependency
		wantDepsLen int
		wantErr     bool
	}{
		{
			name:        "single tool",
			content:     "nodejs 18.16.0\n",
			wantDepsLen: 1,
			wantDeps: []engine.Dependency{
				{Name: "nodejs", CurrentVersion: "18.16.0", Type: "runtime"},
			},
			wantErr: false,
		},
		{
			name:        "multiple tools",
			content:     "nodejs 18.16.0\npython 3.11.0\nruby 3.2.0\n",
			wantDepsLen: 3,
			wantDeps: []engine.Dependency{
				{Name: "nodejs", CurrentVersion: "18.16.0", Type: "runtime"},
				{Name: "python", CurrentVersion: "3.11.0", Type: "runtime"},
				{Name: "ruby", CurrentVersion: "3.2.0", Type: "runtime"},
			},
			wantErr: false,
		},
		{
			name:        "with comments",
			content:     "# This is a comment\nnodejs 18.16.0\n# Another comment\npython 3.11.0\n",
			wantDepsLen: 2,
			wantDeps: []engine.Dependency{
				{Name: "nodejs", CurrentVersion: "18.16.0", Type: "runtime"},
				{Name: "python", CurrentVersion: "3.11.0", Type: "runtime"},
			},
			wantErr: false,
		},
		{
			name:        "with empty lines",
			content:     "nodejs 18.16.0\n\npython 3.11.0\n\n",
			wantDepsLen: 2,
			wantDeps: []engine.Dependency{
				{Name: "nodejs", CurrentVersion: "18.16.0", Type: "runtime"},
				{Name: "python", CurrentVersion: "3.11.0", Type: "runtime"},
			},
			wantErr: false,
		},
		{
			name:        "empty file",
			content:     "",
			wantDepsLen: 0,
			wantDeps:    []engine.Dependency{},
			wantErr:     false,
		},
		{
			name:        "comments only",
			content:     "# Comment 1\n# Comment 2\n",
			wantDepsLen: 0,
			wantDeps:    []engine.Dependency{},
			wantErr:     false,
		},
		{
			name:        "whitespace only",
			content:     "   \n\t\n   \n",
			wantDepsLen: 0,
			wantDeps:    []engine.Dependency{},
			wantErr:     false,
		},
		{
			name:        "invalid line with single field",
			content:     "nodejs\n",
			wantDepsLen: 0,
			wantDeps:    []engine.Dependency{},
			wantErr:     false,
		},
		{
			name:        "mixed valid and invalid lines",
			content:     "nodejs 18.16.0\ninvalid\npython 3.11.0\n",
			wantDepsLen: 2,
			wantDeps: []engine.Dependency{
				{Name: "nodejs", CurrentVersion: "18.16.0", Type: "runtime"},
				{Name: "python", CurrentVersion: "3.11.0", Type: "runtime"},
			},
			wantErr: false,
		},
		{
			name:        "extra fields ignored",
			content:     "nodejs 18.16.0 extra fields here\n",
			wantDepsLen: 1,
			wantDeps: []engine.Dependency{
				{Name: "nodejs", CurrentVersion: "18.16.0", Type: "runtime"},
			},
			wantErr: false,
		},
		{
			name:        "version with special characters",
			content:     "erlang 26.0.2\nnode 20.5.1\nruby 3.2.2\n",
			wantDepsLen: 3,
			wantDeps: []engine.Dependency{
				{Name: "erlang", CurrentVersion: "26.0.2", Type: "runtime"},
				{Name: "node", CurrentVersion: "20.5.1", Type: "runtime"},
				{Name: "ruby", CurrentVersion: "3.2.2", Type: "runtime"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			integration := New()
			manifest := &engine.Manifest{
				Path:         "/test/.tool-versions",
				Type:         "asdf",
				Dependencies: []engine.Dependency{},
				Content:      []byte(tt.content),
				Metadata:     map[string]interface{}{},
			}

			result, err := integration.parseToolVersions(manifest, []byte(tt.content))

			if (err != nil) != tt.wantErr {
				t.Errorf("parseToolVersions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			if got := len(result.Dependencies); got != tt.wantDepsLen {
				t.Errorf("parseToolVersions() got %d dependencies, want %d", got, tt.wantDepsLen)
			}

			// Verify each expected dependency
			for i, wantDep := range tt.wantDeps {
				if i >= len(result.Dependencies) {
					t.Errorf("Missing dependency at index %d", i)
					continue
				}
				gotDep := result.Dependencies[i]
				if gotDep.Name != wantDep.Name {
					t.Errorf("Dependency[%d].Name = %q, want %q", i, gotDep.Name, wantDep.Name)
				}
				if gotDep.CurrentVersion != wantDep.CurrentVersion {
					t.Errorf("Dependency[%d].CurrentVersion = %q, want %q", i, gotDep.CurrentVersion, wantDep.CurrentVersion)
				}
				if gotDep.Type != wantDep.Type {
					t.Errorf("Dependency[%d].Type = %q, want %q", i, gotDep.Type, wantDep.Type)
				}
			}
		})
	}
}

func TestParseManifest(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantDepsLen int
		wantErr     bool
	}{
		{
			name:        "valid manifest",
			content:     "nodejs 18.16.0\npython 3.11.0\n",
			wantDepsLen: 2,
			wantErr:     false,
		},
		{
			name:        "empty manifest",
			content:     "",
			wantDepsLen: 0,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "asdf-parse-test-*")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			manifestPath := filepath.Join(tmpDir, ".tool-versions")
			err = os.WriteFile(manifestPath, []byte(tt.content), 0o644)
			if err != nil {
				t.Fatal(err)
			}

			integration := New()
			manifest, err := integration.parseManifest(manifestPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseManifest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			if manifest == nil {
				t.Fatal("parseManifest() returned nil manifest")
			}

			if manifest.Type != "asdf" {
				t.Errorf("Manifest.Type = %q, want %q", manifest.Type, "asdf")
			}

			if manifest.Path != manifestPath {
				t.Errorf("Manifest.Path = %q, want %q", manifest.Path, manifestPath)
			}

			if got := len(manifest.Dependencies); got != tt.wantDepsLen {
				t.Errorf("parseManifest() got %d dependencies, want %d", got, tt.wantDepsLen)
			}

			if len(manifest.Content) == 0 && tt.content != "" {
				t.Error("Manifest.Content is empty but should contain file content")
			}
		})
	}
}

func TestPlan(t *testing.T) {
	tests := []struct {
		name    string
		deps    []engine.Dependency
		wantErr bool
	}{
		{
			name: "returns empty plan",
			deps: []engine.Dependency{
				{Name: "nodejs", CurrentVersion: "18.16.0", Type: "runtime"},
			},
			wantErr: false,
		},
		{
			name:    "handles manifest with no dependencies",
			deps:    []engine.Dependency{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			integration := New()
			manifest := &engine.Manifest{
				Path:         "/test/.tool-versions",
				Type:         "asdf",
				Dependencies: tt.deps,
				Content:      []byte(""),
				Metadata:     map[string]interface{}{},
			}

			plan, err := integration.Plan(context.Background(), manifest, nil)

			if (err != nil) != tt.wantErr {
				t.Errorf("Plan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			if plan == nil {
				t.Fatal("Plan() returned nil")
			}

			// Currently returns empty updates (not implemented)
			if len(plan.Updates) != 0 {
				t.Errorf("Plan() returned %d updates, expected 0 (not yet implemented)", len(plan.Updates))
			}

			if plan.Strategy != "native_command" {
				t.Errorf("Plan().Strategy = %q, want %q", plan.Strategy, "native_command")
			}

			if plan.Manifest != manifest {
				t.Error("Plan().Manifest does not match input manifest")
			}
		})
	}
}

func TestApply(t *testing.T) {
	tests := []struct {
		name        string
		updates     []engine.Update
		wantApplied int
		wantFailed  int
		wantErr     bool
	}{
		{
			name:        "no updates returns success",
			updates:     []engine.Update{},
			wantApplied: 0,
			wantFailed:  0,
			wantErr:     false,
		},
		{
			name: "with updates returns experimental error",
			updates: []engine.Update{
				{
					Dependency:    engine.Dependency{Name: "nodejs", CurrentVersion: "18.16.0"},
					TargetVersion: "20.0.0",
				},
			},
			wantApplied: 0,
			wantFailed:  1,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			integration := New()
			manifest := &engine.Manifest{
				Path:         "/test/.tool-versions",
				Type:         "asdf",
				Dependencies: []engine.Dependency{},
				Content:      []byte(""),
				Metadata:     map[string]interface{}{},
			}

			plan := &engine.UpdatePlan{
				Manifest: manifest,
				Updates:  tt.updates,
				Strategy: "native_command",
			}

			result, err := integration.Apply(context.Background(), plan)

			if (err != nil) != tt.wantErr {
				t.Errorf("Apply() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			if result == nil {
				t.Fatal("Apply() returned nil")
			}

			if result.Applied != tt.wantApplied {
				t.Errorf("Apply().Applied = %d, want %d", result.Applied, tt.wantApplied)
			}

			if result.Failed != tt.wantFailed {
				t.Errorf("Apply().Failed = %d, want %d", result.Failed, tt.wantFailed)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	integration := New()
	manifest := &engine.Manifest{
		Path:         "/test/.tool-versions",
		Type:         "asdf",
		Dependencies: []engine.Dependency{},
		Content:      []byte(""),
		Metadata:     map[string]interface{}{},
	}

	err := integration.Validate(context.Background(), manifest)
	if err != nil {
		t.Errorf("Validate() returned error: %v", err)
	}
}

func TestIntegration_EndToEnd(t *testing.T) {
	// Create temporary directory structure
	tmpDir, err := os.MkdirTemp("", "asdf-e2e-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .tool-versions file
	content := []byte("nodejs 18.16.0\npython 3.11.0\nruby 3.2.0\n")
	toolVersionsPath := filepath.Join(tmpDir, ".tool-versions")
	err = os.WriteFile(toolVersionsPath, content, 0o644)
	if err != nil {
		t.Fatal(err)
	}

	integration := New()
	ctx := context.Background()

	// Step 1: Detect
	manifests, err := integration.Detect(ctx, tmpDir)
	if err != nil {
		t.Fatalf("Detect() error: %v", err)
	}
	if len(manifests) != 1 {
		t.Fatalf("Detect() found %d manifests, want 1", len(manifests))
	}

	manifest := manifests[0]
	if len(manifest.Dependencies) != 3 {
		t.Errorf("Manifest has %d dependencies, want 3", len(manifest.Dependencies))
	}

	// Step 2: Plan
	plan, err := integration.Plan(ctx, manifest, nil)
	if err != nil {
		t.Fatalf("Plan() error: %v", err)
	}
	if plan == nil {
		t.Fatal("Plan() returned nil")
	}

	// Step 3: Apply (currently returns experimental error)
	result, err := integration.Apply(ctx, plan)
	if err != nil {
		t.Fatalf("Apply() error: %v", err)
	}
	if result == nil {
		t.Fatal("Apply() returned nil")
	}

	// Step 4: Validate
	err = integration.Validate(ctx, manifest)
	if err != nil {
		t.Errorf("Validate() error: %v", err)
	}
}

package mise

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
	if got := integration.Name(); got != "mise" {
		t.Errorf("Name() = %q, want %q", got, "mise")
	}
}

func TestDetect(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, dir string)
		wantCount int
		wantErr   bool
	}{
		{
			name: "finds mise.toml in root",
			setup: func(t *testing.T, dir string) {
				content := []byte("[tools]\nnodejs = \"18.16.0\"\npython = \"3.11.0\"\n")
				if err := os.WriteFile(filepath.Join(dir, "mise.toml"), content, 0644); err != nil {
					t.Fatal(err)
				}
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "finds .mise.toml in root",
			setup: func(t *testing.T, dir string) {
				content := []byte("[tools]\nnodejs = \"18.16.0\"\n")
				if err := os.WriteFile(filepath.Join(dir, ".mise.toml"), content, 0644); err != nil {
					t.Fatal(err)
				}
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "finds both mise.toml and .mise.toml",
			setup: func(t *testing.T, dir string) {
				content1 := []byte("[tools]\nnodejs = \"18.16.0\"\n")
				content2 := []byte("[tools]\npython = \"3.11.0\"\n")

				if err := os.WriteFile(filepath.Join(dir, "mise.toml"), content1, 0644); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, ".mise.toml"), content2, 0644); err != nil {
					t.Fatal(err)
				}
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "finds multiple mise.toml files in subdirectories",
			setup: func(t *testing.T, dir string) {
				content1 := []byte("[tools]\nnodejs = \"18.16.0\"\n")
				content2 := []byte("[tools]\npython = \"3.11.0\"\n")

				if err := os.WriteFile(filepath.Join(dir, "mise.toml"), content1, 0644); err != nil {
					t.Fatal(err)
				}

				subdir := filepath.Join(dir, "subproject")
				if err := os.Mkdir(subdir, 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(subdir, "mise.toml"), content2, 0644); err != nil {
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
				if err := os.Mkdir(hiddenDir, 0755); err != nil {
					t.Fatal(err)
				}
				content := []byte("[tools]\nnodejs = \"18.16.0\"\n")
				if err := os.WriteFile(filepath.Join(hiddenDir, "mise.toml"), content, 0644); err != nil {
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
				if err := os.Mkdir(nmDir, 0755); err != nil {
					t.Fatal(err)
				}
				content := []byte("[tools]\nnodejs = \"18.16.0\"\n")
				if err := os.WriteFile(filepath.Join(nmDir, "mise.toml"), content, 0644); err != nil {
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
			name: "empty mise.toml file",
			setup: func(t *testing.T, dir string) {
				content := []byte("")
				if err := os.WriteFile(filepath.Join(dir, "mise.toml"), content, 0644); err != nil {
					t.Fatal(err)
				}
			},
			wantCount: 1, // File is found, but has zero dependencies
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "mise-test-*")
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
				if m.Type != "mise" {
					t.Errorf("Manifest type = %q, want %q", m.Type, "mise")
				}
			}
		})
	}
}

func TestParseMiseToml(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantDepsLen int
		wantDeps    []engine.Dependency
		wantErr     bool
	}{
		{
			name:        "simple string version",
			content:     "[tools]\nnodejs = \"18.16.0\"\n",
			wantDepsLen: 1,
			wantDeps: []engine.Dependency{
				{Name: "nodejs", CurrentVersion: "18.16.0", Type: "runtime"},
			},
			wantErr: false,
		},
		{
			name:        "multiple tools with string versions",
			content:     "[tools]\nnodejs = \"18.16.0\"\npython = \"3.11.0\"\nruby = \"3.2.0\"\n",
			wantDepsLen: 3,
			wantDeps: []engine.Dependency{
				{Name: "nodejs", CurrentVersion: "18.16.0", Type: "runtime"},
				{Name: "python", CurrentVersion: "3.11.0", Type: "runtime"},
				{Name: "ruby", CurrentVersion: "3.2.0", Type: "runtime"},
			},
			wantErr: false,
		},
		{
			name: "map format with version key",
			content: `[tools]
nodejs = { version = "18.16.0" }
`,
			wantDepsLen: 1,
			wantDeps: []engine.Dependency{
				{Name: "nodejs", CurrentVersion: "18.16.0", Type: "runtime"},
			},
			wantErr: false,
		},
		{
			name: "mixed string and map formats",
			content: `[tools]
nodejs = "18.16.0"
python = { version = "3.11.0" }
ruby = "3.2.0"
`,
			wantDepsLen: 3,
			wantDeps: []engine.Dependency{
				{Name: "nodejs", CurrentVersion: "18.16.0", Type: "runtime"},
				{Name: "python", CurrentVersion: "3.11.0", Type: "runtime"},
				{Name: "ruby", CurrentVersion: "3.2.0", Type: "runtime"},
			},
			wantErr: false,
		},
		{
			name: "map without version key ignored",
			content: `[tools]
nodejs = { other_key = "value" }
python = "3.11.0"
`,
			wantDepsLen: 1,
			wantDeps: []engine.Dependency{
				{Name: "python", CurrentVersion: "3.11.0", Type: "runtime"},
			},
			wantErr: false,
		},
		{
			name:        "empty tools section",
			content:     "[tools]\n",
			wantDepsLen: 0,
			wantDeps:    []engine.Dependency{},
			wantErr:     false,
		},
		{
			name:        "no tools section",
			content:     "[other]\nkey = \"value\"\n",
			wantDepsLen: 0,
			wantDeps:    []engine.Dependency{},
			wantErr:     false,
		},
		{
			name:        "empty file",
			content:     "",
			wantDepsLen: 0,
			wantDeps:    []engine.Dependency{},
			wantErr:     false,
		},
		{
			name:        "invalid toml",
			content:     "[tools\ninvalid toml",
			wantDepsLen: 0,
			wantDeps:    nil,
			wantErr:     true,
		},
		{
			name: "complex configuration with comments",
			content: `# This is a comment
[tools]
# Node.js version
nodejs = "18.16.0"
# Python with additional config
python = { version = "3.11.0" }
`,
			wantDepsLen: 2,
			wantDeps: []engine.Dependency{
				{Name: "nodejs", CurrentVersion: "18.16.0", Type: "runtime"},
				{Name: "python", CurrentVersion: "3.11.0", Type: "runtime"},
			},
			wantErr: false,
		},
		{
			name: "version with semantic versioning",
			content: `[tools]
erlang = "26.0.2"
node = "20.5.1"
ruby = "3.2.2"
`,
			wantDepsLen: 3,
			wantDeps: []engine.Dependency{
				{Name: "erlang", CurrentVersion: "26.0.2", Type: "runtime"},
				{Name: "node", CurrentVersion: "20.5.1", Type: "runtime"},
				{Name: "ruby", CurrentVersion: "3.2.2", Type: "runtime"},
			},
			wantErr: false,
		},
		{
			name: "duplicate tools section is invalid",
			content: `[tools]
nodejs = "18.16.0"

[env]
MY_VAR = "value"

[tools]
python = "3.11.0"
`,
			wantDepsLen: 0,
			wantDeps:    nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			integration := New()
			manifest := &engine.Manifest{
				Path:         "/test/mise.toml",
				Type:         "mise",
				Dependencies: []engine.Dependency{},
				Content:      []byte(tt.content),
				Metadata:     map[string]interface{}{},
			}

			result, err := integration.parseMiseToml(manifest, []byte(tt.content))

			if (err != nil) != tt.wantErr {
				t.Errorf("parseMiseToml() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			if got := len(result.Dependencies); got != tt.wantDepsLen {
				t.Errorf("parseMiseToml() got %d dependencies, want %d", got, tt.wantDepsLen)
				// Print what we got for debugging
				if got > 0 {
					t.Logf("Got dependencies:")
					for i, dep := range result.Dependencies {
						t.Logf("  [%d] %s = %s", i, dep.Name, dep.CurrentVersion)
					}
				}
			}

			// Verify each expected dependency exists (order may vary due to map iteration)
			for _, wantDep := range tt.wantDeps {
				found := false
				for _, gotDep := range result.Dependencies {
					if gotDep.Name == wantDep.Name {
						found = true
						if gotDep.CurrentVersion != wantDep.CurrentVersion {
							t.Errorf("Dependency %q: CurrentVersion = %q, want %q",
								gotDep.Name, gotDep.CurrentVersion, wantDep.CurrentVersion)
						}
						if gotDep.Type != wantDep.Type {
							t.Errorf("Dependency %q: Type = %q, want %q",
								gotDep.Name, gotDep.Type, wantDep.Type)
						}
						break
					}
				}
				if !found {
					t.Errorf("Expected dependency %q not found", wantDep.Name)
				}
			}
		})
	}
}

func TestParseManifest(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		content     string
		wantDepsLen int
		wantErr     bool
	}{
		{
			name:        "valid mise.toml",
			filename:    "mise.toml",
			content:     "[tools]\nnodejs = \"18.16.0\"\npython = \"3.11.0\"\n",
			wantDepsLen: 2,
			wantErr:     false,
		},
		{
			name:        "valid .mise.toml",
			filename:    ".mise.toml",
			content:     "[tools]\nnodejs = \"18.16.0\"\n",
			wantDepsLen: 1,
			wantErr:     false,
		},
		{
			name:        "empty manifest",
			filename:    "mise.toml",
			content:     "",
			wantDepsLen: 0,
			wantErr:     false,
		},
		{
			name:        "invalid toml",
			filename:    "mise.toml",
			content:     "[tools\ninvalid",
			wantDepsLen: 0,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "mise-parse-test-*")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			manifestPath := filepath.Join(tmpDir, tt.filename)
			if err := os.WriteFile(manifestPath, []byte(tt.content), 0644); err != nil {
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

			if manifest.Type != "mise" {
				t.Errorf("Manifest.Type = %q, want %q", manifest.Type, "mise")
			}

			if manifest.Path != manifestPath {
				t.Errorf("Manifest.Path = %q, want %q", manifest.Path, manifestPath)
			}

			if got := len(manifest.Dependencies); got != tt.wantDepsLen {
				t.Errorf("parseManifest() got %d dependencies, want %d", got, tt.wantDepsLen)
			}

			if len(manifest.Content) == 0 && len(tt.content) > 0 {
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
				Path:         "/test/mise.toml",
				Type:         "mise",
				Dependencies: tt.deps,
				Content:      []byte(""),
				Metadata:     map[string]interface{}{},
			}

			plan, err := integration.Plan(context.Background(), manifest)

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

			if plan.Strategy != "custom_rewrite" {
				t.Errorf("Plan().Strategy = %q, want %q", plan.Strategy, "custom_rewrite")
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
				Path:         "/test/mise.toml",
				Type:         "mise",
				Dependencies: []engine.Dependency{},
				Content:      []byte(""),
				Metadata:     map[string]interface{}{},
			}

			plan := &engine.UpdatePlan{
				Manifest: manifest,
				Updates:  tt.updates,
				Strategy: "custom_rewrite",
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
		Path:         "/test/mise.toml",
		Type:         "mise",
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
	tmpDir, err := os.MkdirTemp("", "mise-e2e-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create mise.toml file
	content := []byte(`[tools]
nodejs = "18.16.0"
python = "3.11.0"
ruby = "3.2.0"
`)
	miseTomlPath := filepath.Join(tmpDir, "mise.toml")
	if err := os.WriteFile(miseTomlPath, content, 0644); err != nil {
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
	plan, err := integration.Plan(ctx, manifest)
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

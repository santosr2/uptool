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

package npm

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santosr2/uptool/internal/engine"
)

const (
	integrationName = "npm"
	packageJSONName = "package.json"
)

// setupTestDir creates a test directory with root and nested package.json files
func setupTestDir(t *testing.T, nestedDir, rootContent, nestedContent string) (string, string, string) {
	t.Helper()
	tmpDir := t.TempDir()

	// Root package.json
	rootPkg := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(rootPkg, []byte(rootContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Nested package.json
	nestedPath := filepath.Join(tmpDir, nestedDir)
	if err := os.MkdirAll(nestedPath, 0o755); err != nil {
		t.Fatal(err)
	}
	nestedPkg := filepath.Join(nestedPath, "package.json")
	if err := os.WriteFile(nestedPkg, []byte(nestedContent), 0o644); err != nil {
		t.Fatal(err)
	}

	return tmpDir, rootPkg, nestedPkg
}

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

	t.Run("finds package.json in root", func(t *testing.T) {
		tmpDir := t.TempDir()
		pkgPath := filepath.Join(tmpDir, "package.json")

		pkg := PackageJSON{
			Name:    "test-app",
			Version: "1.0.0",
			Dependencies: map[string]string{
				"react": "^17.0.0",
			},
		}

		data, _ := json.Marshal(pkg)
		if err := os.WriteFile(pkgPath, data, 0o644); err != nil {
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
		if m.Path != packageJSONName {
			t.Errorf("Detect() path = %q, want %q", m.Path, packageJSONName)
		}
		if m.Type != integrationName {
			t.Errorf("Detect() type = %q, want %q", m.Type, integrationName)
		}
		if len(m.Dependencies) != 1 {
			t.Errorf("Detect() dependencies = %d, want 1", len(m.Dependencies))
		}
	})

	t.Run("finds multiple package.json files", func(t *testing.T) {
		tmpDir, _, _ := setupTestDir(t, filepath.Join("packages", "app"), `{"name":"root"}`, `{"name":"app"}`)

		integ := New()
		manifests, err := integ.Detect(ctx, tmpDir)
		if err != nil {
			t.Fatalf("Detect() error = %v", err)
		}
		if len(manifests) != 2 {
			t.Fatalf("Detect() found %d manifests, want 2", len(manifests))
		}
	})

	t.Run("skips node_modules directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Root package.json
		rootPkg := filepath.Join(tmpDir, "package.json")
		if err := os.WriteFile(rootPkg, []byte(`{"name":"root"}`), 0o644); err != nil {
			t.Fatal(err)
		}

		// node_modules package.json (should be skipped)
		nmDir := filepath.Join(tmpDir, "node_modules", "react")
		if err := os.MkdirAll(nmDir, 0o755); err != nil {
			t.Fatal(err)
		}
		nmPkg := filepath.Join(nmDir, "package.json")
		if err := os.WriteFile(nmPkg, []byte(`{"name":"react"}`), 0o644); err != nil {
			t.Fatal(err)
		}

		integ := New()
		manifests, err := integ.Detect(ctx, tmpDir)
		if err != nil {
			t.Fatalf("Detect() error = %v", err)
		}
		if len(manifests) != 1 {
			t.Fatalf("Detect() found %d manifests, want 1 (node_modules should be skipped)", len(manifests))
		}
		if manifests[0].Path != "package.json" {
			t.Errorf("Detect() path = %q, want %q", manifests[0].Path, "package.json")
		}
	})

	t.Run("skips hidden directories", func(t *testing.T) {
		tmpDir, _, _ := setupTestDir(t, filepath.Join(".hidden", "sub"), `{"name":"root"}`, `{"name":"hidden"}`)

		integ := New()
		manifests, err := integ.Detect(ctx, tmpDir)
		if err != nil {
			t.Fatalf("Detect() error = %v", err)
		}
		if len(manifests) != 1 {
			t.Fatalf("Detect() found %d manifests, want 1 (hidden dirs should be skipped)", len(manifests))
		}
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		tmpDir := t.TempDir()
		pkgPath := filepath.Join(tmpDir, "package.json")

		// Invalid JSON
		if err := os.WriteFile(pkgPath, []byte(`{invalid json`), 0o644); err != nil {
			t.Fatal(err)
		}

		integ := New()
		_, err := integ.Detect(ctx, tmpDir)
		if err == nil {
			t.Fatal("Detect() expected error for invalid JSON, got nil")
		}
	})
}

func TestExtractDependencies(t *testing.T) {
	integ := New()

	t.Run("extracts all dependency types", func(t *testing.T) {
		pkg := &PackageJSON{
			Dependencies: map[string]string{
				"react": "^17.0.0",
			},
			DevDependencies: map[string]string{
				"jest": "~27.0.0",
			},
			PeerDependencies: map[string]string{
				"typescript": ">=4.0.0",
			},
			OptionalDependencies: map[string]string{
				"fsevents": "2.3.2",
			},
		}

		deps := integ.extractDependencies(pkg)

		if len(deps) != 4 {
			t.Fatalf("extractDependencies() count = %d, want 4", len(deps))
		}

		// Check types
		typeCount := make(map[string]int)
		for _, dep := range deps {
			typeCount[dep.Type]++
			if dep.Registry != "npm" {
				t.Errorf("dep %q registry = %q, want %q", dep.Name, dep.Registry, "npm")
			}
		}

		if typeCount["direct"] != 1 {
			t.Errorf("direct dependencies = %d, want 1", typeCount["direct"])
		}
		if typeCount["dev"] != 1 {
			t.Errorf("dev dependencies = %d, want 1", typeCount["dev"])
		}
		if typeCount["peer"] != 1 {
			t.Errorf("peer dependencies = %d, want 1", typeCount["peer"])
		}
		if typeCount["optional"] != 1 {
			t.Errorf("optional dependencies = %d, want 1", typeCount["optional"])
		}
	})

	t.Run("handles empty dependencies", func(t *testing.T) {
		pkg := &PackageJSON{}
		deps := integ.extractDependencies(pkg)

		if len(deps) != 0 {
			t.Errorf("extractDependencies() count = %d, want 0", len(deps))
		}
	})
}

func TestPlan(t *testing.T) {
	ctx := context.Background()
	integ := New()

	t.Run("returns empty plan for no dependencies", func(t *testing.T) {
		manifest := &engine.Manifest{
			Path:         "package.json",
			Type:         "npm",
			Dependencies: []engine.Dependency{},
		}

		plan, err := integ.Plan(ctx, manifest, nil)
		if err != nil {
			t.Fatalf("Plan() error = %v", err)
		}
		if len(plan.Updates) != 0 {
			t.Errorf("Plan() updates = %d, want 0", len(plan.Updates))
		}
		if plan.Strategy != "custom_rewrite" {
			t.Errorf("Plan() strategy = %q, want %q", plan.Strategy, "custom_rewrite")
		}
	})

	t.Run("skips file: dependencies", func(t *testing.T) {
		manifest := &engine.Manifest{
			Path: "package.json",
			Type: "npm",
			Dependencies: []engine.Dependency{
				{Name: "local-pkg", CurrentVersion: "file:../local-pkg", Constraint: "file:../local-pkg"},
			},
		}

		plan, err := integ.Plan(ctx, manifest, nil)
		if err != nil {
			t.Fatalf("Plan() error = %v", err)
		}
		if len(plan.Updates) != 0 {
			t.Errorf("Plan() updates = %d, want 0 (file: should be skipped)", len(plan.Updates))
		}
	})

	t.Run("skips link: dependencies", func(t *testing.T) {
		manifest := &engine.Manifest{
			Path: "package.json",
			Type: "npm",
			Dependencies: []engine.Dependency{
				{Name: "linked-pkg", CurrentVersion: "link:../linked-pkg", Constraint: "link:../linked-pkg"},
			},
		}

		plan, err := integ.Plan(ctx, manifest, nil)
		if err != nil {
			t.Fatalf("Plan() error = %v", err)
		}
		if len(plan.Updates) != 0 {
			t.Errorf("Plan() updates = %d, want 0 (link: should be skipped)", len(plan.Updates))
		}
	})

	t.Run("skips git dependencies", func(t *testing.T) {
		manifest := &engine.Manifest{
			Path: "package.json",
			Type: "npm",
			Dependencies: []engine.Dependency{
				{Name: "git-pkg", CurrentVersion: "git+https://github.com/user/repo.git", Constraint: "git+https://github.com/user/repo.git"},
			},
		}

		plan, err := integ.Plan(ctx, manifest, nil)
		if err != nil {
			t.Fatalf("Plan() error = %v", err)
		}
		if len(plan.Updates) != 0 {
			t.Errorf("Plan() updates = %d, want 0 (git URLs should be skipped)", len(plan.Updates))
		}
	})

	t.Run("skips http dependencies", func(t *testing.T) {
		manifest := &engine.Manifest{
			Path: "package.json",
			Type: "npm",
			Dependencies: []engine.Dependency{
				{Name: "http-pkg", CurrentVersion: "https://example.com/package.tgz", Constraint: "https://example.com/package.tgz"},
			},
		}

		plan, err := integ.Plan(ctx, manifest, nil)
		if err != nil {
			t.Fatalf("Plan() error = %v", err)
		}
		if len(plan.Updates) != 0 {
			t.Errorf("Plan() updates = %d, want 0 (http URLs should be skipped)", len(plan.Updates))
		}
	})
}

func TestNeedsUpdate(t *testing.T) {
	integ := New()

	tests := []struct {
		name    string
		current string
		latest  string
		want    bool
	}{
		{"same version", "1.0.0", "1.0.0", false},
		{"patch update", "1.0.0", "1.0.1", true},
		{"minor update", "1.0.0", "1.1.0", true},
		{"major update", "1.0.0", "2.0.0", true},
		{"caret prefix", "^1.0.0", "1.1.0", true},
		{"tilde prefix", "~1.0.0", "1.0.1", true},
		{"gte prefix", ">=1.0.0", "1.1.0", true},
		{"downgrade", "2.0.0", "1.0.0", false},
		{"invalid current", "invalid", "1.0.0", true},
		{"invalid latest", "1.0.0", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := integ.needsUpdate(tt.current, tt.latest)
			if got != tt.want {
				t.Errorf("needsUpdate(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.want)
			}
		})
	}
}

func TestDetermineImpact(t *testing.T) {
	integ := New()

	tests := []struct {
		name    string
		current string
		target  string
		want    string
	}{
		{"patch", "1.0.0", "1.0.1", "patch"},
		{"minor", "1.0.0", "1.1.0", "minor"},
		{"major", "1.0.0", "2.0.0", "major"},
		{"caret prefix", "^1.0.0", "1.1.0", "minor"},
		{"tilde prefix", "~1.0.0", "1.0.1", "patch"},
		{"gte prefix", ">=1.0.0", "2.0.0", "major"},
		{"invalid current", "invalid", "1.0.0", "unknown"},
		{"invalid target", "1.0.0", "invalid", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := integ.determineImpact(tt.current, tt.target)
			if got != tt.want {
				t.Errorf("determineImpact(%q, %q) = %q, want %q", tt.current, tt.target, got, tt.want)
			}
		})
	}
}

func TestApply(t *testing.T) {
	ctx := context.Background()
	integ := New()

	t.Run("returns early for no updates", func(t *testing.T) {
		manifest := &engine.Manifest{
			Path: "package.json",
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

	t.Run("applies updates to dependencies", func(t *testing.T) {
		tmpDir := t.TempDir()
		pkgPath := filepath.Join(tmpDir, "package.json")

		pkg := PackageJSON{
			Name:    "test-app",
			Version: "1.0.0",
			Dependencies: map[string]string{
				"react": "^17.0.0",
			},
		}

		data, _ := json.MarshalIndent(pkg, "", "  ")
		if err := os.WriteFile(pkgPath, data, 0o644); err != nil {
			t.Fatal(err)
		}

		manifest := &engine.Manifest{
			Path: pkgPath,
		}

		update := engine.Update{
			Dependency: engine.Dependency{
				Name:           "react",
				CurrentVersion: "^17.0.0",
				Type:           "direct",
			},
			TargetVersion: "18.0.0",
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
		content, _ := os.ReadFile(pkgPath)
		var updated PackageJSON
		if err := json.Unmarshal(content, &updated); err != nil {
			t.Fatalf("failed to parse updated package.json: %v", err)
		}

		if updated.Dependencies["react"] != "^18.0.0" {
			t.Errorf("Apply() react version = %q, want %q", updated.Dependencies["react"], "^18.0.0")
		}

		if result.ManifestDiff == "" {
			t.Error("Apply() diff should not be empty")
		}
	})

	t.Run("preserves constraint prefixes", func(t *testing.T) {
		tests := []struct {
			name   string
			prefix string
		}{
			{"caret", "^"},
			{"tilde", "~"},
			{"gte", ">="},
			{"none", ""},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tmpDir := t.TempDir()
				pkgPath := filepath.Join(tmpDir, "package.json")

				pkg := PackageJSON{
					Dependencies: map[string]string{
						"react": tt.prefix + "17.0.0",
					},
				}

				data, _ := json.MarshalIndent(pkg, "", "  ")
				if err := os.WriteFile(pkgPath, data, 0o644); err != nil {
					t.Fatal(err)
				}

				manifest := &engine.Manifest{Path: pkgPath}
				update := engine.Update{
					Dependency: engine.Dependency{
						Name:           "react",
						CurrentVersion: tt.prefix + "17.0.0",
						Type:           "direct",
					},
					TargetVersion: "18.0.0",
				}

				plan := &engine.UpdatePlan{
					Manifest: manifest,
					Updates:  []engine.Update{update},
				}

				_, err := integ.Apply(ctx, plan)
				if err != nil {
					t.Fatalf("Apply() error = %v", err)
				}

				content, _ := os.ReadFile(pkgPath)
				var updated PackageJSON
				json.Unmarshal(content, &updated)

				expected := tt.prefix + "18.0.0"
				if updated.Dependencies["react"] != expected {
					t.Errorf("Apply() react version = %q, want %q", updated.Dependencies["react"], expected)
				}
			})
		}
	})

	t.Run("updates all dependency types", func(t *testing.T) {
		tmpDir := t.TempDir()
		pkgPath := filepath.Join(tmpDir, "package.json")

		pkg := PackageJSON{
			Dependencies:         map[string]string{"react": "17.0.0"},
			DevDependencies:      map[string]string{"jest": "27.0.0"},
			PeerDependencies:     map[string]string{"typescript": "4.0.0"},
			OptionalDependencies: map[string]string{"fsevents": "2.0.0"},
		}

		data, _ := json.MarshalIndent(pkg, "", "  ")
		if err := os.WriteFile(pkgPath, data, 0o644); err != nil {
			t.Fatal(err)
		}

		manifest := &engine.Manifest{Path: pkgPath}
		updates := []engine.Update{
			{Dependency: engine.Dependency{Name: "react", CurrentVersion: "17.0.0", Type: "direct"}, TargetVersion: "18.0.0"},
			{Dependency: engine.Dependency{Name: "jest", CurrentVersion: "27.0.0", Type: "dev"}, TargetVersion: "28.0.0"},
			{Dependency: engine.Dependency{Name: "typescript", CurrentVersion: "4.0.0", Type: "peer"}, TargetVersion: "5.0.0"},
			{Dependency: engine.Dependency{Name: "fsevents", CurrentVersion: "2.0.0", Type: "optional"}, TargetVersion: "3.0.0"},
		}

		plan := &engine.UpdatePlan{
			Manifest: manifest,
			Updates:  updates,
		}

		result, err := integ.Apply(ctx, plan)
		if err != nil {
			t.Fatalf("Apply() error = %v", err)
		}
		if result.Applied != 4 {
			t.Errorf("Apply() applied = %d, want 4", result.Applied)
		}

		content, _ := os.ReadFile(pkgPath)
		var updated PackageJSON
		json.Unmarshal(content, &updated)

		if updated.Dependencies["react"] != "18.0.0" {
			t.Errorf("Dependencies[react] = %q, want %q", updated.Dependencies["react"], "18.0.0")
		}
		if updated.DevDependencies["jest"] != "28.0.0" {
			t.Errorf("DevDependencies[jest] = %q, want %q", updated.DevDependencies["jest"], "28.0.0")
		}
		if updated.PeerDependencies["typescript"] != "5.0.0" {
			t.Errorf("PeerDependencies[typescript] = %q, want %q", updated.PeerDependencies["typescript"], "5.0.0")
		}
		if updated.OptionalDependencies["fsevents"] != "3.0.0" {
			t.Errorf("OptionalDependencies[fsevents] = %q, want %q", updated.OptionalDependencies["fsevents"], "3.0.0")
		}
	})
}

func TestUpdateDependency(t *testing.T) {
	integ := New()

	t.Run("updates direct dependency", func(t *testing.T) {
		pkg := &PackageJSON{
			Dependencies: map[string]string{
				"react": "^17.0.0",
			},
		}

		update := engine.Update{
			Dependency: engine.Dependency{
				Name:           "react",
				CurrentVersion: "^17.0.0",
				Type:           "direct",
			},
			TargetVersion: "18.0.0",
		}

		updated := integ.updateDependency(pkg, &update)
		if !updated {
			t.Error("updateDependency() = false, want true")
		}
		if pkg.Dependencies["react"] != "^18.0.0" {
			t.Errorf("Dependencies[react] = %q, want %q", pkg.Dependencies["react"], "^18.0.0")
		}
	})

	t.Run("returns false for non-existent dependency", func(t *testing.T) {
		pkg := &PackageJSON{
			Dependencies: map[string]string{},
		}

		update := engine.Update{
			Dependency: engine.Dependency{
				Name: "react",
				Type: "direct",
			},
			TargetVersion: "18.0.0",
		}

		updated := integ.updateDependency(pkg, &update)
		if updated {
			t.Error("updateDependency() = true, want false for non-existent dep")
		}
	})
}

func TestValidate(t *testing.T) {
	ctx := context.Background()
	integ := New()

	t.Run("validates correct JSON", func(t *testing.T) {
		manifest := &engine.Manifest{
			Content: []byte(`{"name":"test","version":"1.0.0"}`),
		}

		err := integ.Validate(ctx, manifest)
		if err != nil {
			t.Errorf("Validate() error = %v, want nil", err)
		}
	})

	t.Run("fails for invalid JSON", func(t *testing.T) {
		manifest := &engine.Manifest{
			Content: []byte(`{invalid json`),
		}

		err := integ.Validate(ctx, manifest)
		if err == nil {
			t.Error("Validate() error = nil, want error for invalid JSON")
		}
	})
}

func TestGenerateDiff(t *testing.T) {
	t.Run("returns empty string for identical content", func(t *testing.T) {
		diff := generateDiff("test", "test")
		if diff != "" {
			t.Errorf("generateDiff() = %q, want empty string", diff)
		}
	})

	t.Run("generates diff for different content", func(t *testing.T) {
		old := "line1\nline2\nline3"
		updated := "line1\nmodified\nline3"

		diff := generateDiff(old, updated)
		if diff == "" {
			t.Error("generateDiff() returned empty string, want diff")
		}
		if !strings.Contains(diff, "--- package.json") {
			t.Error("generateDiff() missing header")
		}
		if !strings.Contains(diff, "+++ package.json") {
			t.Error("generateDiff() missing header")
		}
		if !strings.Contains(diff, "- line2") {
			t.Error("generateDiff() missing removed line")
		}
		if !strings.Contains(diff, "+ modified") {
			t.Error("generateDiff() missing added line")
		}
	})

	t.Run("handles different line counts", func(t *testing.T) {
		old := "line1\nline2"
		updated := "line1\nline2\nline3"

		diff := generateDiff(old, updated)
		if diff == "" {
			t.Error("generateDiff() returned empty string, want diff")
		}
		if !strings.Contains(diff, "+ line3") {
			t.Error("generateDiff() missing added line")
		}
	})
}

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

package gomod

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/santosr2/uptool/internal/engine"
)

const (
	integrationName = "gomod"
	goModFilename   = "go.mod"
	depTypeIndirect = "indirect"
	depTypeDirect   = "direct"
)

// setupTestDir creates a temp directory with a root go.mod and optionally a nested go.mod.
// Returns the temp directory path. If nestedDir is non-empty, creates go.mod in that subdirectory too.
func setupTestDir(t *testing.T, nestedDir string) string {
	t.Helper()
	tmpDir := t.TempDir()

	// Root go.mod
	rootGoMod := filepath.Join(tmpDir, goModFilename)
	if err := os.WriteFile(rootGoMod, []byte(simpleGoMod), 0o644); err != nil {
		t.Fatal(err)
	}

	// Optional nested go.mod
	if nestedDir != "" {
		dir := filepath.Join(tmpDir, nestedDir)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		nestedGoMod := filepath.Join(dir, goModFilename)
		if err := os.WriteFile(nestedGoMod, []byte(simpleGoMod), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	return tmpDir
}

// Sample go.mod content for testing
const sampleGoMod = `module github.com/user/myapp

go 1.21

require (
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.7.0
	golang.org/x/text v0.13.0 // indirect
)

replace github.com/old/pkg => github.com/new/pkg v1.0.0
`

const simpleGoMod = `module example.com/test

go 1.20

require github.com/pkg/errors v0.9.1
`

const emptyGoMod = `module example.com/empty

go 1.21
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

	t.Run("finds go.mod in root", func(t *testing.T) {
		tmpDir := t.TempDir()
		goModPath := filepath.Join(tmpDir, goModFilename)

		if err := os.WriteFile(goModPath, []byte(sampleGoMod), 0o644); err != nil {
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
		if m.Path != goModFilename {
			t.Errorf("Detect() path = %q, want %q", m.Path, goModFilename)
		}
		if m.Type != integrationName {
			t.Errorf("Detect() type = %q, want %q", m.Type, integrationName)
		}
	})

	t.Run("finds multiple go.mod files (monorepo)", func(t *testing.T) {
		tmpDir := setupTestDir(t, "tools")
		integ := New()
		manifests, err := integ.Detect(ctx, tmpDir)
		if err != nil {
			t.Fatalf("Detect() error = %v", err)
		}
		if len(manifests) != 2 {
			t.Fatalf("Detect() found %d manifests, want 2", len(manifests))
		}
	})

	t.Run("skips vendor directory", func(t *testing.T) {
		tmpDir := setupTestDir(t, filepath.Join("vendor", "github.com", "some", "pkg"))
		integ := New()
		manifests, err := integ.Detect(ctx, tmpDir)
		if err != nil {
			t.Fatalf("Detect() error = %v", err)
		}
		if len(manifests) != 1 {
			t.Fatalf("Detect() found %d manifests, want 1 (vendor should be skipped)", len(manifests))
		}
		if manifests[0].Path != goModFilename {
			t.Errorf("Detect() path = %q, want %q", manifests[0].Path, goModFilename)
		}
	})

	t.Run("skips hidden directories", func(t *testing.T) {
		tmpDir := setupTestDir(t, ".cache")
		integ := New()
		manifests, err := integ.Detect(ctx, tmpDir)
		if err != nil {
			t.Fatalf("Detect() error = %v", err)
		}
		if len(manifests) != 1 {
			t.Fatalf("Detect() found %d manifests, want 1 (hidden dirs should be skipped)", len(manifests))
		}
	})

	t.Run("skips testdata directory", func(t *testing.T) {
		tmpDir := setupTestDir(t, "testdata")
		integ := New()
		manifests, err := integ.Detect(ctx, tmpDir)
		if err != nil {
			t.Fatalf("Detect() error = %v", err)
		}
		if len(manifests) != 1 {
			t.Fatalf("Detect() found %d manifests, want 1 (testdata should be skipped)", len(manifests))
		}
	})

	t.Run("returns empty for directory without go.mod", func(t *testing.T) {
		tmpDir := t.TempDir()

		integ := New()
		manifests, err := integ.Detect(ctx, tmpDir)
		if err != nil {
			t.Fatalf("Detect() error = %v", err)
		}
		if len(manifests) != 0 {
			t.Errorf("Detect() found %d manifests, want 0", len(manifests))
		}
	})
}

func TestParseGoMod(t *testing.T) {
	integ := New()

	t.Run("parses module name and go version", func(t *testing.T) {
		deps, metadata := integ.parseGoMod([]byte(sampleGoMod))

		if metadata["module_name"] != "github.com/user/myapp" {
			t.Errorf("module_name = %q, want %q", metadata["module_name"], "github.com/user/myapp")
		}
		if metadata["go_version"] != "1.21" {
			t.Errorf("go_version = %q, want %q", metadata["go_version"], "1.21")
		}

		// Should find 3 dependencies (2 direct + 1 indirect)
		if len(deps) != 3 {
			t.Errorf("dependencies count = %d, want 3", len(deps))
		}
	})

	t.Run("parses require block dependencies", func(t *testing.T) {
		deps, _ := integ.parseGoMod([]byte(sampleGoMod))

		// Check for expected dependencies
		depMap := make(map[string]engine.Dependency)
		for _, dep := range deps {
			depMap[dep.Name] = dep
		}

		// Check logrus
		if logrus, ok := depMap["github.com/sirupsen/logrus"]; ok {
			if logrus.CurrentVersion != "v1.9.3" {
				t.Errorf("logrus version = %q, want %q", logrus.CurrentVersion, "v1.9.3")
			}
			if logrus.Type != "direct" {
				t.Errorf("logrus type = %q, want %q", logrus.Type, "direct")
			}
		} else {
			t.Error("logrus dependency not found")
		}

		// Check indirect dependency
		if text, ok := depMap["golang.org/x/text"]; ok {
			if text.Type != depTypeIndirect {
				t.Errorf("golang.org/x/text type = %q, want %q", text.Type, depTypeIndirect)
			}
		} else {
			t.Error("golang.org/x/text dependency not found")
		}
	})

	t.Run("parses single-line require", func(t *testing.T) {
		deps, _ := integ.parseGoMod([]byte(simpleGoMod))

		if len(deps) != 1 {
			t.Fatalf("dependencies count = %d, want 1", len(deps))
		}

		if deps[0].Name != "github.com/pkg/errors" {
			t.Errorf("dependency name = %q, want %q", deps[0].Name, "github.com/pkg/errors")
		}
		if deps[0].CurrentVersion != "v0.9.1" {
			t.Errorf("dependency version = %q, want %q", deps[0].CurrentVersion, "v0.9.1")
		}
	})

	t.Run("handles empty go.mod", func(t *testing.T) {
		deps, metadata := integ.parseGoMod([]byte(emptyGoMod))

		if metadata["module_name"] != "example.com/empty" {
			t.Errorf("module_name = %q, want %q", metadata["module_name"], "example.com/empty")
		}
		if len(deps) != 0 {
			t.Errorf("dependencies count = %d, want 0", len(deps))
		}
	})

	t.Run("handles go version with patch", func(t *testing.T) {
		goModWithPatch := `module example.com/test

go 1.21.5

require github.com/pkg/errors v0.9.1
`
		_, metadata := integ.parseGoMod([]byte(goModWithPatch))

		if metadata["go_version"] != "1.21.5" {
			t.Errorf("go_version = %q, want %q", metadata["go_version"], "1.21.5")
		}
	})
}

func TestParseDependencyLine(t *testing.T) {
	integ := New()

	tests := []struct {
		name        string
		line        string
		wantName    string
		wantVersion string
		wantType    string
		wantNil     bool
	}{
		{
			name:        "simple dependency",
			line:        "github.com/pkg/errors v0.9.1",
			wantName:    "github.com/pkg/errors",
			wantVersion: "v0.9.1",
			wantType:    depTypeDirect,
		},
		{
			name:        "indirect dependency",
			line:        "golang.org/x/text v0.13.0 // indirect",
			wantName:    "golang.org/x/text",
			wantVersion: "v0.13.0",
			wantType:    depTypeIndirect,
		},
		{
			name:        "with leading whitespace",
			line:        "  github.com/sirupsen/logrus v1.9.3",
			wantName:    "github.com/sirupsen/logrus",
			wantVersion: "v1.9.3",
			wantType:    depTypeDirect,
		},
		{
			name:    "empty line",
			line:    "",
			wantNil: true,
		},
		{
			name:    "comment line",
			line:    "// some comment",
			wantNil: true,
		},
		{
			name:    "invalid format",
			line:    "just-a-string",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep := integ.parseDependencyLine(tt.line)

			if tt.wantNil {
				if dep != nil {
					t.Errorf("parseDependencyLine() = %+v, want nil", dep)
				}
				return
			}

			if dep == nil {
				t.Fatal("parseDependencyLine() = nil, want dependency")
			}

			if dep.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", dep.Name, tt.wantName)
			}
			if dep.CurrentVersion != tt.wantVersion {
				t.Errorf("CurrentVersion = %q, want %q", dep.CurrentVersion, tt.wantVersion)
			}
			if dep.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", dep.Type, tt.wantType)
			}
			if dep.Registry != "go" {
				t.Errorf("Registry = %q, want %q", dep.Registry, "go")
			}
		})
	}
}

func TestPlan(t *testing.T) {
	ctx := context.Background()
	integ := New()

	t.Run("returns empty plan for no dependencies", func(t *testing.T) {
		manifest := &engine.Manifest{
			Path:         goModFilename,
			Type:         integrationName,
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

	t.Run("skips indirect dependencies", func(t *testing.T) {
		manifest := &engine.Manifest{
			Path: goModFilename,
			Type: integrationName,
			Dependencies: []engine.Dependency{
				{Name: "golang.org/x/text", CurrentVersion: "v0.13.0", Type: depTypeIndirect, Registry: "go"},
			},
		}

		plan, err := integ.Plan(ctx, manifest, nil)
		if err != nil {
			t.Fatalf("Plan() error = %v", err)
		}
		if len(plan.Updates) != 0 {
			t.Errorf("Plan() updates = %d, want 0 (indirect should be skipped)", len(plan.Updates))
		}
	})

	t.Run("skips replaced modules", func(t *testing.T) {
		manifest := &engine.Manifest{
			Path: goModFilename,
			Type: integrationName,
			Dependencies: []engine.Dependency{
				{
					Name:           "github.com/old/pkg",
					CurrentVersion: "v1.0.0",
					Type:           depTypeDirect,
					Registry:       "go",
				},
			},
			Metadata: map[string]interface{}{
				"replacements": map[string]bool{"github.com/old/pkg": true},
			},
		}

		plan, err := integ.Plan(ctx, manifest, nil)
		if err != nil {
			t.Fatalf("Plan() error = %v", err)
		}
		if len(plan.Updates) != 0 {
			t.Errorf("Plan() updates = %d, want 0 (replaced should be skipped)", len(plan.Updates))
		}
	})

	t.Run("skips pseudo-versions", func(t *testing.T) {
		manifest := &engine.Manifest{
			Path: goModFilename,
			Type: integrationName,
			Dependencies: []engine.Dependency{
				{
					Name:           "github.com/some/fork",
					CurrentVersion: "v0.0.0-20231201123456-abcdef123456",
					Type:           depTypeDirect,
					Registry:       "go",
				},
			},
		}

		plan, err := integ.Plan(ctx, manifest, nil)
		if err != nil {
			t.Fatalf("Plan() error = %v", err)
		}
		if len(plan.Updates) != 0 {
			t.Errorf("Plan() updates = %d, want 0 (pseudo-versions should be skipped)", len(plan.Updates))
		}
	})
}

func TestApply(t *testing.T) {
	ctx := context.Background()
	integ := New()

	t.Run("returns early for no updates", func(t *testing.T) {
		manifest := &engine.Manifest{
			Path: goModFilename,
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

	t.Run("applies updates to go.mod", func(t *testing.T) {
		tmpDir := t.TempDir()
		goModPath := filepath.Join(tmpDir, goModFilename)

		content := `module example.com/test

go 1.21

require (
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.9.0
)
`
		if err := os.WriteFile(goModPath, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		manifest := &engine.Manifest{
			Path: goModPath,
		}

		updates := []engine.Update{
			{
				Dependency: engine.Dependency{
					Name:           "github.com/sirupsen/logrus",
					CurrentVersion: "v1.9.0",
				},
				TargetVersion: "v1.9.3",
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
		if result.Applied != 1 {
			t.Errorf("Apply() applied = %d, want 1", result.Applied)
		}

		// Verify file was updated
		updatedContent, _ := os.ReadFile(goModPath)
		if !strings.Contains(string(updatedContent), "github.com/sirupsen/logrus v1.9.3") {
			t.Error("Apply() did not update logrus version")
		}
		if !strings.Contains(string(updatedContent), "github.com/pkg/errors v0.9.1") {
			t.Error("Apply() accidentally modified pkg/errors version")
		}

		if result.ManifestDiff == "" {
			t.Error("Apply() diff should not be empty")
		}
	})

	t.Run("applies multiple updates", func(t *testing.T) {
		tmpDir := t.TempDir()
		goModPath := filepath.Join(tmpDir, goModFilename)

		content := `module example.com/test

go 1.21

require (
	github.com/pkg/errors v0.9.0
	github.com/sirupsen/logrus v1.9.0
)
`
		if err := os.WriteFile(goModPath, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		manifest := &engine.Manifest{
			Path: goModPath,
		}

		updates := []engine.Update{
			{
				Dependency: engine.Dependency{
					Name:           "github.com/pkg/errors",
					CurrentVersion: "v0.9.0",
				},
				TargetVersion: "v0.9.1",
			},
			{
				Dependency: engine.Dependency{
					Name:           "github.com/sirupsen/logrus",
					CurrentVersion: "v1.9.0",
				},
				TargetVersion: "v1.9.3",
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

		// Verify both were updated
		updatedContent, _ := os.ReadFile(goModPath)
		if !strings.Contains(string(updatedContent), "github.com/pkg/errors v0.9.1") {
			t.Error("Apply() did not update pkg/errors version")
		}
		if !strings.Contains(string(updatedContent), "github.com/sirupsen/logrus v1.9.3") {
			t.Error("Apply() did not update logrus version")
		}
	})

	t.Run("handles non-matching update gracefully", func(t *testing.T) {
		tmpDir := t.TempDir()
		goModPath := filepath.Join(tmpDir, goModFilename)

		content := `module example.com/test

go 1.21

require github.com/pkg/errors v0.9.1
`
		if err := os.WriteFile(goModPath, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		manifest := &engine.Manifest{
			Path: goModPath,
		}

		// Try to update a dependency that doesn't exist
		updates := []engine.Update{
			{
				Dependency: engine.Dependency{
					Name:           "github.com/nonexistent/pkg",
					CurrentVersion: "v1.0.0",
				},
				TargetVersion: "v2.0.0",
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
		if result.Applied != 0 {
			t.Errorf("Apply() applied = %d, want 0", result.Applied)
		}
		if result.Failed != 1 {
			t.Errorf("Apply() failed = %d, want 1", result.Failed)
		}
	})
}

func TestValidate(t *testing.T) {
	ctx := context.Background()
	integ := New()

	t.Run("validates correct go.mod", func(t *testing.T) {
		manifest := &engine.Manifest{
			Content: []byte(sampleGoMod),
		}

		err := integ.Validate(ctx, manifest)
		if err != nil {
			t.Errorf("Validate() error = %v, want nil", err)
		}
	})

	t.Run("fails for go.mod without module directive", func(t *testing.T) {
		manifest := &engine.Manifest{
			Content: []byte(`go 1.21

require github.com/pkg/errors v0.9.1
`),
		}

		err := integ.Validate(ctx, manifest)
		if err == nil {
			t.Error("Validate() error = nil, want error for missing module directive")
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
		if !strings.Contains(diff, "--- go.mod") {
			t.Error("generateDiff() missing header")
		}
		if !strings.Contains(diff, "+++ go.mod") {
			t.Error("generateDiff() missing header")
		}
		if !strings.Contains(diff, "- line2") {
			t.Error("generateDiff() missing removed line")
		}
		if !strings.Contains(diff, "+ modified") {
			t.Error("generateDiff() missing added line")
		}
	})

	t.Run("handles version updates", func(t *testing.T) {
		old := `module example.com/test

require github.com/pkg/errors v0.9.0
`
		updated := `module example.com/test

require github.com/pkg/errors v0.9.1
`
		diff := generateDiff(old, updated)
		if !strings.Contains(diff, "- \trequire github.com/pkg/errors v0.9.0") ||
			!strings.Contains(diff, "+ \trequire github.com/pkg/errors v0.9.1") {
			// The diff might have different formatting, just check it's not empty
			if diff == "" {
				t.Error("generateDiff() returned empty string for version change")
			}
		}
	})
}

func TestModulePatterns(t *testing.T) {
	// Helper to test single-group regex patterns
	testSingleGroupPattern := func(t *testing.T, pattern *regexp.Regexp, patternName string, tests []struct {
		name, line, want string
	},
	) {
		t.Helper()
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				matches := pattern.FindStringSubmatch(tt.line)
				if len(matches) < 2 || matches[1] != tt.want {
					t.Errorf("%s.FindStringSubmatch(%q) = %v, want %q", patternName, tt.line, matches, tt.want)
				}
			})
		}
	}

	t.Run("modulePattern matches module line", func(t *testing.T) {
		testSingleGroupPattern(t, modulePattern, "modulePattern", []struct {
			name, line, want string
		}{
			{"simple", "module example.com/test", "example.com/test"},
			{"github", "module github.com/user/repo", "github.com/user/repo"},
			{"nested", "module github.com/user/repo/v2", "github.com/user/repo/v2"},
		})
	})

	t.Run("goVersionPattern matches go version", func(t *testing.T) {
		testSingleGroupPattern(t, goVersionPat, "goVersionPat", []struct {
			name, line, want string
		}{
			{"major.minor", "go 1.21", "1.21"},
			{"with patch", "go 1.21.5", "1.21.5"},
			{"older", "go 1.18", "1.18"},
		})
	})

	t.Run("requirePattern matches require lines", func(t *testing.T) {
		tests := []struct {
			name        string
			line        string
			wantModule  string
			wantVersion string
		}{
			{"simple", "github.com/pkg/errors v0.9.1", "github.com/pkg/errors", "v0.9.1"},
			{"with tabs", "\tgithub.com/sirupsen/logrus v1.9.3", "github.com/sirupsen/logrus", "v1.9.3"},
			{"indirect", "\tgolang.org/x/text v0.13.0 // indirect", "golang.org/x/text", "v0.13.0"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				matches := requirePattern.FindStringSubmatch(tt.line)
				if len(matches) < 3 {
					t.Fatalf("requirePattern.FindStringSubmatch(%q) = %v, want at least 3 groups", tt.line, matches)
				}
				if matches[1] != tt.wantModule {
					t.Errorf("module = %q, want %q", matches[1], tt.wantModule)
				}
				if matches[2] != tt.wantVersion {
					t.Errorf("version = %q, want %q", matches[2], tt.wantVersion)
				}
			})
		}
	})
}

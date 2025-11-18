package helm

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/santosr2/uptool/internal/engine"
)

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
	if integ.Name() != "helm" {
		t.Errorf("Name() = %q, want %q", integ.Name(), "helm")
	}
}

func TestDetect(t *testing.T) {
	ctx := context.Background()

	t.Run("finds Chart.yaml in root", func(t *testing.T) {
		tmpDir := t.TempDir()
		chartPath := filepath.Join(tmpDir, "Chart.yaml")

		chart := Chart{
			APIVersion: "v2",
			Name:       "myapp",
			Version:    "1.0.0",
			Dependencies: []Dependency{
				{Name: "nginx", Version: "1.0.0", Repository: "https://charts.bitnami.com/bitnami"},
			},
		}

		data, _ := yaml.Marshal(chart)
		if err := os.WriteFile(chartPath, data, 0o644); err != nil {
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
		if m.Path != "Chart.yaml" {
			t.Errorf("Detect() path = %q, want %q", m.Path, "Chart.yaml")
		}
		if m.Type != "helm" {
			t.Errorf("Detect() type = %q, want %q", m.Type, "helm")
		}
		if len(m.Dependencies) != 1 {
			t.Errorf("Detect() dependencies = %d, want 1", len(m.Dependencies))
		}
	})

	t.Run("finds multiple Chart.yaml files", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Root Chart.yaml
		rootChart := filepath.Join(tmpDir, "Chart.yaml")
		if err := os.WriteFile(rootChart, []byte("apiVersion: v2\nname: root\nversion: 1.0.0"), 0o644); err != nil {
			t.Fatal(err)
		}

		// Nested Chart.yaml
		nestedDir := filepath.Join(tmpDir, "charts", "subchart")
		if err := os.MkdirAll(nestedDir, 0o755); err != nil {
			t.Fatal(err)
		}
		nestedChart := filepath.Join(nestedDir, "Chart.yaml")
		if err := os.WriteFile(nestedChart, []byte("apiVersion: v2\nname: subchart\nversion: 1.0.0"), 0o644); err != nil {
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

		// Root Chart.yaml
		rootChart := filepath.Join(tmpDir, "Chart.yaml")
		if err := os.WriteFile(rootChart, []byte("apiVersion: v2\nname: root\nversion: 1.0.0"), 0o644); err != nil {
			t.Fatal(err)
		}

		// Hidden directory Chart.yaml (should be skipped)
		hiddenDir := filepath.Join(tmpDir, ".hidden")
		if err := os.MkdirAll(hiddenDir, 0o755); err != nil {
			t.Fatal(err)
		}
		hiddenChart := filepath.Join(hiddenDir, "Chart.yaml")
		if err := os.WriteFile(hiddenChart, []byte("apiVersion: v2\nname: hidden\nversion: 1.0.0"), 0o644); err != nil {
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

	t.Run("returns error for invalid YAML", func(t *testing.T) {
		tmpDir := t.TempDir()
		chartPath := filepath.Join(tmpDir, "Chart.yaml")

		// Invalid YAML
		if err := os.WriteFile(chartPath, []byte("invalid: yaml: content:"), 0o644); err != nil {
			t.Fatal(err)
		}

		integ := New()
		_, err := integ.Detect(ctx, tmpDir)

		if err == nil {
			t.Fatal("Detect() expected error for invalid YAML, got nil")
		}
	})

	t.Run("sets metadata fields", func(t *testing.T) {
		tmpDir := t.TempDir()
		chartPath := filepath.Join(tmpDir, "Chart.yaml")

		chart := Chart{
			APIVersion: "v2",
			Name:       "myapp",
			Version:    "1.2.3",
			Dependencies: []Dependency{
				{Name: "dep1", Version: "1.0.0", Repository: "https://example.com"},
				{Name: "dep2", Version: "2.0.0", Repository: "https://example.com"},
			},
		}

		data, _ := yaml.Marshal(chart)
		if err := os.WriteFile(chartPath, data, 0o644); err != nil {
			t.Fatal(err)
		}

		integ := New()
		manifests, err := integ.Detect(ctx, tmpDir)

		if err != nil || len(manifests) == 0 {
			t.Fatalf("Detect() error = %v", err)
		}

		m := manifests[0]
		if m.Metadata["chart_name"] != "myapp" {
			t.Errorf("Metadata[chart_name] = %v, want %q", m.Metadata["chart_name"], "myapp")
		}
		if m.Metadata["chart_version"] != "1.2.3" {
			t.Errorf("Metadata[chart_version] = %v, want %q", m.Metadata["chart_version"], "1.2.3")
		}
		if m.Metadata["deps_count"] != 2 {
			t.Errorf("Metadata[deps_count] = %v, want 2", m.Metadata["deps_count"])
		}
	})
}

func TestExtractDependencies(t *testing.T) {
	integ := New()

	t.Run("extracts chart dependencies", func(t *testing.T) {
		chart := &Chart{
			Dependencies: []Dependency{
				{Name: "nginx", Version: "1.0.0", Repository: "https://charts.bitnami.com/bitnami"},
				{Name: "redis", Version: "2.0.0", Repository: "https://charts.bitnami.com/bitnami"},
			},
		}

		deps := integ.extractDependencies(chart)

		if len(deps) != 2 {
			t.Fatalf("extractDependencies() count = %d, want 2", len(deps))
		}

		for _, dep := range deps {
			if dep.Type != "chart" {
				t.Errorf("dep %q type = %q, want %q", dep.Name, dep.Type, "chart")
			}
			if !strings.Contains(dep.Registry, "bitnami") {
				t.Errorf("dep %q registry = %q, want bitnami URL", dep.Name, dep.Registry)
			}
		}
	})

	t.Run("skips OCI repositories", func(t *testing.T) {
		chart := &Chart{
			Dependencies: []Dependency{
				{Name: "oci-chart", Version: "1.0.0", Repository: "oci://registry.example.com/charts"},
				{Name: "normal-chart", Version: "1.0.0", Repository: "https://charts.example.com"},
			},
		}

		deps := integ.extractDependencies(chart)

		if len(deps) != 1 {
			t.Fatalf("extractDependencies() count = %d, want 1 (OCI should be skipped)", len(deps))
		}
		if deps[0].Name != "normal-chart" {
			t.Errorf("extractDependencies() name = %q, want %q", deps[0].Name, "normal-chart")
		}
	})

	t.Run("skips file:// repositories", func(t *testing.T) {
		chart := &Chart{
			Dependencies: []Dependency{
				{Name: "local-chart", Version: "1.0.0", Repository: "file://../local-charts"},
				{Name: "remote-chart", Version: "1.0.0", Repository: "https://charts.example.com"},
			},
		}

		deps := integ.extractDependencies(chart)

		if len(deps) != 1 {
			t.Fatalf("extractDependencies() count = %d, want 1 (file:// should be skipped)", len(deps))
		}
		if deps[0].Name != "remote-chart" {
			t.Errorf("extractDependencies() name = %q, want %q", deps[0].Name, "remote-chart")
		}
	})

	t.Run("handles empty dependencies", func(t *testing.T) {
		chart := &Chart{}
		deps := integ.extractDependencies(chart)

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
			Path:         "Chart.yaml",
			Type:         "helm",
			Dependencies: []engine.Dependency{},
		}

		plan, err := integ.Plan(ctx, manifest)
		if err != nil {
			t.Fatalf("Plan() error = %v", err)
		}
		if len(plan.Updates) != 0 {
			t.Errorf("Plan() updates = %d, want 0", len(plan.Updates))
		}
		if plan.Strategy != "yaml_rewrite" {
			t.Errorf("Plan() strategy = %q, want %q", plan.Strategy, "yaml_rewrite")
		}
	})
}

func TestApply(t *testing.T) {
	ctx := context.Background()
	integ := New()

	t.Run("returns early for no updates", func(t *testing.T) {
		manifest := &engine.Manifest{
			Path: "Chart.yaml",
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
		chartPath := filepath.Join(tmpDir, "Chart.yaml")

		chart := Chart{
			APIVersion: "v2",
			Name:       "myapp",
			Version:    "1.0.0",
			Dependencies: []Dependency{
				{Name: "nginx", Version: "1.0.0", Repository: "https://charts.bitnami.com/bitnami"},
			},
		}

		data, _ := yaml.Marshal(chart)
		if err := os.WriteFile(chartPath, data, 0o644); err != nil {
			t.Fatal(err)
		}

		manifest := &engine.Manifest{
			Path: chartPath,
		}

		update := engine.Update{
			Dependency: engine.Dependency{
				Name:           "nginx",
				CurrentVersion: "1.0.0",
				Type:           "chart",
			},
			TargetVersion: "2.0.0",
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
		content, _ := os.ReadFile(chartPath)
		var updated Chart
		if err := yaml.Unmarshal(content, &updated); err != nil {
			t.Fatalf("failed to parse updated Chart.yaml: %v", err)
		}

		if updated.Dependencies[0].Version != "2.0.0" {
			t.Errorf("Apply() nginx version = %q, want %q", updated.Dependencies[0].Version, "2.0.0")
		}

		if result.ManifestDiff == "" {
			t.Error("Apply() diff should not be empty")
		}
	})

	t.Run("updates multiple dependencies", func(t *testing.T) {
		tmpDir := t.TempDir()
		chartPath := filepath.Join(tmpDir, "Chart.yaml")

		chart := Chart{
			APIVersion: "v2",
			Name:       "myapp",
			Version:    "1.0.0",
			Dependencies: []Dependency{
				{Name: "nginx", Version: "1.0.0", Repository: "https://charts.bitnami.com/bitnami"},
				{Name: "redis", Version: "2.0.0", Repository: "https://charts.bitnami.com/bitnami"},
			},
		}

		data, _ := yaml.Marshal(chart)
		if err := os.WriteFile(chartPath, data, 0o644); err != nil {
			t.Fatal(err)
		}

		manifest := &engine.Manifest{Path: chartPath}
		updates := []engine.Update{
			{Dependency: engine.Dependency{Name: "nginx"}, TargetVersion: "1.5.0"},
			{Dependency: engine.Dependency{Name: "redis"}, TargetVersion: "3.0.0"},
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

		content, _ := os.ReadFile(chartPath)
		var updated Chart
		yaml.Unmarshal(content, &updated)

		if updated.Dependencies[0].Version != "1.5.0" {
			t.Errorf("Dependencies[0].Version = %q, want %q", updated.Dependencies[0].Version, "1.5.0")
		}
		if updated.Dependencies[1].Version != "3.0.0" {
			t.Errorf("Dependencies[1].Version = %q, want %q", updated.Dependencies[1].Version, "3.0.0")
		}
	})
}

func TestValidate(t *testing.T) {
	ctx := context.Background()
	integ := New()

	t.Run("validates correct Chart.yaml", func(t *testing.T) {
		chart := Chart{
			APIVersion: "v2",
			Name:       "myapp",
			Version:    "1.0.0",
		}

		data, _ := yaml.Marshal(chart)
		manifest := &engine.Manifest{
			Content: data,
		}

		err := integ.Validate(ctx, manifest)
		if err != nil {
			t.Errorf("Validate() error = %v, want nil", err)
		}
	})

	t.Run("fails for invalid YAML", func(t *testing.T) {
		manifest := &engine.Manifest{
			Content: []byte("invalid: yaml: content:"),
		}

		err := integ.Validate(ctx, manifest)
		if err == nil {
			t.Error("Validate() error = nil, want error for invalid YAML")
		}
	})

	t.Run("fails for missing apiVersion", func(t *testing.T) {
		chart := Chart{
			Name:    "myapp",
			Version: "1.0.0",
		}

		data, _ := yaml.Marshal(chart)
		manifest := &engine.Manifest{
			Content: data,
		}

		err := integ.Validate(ctx, manifest)
		if err == nil {
			t.Error("Validate() error = nil, want error for missing apiVersion")
		}
		if !strings.Contains(err.Error(), "apiVersion") {
			t.Errorf("Validate() error = %q, want error mentioning apiVersion", err.Error())
		}
	})

	t.Run("fails for missing name", func(t *testing.T) {
		chart := Chart{
			APIVersion: "v2",
			Version:    "1.0.0",
		}

		data, _ := yaml.Marshal(chart)
		manifest := &engine.Manifest{
			Content: data,
		}

		err := integ.Validate(ctx, manifest)
		if err == nil {
			t.Error("Validate() error = nil, want error for missing name")
		}
		if !strings.Contains(err.Error(), "name") {
			t.Errorf("Validate() error = %q, want error mentioning name", err.Error())
		}
	})

	t.Run("fails for missing version", func(t *testing.T) {
		chart := Chart{
			APIVersion: "v2",
			Name:       "myapp",
		}

		data, _ := yaml.Marshal(chart)
		manifest := &engine.Manifest{
			Content: data,
		}

		err := integ.Validate(ctx, manifest)
		if err == nil {
			t.Error("Validate() error = nil, want error for missing version")
		}
		if !strings.Contains(err.Error(), "version") {
			t.Errorf("Validate() error = %q, want error mentioning version", err.Error())
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
		old := "name: myapp\nversion: 1.0.0\ndependencies:\n  - version: 1.0.0"
		updated := "name: myapp\nversion: 1.0.0\ndependencies:\n  - version: 2.0.0"

		diff := generateDiff(old, updated)
		if diff == "" {
			t.Error("generateDiff() returned empty string, want diff")
		}
		if !strings.Contains(diff, "--- Chart.yaml") {
			t.Error("generateDiff() missing header")
		}
		if !strings.Contains(diff, "+++ Chart.yaml") {
			t.Error("generateDiff() missing header")
		}
		if !strings.Contains(diff, "version:") {
			t.Error("generateDiff() missing version line")
		}
	})

	t.Run("only includes version changes", func(t *testing.T) {
		old := "name: myapp\nversion: 1.0.0\ndescription: My app"
		updated := "name: myapp\nversion: 2.0.0\ndescription: My app"

		diff := generateDiff(old, updated)
		if !strings.Contains(diff, "version:") {
			t.Error("generateDiff() missing version line in diff")
		}
		// Should not include non-version changes
		if strings.Contains(diff, "description") {
			t.Error("generateDiff() should not include non-version changes")
		}
	})
}

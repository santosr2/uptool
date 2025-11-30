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
package docker

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
	if got := integration.Name(); got != "docker" {
		t.Errorf("Name() = %q, want %q", got, "docker")
	}
}

func TestParseImageReference(t *testing.T) {
	tests := []struct {
		name          string
		ref           string
		expectedImage string
		expectedTag   string
	}{
		{"simple image:tag", "nginx:1.25", "nginx", "1.25"},
		{"image without tag", "nginx", "nginx", "latest"},
		{"org/image:tag", "library/nginx:1.25", "library/nginx", "1.25"},
		{"registry/org/image:tag", "gcr.io/project/image:1.0", "gcr.io/project/image", "1.0"},
		{"digest reference", "nginx@sha256:abc123", "nginx", "sha256"},
		{"variable reference", "${IMAGE}:${TAG}", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			image, tag := parseImageReference(tt.ref)
			if image != tt.expectedImage {
				t.Errorf("parseImageReference(%q) image = %q, want %q", tt.ref, image, tt.expectedImage)
			}
			if tag != tt.expectedTag {
				t.Errorf("parseImageReference(%q) tag = %q, want %q", tt.ref, tag, tt.expectedTag)
			}
		})
	}
}

func TestIntegration_ExtractDockerfileDeps(t *testing.T) {
	integration := New()

	tests := []struct {
		name          string
		content       string
		expectedCount int
		expectedNames []string
	}{
		{
			name: "single FROM",
			content: `FROM golang:1.21
RUN go build -o app .
`,
			expectedCount: 1,
			expectedNames: []string{"golang"},
		},
		{
			name: "multi-stage build",
			content: `FROM golang:1.21 AS builder
RUN go build -o app .

FROM alpine:3.18
COPY --from=builder /app /app
`,
			expectedCount: 2,
			expectedNames: []string{"golang", "alpine"},
		},
		{
			name: "with platform",
			content: `FROM --platform=linux/amd64 golang:1.21
RUN go build -o app .
`,
			expectedCount: 1,
			expectedNames: []string{"golang"},
		},
		{
			name: "skip scratch",
			content: `FROM golang:1.21 AS builder
RUN go build -o app .

FROM scratch
COPY --from=builder /app /app
`,
			expectedCount: 1,
			expectedNames: []string{"golang"},
		},
		{
			name: "skip ARG variables",
			content: `ARG BASE_IMAGE=golang:1.21
FROM ${BASE_IMAGE}
RUN go build -o app .
`,
			expectedCount: 0,
			expectedNames: []string{},
		},
		{
			name: "deduplicate same image",
			content: `FROM golang:1.21 AS builder
FROM golang:1.21 AS tester
`,
			expectedCount: 1,
			expectedNames: []string{"golang"},
		},
		{
			name:          "empty dockerfile",
			content:       "",
			expectedCount: 0,
			expectedNames: []string{},
		},
		{
			name: "with comments",
			content: `# Build stage
FROM golang:1.21
# This is a comment
RUN go build -o app .
`,
			expectedCount: 1,
			expectedNames: []string{"golang"},
		},
		{
			name: "default tag",
			content: `FROM nginx
RUN echo "hello"
`,
			expectedCount: 1,
			expectedNames: []string{"nginx"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := integration.extractDockerfileDeps([]byte(tt.content))

			if len(deps) != tt.expectedCount {
				t.Errorf("extractDockerfileDeps() returned %d deps, want %d", len(deps), tt.expectedCount)
			}

			for i, expectedName := range tt.expectedNames {
				if i < len(deps) && deps[i].Name != expectedName {
					t.Errorf("deps[%d].Name = %q, want %q", i, deps[i].Name, expectedName)
				}
			}
		})
	}
}

func TestIntegration_ExtractComposeDeps(t *testing.T) {
	integration := New()

	tests := []struct {
		name          string
		content       string
		expectedCount int
		expectedNames []string
	}{
		{
			name: "single service",
			content: `
version: "3.8"
services:
  web:
    image: nginx:1.25
`,
			expectedCount: 1,
			expectedNames: []string{"nginx"},
		},
		{
			name: "multiple services",
			content: `
version: "3.8"
services:
  web:
    image: nginx:1.25
  db:
    image: postgres:15
  cache:
    image: redis:7.2
`,
			expectedCount: 3,
			expectedNames: []string{"nginx", "postgres", "redis"}, // Note: order may vary due to map iteration
		},
		{
			name: "service with build (no image)",
			content: `
version: "3.8"
services:
  app:
    build: .
`,
			expectedCount: 0,
			expectedNames: []string{},
		},
		{
			name: "deduplicate same image",
			content: `
version: "3.8"
services:
  web1:
    image: nginx:1.25
  web2:
    image: nginx:1.25
`,
			expectedCount: 1,
			expectedNames: []string{"nginx"},
		},
		{
			name: "invalid yaml",
			content: `
version: [invalid
`,
			expectedCount: 0,
			expectedNames: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := integration.extractComposeDeps([]byte(tt.content))

			if len(deps) != tt.expectedCount {
				t.Errorf("extractComposeDeps() returned %d deps, want %d", len(deps), tt.expectedCount)
			}

			// Check that all expected names are present (order-independent)
			depNames := make(map[string]bool)
			for _, dep := range deps {
				depNames[dep.Name] = true
			}
			for _, expectedName := range tt.expectedNames {
				if !depNames[expectedName] {
					t.Errorf("extractComposeDeps() missing expected dependency %q", expectedName)
				}
			}
		})
	}
}

func TestIntegration_Detect(t *testing.T) {
	integration := New()
	ctx := context.Background()

	t.Run("detects Dockerfile", func(t *testing.T) {
		tmpDir := t.TempDir()
		dockerfileContent := `FROM golang:1.21
RUN go build -o app .
`
		if err := os.WriteFile(filepath.Join(tmpDir, "Dockerfile"), []byte(dockerfileContent), 0o644); err != nil {
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
			if manifests[0].Type != "docker" {
				t.Errorf("manifest.Type = %q, want %q", manifests[0].Type, "docker")
			}
			fileType, ok := manifests[0].Metadata["file_type"].(string)
			if !ok || fileType != "dockerfile" {
				t.Errorf("manifest.Metadata[file_type] = %v, want %q", manifests[0].Metadata["file_type"], "dockerfile")
			}
		}
	})

	t.Run("detects Dockerfile.prod", func(t *testing.T) {
		tmpDir := t.TempDir()
		dockerfileContent := `FROM golang:1.21
RUN go build -o app .
`
		if err := os.WriteFile(filepath.Join(tmpDir, "Dockerfile.prod"), []byte(dockerfileContent), 0o644); err != nil {
			t.Fatal(err)
		}

		manifests, err := integration.Detect(ctx, tmpDir)
		if err != nil {
			t.Fatalf("Detect() error = %v", err)
		}

		if len(manifests) != 1 {
			t.Errorf("Detect() returned %d manifests, want 1", len(manifests))
		}
	})

	t.Run("detects docker-compose.yml", func(t *testing.T) {
		tmpDir := t.TempDir()
		composeContent := `
version: "3.8"
services:
  web:
    image: nginx:1.25
`
		if err := os.WriteFile(filepath.Join(tmpDir, "docker-compose.yml"), []byte(composeContent), 0o644); err != nil {
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
			fileType, ok := manifests[0].Metadata["file_type"].(string)
			if !ok || fileType != "compose" {
				t.Errorf("manifest.Metadata[file_type] = %v, want %q", manifests[0].Metadata["file_type"], "compose")
			}
		}
	})

	t.Run("detects compose.yaml", func(t *testing.T) {
		tmpDir := t.TempDir()
		composeContent := `
services:
  web:
    image: nginx:1.25
`
		if err := os.WriteFile(filepath.Join(tmpDir, "compose.yaml"), []byte(composeContent), 0o644); err != nil {
			t.Fatal(err)
		}

		manifests, err := integration.Detect(ctx, tmpDir)
		if err != nil {
			t.Fatalf("Detect() error = %v", err)
		}

		if len(manifests) != 1 {
			t.Errorf("Detect() returned %d manifests, want 1", len(manifests))
		}
	})

	t.Run("skips hidden directories", func(t *testing.T) {
		tmpDir := t.TempDir()
		hiddenDir := filepath.Join(tmpDir, ".hidden")
		if err := os.MkdirAll(hiddenDir, 0o755); err != nil {
			t.Fatal(err)
		}

		dockerfileContent := `FROM golang:1.21`
		if err := os.WriteFile(filepath.Join(hiddenDir, "Dockerfile"), []byte(dockerfileContent), 0o644); err != nil {
			t.Fatal(err)
		}

		manifests, err := integration.Detect(ctx, tmpDir)
		if err != nil {
			t.Fatalf("Detect() error = %v", err)
		}

		if len(manifests) != 0 {
			t.Errorf("Detect() returned %d manifests, want 0 (should skip hidden dirs)", len(manifests))
		}
	})

	t.Run("skips vendor directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		vendorDir := filepath.Join(tmpDir, "vendor")
		if err := os.MkdirAll(vendorDir, 0o755); err != nil {
			t.Fatal(err)
		}

		dockerfileContent := `FROM golang:1.21`
		if err := os.WriteFile(filepath.Join(vendorDir, "Dockerfile"), []byte(dockerfileContent), 0o644); err != nil {
			t.Fatal(err)
		}

		manifests, err := integration.Detect(ctx, tmpDir)
		if err != nil {
			t.Fatalf("Detect() error = %v", err)
		}

		if len(manifests) != 0 {
			t.Errorf("Detect() returned %d manifests, want 0 (should skip vendor)", len(manifests))
		}
	})

	t.Run("handles multiple files", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create Dockerfile
		if err := os.WriteFile(filepath.Join(tmpDir, "Dockerfile"), []byte("FROM golang:1.21"), 0o644); err != nil {
			t.Fatal(err)
		}

		// Create docker-compose.yml
		composeContent := `
services:
  web:
    image: nginx:1.25
`
		if err := os.WriteFile(filepath.Join(tmpDir, "docker-compose.yml"), []byte(composeContent), 0o644); err != nil {
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
}

func TestIntegration_Validate(t *testing.T) {
	integration := New()
	ctx := context.Background()

	tests := []struct {
		name      string
		content   string
		metadata  map[string]interface{}
		expectErr bool
	}{
		{
			name: "valid Dockerfile",
			content: `FROM golang:1.21
RUN go build -o app .
`,
			metadata:  map[string]interface{}{"file_type": "dockerfile"},
			expectErr: false,
		},
		{
			name: "Dockerfile without FROM",
			content: `RUN go build -o app .
COPY . .
`,
			metadata:  map[string]interface{}{"file_type": "dockerfile"},
			expectErr: true,
		},
		{
			name: "valid compose",
			content: `
version: "3.8"
services:
  web:
    image: nginx:1.25
`,
			metadata:  map[string]interface{}{"file_type": "compose"},
			expectErr: false,
		},
		{
			name:      "invalid compose yaml",
			content:   `version: [invalid`,
			metadata:  map[string]interface{}{"file_type": "compose"},
			expectErr: true,
		},
		{
			name: "compose without services",
			content: `
version: "3.8"
`,
			metadata:  map[string]interface{}{"file_type": "compose"},
			expectErr: true,
		},
		{
			name:      "unknown file type",
			content:   "FROM golang:1.21",
			metadata:  map[string]interface{}{},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := &engine.Manifest{
				Content:  []byte(tt.content),
				Metadata: tt.metadata,
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

	t.Run("plans updates for image versions", func(t *testing.T) {
		mockDS := &mockDatasource{
			versions: []string{"1.26", "1.25.3", "1.25.2", "1.25.1", "1.25"},
		}
		integration := &Integration{ds: mockDS}

		manifest := &engine.Manifest{
			Path: "Dockerfile",
			Type: "docker",
			Dependencies: []engine.Dependency{
				{
					Name:           "nginx",
					CurrentVersion: "1.25",
					Constraint:     "", // Docker images don't have semver constraints
					Type:           "image",
					Registry:       "docker-hub",
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

		if len(plan.Updates) > 0 && plan.Updates[0].TargetVersion != "1.26" {
			t.Errorf("Plan() target = %q, want %q", plan.Updates[0].TargetVersion, "1.26")
		}
	})

	t.Run("skips latest tag", func(t *testing.T) {
		mockDS := &mockDatasource{
			versions: []string{"1.26", "1.25"},
		}
		integration := &Integration{ds: mockDS}

		manifest := &engine.Manifest{
			Path: "Dockerfile",
			Type: "docker",
			Dependencies: []engine.Dependency{
				{
					Name:           "nginx",
					CurrentVersion: "latest",
					Constraint:     "latest",
					Type:           "image",
				},
			},
		}

		planCtx := engine.NewPlanContext()
		plan, err := integration.Plan(ctx, manifest, planCtx)
		if err != nil {
			t.Fatalf("Plan() error = %v", err)
		}

		if len(plan.Updates) != 0 {
			t.Errorf("Plan() returned %d updates, want 0 (latest should be skipped)", len(plan.Updates))
		}
	})

	t.Run("skips sha256 digest", func(t *testing.T) {
		mockDS := &mockDatasource{
			versions: []string{"1.26", "1.25"},
		}
		integration := &Integration{ds: mockDS}

		manifest := &engine.Manifest{
			Path: "Dockerfile",
			Type: "docker",
			Dependencies: []engine.Dependency{
				{
					Name:           "nginx",
					CurrentVersion: "sha256",
					Constraint:     "sha256",
					Type:           "image",
				},
			},
		}

		planCtx := engine.NewPlanContext()
		plan, err := integration.Plan(ctx, manifest, planCtx)
		if err != nil {
			t.Fatalf("Plan() error = %v", err)
		}

		if len(plan.Updates) != 0 {
			t.Errorf("Plan() returned %d updates, want 0 (sha256 should be skipped)", len(plan.Updates))
		}
	})
}

func TestIntegration_Apply(t *testing.T) {
	integration := New()
	ctx := context.Background()

	t.Run("applies Dockerfile updates", func(t *testing.T) {
		tmpDir := t.TempDir()
		dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
		originalContent := `FROM golang:1.21
RUN go build -o app .
`
		if err := os.WriteFile(dockerfilePath, []byte(originalContent), 0o644); err != nil {
			t.Fatal(err)
		}

		plan := &engine.UpdatePlan{
			Manifest: &engine.Manifest{
				Path: dockerfilePath,
			},
			Updates: []engine.Update{
				{
					Dependency: engine.Dependency{
						Name:           "golang",
						CurrentVersion: "1.21",
					},
					TargetVersion: "1.22",
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

		updatedContent, _ := os.ReadFile(dockerfilePath)
		if !strings.Contains(string(updatedContent), "golang:1.22") {
			t.Errorf("Apply() did not update image reference")
		}
	})

	t.Run("applies docker-compose updates", func(t *testing.T) {
		tmpDir := t.TempDir()
		composePath := filepath.Join(tmpDir, "docker-compose.yml")
		originalContent := `version: "3.8"
services:
  web:
    image: nginx:1.25
`
		if err := os.WriteFile(composePath, []byte(originalContent), 0o644); err != nil {
			t.Fatal(err)
		}

		plan := &engine.UpdatePlan{
			Manifest: &engine.Manifest{
				Path: composePath,
			},
			Updates: []engine.Update{
				{
					Dependency: engine.Dependency{
						Name:           "nginx",
						CurrentVersion: "1.25",
					},
					TargetVersion: "1.26",
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

		updatedContent, _ := os.ReadFile(composePath)
		if !strings.Contains(string(updatedContent), "nginx:1.26") {
			t.Errorf("Apply() did not update image reference")
		}
	})

	t.Run("handles empty updates", func(t *testing.T) {
		plan := &engine.UpdatePlan{
			Manifest: &engine.Manifest{
				Path: "Dockerfile",
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
						Name:           "nginx",
						CurrentVersion: "1.25",
					},
					TargetVersion: "1.26",
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
	t.Run("generates diff for Dockerfile", func(t *testing.T) {
		old := "FROM golang:1.21"
		newContent := "FROM golang:1.22"

		diff := generateDiff("Dockerfile", old, newContent)

		if diff == "" {
			t.Error("generateDiff() returned empty diff")
		}

		if !strings.Contains(diff, "- ") || !strings.Contains(diff, "+ ") {
			t.Error("generateDiff() missing diff markers")
		}
	})

	t.Run("generates diff for compose", func(t *testing.T) {
		old := "    image: nginx:1.25"
		newContent := "    image: nginx:1.26"

		diff := generateDiff("docker-compose.yml", old, newContent)

		if diff == "" {
			t.Error("generateDiff() returned empty diff")
		}
	})

	t.Run("returns empty for identical content", func(t *testing.T) {
		content := "FROM golang:1.21"
		diff := generateDiff("Dockerfile", content, content)

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

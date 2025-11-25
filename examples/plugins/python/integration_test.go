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

package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/santosr2/uptool/internal/engine"
)

// TestParseRequirements tests the requirements.txt parser
func TestParseRequirements(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantDeps int
		wantErr  bool
	}{
		{
			name: "simple requirements",
			content: `requests==2.28.0
flask==2.2.0
pytest==7.0.0`,
			wantDeps: 3,
			wantErr:  false,
		},
		{
			name: "with comments",
			content: `# Web frameworks
flask==2.2.0
# HTTP library
requests==2.28.0  # Inline comment`,
			wantDeps: 2,
			wantErr:  false,
		},
		{
			name: "version constraints",
			content: `requests>=2.28.0
flask~=2.2.0
pytest>=7.0.0,<8.0.0`,
			wantDeps: 3,
			wantErr:  false,
		},
		{
			name: "with extras",
			content: `requests[security]==2.28.0
flask[async]>=2.2.0`,
			wantDeps: 2,
			wantErr:  false,
		},
		{
			name: "empty and blank lines",
			content: `
requests==2.28.0

flask==2.2.0

`,
			wantDeps: 2,
			wantErr:  false,
		},
		{
			name: "pip flags ignored",
			content: `--index-url https://pypi.org/simple
-e git+https://github.com/user/repo.git
requests==2.28.0`,
			wantDeps: 1,
			wantErr:  false,
		},
		{
			name:     "invalid format",
			content:  `invalid-line-without-version`,
			wantDeps: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, err := ParseRequirements(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRequirements() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(deps) != tt.wantDeps {
				t.Errorf("ParseRequirements() got %d dependencies, want %d", len(deps), tt.wantDeps)
			}
		})
	}
}

// TestParseRequirement tests individual requirement parsing
func TestParseRequirement(t *testing.T) {
	tests := []struct {
		name        string
		line        string
		wantName    string
		wantVersion string
		wantOp      string
		wantErr     bool
	}{
		{
			name:        "exact version",
			line:        "requests==2.28.0",
			wantName:    "requests",
			wantVersion: "2.28.0",
			wantOp:      "==",
			wantErr:     false,
		},
		{
			name:        "greater than or equal",
			line:        "flask>=2.2.0",
			wantName:    "flask",
			wantVersion: "2.2.0",
			wantOp:      ">=",
			wantErr:     false,
		},
		{
			name:        "compatible release",
			line:        "pytest~=7.0.0",
			wantName:    "pytest",
			wantVersion: "7.0.0",
			wantOp:      "~=",
			wantErr:     false,
		},
		{
			name:        "with extras",
			line:        "requests[security]==2.28.0",
			wantName:    "requests",
			wantVersion: "2.28.0",
			wantOp:      "==",
			wantErr:     false,
		},
		{
			name:        "package with hyphens",
			line:        "python-dotenv==0.21.0",
			wantName:    "python-dotenv",
			wantVersion: "0.21.0",
			wantOp:      "==",
			wantErr:     false,
		},
		{
			name:        "pre-release version",
			line:        "django==4.2.0rc1",
			wantName:    "django",
			wantVersion: "4.2.0rc1",
			wantOp:      "==",
			wantErr:     false,
		},
		{
			name:    "invalid format",
			line:    "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep, err := parseRequirement(tt.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRequirement() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if dep.Name != tt.wantName {
					t.Errorf("parseRequirement() name = %v, want %v", dep.Name, tt.wantName)
				}
				if dep.CurrentVersion != tt.wantVersion {
					t.Errorf("parseRequirement() version = %v, want %v", dep.CurrentVersion, tt.wantVersion)
				}
				if dep.Constraint != tt.wantOp {
					t.Errorf("parseRequirement() operator = %v, want %v", dep.Constraint, tt.wantOp)
				}
			}
		})
	}
}

// TestIntegrationDetect tests the Detect method
func TestIntegrationDetect(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create test requirements.txt
	requirementsPath := filepath.Join(tmpDir, "requirements.txt")
	content := `requests==2.28.0
flask==2.2.0`
	if err := os.WriteFile(requirementsPath, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test detection
	integration := New().(*Integration)
	ctx := context.Background()
	manifests, err := integration.Detect(ctx, tmpDir)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	if len(manifests) != 1 {
		t.Errorf("Detect() found %d manifests, want 1", len(manifests))
	}

	if len(manifests) > 0 {
		manifest := manifests[0]
		if manifest.Path != requirementsPath {
			t.Errorf("Detect() path = %v, want %v", manifest.Path, requirementsPath)
		}
		if manifest.Type != "python" {
			t.Errorf("Detect() type = %v, want python", manifest.Type)
		}
	}
}

// TestIntegrationValidate tests the Validate method
func TestIntegrationValidate(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name: "valid requirements",
			content: `requests==2.28.0
flask==2.2.0`,
			wantErr: false,
		},
		{
			name:    "empty file",
			content: "",
			wantErr: false, // Empty requirements.txt is technically valid
		},
		{
			name: "only comments",
			content: `# Just comments
# No actual dependencies`,
			wantErr: false, // Comments-only file is valid, just has no dependencies
		},
		{
			name:    "invalid format",
			content: `invalid-line`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := filepath.Join(t.TempDir(), "requirements.txt")
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0o644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			integration := New().(*Integration)
			ctx := context.Background()
			manifest := &engine.Manifest{
				Path: tmpFile,
				Type: "python",
			}

			err := integration.Validate(ctx, manifest)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestIntegrationName tests the Name method
func TestIntegrationName(t *testing.T) {
	integration := New().(*Integration)
	if name := integration.Name(); name != "python" {
		t.Errorf("Name() = %v, want python", name)
	}
}

// TestIntegration_EndToEnd tests the full workflow
func TestIntegration_EndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create temporary directory
	tmpDir := t.TempDir()
	requirementsPath := filepath.Join(tmpDir, "requirements.txt")

	// Create initial requirements.txt
	initialContent := `# Test requirements file
requests==2.28.0
flask==2.2.0
pytest>=7.0.0`
	if err := os.WriteFile(requirementsPath, []byte(initialContent), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	integration := New().(*Integration)
	ctx := context.Background()

	// Test Detect
	manifests, err := integration.Detect(ctx, tmpDir)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}
	if len(manifests) != 1 {
		t.Fatalf("Detect() found %d manifests, want 1", len(manifests))
	}

	manifest := manifests[0]

	// Test that dependencies were parsed during Detect
	if len(manifest.Dependencies) != 3 {
		t.Fatalf("Detect() found %d dependencies, want 3", len(manifest.Dependencies))
	}

	// Test Validate
	if err := integration.Validate(ctx, manifest); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	// Verify file content is preserved
	content, err := os.ReadFile(requirementsPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(content) != initialContent {
		t.Error("File content was modified unexpectedly")
	}
}

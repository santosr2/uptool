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

package rewrite

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestReplaceYAMLValue(t *testing.T) {
	tests := []struct {
		matcher   func(*yaml.Node) bool
		name      string
		content   string
		oldValue  string
		newValue  string
		wantValue string
		path      []string
		wantErr   bool
	}{
		{
			name: "simple field replacement",
			content: `
name: myapp
version: 1.0.0
`,
			path:      []string{"version"},
			oldValue:  "1.0.0",
			newValue:  "2.0.0",
			wantErr:   false,
			wantValue: "2.0.0",
		},
		{
			name: "nested field replacement",
			content: `
metadata:
  name: myapp
  version: 1.0.0
`,
			path:      []string{"metadata", "version"},
			oldValue:  "1.0.0",
			newValue:  "2.0.0",
			wantErr:   false,
			wantValue: "2.0.0",
		},
		{
			name: "array wildcard replacement",
			content: `
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.0.0
  - repo: https://github.com/psf/black
    rev: 22.3.0
`,
			path:      []string{"repos", "*", "rev"},
			oldValue:  "v4.0.0",
			newValue:  "v4.5.0",
			wantErr:   false,
			wantValue: "v4.5.0",
		},
		{
			name: "value not found",
			content: `
name: myapp
version: 1.0.0
`,
			path:     []string{"version"},
			oldValue: "nonexistent",
			newValue: "2.0.0",
			wantErr:  true,
		},
		{
			name: "with matcher function",
			content: `
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.0.0
  - repo: https://github.com/psf/black
    rev: v4.0.0
`,
			path:     []string{"repos", "*", "rev"},
			oldValue: "v4.0.0",
			newValue: "v4.5.0",
			matcher: func(node *yaml.Node) bool {
				// Only update if repo contains "pre-commit"
				for i := 0; i < len(node.Content); i += 2 {
					if node.Content[i].Value == "repo" &&
						strings.Contains(node.Content[i+1].Value, "pre-commit") {
						return true
					}
				}
				return false
			},
			wantErr:   false,
			wantValue: "v4.5.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReplaceYAMLValue(tt.content, tt.path, tt.oldValue, tt.newValue, tt.matcher)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReplaceYAMLValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !strings.Contains(got, tt.wantValue) {
				t.Errorf("ReplaceYAMLValue() result should contain %q, got:\n%s", tt.wantValue, got)
			}
		})
	}
}

func TestUpdateYAMLField(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		newValue  string
		wantValue string
		path      []string
		wantErr   bool
	}{
		{
			name: "update top-level field",
			content: `
name: myapp
version: 1.0.0
`,
			path:      []string{"version"},
			newValue:  "2.0.0",
			wantErr:   false,
			wantValue: "2.0.0",
		},
		{
			name: "update nested field",
			content: `
app:
  metadata:
    version: 1.0.0
    name: myapp
`,
			path:      []string{"app", "metadata", "version"},
			newValue:  "3.0.0",
			wantErr:   false,
			wantValue: "3.0.0",
		},
		{
			name: "path not found",
			content: `
name: myapp
version: 1.0.0
`,
			path:     []string{"nonexistent", "field"},
			newValue: "value",
			wantErr:  true,
		},
		{
			name: "update with special characters",
			content: `
image:
  tag: v1.0.0
`,
			path:      []string{"image", "tag"},
			newValue:  "v2.0.0-beta.1",
			wantErr:   false,
			wantValue: "v2.0.0-beta.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UpdateYAMLField(tt.content, tt.path, tt.newValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateYAMLField() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !strings.Contains(got, tt.wantValue) {
				t.Errorf("UpdateYAMLField() result should contain %q, got:\n%s", tt.wantValue, got)
			}
		})
	}
}

func TestReplaceYAMLValue_PreservesFormatting(t *testing.T) {
	content := `# Comment at top
repos:
  # Pre-commit hooks
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.0.0
    hooks:
      - id: trailing-whitespace
`

	got, err := ReplaceYAMLValue(content, []string{"repos", "*", "rev"}, "v4.0.0", "v4.5.0", nil)
	if err != nil {
		t.Fatalf("ReplaceYAMLValue() error = %v", err)
	}

	// Verify the new value is present
	if !strings.Contains(got, "v4.5.0") {
		t.Error("ReplaceYAMLValue() should contain new value v4.5.0")
	}

	// Verify structure is preserved (though comments might be lost in gopkg.in/yaml.v3)
	if !strings.Contains(got, "repos:") {
		t.Error("ReplaceYAMLValue() should preserve YAML structure")
	}
	if !strings.Contains(got, "hooks:") {
		t.Error("ReplaceYAMLValue() should preserve nested structure")
	}
}

func TestReplaceYAMLValue_InvalidYAML(t *testing.T) {
	invalidYAML := `
this is not valid YAML: [unclosed bracket
`

	_, err := ReplaceYAMLValue(invalidYAML, []string{"key"}, "old", "new", nil)
	if err == nil {
		t.Error("ReplaceYAMLValue() should return error for invalid YAML")
	}
}

func TestUpdateYAMLField_InvalidYAML(t *testing.T) {
	invalidYAML := `{this is definitely not valid YAML!!`

	_, err := UpdateYAMLField(invalidYAML, []string{"name"}, "newvalue")
	if err == nil {
		t.Error("UpdateYAMLField() should return error for invalid YAML")
	}
}

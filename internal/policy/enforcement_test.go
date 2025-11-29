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

package policy

import (
	"context"
	"testing"
)

func TestNewEnforcer(t *testing.T) {
	config := &Config{
		Version: 1,
		OrgPolicy: &OrgPolicy{
			RequireSignoffFrom: []string{"team@example.com"},
		},
	}

	enforcer := NewEnforcer(config)

	if enforcer == nil {
		t.Fatal("NewEnforcer() returned nil")
	}

	if enforcer.config != config {
		t.Error("NewEnforcer() did not store config")
	}
}

func TestEnforcer_NoPolicies(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		config *Config
		name   string
	}{
		{
			name:   "nil config",
			config: nil,
		},
		{
			name: "no org policy",
			config: &Config{
				Version: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enforcer := NewEnforcer(tt.config)
			result, err := enforcer.Enforce(ctx)
			if err != nil {
				t.Fatalf("Enforce() error = %v, want nil", err)
			}

			if result == nil {
				t.Fatal("Enforce() returned nil result")
			}

			if !result.SignoffValid {
				t.Error("Enforce() SignoffValid = false, want true (no policy)")
			}

			if !result.CosignValid {
				t.Error("Enforce() CosignValid = false, want true (no policy)")
			}

			if !result.AutoMergeAllowed {
				t.Error("Enforce() AutoMergeAllowed = false, want true (no policy)")
			}
		})
	}
}

func TestEnforcer_SignoffNoGitHubEnv(t *testing.T) {
	ctx := context.Background()

	config := &Config{
		Version: 1,
		OrgPolicy: &OrgPolicy{
			RequireSignoffFrom: []string{"team@example.com"},
		},
	}

	// Ensure GitHub env vars are not set
	t.Setenv("GITHUB_REPOSITORY", "")
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GITHUB_PR_NUMBER", "")

	enforcer := NewEnforcer(config)
	result, err := enforcer.Enforce(ctx)
	if err != nil {
		t.Fatalf("Enforce() error = %v, want nil", err)
	}

	if result.SignoffValid {
		t.Error("Enforce() SignoffValid = true, want false (missing GitHub env)")
	}

	if len(result.SignoffErrors) == 0 {
		t.Error("Enforce() SignoffErrors empty, want error about missing GITHUB_REPOSITORY")
	}
}

func TestEnforcer_CosignNoBinary(t *testing.T) {
	ctx := context.Background()

	config := &Config{
		Version: 1,
		OrgPolicy: &OrgPolicy{
			Signing: &SigningConfig{
				CosignVerify: true,
			},
		},
	}

	enforcer := NewEnforcer(config)
	result, err := enforcer.Enforce(ctx)
	if err != nil {
		t.Fatalf("Enforce() error = %v, want nil", err)
	}

	// If cosign is not installed, should fail gracefully
	if result.CosignValid && len(result.CosignErrors) > 0 {
		t.Error("Enforce() CosignValid = true but has errors")
	}
}

func TestEnforcer_AutoMergeNoGitHubEnv(t *testing.T) {
	ctx := context.Background()

	config := &Config{
		Version: 1,
		OrgPolicy: &OrgPolicy{
			AutoMerge: &AutoMergeConfig{
				Enabled: true,
				Guards:  []string{"ci-green"},
			},
		},
	}

	// Ensure GitHub env vars are not set
	t.Setenv("GITHUB_REPOSITORY", "")
	t.Setenv("GITHUB_TOKEN", "")

	enforcer := NewEnforcer(config)
	result, err := enforcer.Enforce(ctx)
	if err != nil {
		t.Fatalf("Enforce() error = %v, want nil", err)
	}

	if result.AutoMergeAllowed {
		t.Error("Enforce() AutoMergeAllowed = true, want false (missing GitHub env)")
	}

	if len(result.AutoMergeErrors) == 0 {
		t.Error("Enforce() AutoMergeErrors empty, want error about missing GitHub env")
	}
}

func TestEnforcementResult_Structure(t *testing.T) {
	result := &EnforcementResult{
		SignoffValid: true,
		CosignValid:  false,
		CosignErrors: []string{"test error"},
		GuardsStatus: map[string]bool{"ci-green": true},
	}

	if !result.SignoffValid {
		t.Error("SignoffValid should be true")
	}

	if result.CosignValid {
		t.Error("CosignValid should be false")
	}

	if len(result.CosignErrors) != 1 {
		t.Errorf("CosignErrors count = %d, want 1", len(result.CosignErrors))
	}

	if !result.GuardsStatus["ci-green"] {
		t.Error("GuardsStatus[ci-green] should be true")
	}
}

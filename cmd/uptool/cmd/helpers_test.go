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

package cmd

import (
	"testing"

	"github.com/santosr2/uptool/internal/engine"
	"github.com/santosr2/uptool/internal/policy"
)

func TestBuildPolicies(t *testing.T) {
	tests := []struct {
		config        *policy.Config
		wantPolicies  map[string]bool
		wantUpdateFor map[string]string
		name          string
	}{
		{
			name: "policy enabled - includes policy",
			config: &policy.Config{
				Version: 1,
				Integrations: []policy.IntegrationConfig{
					{
						ID:      "npm",
						Enabled: true,
						Policy: engine.IntegrationPolicy{
							Enabled:         true,
							Update:          "minor",
							AllowPrerelease: false,
							Pin:             false,
						},
					},
				},
			},
			wantPolicies: map[string]bool{
				"npm": true,
			},
			wantUpdateFor: map[string]string{
				"npm": "minor",
			},
		},
		{
			name: "policy disabled - excludes policy",
			config: &policy.Config{
				Version: 1,
				Integrations: []policy.IntegrationConfig{
					{
						ID:      "helm",
						Enabled: true,
						Policy: engine.IntegrationPolicy{
							Enabled:         false, // Policy disabled
							Update:          "patch",
							AllowPrerelease: false,
							Pin:             true,
						},
					},
				},
			},
			wantPolicies: map[string]bool{
				"helm": false, // Should NOT be in policies map
			},
		},
		{
			name: "integration disabled - excludes policy",
			config: &policy.Config{
				Version: 1,
				Integrations: []policy.IntegrationConfig{
					{
						ID:      "terraform",
						Enabled: false, // Integration disabled
						Policy: engine.IntegrationPolicy{
							Enabled:         true,
							Update:          "major",
							AllowPrerelease: true,
							Pin:             true,
						},
					},
				},
			},
			wantPolicies: map[string]bool{
				"terraform": false, // Should NOT be in policies map
			},
		},
		{
			name: "mixed enabled and disabled policies",
			config: &policy.Config{
				Version: 1,
				Integrations: []policy.IntegrationConfig{
					{
						ID:      "npm",
						Enabled: true,
						Policy: engine.IntegrationPolicy{
							Enabled:         true,
							Update:          "minor",
							AllowPrerelease: false,
						},
					},
					{
						ID:      "helm",
						Enabled: true,
						Policy: engine.IntegrationPolicy{
							Enabled:         false, // Policy disabled
							Update:          "patch",
							AllowPrerelease: false,
						},
					},
					{
						ID:      "terraform",
						Enabled: true,
						Policy: engine.IntegrationPolicy{
							Enabled:         true,
							Update:          "major",
							AllowPrerelease: true,
						},
					},
				},
			},
			wantPolicies: map[string]bool{
				"npm":       true,  // Enabled
				"helm":      false, // Policy disabled
				"terraform": true,  // Enabled
			},
			wantUpdateFor: map[string]string{
				"npm":       "minor",
				"terraform": "major",
			},
		},
		{
			name: "policy enabled but no settings - not included",
			config: &policy.Config{
				Version: 1,
				Integrations: []policy.IntegrationConfig{
					{
						ID:      "npm",
						Enabled: true,
						Policy: engine.IntegrationPolicy{
							Enabled: true,
							// No Update, no AllowPrerelease
						},
					},
				},
			},
			wantPolicies: map[string]bool{
				"npm": false, // No settings, so not included
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPolicies := buildPolicies(tt.config)

			// Check which integrations have policies
			for id, shouldHavePolicy := range tt.wantPolicies {
				_, hasPolicy := gotPolicies[id]
				if hasPolicy != shouldHavePolicy {
					t.Errorf("Integration %s: hasPolicy = %v, want %v", id, hasPolicy, shouldHavePolicy)
				}
			}

			// Check update levels for integrations that should have policies
			for id, wantUpdate := range tt.wantUpdateFor {
				p, ok := gotPolicies[id]
				if !ok {
					t.Errorf("Integration %s: policy not found", id)
					continue
				}
				if p.Update != wantUpdate {
					t.Errorf("Integration %s: Update = %q, want %q", id, p.Update, wantUpdate)
				}
			}
		})
	}
}

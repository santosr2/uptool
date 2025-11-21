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
	"os"
	"path/filepath"
	"testing"

	"github.com/santosr2/uptool/internal/engine"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "uptool.yaml")

	configContent := `version: 1
integrations:
  - id: precommit
    enabled: true
    policy:
      enabled: true
      update: minor
      allow_prerelease: false
      pin: true
      cadence: weekly
`

	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if config.Version != 1 {
		t.Errorf("Version = %d, want 1", config.Version)
	}

	if len(config.Integrations) != 1 {
		t.Fatalf("len(Integrations) = %d, want 1", len(config.Integrations))
	}

	integ := config.Integrations[0]
	if integ.ID != "precommit" {
		t.Errorf("Integration.ID = %q, want %q", integ.ID, "precommit")
	}

	if !integ.Enabled {
		t.Error("Integration.Enabled = false, want true")
	}

	if integ.Policy.Update != "minor" {
		t.Errorf("Policy.Update = %q, want %q", integ.Policy.Update, "minor")
	}
}

func TestValidateIntegrationPolicy(t *testing.T) {
	tests := []struct {
		name    string
		policy  engine.IntegrationPolicy
		wantErr bool
	}{
		{
			name: "valid policy",
			policy: engine.IntegrationPolicy{
				Update:  "minor",
				Cadence: "weekly",
			},
			wantErr: false,
		},
		{
			name: "invalid update strategy",
			policy: engine.IntegrationPolicy{
				Update: "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid cadence",
			policy: engine.IntegrationPolicy{
				Update:  "minor",
				Cadence: "invalid",
			},
			wantErr: true,
		},
		{
			name: "all valid update strategies",
			policy: engine.IntegrationPolicy{
				Update: "none",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIntegrationPolicy(&tt.policy)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIntegrationPolicy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Version != 1 {
		t.Errorf("DefaultConfig().Version = %d, want 1", config.Version)
	}

	if len(config.Integrations) == 0 {
		t.Error("DefaultConfig() has no integrations")
	}

	// Check that all default integrations have valid IDs
	seenIDs := make(map[string]bool)
	for _, integ := range config.Integrations {
		if integ.ID == "" {
			t.Error("DefaultConfig() integration has empty ID")
		}
		if seenIDs[integ.ID] {
			t.Errorf("DefaultConfig() has duplicate integration ID: %s", integ.ID)
		}
		seenIDs[integ.ID] = true
	}
}

func TestEnabledIntegrations(t *testing.T) {
	config := &Config{
		Version: 1,
		Integrations: []IntegrationConfig{
			{ID: "a", Enabled: true},
			{ID: "b", Enabled: false},
			{ID: "c", Enabled: true},
		},
	}

	enabled := config.EnabledIntegrations()

	if len(enabled) != 2 {
		t.Fatalf("EnabledIntegrations() returned %d items, want 2", len(enabled))
	}

	if enabled[0] != "a" || enabled[1] != "c" {
		t.Errorf("EnabledIntegrations() = %v, want [a c]", enabled)
	}
}

func TestToPolicyMap(t *testing.T) {
	config := &Config{
		Version: 1,
		Integrations: []IntegrationConfig{
			{
				ID:      "precommit",
				Enabled: true,
				Policy: engine.IntegrationPolicy{
					Update: "minor",
				},
			},
			{
				ID:      "asdf",
				Enabled: true,
				Policy: engine.IntegrationPolicy{
					Update: "patch",
				},
			},
		},
	}

	policyMap := config.ToPolicyMap()

	if len(policyMap) != 2 {
		t.Fatalf("ToPolicyMap() returned %d items, want 2", len(policyMap))
	}

	if policy, ok := policyMap["precommit"]; !ok {
		t.Error("ToPolicyMap() missing 'precommit' key")
	} else if policy.Update != "minor" {
		t.Errorf("precommit policy.Update = %q, want 'minor'", policy.Update)
	}

	if policy, ok := policyMap["asdf"]; !ok {
		t.Error("ToPolicyMap() missing 'asdf' key")
	} else if policy.Update != "patch" {
		t.Errorf("asdf policy.Update = %q, want 'patch'", policy.Update)
	}
}

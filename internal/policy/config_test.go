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

func TestConfig_OrgPolicyMethods(t *testing.T) {
	tests := []struct {
		config                   *Config
		name                     string
		wantAutoMergeGuardsCount int
		wantRequiresSignoff      bool
		wantRequiresCosign       bool
		wantAutoMergeEnabled     bool
	}{
		{
			name: "no org policy",
			config: &Config{
				Version: 1,
			},
			wantRequiresSignoff:      false,
			wantRequiresCosign:       false,
			wantAutoMergeEnabled:     false,
			wantAutoMergeGuardsCount: 0,
		},
		{
			name: "with require_signoff_from",
			config: &Config{
				Version: 1,
				OrgPolicy: &OrgPolicy{
					RequireSignoffFrom: []string{"team@example.com"},
				},
			},
			wantRequiresSignoff:      true,
			wantRequiresCosign:       false,
			wantAutoMergeEnabled:     false,
			wantAutoMergeGuardsCount: 0,
		},
		{
			name: "with cosign verification",
			config: &Config{
				Version: 1,
				OrgPolicy: &OrgPolicy{
					Signing: &SigningConfig{
						CosignVerify: true,
					},
				},
			},
			wantRequiresSignoff:      false,
			wantRequiresCosign:       true,
			wantAutoMergeEnabled:     false,
			wantAutoMergeGuardsCount: 0,
		},
		{
			name: "with auto merge enabled",
			config: &Config{
				Version: 1,
				OrgPolicy: &OrgPolicy{
					AutoMerge: &AutoMergeConfig{
						Enabled: true,
						Guards:  []string{"ci-green", "codeowners-approve"},
					},
				},
			},
			wantRequiresSignoff:      false,
			wantRequiresCosign:       false,
			wantAutoMergeEnabled:     true,
			wantAutoMergeGuardsCount: 2,
		},
		{
			name: "complete org policy",
			config: &Config{
				Version: 1,
				OrgPolicy: &OrgPolicy{
					RequireSignoffFrom: []string{"platform-team@example.com", "security-team@example.com"},
					Signing: &SigningConfig{
						CosignVerify: true,
					},
					AutoMerge: &AutoMergeConfig{
						Enabled: true,
						Guards:  []string{"ci-green", "codeowners-approve", "security-scan"},
					},
				},
			},
			wantRequiresSignoff:      true,
			wantRequiresCosign:       true,
			wantAutoMergeEnabled:     true,
			wantAutoMergeGuardsCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.RequiresSignoff(); got != tt.wantRequiresSignoff {
				t.Errorf("RequiresSignoff() = %v, want %v", got, tt.wantRequiresSignoff)
			}

			if got := tt.config.RequiresCosignVerification(); got != tt.wantRequiresCosign {
				t.Errorf("RequiresCosignVerification() = %v, want %v", got, tt.wantRequiresCosign)
			}

			if got := tt.config.IsAutoMergeEnabled(); got != tt.wantAutoMergeEnabled {
				t.Errorf("IsAutoMergeEnabled() = %v, want %v", got, tt.wantAutoMergeEnabled)
			}

			guards := tt.config.GetAutoMergeGuards()
			if len(guards) != tt.wantAutoMergeGuardsCount {
				t.Errorf("GetAutoMergeGuards() count = %d, want %d", len(guards), tt.wantAutoMergeGuardsCount)
			}

			orgPolicy := tt.config.GetOrgPolicy()
			if (orgPolicy != nil) != (tt.config.OrgPolicy != nil) {
				t.Errorf("GetOrgPolicy() returned %v, want %v", orgPolicy != nil, tt.config.OrgPolicy != nil)
			}
		})
	}
}

func TestConfig_ToMatchConfigMap(t *testing.T) {
	config := &Config{
		Version: 1,
		Integrations: []IntegrationConfig{
			{
				ID:      "npm",
				Enabled: true,
				Match: &MatchConfig{
					Files: []string{"package.json", "apps/*/package.json"},
				},
			},
			{
				ID:      "terraform",
				Enabled: true,
				Match: &MatchConfig{
					Files: []string{"*.tf", "modules/**/*.tf"},
				},
			},
			{
				ID:      "helm",
				Enabled: true,
				// No match config
			},
		},
	}

	matchMap := config.ToMatchConfigMap()

	if len(matchMap) != 2 {
		t.Errorf("ToMatchConfigMap() returned %d items, want 2", len(matchMap))
	}

	if matchConfig, ok := matchMap["npm"]; !ok {
		t.Error("ToMatchConfigMap() missing 'npm' key")
	} else if len(matchConfig.Files) != 2 {
		t.Errorf("npm patterns count = %d, want 2", len(matchConfig.Files))
	}

	if matchConfig, ok := matchMap["terraform"]; !ok {
		t.Error("ToMatchConfigMap() missing 'terraform' key")
	} else if len(matchConfig.Files) != 2 {
		t.Errorf("terraform patterns count = %d, want 2", len(matchConfig.Files))
	}

	if _, ok := matchMap["helm"]; ok {
		t.Error("ToMatchConfigMap() should not include 'helm' (no match config)")
	}
}

func TestConfig_ToMatchConfigMap_WithExclude(t *testing.T) {
	config := &Config{
		Version: 1,
		Integrations: []IntegrationConfig{
			{
				ID:      "npm",
				Enabled: true,
				Match: &MatchConfig{
					Files:   []string{"package.json", "apps/*/package.json", "libs/*/package.json"},
					Exclude: []string{"libs/*/package.json", "node_modules/**/package.json"},
				},
			},
			{
				ID:      "terraform",
				Enabled: true,
				Match: &MatchConfig{
					Files:   []string{"*.tf", "modules/**/*.tf"},
					Exclude: []string{".terraform/**/*.tf"},
				},
			},
		},
	}

	matchMap := config.ToMatchConfigMap()

	if len(matchMap) != 2 {
		t.Errorf("ToMatchConfigMap() returned %d items, want 2", len(matchMap))
	}

	npmConfig, ok := matchMap["npm"]
	if !ok {
		t.Fatal("ToMatchConfigMap() missing 'npm' key")
	}
	if len(npmConfig.Files) != 3 {
		t.Errorf("npm files count = %d, want 3", len(npmConfig.Files))
	}
	if len(npmConfig.Exclude) != 2 {
		t.Errorf("npm exclude count = %d, want 2", len(npmConfig.Exclude))
	}

	tfConfig, ok := matchMap["terraform"]
	if !ok {
		t.Fatal("ToMatchConfigMap() missing 'terraform' key")
	}
	if len(tfConfig.Files) != 2 {
		t.Errorf("terraform files count = %d, want 2", len(tfConfig.Files))
	}
	if len(tfConfig.Exclude) != 1 {
		t.Errorf("terraform exclude count = %d, want 1", len(tfConfig.Exclude))
	}
}

func TestLoadConfig_NonExistentFile(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/uptool.yaml")
	if err == nil {
		t.Error("LoadConfig() should error for non-existent file")
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "uptool.yaml")

	// Write invalid YAML
	invalidContent := `version: 1
integrations:
  - id: test
    enabled: true
    policy:
      update: !!!invalid yaml
`
	if err := os.WriteFile(configPath, []byte(invalidContent), 0o644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("LoadConfig() should error for invalid YAML")
	}
}

func TestLoadConfig_InvalidVersion(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "uptool.yaml")

	// Write config with invalid version
	invalidContent := `version: 999
integrations:
  - id: test
    enabled: true
`
	if err := os.WriteFile(configPath, []byte(invalidContent), 0o644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("LoadConfig() should error for invalid version")
	}
}

func TestValidate_EmptyIntegrationID(t *testing.T) {
	config := &Config{
		Version: 1,
		Integrations: []IntegrationConfig{
			{
				ID:      "", // Empty ID
				Enabled: true,
			},
		},
	}

	err := config.Validate()
	if err == nil {
		t.Error("Validate() should error for empty integration ID")
	}
}

func TestValidate_DuplicateIntegrationID(t *testing.T) {
	config := &Config{
		Version: 1,
		Integrations: []IntegrationConfig{
			{ID: "npm", Enabled: true},
			{ID: "npm", Enabled: true}, // Duplicate
		},
	}

	err := config.Validate()
	if err == nil {
		t.Error("Validate() should error for duplicate integration ID")
	}
}

func TestValidate_InvalidPolicy(t *testing.T) {
	config := &Config{
		Version: 1,
		Integrations: []IntegrationConfig{
			{
				ID:      "npm",
				Enabled: true,
				Policy: engine.IntegrationPolicy{
					Update: "invalid-strategy",
				},
			},
		},
	}

	err := config.Validate()
	if err == nil {
		t.Error("Validate() should error for invalid policy update strategy")
	}
}

func TestValidateIntegrationPolicy_AllValidUpdateStrategies(t *testing.T) {
	// Note: empty string is NOT a valid update strategy
	validStrategies := []string{"none", "patch", "minor", "major"}

	for _, strategy := range validStrategies {
		policy := &engine.IntegrationPolicy{Update: strategy}
		if err := ValidateIntegrationPolicy(policy); err != nil {
			t.Errorf("ValidateIntegrationPolicy() error for valid strategy %q: %v", strategy, err)
		}
	}
}

func TestValidateIntegrationPolicy_AllValidCadences(t *testing.T) {
	// Empty string is a valid cadence (means no cadence restriction)
	validCadences := []string{"", "daily", "weekly", "monthly"}

	for _, cadence := range validCadences {
		// Need a valid update strategy to test cadence
		policy := &engine.IntegrationPolicy{Update: "minor", Cadence: cadence}
		if err := ValidateIntegrationPolicy(policy); err != nil {
			t.Errorf("ValidateIntegrationPolicy() error for valid cadence %q: %v", cadence, err)
		}
	}
}

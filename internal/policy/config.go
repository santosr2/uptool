// Package policy handles configuration file parsing and policy management.
// It defines the structure for uptool.yaml configuration files, including integration-specific
// policies, update strategies, and organization-level governance settings.
package policy

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/santosr2/uptool/internal/engine"
	"github.com/santosr2/uptool/internal/secureio"
)

// Config represents the uptool.yaml configuration file.
type Config struct {
	Version      int                 `yaml:"version"`
	Integrations []IntegrationConfig `yaml:"integrations"`
	OrgPolicy    *OrgPolicy          `yaml:"org_policy,omitempty"`
}

// IntegrationConfig defines configuration for a specific integration.
type IntegrationConfig struct {
	ID      string                   `yaml:"id"`
	Enabled bool                     `yaml:"enabled"`
	Match   *MatchConfig             `yaml:"match,omitempty"`
	Policy  engine.IntegrationPolicy `yaml:"policy"`
}

// MatchConfig specifies file patterns for integration detection.
type MatchConfig struct {
	Files []string `yaml:"files"`
}

// OrgPolicy contains organization-level policies and governance settings.
type OrgPolicy struct {
	RequireSignoffFrom []string         `yaml:"require_signoff_from,omitempty"`
	Signing            *SigningConfig   `yaml:"signing,omitempty"`
	AutoMerge          *AutoMergeConfig `yaml:"auto_merge,omitempty"`
}

// SigningConfig controls artifact signing verification.
type SigningConfig struct {
	CosignVerify bool `yaml:"cosign_verify"`
}

// AutoMergeConfig controls automatic PR merging.
type AutoMergeConfig struct {
	Enabled bool     `yaml:"enabled"`
	Guards  []string `yaml:"guards"`
}

// LoadConfig reads and parses the configuration file.
func LoadConfig(path string) (*Config, error) {
	data, err := secureio.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

// Validate checks that the configuration is valid.
func (c *Config) Validate() error {
	if c.Version != 1 {
		return fmt.Errorf("unsupported version: %d (expected 1)", c.Version)
	}

	seenIDs := make(map[string]bool)
	for i, integ := range c.Integrations {
		if integ.ID == "" {
			return fmt.Errorf("integration[%d]: id is required", i)
		}
		if seenIDs[integ.ID] {
			return fmt.Errorf("integration[%d]: duplicate id %q", i, integ.ID)
		}
		seenIDs[integ.ID] = true

		if err := ValidateIntegrationPolicy(&integ.Policy); err != nil {
			return fmt.Errorf("integration[%d] (%s): %w", i, integ.ID, err)
		}
	}

	return nil
}

// ValidateIntegrationPolicy checks that an integration policy is valid.
func ValidateIntegrationPolicy(p *engine.IntegrationPolicy) error {
	validUpdates := map[string]bool{
		"none":  true,
		"patch": true,
		"minor": true,
		"major": true,
	}
	if !validUpdates[p.Update] {
		return fmt.Errorf("invalid update strategy %q (must be: none, patch, minor, major)", p.Update)
	}

	validCadences := map[string]bool{
		"":        true,
		"daily":   true,
		"weekly":  true,
		"monthly": true,
	}
	if !validCadences[p.Cadence] {
		return fmt.Errorf("invalid cadence %q (must be: daily, weekly, monthly)", p.Cadence)
	}

	return nil
}

// ToPolicyMap converts the configuration into a map of integration policies.
func (c *Config) ToPolicyMap() map[string]engine.IntegrationPolicy {
	result := make(map[string]engine.IntegrationPolicy)
	for _, integ := range c.Integrations {
		result[integ.ID] = integ.Policy
	}
	return result
}

// EnabledIntegrations returns the IDs of all enabled integrations.
func (c *Config) EnabledIntegrations() []string {
	result := make([]string, 0, len(c.Integrations))
	for _, integ := range c.Integrations {
		if integ.Enabled {
			result = append(result, integ.ID)
		}
	}
	return result
}

// DefaultConfig returns a default configuration with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Version: 1,
		Integrations: []IntegrationConfig{
			{
				ID:      "precommit",
				Enabled: true,
				Policy: engine.IntegrationPolicy{
					Enabled:         true,
					Update:          "minor",
					AllowPrerelease: false,
					Pin:             true,
					Cadence:         "weekly",
				},
			},
			{
				ID:      "tflint",
				Enabled: true,
				Policy: engine.IntegrationPolicy{
					Enabled:         true,
					Update:          "minor",
					AllowPrerelease: false,
					Pin:             true,
				},
			},
			{
				ID:      "terraform",
				Enabled: true,
				Policy: engine.IntegrationPolicy{
					Enabled:         true,
					Update:          "minor",
					AllowPrerelease: false,
					Pin:             true,
				},
			},
			{
				ID:      "asdf",
				Enabled: true,
				Policy: engine.IntegrationPolicy{
					Enabled:         true,
					Update:          "patch",
					AllowPrerelease: false,
					Pin:             true,
				},
			},
			{
				ID:      "mise",
				Enabled: true,
				Policy: engine.IntegrationPolicy{
					Enabled:         true,
					Update:          "patch",
					AllowPrerelease: false,
					Pin:             true,
				},
			},
			{
				ID:      "helm",
				Enabled: true,
				Policy: engine.IntegrationPolicy{
					Enabled:         true,
					Update:          "minor",
					AllowPrerelease: false,
					Pin:             true,
				},
			},
		},
	}
}

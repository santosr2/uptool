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
	OrgPolicy    *OrgPolicy          `yaml:"org_policy,omitempty"`
	Integrations []IntegrationConfig `yaml:"integrations"`
	Version      int                 `yaml:"version"`
}

// IntegrationConfig defines configuration for a specific integration.
type IntegrationConfig struct {
	Match   *MatchConfig             `yaml:"match,omitempty"`
	ID      string                   `yaml:"id"`
	Policy  engine.IntegrationPolicy `yaml:"policy"`
	Enabled bool                     `yaml:"enabled"`
}

// MatchConfig specifies file patterns for integration detection.
type MatchConfig struct {
	Files []string `yaml:"files"`
}

// OrgPolicy contains organization-level policies and governance settings.
type OrgPolicy struct {
	Signing            *SigningConfig   `yaml:"signing,omitempty"`
	AutoMerge          *AutoMergeConfig `yaml:"auto_merge,omitempty"`
	RequireSignoffFrom []string         `yaml:"require_signoff_from,omitempty"`
}

// SigningConfig controls artifact signing verification.
type SigningConfig struct {
	CosignVerify bool `yaml:"cosign_verify"`
}

// AutoMergeConfig controls automatic PR merging.
type AutoMergeConfig struct {
	Guards  []string `yaml:"guards"`
	Enabled bool     `yaml:"enabled"`
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

// ToMatchConfigMap converts the configuration into a map of file patterns per integration.
func (c *Config) ToMatchConfigMap() map[string][]string {
	result := make(map[string][]string)
	for _, integ := range c.Integrations {
		if integ.Match != nil && len(integ.Match.Files) > 0 {
			result[integ.ID] = integ.Match.Files
		}
	}
	return result
}

// GetOrgPolicy returns the organization-level policy settings if configured.
func (c *Config) GetOrgPolicy() *OrgPolicy {
	return c.OrgPolicy
}

// RequiresSignoff returns whether the organization policy requires signoff for changes.
func (c *Config) RequiresSignoff() bool {
	return c.OrgPolicy != nil && len(c.OrgPolicy.RequireSignoffFrom) > 0
}

// RequiresCosignVerification returns whether cosign verification is required.
func (c *Config) RequiresCosignVerification() bool {
	return c.OrgPolicy != nil && c.OrgPolicy.Signing != nil && c.OrgPolicy.Signing.CosignVerify
}

// IsAutoMergeEnabled returns whether auto-merge is enabled.
func (c *Config) IsAutoMergeEnabled() bool {
	return c.OrgPolicy != nil && c.OrgPolicy.AutoMerge != nil && c.OrgPolicy.AutoMerge.Enabled
}

// GetAutoMergeGuards returns the list of required guards for auto-merge.
func (c *Config) GetAutoMergeGuards() []string {
	if c.OrgPolicy == nil || c.OrgPolicy.AutoMerge == nil {
		return nil
	}
	return c.OrgPolicy.AutoMerge.Guards
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

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
//
// # Overview
//
// This package defines the structure for uptool.yaml configuration files, which control:
//   - Integration-specific policies (update strategies, version pinning, cadence)
//   - Organization-level governance (signoffs, signing, auto-merge guards)
//   - File matching patterns for integration detection
//
// # Configuration Structure
//
// uptool.yaml has two main sections:
//
//  1. integrations[] - Per-integration update policies
//  2. org_policy - Organization-level governance policies
//
// # Policy Precedence
//
// Update policies follow this precedence order (highest to lowest):
//  1. CLI flags (--update-level, --allow-prerelease)
//  2. uptool.yaml integration policy (integrations[*].policy)
//  3. Manifest constraints (^, ~, >=, etc.)
//  4. Default behavior
//
// # Example Configuration
//
//	version: 1
//
//	integrations:
//	  - id: npm
//	    enabled: true
//	    policy:
//	      update: minor          # Allow patch + minor updates
//	      allow_prerelease: false
//	      pin: false             # Preserve version ranges
//	      cadence: weekly
//
//	org_policy:
//	  require_signoff_from:
//	    - "@security-team"
//	  auto_merge:
//	    enabled: true
//	    guards:
//	      - "ci-green"
//	      - "codeowners-approve"
//
// See docs/configuration.md and docs/policy.md for comprehensive documentation.
package policy

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/santosr2/uptool/internal/engine"
	"github.com/santosr2/uptool/internal/secureio"
)

// Config represents the complete uptool.yaml configuration file.
//
// The configuration file controls both integration-specific update policies and
// organization-level governance settings.
//
// Example:
//
//	version: 1
//	integrations:
//	  - id: npm
//	    enabled: true
//	    policy:
//	      update: minor
//	org_policy:
//	  auto_merge:
//	    enabled: true
//	    guards: ["ci-green"]
type Config struct {
	// OrgPolicy contains organization-level governance policies (signoffs, signing, auto-merge).
	// This field is optional - if omitted, no org-level policies are enforced.
	OrgPolicy *OrgPolicy `yaml:"org_policy,omitempty"`

	// Integrations contains per-integration configuration (update policies, file patterns).
	// Each integration can be individually enabled/disabled and configured with its own policy.
	Integrations []IntegrationConfig `yaml:"integrations"`

	// Version specifies the configuration format version.
	// Currently only version 1 is supported. This field is required.
	Version int `yaml:"version"`
}

// IntegrationConfig defines configuration for a specific integration (npm, helm, terraform, etc.).
//
// Each integration can be independently configured with custom update policies,
// file matching patterns, and enable/disable state.
//
// Example:
//
//   - id: npm
//     enabled: true
//     match:
//     files: ["package.json", "apps/*/package.json"]
//     policy:
//     update: minor
//     allow_prerelease: false
//     pin: false
//     cadence: weekly
type IntegrationConfig struct {
	// Match specifies custom file patterns for this integration.
	// If omitted, the integration uses its default file patterns.
	Match *MatchConfig `yaml:"match,omitempty"`

	// ID is the integration identifier (e.g., "npm", "helm", "terraform").
	// Must match one of the registered integration names. Required.
	ID string `yaml:"id"`

	// Policy contains update policy settings for this integration.
	// Controls which updates are allowed (patch/minor/major), version pinning, etc.
	Policy engine.IntegrationPolicy `yaml:"policy"`

	// Enabled controls whether this integration runs during scan/plan/update.
	// Default: true. Can be overridden by CLI flags (--only, --exclude).
	Enabled bool `yaml:"enabled"`
}

// MatchConfig specifies file glob patterns for integration detection.
//
// Use this to customize which files an integration should process, particularly
// useful for monorepos or non-standard project structures.
//
// Example:
//
//	match:
//	  files:
//	    - "package.json"           # Root package
//	    - "apps/*/package.json"    # App packages
//	    - "packages/*/package.json" # Library packages
//	  exclude:
//	    - "node_modules/**/package.json"  # Ignore dependencies
//	    - "dist/**/package.json"          # Ignore build artifacts
type MatchConfig struct {
	// Files is a list of glob patterns matching manifest files.
	// Patterns support standard glob syntax: *, **, ?, [abc], {a,b,c}.
	Files []string `yaml:"files"`

	// Exclude is a list of glob patterns to exclude from matches.
	// Files matching any exclude pattern are filtered out even if they match a files pattern.
	// Patterns support standard glob syntax: *, **, ?, [abc], {a,b,c}.
	//
	// Exclude patterns are applied AFTER files patterns, providing fine-grained control.
	//
	// Common use cases:
	//   - Exclude vendor directories: "vendor/**"
	//   - Exclude build artifacts: "dist/**", "build/**"
	//   - Exclude test fixtures: "testdata/**", "fixtures/**"
	//   - Exclude specific paths: "legacy/old-app/**"
	Exclude []string `yaml:"exclude,omitempty"`
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

	// Validate schedule if present
	if p.Schedule != nil {
		if err := validateSchedule(p.Schedule); err != nil {
			return fmt.Errorf("invalid schedule: %w", err)
		}
	}

	// Validate versioning strategy if present
	if p.VersioningStrategy != "" {
		if err := validateVersioningStrategy(p.VersioningStrategy); err != nil {
			return err
		}
	}

	// Validate open pull requests limit
	if p.OpenPullRequestsLimit < 0 || p.OpenPullRequestsLimit > 10 {
		if p.OpenPullRequestsLimit != 0 {
			return fmt.Errorf("open_pull_requests_limit must be between 0 and 10")
		}
	}

	// Validate groups
	for name, group := range p.Groups {
		if err := validateDependencyGroup(name, group); err != nil {
			return err
		}
	}

	// Validate cooldown
	if p.Cooldown != nil {
		if err := validateCooldown(p.Cooldown); err != nil {
			return fmt.Errorf("invalid cooldown: %w", err)
		}
	}

	// Validate commit message
	if p.CommitMessage != nil {
		if err := validateCommitMessage(p.CommitMessage); err != nil {
			return fmt.Errorf("invalid commit_message: %w", err)
		}
	}

	return nil
}

// validateSchedule validates a Schedule configuration.
func validateSchedule(s *engine.Schedule) error {
	validIntervals := map[string]bool{
		"daily": true, "weekly": true, "monthly": true,
		"quarterly": true, "semiannually": true, "yearly": true, "cron": true,
	}
	if s.Interval != "" && !validIntervals[s.Interval] {
		return fmt.Errorf("invalid interval %q (must be: daily, weekly, monthly, quarterly, semiannually, yearly, cron)", s.Interval)
	}

	if s.Interval == "weekly" && s.Day != "" {
		validDays := map[string]bool{
			"monday": true, "tuesday": true, "wednesday": true, "thursday": true,
			"friday": true, "saturday": true, "sunday": true,
		}
		if !validDays[s.Day] {
			return fmt.Errorf("invalid day %q (must be: monday-sunday)", s.Day)
		}
	}

	if s.Interval == "cron" && s.Cron == "" {
		return fmt.Errorf("cron expression is required when interval is 'cron'")
	}

	return nil
}

// validateVersioningStrategy validates a versioning strategy.
func validateVersioningStrategy(strategy string) error {
	valid := map[string]bool{
		"auto": true, "increase": true, "increase-if-necessary": true,
		"lockfile-only": true, "widen": true,
	}
	if !valid[strategy] {
		return fmt.Errorf("invalid versioning_strategy %q (must be: auto, increase, increase-if-necessary, lockfile-only, widen)", strategy)
	}
	return nil
}

// validateDependencyGroup validates a DependencyGroup configuration.
func validateDependencyGroup(name string, g *engine.DependencyGroup) error {
	if g == nil {
		return nil
	}

	if g.AppliesTo != "" {
		validAppliesTo := map[string]bool{
			"version-updates": true, "security-updates": true,
		}
		if !validAppliesTo[g.AppliesTo] {
			return fmt.Errorf("group %q: invalid applies_to %q (must be: version-updates, security-updates)", name, g.AppliesTo)
		}
	}

	if g.DependencyType != "" {
		validTypes := map[string]bool{
			"production": true, "development": true,
		}
		if !validTypes[g.DependencyType] {
			return fmt.Errorf("group %q: invalid dependency_type %q (must be: production, development)", name, g.DependencyType)
		}
	}

	for _, ut := range g.UpdateTypes {
		validUpdateTypes := map[string]bool{
			"major": true, "minor": true, "patch": true,
		}
		if !validUpdateTypes[ut] {
			return fmt.Errorf("group %q: invalid update_type %q (must be: major, minor, patch)", name, ut)
		}
	}

	return nil
}

// validateCooldown validates a CooldownConfig.
func validateCooldown(c *engine.CooldownConfig) error {
	if c.DefaultDays < 0 {
		return fmt.Errorf("default_days cannot be negative")
	}
	if c.SemverMajorDays < 0 {
		return fmt.Errorf("semver_major_days cannot be negative")
	}
	if c.SemverMinorDays < 0 {
		return fmt.Errorf("semver_minor_days cannot be negative")
	}
	if c.SemverPatchDays < 0 {
		return fmt.Errorf("semver_patch_days cannot be negative")
	}
	return nil
}

// validateCommitMessage validates a CommitMessageConfig.
func validateCommitMessage(c *engine.CommitMessageConfig) error {
	if len(c.Prefix) > 50 {
		return fmt.Errorf("prefix cannot exceed 50 characters")
	}
	if len(c.PrefixDevelopment) > 50 {
		return fmt.Errorf("prefix_development cannot exceed 50 characters")
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

// ToMatchConfigMap converts the configuration into a map of match configs per integration.
// Returns only integrations that have match configuration specified.
func (c *Config) ToMatchConfigMap() map[string]*engine.MatchConfig {
	result := make(map[string]*engine.MatchConfig)
	for _, integ := range c.Integrations {
		if integ.Match != nil && (len(integ.Match.Files) > 0 || len(integ.Match.Exclude) > 0) {
			result[integ.ID] = &engine.MatchConfig{
				Files:   integ.Match.Files,
				Exclude: integ.Match.Exclude,
			}
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
				ID:      "npm",
				Enabled: true,
				Policy: engine.IntegrationPolicy{
					Enabled:         true,
					Update:          "minor",
					AllowPrerelease: false,
					Pin:             false, // Preserve version ranges for npm
				},
			},
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

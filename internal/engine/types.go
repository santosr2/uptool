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

// Package engine provides the core orchestration layer for uptool's dependency scanning and updating.
// It defines the fundamental types and interfaces used across all integrations, including Manifest,
// Dependency, UpdatePlan, and the Integration interface.
package engine

import (
	"context"
	"time"
)

// PlanContext provides policy and configuration context for planning operations.
// It implements a precedence order: CLI flags > uptool.yaml policy > manifest constraints.
// This allows fine-grained control over which updates are allowed.
type PlanContext struct {
	// Policy contains the integration-specific policy from uptool.yaml.
	// This has medium precedence (after CLI flags) when determining allowed updates.
	Policy *IntegrationPolicy

	// CLIFlags contains any command-line overrides.
	// These have the highest precedence and override all other policy sources.
	CLIFlags *CLIFlags

	// RespectConstraints indicates whether manifest constraints (e.g., ~> 5.0)
	// should be respected when no policy or flag overrides them.
	// Defaults to true.
	RespectConstraints bool
}

// CLIFlags represents command-line flag overrides for update behavior.
type CLIFlags struct {
	AllowPrerelease *bool
	UpdateLevel     string
}

// NewPlanContext creates a new PlanContext with default settings.
// By default, constraints are respected when no policy overrides them.
func NewPlanContext() *PlanContext {
	return &PlanContext{
		RespectConstraints: true,
	}
}

// WithPolicy returns a copy of the context with the given policy.
func (pc *PlanContext) WithPolicy(p *IntegrationPolicy) *PlanContext {
	if pc == nil {
		pc = NewPlanContext()
	}
	newCtx := *pc
	newCtx.Policy = p
	return &newCtx
}

// WithCLIFlags returns a copy of the context with the given CLI flags.
func (pc *PlanContext) WithCLIFlags(flags *CLIFlags) *PlanContext {
	if pc == nil {
		pc = NewPlanContext()
	}
	newCtx := *pc
	newCtx.CLIFlags = flags
	return &newCtx
}

// EffectiveUpdateLevel returns the update level to use, following precedence:
// 1. uptool.yaml policy (highest)
// 2. CLI flags
// 3. Default ("major" - allow all updates, let constraints filter)
func (pc *PlanContext) EffectiveUpdateLevel() string {
	if pc == nil {
		return "major"
	}

	// Highest precedence: CLI flags
	if pc.CLIFlags != nil && pc.CLIFlags.UpdateLevel != "" {
		return pc.CLIFlags.UpdateLevel
	}

	// Second precedence: uptool.yaml policy
	if pc.Policy != nil && pc.Policy.Update != "" {
		return pc.Policy.Update
	}

	// Default: allow all update levels (let constraints filter)
	return "major"
}

// EffectiveAllowPrerelease returns whether prereleases are allowed, following precedence:
// 1. CLI flags (highest)
// 2. uptool.yaml policy
// 3. Default (false)
func (pc *PlanContext) EffectiveAllowPrerelease() bool {
	if pc == nil {
		return false
	}

	// Highest precedence: CLI flags
	if pc.CLIFlags != nil && pc.CLIFlags.AllowPrerelease != nil {
		return *pc.CLIFlags.AllowPrerelease
	}

	// Second precedence: uptool.yaml policy
	if pc.Policy != nil {
		return pc.Policy.AllowPrerelease
	}

	// Default: no prereleases
	return false
}

// ShouldRespectConstraints returns whether manifest constraints should be respected.
// Constraints are always respected unless explicitly disabled.
func (pc *PlanContext) ShouldRespectConstraints() bool {
	if pc == nil {
		return true
	}
	return pc.RespectConstraints
}

// GetPolicySource determines the source of the effective policy based on precedence.
// Returns the policy source following the precedence order:
// 1. CLI flags (highest) - overrides all other sources
// 2. uptool.yaml policy - overrides constraints
// 3. Manifest constraints (when no higher precedence policy exists)
// 4. Default
func (pc *PlanContext) GetPolicySource() PolicySource {
	if pc == nil {
		return PolicySourceDefault
	}

	// Highest precedence: CLI flags (overrides all)
	if pc.CLIFlags != nil && pc.CLIFlags.UpdateLevel != "" {
		return PolicySourceCLIFlag
	}

	// Second precedence: uptool.yaml policy (overrides constraints)
	if pc.Policy != nil && pc.Policy.Update != "" {
		return PolicySourceUptoolYAML
	}

	// Third precedence: Manifest constraints (only when no policy/flags override)
	// Constraints are respected by default when no higher precedence policy exists
	return PolicySourceConstraint
}

// Manifest represents a dependency manifest file.
type Manifest struct {
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Path         string                 `json:"path"`
	Type         string                 `json:"type"`
	Dependencies []Dependency           `json:"dependencies"`
	Content      []byte                 `json:"-"`
}

// Dependency represents a single dependency in a manifest.
type Dependency struct {
	Name           string `json:"name"`
	CurrentVersion string `json:"current_version"`
	Constraint     string `json:"constraint,omitempty"`
	Type           string `json:"type"` // direct, dev, peer, optional
	Registry       string `json:"registry,omitempty"`
}

// IntegrationPolicy contains policy settings that apply to a specific integration.
type IntegrationPolicy struct {
	Custom          map[string]interface{} `yaml:",inline" json:"custom,omitempty"`
	Update          string                 `yaml:"update" json:"update"`
	Cadence         string                 `yaml:"cadence,omitempty" json:"cadence,omitempty"`
	Enabled         bool                   `yaml:"enabled" json:"enabled"`
	AllowPrerelease bool                   `yaml:"allow_prerelease" json:"allow_prerelease"`
	Pin             bool                   `yaml:"pin" json:"pin"`
}

// Impact describes the severity of an update.
type Impact string

// Impact levels for update severity
const (
	ImpactNone  Impact = "none"
	ImpactPatch Impact = "patch"
	ImpactMinor Impact = "minor"
	ImpactMajor Impact = "major"
)

// PolicySource indicates where the update policy originated from.
type PolicySource string

// PolicySource values for policy precedence tracking
const (
	// PolicySourceUptoolYAML indicates the policy came from uptool.yaml (highest precedence)
	PolicySourceUptoolYAML PolicySource = "uptool.yaml"

	// PolicySourceCLIFlag indicates the policy came from a CLI flag
	PolicySourceCLIFlag PolicySource = "cli-flag"

	// PolicySourceConstraint indicates the policy came from manifest constraints (e.g., ~> 5.0)
	PolicySourceConstraint PolicySource = "constraint"

	// PolicySourceDefault indicates the default policy was used
	PolicySourceDefault PolicySource = "default"
)

// UpdatePlan describes planned updates for a manifest.
type UpdatePlan struct {
	Manifest *Manifest `json:"manifest"`
	Strategy string    `json:"strategy"`
	Updates  []Update  `json:"updates"`
}

// Update represents a planned update for a dependency.
type Update struct {
	Dependency    Dependency   `json:"dependency"`
	TargetVersion string       `json:"target_version"`
	Impact        string       `json:"impact"` // patch, minor, major
	ChangelogURL  string       `json:"changelog_url,omitempty"`
	Breaking      bool         `json:"breaking"`
	PolicySource  PolicySource `json:"policy_source,omitempty"` // where the policy originated from
}

// ApplyResult contains the outcome of applying updates.
type ApplyResult struct {
	Manifest     *Manifest `json:"manifest"`
	ManifestDiff string    `json:"manifest_diff,omitempty"`
	LockfileDiff string    `json:"lockfile_diff,omitempty"`
	Errors       []string  `json:"errors,omitempty"`
	Applied      int       `json:"applied"`
	Failed       int       `json:"failed"`
}

// Integration defines the interface for ecosystem integrations.
type Integration interface {
	// Name returns the integration identifier
	Name() string

	// Detect finds manifest files for this integration
	Detect(ctx context.Context, repoRoot string) ([]*Manifest, error)

	// Plan determines available updates for a manifest.
	// The planCtx parameter provides policy configuration following the precedence order:
	// uptool.yaml policy > CLI flags > manifest constraints.
	// If planCtx is nil, the integration should use default behavior (respect constraints only).
	Plan(ctx context.Context, manifest *Manifest, planCtx *PlanContext) (*UpdatePlan, error)

	// Apply executes the update plan
	Apply(ctx context.Context, plan *UpdatePlan) (*ApplyResult, error)

	// Validate checks if changes are valid (optional)
	Validate(ctx context.Context, manifest *Manifest) error
}

// ScanResult aggregates all discovered manifests.
type ScanResult struct {
	Manifests []*Manifest `json:"manifests"`
	Timestamp time.Time   `json:"timestamp"`
	RepoRoot  string      `json:"repo_root"`
	Errors    []string    `json:"errors,omitempty"`
}

// PlanResult aggregates all update plans.
type PlanResult struct {
	Plans     []*UpdatePlan `json:"plans"`
	Timestamp time.Time     `json:"timestamp"`
	Errors    []string      `json:"errors,omitempty"`
}

// UpdateResult aggregates all apply results.
type UpdateResult struct {
	Results   []*ApplyResult `json:"results"`
	Timestamp time.Time      `json:"timestamp"`
	Errors    []string       `json:"errors,omitempty"`
}

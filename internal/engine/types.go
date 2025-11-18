// Package engine provides the core orchestration layer for uptool's dependency scanning and updating.
// It defines the fundamental types and interfaces used across all integrations, including Manifest,
// Dependency, UpdatePlan, and the Integration interface.
package engine

import (
	"context"
	"time"
)

// Manifest represents a dependency manifest file.
type Manifest struct {
	Path         string                 `json:"path"`
	Type         string                 `json:"type"` // npm, pre-commit, terraform, etc.
	Dependencies []Dependency           `json:"dependencies"`
	Content      []byte                 `json:"-"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
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
	// Enabled controls whether this integration is active
	Enabled bool `yaml:"enabled" json:"enabled"`

	// Update specifies the maximum semver change allowed: none, patch, minor, or major
	Update string `yaml:"update" json:"update"`

	// AllowPrerelease allows selection of pre-release versions
	AllowPrerelease bool `yaml:"allow_prerelease" json:"allow_prerelease"`

	// Pin controls whether to write exact versions or ranges
	Pin bool `yaml:"pin" json:"pin"`

	// Cadence suggests update frequency (daily, weekly, monthly)
	Cadence string `yaml:"cadence,omitempty" json:"cadence,omitempty"`

	// Custom contains integration-specific policy fields
	Custom map[string]interface{} `yaml:",inline" json:"custom,omitempty"`
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

// UpdatePlan describes planned updates for a manifest.
type UpdatePlan struct {
	Manifest *Manifest `json:"manifest"`
	Updates  []Update  `json:"updates"`
	Strategy string    `json:"strategy"` // native_command or custom_rewrite
}

// Update represents a planned update for a dependency.
type Update struct {
	Dependency    Dependency `json:"dependency"`
	TargetVersion string     `json:"target_version"`
	Impact        string     `json:"impact"` // patch, minor, major
	ChangelogURL  string     `json:"changelog_url,omitempty"`
	Breaking      bool       `json:"breaking"`
}

// ApplyResult contains the outcome of applying updates.
type ApplyResult struct {
	Manifest     *Manifest `json:"manifest"`
	Applied      int       `json:"applied"`
	Failed       int       `json:"failed"`
	ManifestDiff string    `json:"manifest_diff,omitempty"`
	LockfileDiff string    `json:"lockfile_diff,omitempty"`
	Errors       []string  `json:"errors,omitempty"`
}

// Integration defines the interface for ecosystem integrations.
type Integration interface {
	// Name returns the integration identifier
	Name() string

	// Detect finds manifest files for this integration
	Detect(ctx context.Context, repoRoot string) ([]*Manifest, error)

	// Plan determines available updates for a manifest
	Plan(ctx context.Context, manifest *Manifest) (*UpdatePlan, error)

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

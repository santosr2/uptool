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

// Package dependabot provides Dependabot configuration parsing and migration support.
// It enables users to migrate from Dependabot to uptool by reading existing dependabot.yml
// files and converting them to uptool.yaml format.
//
// # Supported Features
//
// This package supports parsing all major Dependabot configuration options:
//   - package-ecosystem: Maps to uptool integrations
//   - directory/directories: Maps to match.files patterns
//   - schedule: Maps to cadence with extended cron support
//   - groups: Dependency grouping for combined PRs
//   - allow/ignore: Dependency filtering patterns
//   - versioning-strategy: Maps to update policy
//   - commit-message: Commit message customization
//   - labels, assignees, reviewers: PR metadata
//   - open-pull-requests-limit: Concurrent PR limits
//   - cooldown: Delayed update support
//
// # Migration Example
//
//	// Load existing dependabot.yml
//	depConfig, err := dependabot.LoadConfig(".github/dependabot.yml")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Convert to uptool.yaml format
//	uptoolConfig := depConfig.ToUptoolConfig()
//
//	// Save as uptool.yaml
//	data, _ := yaml.Marshal(uptoolConfig)
//	os.WriteFile("uptool.yaml", data, 0644)
package dependabot

import (
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/santosr2/uptool/internal/secureio"
)

// Ecosystem constants for consistent string usage.
const (
	ecosystemGitHubActions = "github-actions"
	intervalWeekly         = "weekly"
)

// Config represents a complete dependabot.yml configuration file.
// Reference: https://docs.github.com/en/code-security/dependabot/dependabot-version-updates/configuration-options-for-the-dependabot.yml-file
type Config struct {
	Registries           map[string]Registry            `yaml:"registries,omitempty"`
	MultiEcosystemGroups map[string]MultiEcosystemGroup `yaml:"multi-ecosystem-groups,omitempty"`
	Updates              []UpdateConfig                 `yaml:"updates"`
	Version              int                            `yaml:"version"`
	EnableBetaEcosystems bool                           `yaml:"enable-beta-ecosystems,omitempty"`
}

// UpdateConfig defines configuration for a specific package ecosystem.
type UpdateConfig struct {
	Registries                    interface{}      `yaml:"registries,omitempty"`
	CommitMessage                 *CommitMessage   `yaml:"commit-message,omitempty"`
	Cooldown                      *Cooldown        `yaml:"cooldown,omitempty"`
	Groups                        map[string]Group `yaml:"groups,omitempty"`
	PullRequestBranchName         *BranchName      `yaml:"pull-request-branch-name,omitempty"`
	Schedule                      Schedule         `yaml:"schedule"`
	MultiEcosystemGroup           string           `yaml:"multi-ecosystem-group,omitempty"`
	PackageEcosystem              string           `yaml:"package-ecosystem"`
	VersioningStrategy            string           `yaml:"versioning-strategy,omitempty"`
	InsecureExternalCodeExecution string           `yaml:"insecure-external-code-execution,omitempty"`
	TargetBranch                  string           `yaml:"target-branch,omitempty"`
	Directory                     string           `yaml:"directory,omitempty"`
	RebaseStrategy                string           `yaml:"rebase-strategy,omitempty"`
	Ignore                        []IgnoreRule     `yaml:"ignore,omitempty"`
	Reviewers                     []string         `yaml:"reviewers,omitempty"`
	Assignees                     []string         `yaml:"assignees,omitempty"`
	Labels                        []string         `yaml:"labels,omitempty"`
	Allow                         []AllowRule      `yaml:"allow,omitempty"`
	Directories                   []string         `yaml:"directories,omitempty"`
	ExcludePaths                  []string         `yaml:"exclude-paths,omitempty"`
	Milestone                     int              `yaml:"milestone,omitempty"`
	OpenPullRequestsLimit         int              `yaml:"open-pull-requests-limit,omitempty"`
	Vendor                        bool             `yaml:"vendor,omitempty"`
}

// Schedule defines when Dependabot checks for updates.
type Schedule struct {
	// Interval is the update frequency (required)
	// Valid values: daily, weekly, monthly, quarterly, semiannually, yearly, cron
	Interval string `yaml:"interval"`

	// Day specifies the day for weekly updates
	// Valid values: monday, tuesday, wednesday, thursday, friday, saturday, sunday
	Day string `yaml:"day,omitempty"`

	// Time specifies the time for updates in HH:MM format (UTC default)
	Time string `yaml:"time,omitempty"`

	// Timezone is the IANA timezone for schedule (default: UTC)
	Timezone string `yaml:"timezone,omitempty"`

	// Cronjob is a cron expression for "cron" interval type
	Cronjob string `yaml:"cronjob,omitempty"`
}

// AllowRule specifies which dependencies to include.
type AllowRule struct {
	// DependencyName matches dependencies by name (supports * wildcard)
	DependencyName string `yaml:"dependency-name,omitempty"`

	// DependencyType filters by dependency type
	// Valid values: direct, indirect, all, production, development
	DependencyType string `yaml:"dependency-type,omitempty"`
}

// IgnoreRule specifies which dependencies or versions to exclude.
type IgnoreRule struct {
	// DependencyName matches dependencies by name (supports * wildcard)
	DependencyName string `yaml:"dependency-name,omitempty"`

	// Versions specifies version ranges to ignore (package manager syntax)
	Versions []string `yaml:"versions,omitempty"`

	// UpdateTypes specifies update types to ignore
	// Valid values: version-update:semver-major, version-update:semver-minor, version-update:semver-patch
	UpdateTypes []string `yaml:"update-types,omitempty"`
}

// Group defines a dependency grouping rule.
type Group struct {
	// AppliesTo specifies when this group applies
	// Valid values: version-updates, security-updates
	AppliesTo string `yaml:"applies-to,omitempty"`

	// DependencyType filters dependencies
	// Valid values: production, development
	DependencyType string `yaml:"dependency-type,omitempty"`

	// Patterns includes dependencies matching these patterns (supports * wildcard)
	Patterns []string `yaml:"patterns,omitempty"`

	// ExcludePatterns excludes dependencies matching these patterns
	ExcludePatterns []string `yaml:"exclude-patterns,omitempty"`

	// UpdateTypes limits to specific update types
	// Valid values: major, minor, patch
	UpdateTypes []string `yaml:"update-types,omitempty"`
}

// MultiEcosystemGroup defines a group spanning multiple ecosystems.
type MultiEcosystemGroup struct {
	// Schedule for this multi-ecosystem group
	Schedule Schedule `yaml:"schedule,omitempty"`
}

// CommitMessage customizes the commit message format.
type CommitMessage struct {
	// Prefix is prepended to commit messages (max 50 chars)
	Prefix string `yaml:"prefix,omitempty"`

	// PrefixDevelopment is used for dev dependency updates
	PrefixDevelopment string `yaml:"prefix-development,omitempty"`

	// Include adds scope to commit messages
	// Valid values: scope (adds deps or deps-dev)
	Include string `yaml:"include,omitempty"`
}

// BranchName customizes PR branch naming.
type BranchName struct {
	// Separator replaces "/" in branch names
	// Valid values: "-", "_", "/"
	Separator string `yaml:"separator,omitempty"`
}

// Cooldown defines delayed update settings.
type Cooldown struct {
	Include         []string `yaml:"include,omitempty"`
	Exclude         []string `yaml:"exclude,omitempty"`
	DefaultDays     int      `yaml:"default-days,omitempty"`
	SemverMajorDays int      `yaml:"semver-major-days,omitempty"`
	SemverMinorDays int      `yaml:"semver-minor-days,omitempty"`
	SemverPatchDays int      `yaml:"semver-patch-days,omitempty"`
}

// Registry defines a private package registry configuration.
type Registry struct {
	// Type is the registry type
	// Valid values: docker-registry, npm-registry, maven-repository, nuget-feed,
	// python-index, rubygems-server, hex-repository, git-repository
	Type string `yaml:"type"`

	// URL is the registry endpoint
	URL string `yaml:"url,omitempty"`

	// Username for authentication
	Username string `yaml:"username,omitempty"`

	// Password for authentication
	Password string `yaml:"password,omitempty"`

	// Token for authentication (alternative to username/password)
	Token string `yaml:"token,omitempty"`

	// Key for some registry types
	Key string `yaml:"key,omitempty"`

	// ReplacesBase indicates if this replaces the default registry
	ReplacesBase bool `yaml:"replaces-base,omitempty"`
}

// LoadConfig reads and parses a dependabot.yml file.
func LoadConfig(path string) (*Config, error) {
	// Convert to absolute path if relative
	absPath := path
	if !filepath.IsAbs(path) {
		var err error
		absPath, err = filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("resolve path: %w", err)
		}
	}

	data, err := secureio.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("read dependabot config: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parse dependabot config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid dependabot config: %w", err)
	}

	return &config, nil
}

// Validate checks that the configuration is valid.
func (c *Config) Validate() error {
	if c.Version != 2 {
		return fmt.Errorf("unsupported version: %d (expected 2)", c.Version)
	}

	if len(c.Updates) == 0 {
		return fmt.Errorf("at least one update configuration is required")
	}

	for i := range c.Updates {
		update := &c.Updates[i]
		if update.PackageEcosystem == "" {
			return fmt.Errorf("updates[%d]: package-ecosystem is required", i)
		}

		if update.Directory == "" && len(update.Directories) == 0 {
			return fmt.Errorf("updates[%d]: directory or directories is required", i)
		}

		if err := validateSchedule(&update.Schedule); err != nil {
			return fmt.Errorf("updates[%d]: %w", i, err)
		}

		if update.VersioningStrategy != "" {
			if !isValidVersioningStrategy(update.VersioningStrategy) {
				return fmt.Errorf("updates[%d]: invalid versioning-strategy: %s", i, update.VersioningStrategy)
			}
		}

		if update.OpenPullRequestsLimit < 0 || update.OpenPullRequestsLimit > 10 {
			if update.OpenPullRequestsLimit != 0 {
				return fmt.Errorf("updates[%d]: open-pull-requests-limit must be 0-10", i)
			}
		}
	}

	return nil
}

func validateSchedule(s *Schedule) error {
	validIntervals := map[string]bool{
		"daily": true, intervalWeekly: true, "monthly": true,
		"quarterly": true, "semiannually": true, "yearly": true, "cron": true,
	}
	if !validIntervals[s.Interval] {
		return fmt.Errorf("invalid schedule interval: %s", s.Interval)
	}

	if s.Interval == intervalWeekly && s.Day != "" {
		validDays := map[string]bool{
			"monday": true, "tuesday": true, "wednesday": true, "thursday": true,
			"friday": true, "saturday": true, "sunday": true,
		}
		if !validDays[strings.ToLower(s.Day)] {
			return fmt.Errorf("invalid schedule day: %s", s.Day)
		}
	}

	if s.Interval == "cron" && s.Cronjob == "" {
		return fmt.Errorf("cronjob is required when interval is 'cron'")
	}

	return nil
}

func isValidVersioningStrategy(s string) bool {
	valid := map[string]bool{
		"auto": true, "increase": true, "increase-if-necessary": true,
		"lockfile-only": true, "widen": true,
	}
	return valid[s]
}

// EcosystemToIntegration maps Dependabot ecosystem names to uptool integration IDs.
var EcosystemToIntegration = map[string]string{
	ecosystemGitHubActions: "actions",
	"gomod":                "gomod",
	"npm":                  "npm",
	"docker":               "docker",
	"docker-compose":       "docker",
	"helm":                 "helm",
	"terraform":            "terraform",
	"pip":                  "pip",
	"bundler":              "bundler",
	"cargo":                "cargo",
	"composer":             "composer",
	"maven":                "maven",
	"gradle":               "gradle",
	"nuget":                "nuget",
	"mix":                  "hex",
	"pub":                  "pub",
	"swift":                "swift",
	"devcontainers":        "devcontainers",
	"elm":                  "elm",
	"bun":                  "bun",
	"vcpkg":                "vcpkg",
	"uv":                   "uv",
}

// GetIntegrationID converts a Dependabot package-ecosystem to an uptool integration ID.
func GetIntegrationID(ecosystem string) string {
	if id, ok := EcosystemToIntegration[ecosystem]; ok {
		return id
	}
	return ecosystem
}

// GetDirectories returns all directories for this update config.
func (u *UpdateConfig) GetDirectories() []string {
	if len(u.Directories) > 0 {
		return u.Directories
	}
	if u.Directory != "" {
		return []string{u.Directory}
	}
	return nil
}

// GetFilePatterns converts directory paths to glob patterns for uptool.
func (u *UpdateConfig) GetFilePatterns() []string {
	dirs := u.GetDirectories()
	patterns := make([]string, 0, len(dirs))

	for _, dir := range dirs {
		// Clean the directory path
		dir = strings.TrimPrefix(dir, "/")
		if dir == "" {
			dir = "."
		}

		// Generate appropriate pattern based on ecosystem
		pattern := generateFilePattern(u.PackageEcosystem, dir)
		patterns = append(patterns, pattern)
	}

	return patterns
}

func generateFilePattern(ecosystem, dir string) string {
	if dir == "." {
		dir = ""
	} else {
		dir += "/"
	}

	switch ecosystem {
	case "npm", "yarn", "pnpm", "bun":
		return dir + "package.json"
	case "pip", "pipenv", "poetry", "pip-compile", "uv":
		return dir + "requirements*.txt"
	case "bundler":
		return dir + "Gemfile"
	case "cargo":
		return dir + "Cargo.toml"
	case "composer":
		return dir + "composer.json"
	case "maven":
		return dir + "pom.xml"
	case "gradle":
		return dir + "build.gradle*"
	case "nuget":
		return dir + "*.csproj"
	case "gomod":
		return dir + "go.mod"
	case "terraform":
		return dir + "*.tf"
	case "helm":
		return dir + "Chart.yaml"
	case "docker", "docker-compose":
		return dir + "Dockerfile*"
	case ecosystemGitHubActions:
		return ".github/workflows/*.yml"
	case "pub":
		return dir + "pubspec.yaml"
	case "swift":
		return dir + "Package.swift"
	case "mix":
		return dir + "mix.exs"
	case "elm":
		return dir + "elm.json"
	case "devcontainers":
		return ".devcontainer/devcontainer.json"
	default:
		return filepath.Join(dir, "*")
	}
}

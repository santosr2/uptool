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

package dependabot

import (
	"github.com/santosr2/uptool/internal/engine"
	"github.com/santosr2/uptool/internal/policy"
)

// Update policy constants for consistent string usage.
const (
	policyMinor   = "minor"
	policyMajor   = "major"
	policyNone    = "none"
	cadenceWeekly = "weekly"
)

// MigrateToUptool converts a Dependabot configuration to an uptool configuration.
// This enables seamless migration from Dependabot to uptool.
func (c *Config) MigrateToUptool() *policy.Config {
	uptoolConfig := &policy.Config{
		Version:      1,
		Integrations: make([]policy.IntegrationConfig, 0, len(c.Updates)),
	}

	for i := range c.Updates {
		integrationConfig := convertUpdateConfig(&c.Updates[i])
		uptoolConfig.Integrations = append(uptoolConfig.Integrations, integrationConfig)
	}

	return uptoolConfig
}

func convertUpdateConfig(update *UpdateConfig) policy.IntegrationConfig {
	integrationID := GetIntegrationID(update.PackageEcosystem)

	config := policy.IntegrationConfig{
		ID:      integrationID,
		Enabled: true,
		Policy:  convertPolicy(update),
	}

	// Convert file patterns
	patterns := update.GetFilePatterns()
	if len(patterns) > 0 || len(update.ExcludePaths) > 0 {
		config.Match = &policy.MatchConfig{
			Files:   patterns,
			Exclude: update.ExcludePaths,
		}
	}

	return config
}

func convertPolicy(update *UpdateConfig) engine.IntegrationPolicy {
	pol := engine.IntegrationPolicy{
		Enabled:               true,
		Update:                convertVersioningStrategyToUpdate(update.VersioningStrategy),
		Cadence:               convertScheduleToCadence(&update.Schedule),
		Schedule:              convertSchedule(&update.Schedule),
		Groups:                convertGroups(update.Groups),
		Allow:                 convertAllowRules(update.Allow),
		Ignore:                convertIgnoreRules(update.Ignore),
		Cooldown:              convertCooldown(update.Cooldown),
		CommitMessage:         convertCommitMessage(update.CommitMessage),
		Labels:                update.Labels,
		Assignees:             update.Assignees,
		Reviewers:             update.Reviewers,
		OpenPullRequestsLimit: update.OpenPullRequestsLimit,
		VersioningStrategy:    update.VersioningStrategy,
	}

	return pol
}

// convertSchedule converts Dependabot schedule to uptool Schedule.
func convertSchedule(schedule *Schedule) *engine.Schedule {
	if schedule == nil || schedule.Interval == "" {
		return nil
	}

	return &engine.Schedule{
		Interval: schedule.Interval,
		Day:      schedule.Day,
		Time:     schedule.Time,
		Timezone: schedule.Timezone,
		Cron:     schedule.Cronjob,
	}
}

// convertGroups converts Dependabot groups to uptool DependencyGroups.
func convertGroups(groups map[string]Group) map[string]*engine.DependencyGroup {
	if len(groups) == 0 {
		return nil
	}

	result := make(map[string]*engine.DependencyGroup, len(groups))
	for name, group := range groups {
		result[name] = &engine.DependencyGroup{
			AppliesTo:       group.AppliesTo,
			DependencyType:  group.DependencyType,
			Patterns:        group.Patterns,
			ExcludePatterns: group.ExcludePatterns,
			UpdateTypes:     group.UpdateTypes,
		}
	}
	return result
}

// convertAllowRules converts Dependabot allow rules to uptool DependencyRules.
func convertAllowRules(rules []AllowRule) []engine.DependencyRule {
	if len(rules) == 0 {
		return nil
	}

	result := make([]engine.DependencyRule, len(rules))
	for i, rule := range rules {
		result[i] = engine.DependencyRule{
			DependencyName: rule.DependencyName,
			DependencyType: rule.DependencyType,
		}
	}
	return result
}

// convertIgnoreRules converts Dependabot ignore rules to uptool IgnoreRules.
func convertIgnoreRules(rules []IgnoreRule) []engine.IgnoreRule {
	if len(rules) == 0 {
		return nil
	}

	result := make([]engine.IgnoreRule, len(rules))
	for i, rule := range rules {
		result[i] = engine.IgnoreRule{
			DependencyName: rule.DependencyName,
			Versions:       rule.Versions,
			UpdateTypes:    rule.UpdateTypes,
		}
	}
	return result
}

// convertCooldown converts Dependabot cooldown to uptool CooldownConfig.
func convertCooldown(cooldown *Cooldown) *engine.CooldownConfig {
	if cooldown == nil {
		return nil
	}

	return &engine.CooldownConfig{
		DefaultDays:     cooldown.DefaultDays,
		SemverMajorDays: cooldown.SemverMajorDays,
		SemverMinorDays: cooldown.SemverMinorDays,
		SemverPatchDays: cooldown.SemverPatchDays,
		Include:         cooldown.Include,
		Exclude:         cooldown.Exclude,
	}
}

// convertCommitMessage converts Dependabot commit-message to uptool CommitMessageConfig.
func convertCommitMessage(cm *CommitMessage) *engine.CommitMessageConfig {
	if cm == nil {
		return nil
	}

	return &engine.CommitMessageConfig{
		Prefix:            cm.Prefix,
		PrefixDevelopment: cm.PrefixDevelopment,
		IncludeScope:      cm.Include == "scope",
	}
}

// convertVersioningStrategyToUpdate maps Dependabot versioning-strategy to uptool update level.
func convertVersioningStrategyToUpdate(strategy string) string {
	switch strategy {
	case "lockfile-only":
		return policyNone // Only lockfile updates, no manifest changes
	case "increase":
		return policyMajor // Always bump, allow all updates
	case "increase-if-necessary":
		return policyMinor // Conservative, similar to minor
	case "widen":
		return policyMajor // Widen ranges, allow all
	case "auto", "":
		return policyMinor // Default: minor updates only
	default:
		return policyMinor
	}
}

// convertScheduleToCadence maps Dependabot schedule to uptool cadence.
func convertScheduleToCadence(schedule *Schedule) string {
	switch schedule.Interval {
	case "daily":
		return "daily"
	case cadenceWeekly:
		return cadenceWeekly
	case "monthly", "quarterly", "semiannually", "yearly":
		return "monthly"
	default:
		return cadenceWeekly
	}
}

// MigrationReport provides information about the migration process.
type MigrationReport struct {
	// SourceFile is the path to the source dependabot.yml
	SourceFile string

	// EcosystemsMigrated lists the ecosystems that were converted
	EcosystemsMigrated []string

	// UnsupportedFeatures lists features that couldn't be fully migrated
	UnsupportedFeatures []string

	// Warnings contains any warnings encountered during migration
	Warnings []string

	// IntegrationsCreated is the number of integration configs created
	IntegrationsCreated int
}

// MigrateWithReport converts a Dependabot config and returns a detailed report.
func (c *Config) MigrateWithReport(sourceFile string) (*policy.Config, *MigrationReport) {
	report := &MigrationReport{
		SourceFile:          sourceFile,
		EcosystemsMigrated:  make([]string, 0),
		UnsupportedFeatures: make([]string, 0),
		Warnings:            make([]string, 0),
	}

	uptoolConfig := &policy.Config{
		Version:      1,
		Integrations: make([]policy.IntegrationConfig, 0, len(c.Updates)),
	}

	// Track features that need manual attention
	for i := range c.Updates {
		update := &c.Updates[i]
		// Convert the integration
		integrationConfig := convertUpdateConfig(update)
		uptoolConfig.Integrations = append(uptoolConfig.Integrations, integrationConfig)
		report.EcosystemsMigrated = append(report.EcosystemsMigrated, update.PackageEcosystem)

		// Check for unsupported/partially supported features
		checkUnsupportedFeatures(update, report)
	}

	// Check for global unsupported features
	if len(c.Registries) > 0 {
		report.UnsupportedFeatures = append(report.UnsupportedFeatures,
			"private registries (require manual configuration)")
	}

	if len(c.MultiEcosystemGroups) > 0 {
		report.UnsupportedFeatures = append(report.UnsupportedFeatures,
			"multi-ecosystem groups")
	}

	report.IntegrationsCreated = len(uptoolConfig.Integrations)

	return uptoolConfig, report
}

func checkUnsupportedFeatures(update *UpdateConfig, report *MigrationReport) {
	// Note: The following features are NOW FULLY SUPPORTED:
	// - Dependency groups (grouping related updates together)
	// - Allow/Ignore rules (filtering which dependencies to update)
	// - Cooldown settings (delaying updates after release)
	// - Commit message customization
	// - Schedule enforcement (daily, weekly, monthly, cron)
	// - PR metadata (labels, assignees, reviewers) via GitHub Actions

	// PR metadata note (still needs GitHub Actions)
	if len(update.Labels) > 0 || len(update.Assignees) > 0 || len(update.Reviewers) > 0 {
		report.Warnings = append(report.Warnings,
			update.PackageEcosystem+": PR labels/assignees/reviewers will be applied via GitHub Actions")
	}

	// Cooldown note (requires release timestamp data)
	if update.Cooldown != nil {
		report.Warnings = append(report.Warnings,
			update.PackageEcosystem+": cooldown requires release timestamp data from registry")
	}

	// Vendor mode
	if update.Vendor {
		report.UnsupportedFeatures = append(report.UnsupportedFeatures,
			update.PackageEcosystem+": vendor mode requires manual configuration")
	}

	// Check if integration is supported
	if _, ok := EcosystemToIntegration[update.PackageEcosystem]; !ok {
		report.Warnings = append(report.Warnings,
			update.PackageEcosystem+": ecosystem may not be fully supported yet")
	}
}

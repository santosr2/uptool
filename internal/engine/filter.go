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

package engine

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Constants for dependency types.
const (
	depTypeAll         = "all"
	depTypeDevelopment = "development"
)

// parseIntSafe parses a string to int, returning 0 on error.
// This is intentionally lenient for version comparison where non-numeric parts should be treated as 0.
func parseIntSafe(s string) int {
	n, _ := strconv.Atoi(s) //nolint:errcheck // intentionally ignoring error, treating non-numeric as 0
	return n
}

// UpdateFilter applies policy-based filtering to updates.
// It enforces allow/ignore rules, cooldown periods, and versioning strategies.
type UpdateFilter struct {
	policy *IntegrationPolicy
}

// NewUpdateFilter creates a new filter with the given policy.
func NewUpdateFilter(policy *IntegrationPolicy) *UpdateFilter {
	return &UpdateFilter{policy: policy}
}

// FilterUpdates applies all policy filters to a list of updates.
// It returns the filtered updates and the reason each update was filtered (if any).
func (f *UpdateFilter) FilterUpdates(updates []Update, releaseTimestamps map[string]time.Time) ([]Update, map[string]string) {
	if f.policy == nil {
		return updates, nil
	}

	filtered := make([]Update, 0, len(updates))
	reasons := make(map[string]string)

	for i := range updates {
		update := &updates[i]
		depName := update.Dependency.Name

		// Check allow rules (if specified, only matching deps are allowed)
		if len(f.policy.Allow) > 0 && !f.isAllowed(update) {
			reasons[depName] = "not in allow list"
			continue
		}

		// Check ignore rules
		if reason := f.isIgnored(update); reason != "" {
			reasons[depName] = reason
			continue
		}

		// Check cooldown
		if releaseTimestamps != nil {
			if reason := f.checkCooldown(update, releaseTimestamps); reason != "" {
				reasons[depName] = reason
				continue
			}
		}

		filtered = append(filtered, *update)
	}

	return filtered, reasons
}

// isAllowed checks if an update matches any allow rule.
// Returns true if no allow rules are specified or if the update matches at least one rule.
func (f *UpdateFilter) isAllowed(update *Update) bool {
	if len(f.policy.Allow) == 0 {
		return true
	}

	for _, rule := range f.policy.Allow {
		if f.matchesDependencyRule(update, rule) {
			return true
		}
	}

	return false
}

// isIgnored checks if an update matches any ignore rule.
// Returns the reason if ignored, empty string otherwise.
func (f *UpdateFilter) isIgnored(update *Update) string {
	for _, rule := range f.policy.Ignore {
		if reason := f.matchesIgnoreRule(update, rule); reason != "" {
			return reason
		}
	}
	return ""
}

// matchesDependencyRule checks if an update matches a dependency rule.
func (f *UpdateFilter) matchesDependencyRule(update *Update, rule DependencyRule) bool {
	// Check dependency name pattern
	if rule.DependencyName != "" {
		if !matchGlob(rule.DependencyName, update.Dependency.Name) {
			return false
		}
	}

	// Check dependency type
	if rule.DependencyType != "" {
		depType := normalizeDependencyType(update.Dependency.Type)
		ruleType := normalizeDependencyType(rule.DependencyType)

		if ruleType == depTypeAll {
			return true
		}

		if ruleType != depType {
			return false
		}
	}

	return true
}

// matchesIgnoreRule checks if an update matches an ignore rule.
// Returns the reason if ignored, empty string otherwise.
func (f *UpdateFilter) matchesIgnoreRule(update *Update, rule IgnoreRule) string {
	// Check dependency name pattern
	if rule.DependencyName != "" {
		if !matchGlob(rule.DependencyName, update.Dependency.Name) {
			return ""
		}
	}

	// Check version ranges
	if len(rule.Versions) > 0 {
		versionMatched := false
		for _, versionPattern := range rule.Versions {
			if matchVersion(versionPattern, update.TargetVersion) {
				versionMatched = true
				break
			}
		}
		if !versionMatched {
			return ""
		}
		return "version ignored: " + strings.Join(rule.Versions, ", ")
	}

	// Check update types
	if len(rule.UpdateTypes) > 0 {
		updateType := normalizeUpdateType(update.Impact)
		for _, ignoredType := range rule.UpdateTypes {
			normalizedIgnored := normalizeUpdateType(ignoredType)
			if normalizedIgnored == updateType {
				return "update type ignored: " + ignoredType
			}
		}
		return ""
	}

	// If only dependency name matches (no version or update type filter)
	if rule.DependencyName != "" && len(rule.Versions) == 0 && len(rule.UpdateTypes) == 0 {
		return "dependency ignored: " + rule.DependencyName
	}

	return ""
}

// checkCooldown checks if an update should be delayed due to cooldown settings.
// Returns the reason if in cooldown, empty string otherwise.
func (f *UpdateFilter) checkCooldown(update *Update, releaseTimestamps map[string]time.Time) string {
	if f.policy.Cooldown == nil {
		return ""
	}

	cooldown := f.policy.Cooldown
	depName := update.Dependency.Name

	// Check if dependency is excluded from cooldown
	for _, pattern := range cooldown.Exclude {
		if matchGlob(pattern, depName) {
			return ""
		}
	}

	// Check if dependency is included in cooldown (if Include is specified)
	if len(cooldown.Include) > 0 {
		included := false
		for _, pattern := range cooldown.Include {
			if matchGlob(pattern, depName) {
				included = true
				break
			}
		}
		if !included {
			return ""
		}
	}

	// Get release timestamp
	key := depName + "@" + update.TargetVersion
	releaseTime, ok := releaseTimestamps[key]
	if !ok {
		// No timestamp available, can't enforce cooldown
		return ""
	}

	// Determine cooldown days based on update type
	cooldownDays := cooldown.DefaultDays
	switch strings.ToLower(update.Impact) {
	case string(ImpactMajor):
		if cooldown.SemverMajorDays > 0 {
			cooldownDays = cooldown.SemverMajorDays
		}
	case string(ImpactMinor):
		if cooldown.SemverMinorDays > 0 {
			cooldownDays = cooldown.SemverMinorDays
		}
	case string(ImpactPatch):
		if cooldown.SemverPatchDays > 0 {
			cooldownDays = cooldown.SemverPatchDays
		}
	}

	if cooldownDays <= 0 {
		return ""
	}

	// Check if release is old enough
	cooldownEnd := releaseTime.AddDate(0, 0, cooldownDays)
	if time.Now().Before(cooldownEnd) {
		daysRemaining := int(time.Until(cooldownEnd).Hours() / 24)
		return "cooldown: " + strconv.Itoa(daysRemaining) + " days remaining"
	}

	return ""
}

// GetCooldownDays returns the cooldown days for a specific update type.
func (f *UpdateFilter) GetCooldownDays(impact string) int {
	if f.policy == nil || f.policy.Cooldown == nil {
		return 0
	}

	cooldown := f.policy.Cooldown
	switch strings.ToLower(impact) {
	case string(ImpactMajor):
		if cooldown.SemverMajorDays > 0 {
			return cooldown.SemverMajorDays
		}
	case string(ImpactMinor):
		if cooldown.SemverMinorDays > 0 {
			return cooldown.SemverMinorDays
		}
	case string(ImpactPatch):
		if cooldown.SemverPatchDays > 0 {
			return cooldown.SemverPatchDays
		}
	}
	return cooldown.DefaultDays
}

// matchGlob matches a pattern against a string.
// Supports * as wildcard for any characters.
func matchGlob(pattern, str string) bool {
	// Handle exact match
	if pattern == str {
		return true
	}

	// Handle wildcard patterns
	if strings.Contains(pattern, "*") {
		// Convert glob pattern to regex
		regexPattern := "^" + regexp.QuoteMeta(pattern) + "$"
		regexPattern = strings.ReplaceAll(regexPattern, `\*`, ".*")

		matched, err := regexp.MatchString(regexPattern, str)
		if err != nil {
			return false
		}
		return matched
	}

	return false
}

// matchVersion checks if a version matches a version pattern.
// Supports patterns like "4.x", ">= 2.0.0", "< 3.0.0"
func matchVersion(pattern, version string) bool {
	pattern = strings.TrimSpace(pattern)
	version = strings.TrimSpace(version)

	// Remove 'v' prefix if present
	version = strings.TrimPrefix(version, "v")

	// Handle x.x patterns (e.g., "4.x", "4.17.x")
	if strings.Contains(pattern, ".x") {
		// Convert "4.x" to regex "^4\.\d+.*"
		prefix := strings.TrimSuffix(pattern, ".x")
		prefix = strings.TrimSuffix(prefix, "x")

		// Build regex pattern
		regexPattern := "^" + regexp.QuoteMeta(prefix)
		if !strings.HasSuffix(prefix, ".") {
			regexPattern += `\.`
		}
		regexPattern += `\d+`

		matched, err := regexp.MatchString(regexPattern, version)
		if err != nil {
			return false
		}
		return matched
	}

	// Handle comparison operators
	if strings.HasPrefix(pattern, ">=") {
		compareVer := strings.TrimSpace(strings.TrimPrefix(pattern, ">="))
		return compareVersions(version, compareVer) >= 0
	}
	if strings.HasPrefix(pattern, "<=") {
		compareVer := strings.TrimSpace(strings.TrimPrefix(pattern, "<="))
		return compareVersions(version, compareVer) <= 0
	}
	if strings.HasPrefix(pattern, ">") {
		compareVer := strings.TrimSpace(strings.TrimPrefix(pattern, ">"))
		return compareVersions(version, compareVer) > 0
	}
	if strings.HasPrefix(pattern, "<") {
		compareVer := strings.TrimSpace(strings.TrimPrefix(pattern, "<"))
		return compareVersions(version, compareVer) < 0
	}
	if strings.HasPrefix(pattern, "=") {
		compareVer := strings.TrimSpace(strings.TrimPrefix(pattern, "="))
		return version == compareVer
	}

	// Exact match
	return version == pattern
}

// compareVersions compares two semver versions.
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func compareVersions(v1, v2 string) int {
	// Remove 'v' prefix
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	// Split into parts
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	// Compare each part
	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(parts1) {
			// Extract numeric part (ignore prerelease suffixes)
			numStr := strings.Split(parts1[i], "-")[0]
			n1 = parseIntSafe(numStr)
		}
		if i < len(parts2) {
			numStr := strings.Split(parts2[i], "-")[0]
			n2 = parseIntSafe(numStr)
		}

		if n1 < n2 {
			return -1
		}
		if n1 > n2 {
			return 1
		}
	}

	return 0
}

// normalizeDependencyType normalizes dependency type strings.
func normalizeDependencyType(depType string) string {
	depType = strings.ToLower(strings.TrimSpace(depType))
	switch depType {
	case "prod", "production", "direct":
		return "production"
	case "dev", depTypeDevelopment, "devdependencies":
		return depTypeDevelopment
	case "peer", "peerdependencies":
		return "peer"
	case "optional", "optionaldependencies":
		return "optional"
	case "indirect":
		return "indirect"
	case depTypeAll, "*":
		return depTypeAll
	default:
		return depType
	}
}

// normalizeUpdateType normalizes update type strings.
// Handles both simple (major/minor/patch) and Dependabot format (version-update:semver-major).
func normalizeUpdateType(updateType string) string {
	updateType = strings.ToLower(strings.TrimSpace(updateType))

	// Handle Dependabot format (use unconditional TrimPrefix per staticcheck S1017)
	updateType = strings.TrimPrefix(updateType, "version-update:semver-")

	switch updateType {
	case string(ImpactMajor):
		return string(ImpactMajor)
	case string(ImpactMinor):
		return string(ImpactMinor)
	case string(ImpactPatch):
		return string(ImpactPatch)
	default:
		return updateType
	}
}

// GroupUpdates groups updates based on dependency group rules.
// Returns a map of group name to updates, and a slice of ungrouped updates.
func (f *UpdateFilter) GroupUpdates(updates []Update) (map[string][]Update, []Update) {
	if f.policy == nil || len(f.policy.Groups) == 0 {
		return nil, updates
	}

	grouped := make(map[string][]Update)
	ungrouped := make([]Update, 0)

	for i := range updates {
		update := &updates[i]
		groupName := f.findGroup(update)
		if groupName != "" {
			// Mark the update with its group
			update.Group = groupName
			grouped[groupName] = append(grouped[groupName], *update)
		} else {
			ungrouped = append(ungrouped, *update)
		}
	}

	return grouped, ungrouped
}

// findGroup finds the group name for an update.
// Returns empty string if the update doesn't match any group.
func (f *UpdateFilter) findGroup(update *Update) string {
	for groupName, group := range f.policy.Groups {
		if f.matchesGroup(update, group) {
			return groupName
		}
	}
	return ""
}

// matchesGroup checks if an update matches a dependency group.
func (f *UpdateFilter) matchesGroup(update *Update, group *DependencyGroup) bool {
	if group == nil {
		return false
	}

	depName := update.Dependency.Name
	depType := normalizeDependencyType(update.Dependency.Type)
	updateType := normalizeUpdateType(update.Impact)

	// Check dependency type filter
	if group.DependencyType != "" {
		groupDepType := normalizeDependencyType(group.DependencyType)
		if groupDepType != depType {
			return false
		}
	}

	// Check update type filter
	if len(group.UpdateTypes) > 0 {
		typeMatched := false
		for _, ut := range group.UpdateTypes {
			if normalizeUpdateType(ut) == updateType {
				typeMatched = true
				break
			}
		}
		if !typeMatched {
			return false
		}
	}

	// Check exclude patterns first
	for _, pattern := range group.ExcludePatterns {
		if matchGlob(pattern, depName) {
			return false
		}
	}

	// Check include patterns
	if len(group.Patterns) == 0 {
		// No patterns means match all (that passed other filters)
		return true
	}

	for _, pattern := range group.Patterns {
		if matchGlob(pattern, depName) {
			return true
		}
	}

	return false
}

// ApplyVersioningStrategy adjusts the target version based on the versioning strategy.
// Returns the adjusted version and whether the update should be applied.
func (f *UpdateFilter) ApplyVersioningStrategy(update *Update, currentConstraint string) (string, bool) {
	if f.policy == nil || f.policy.VersioningStrategy == "" {
		return update.TargetVersion, true
	}

	strategy := strings.ToLower(f.policy.VersioningStrategy)

	switch strategy {
	case "auto":
		// Default behavior - let the integration decide
		return update.TargetVersion, true

	case "lockfile-only":
		// Only update lock files, don't change manifests
		// Return false to indicate manifest shouldn't be updated
		return update.TargetVersion, false

	case "increase":
		// Always bump minimum version requirement
		return update.TargetVersion, true

	case "increase-if-necessary":
		// Only increase if current range doesn't accommodate new version
		if constraintAllowsVersion(currentConstraint, update.TargetVersion) {
			// Current constraint allows the new version, no manifest change needed
			return update.TargetVersion, false
		}
		return update.TargetVersion, true

	case "widen":
		// Widen the version range to include both old and new
		widenedConstraint := widenConstraint(currentConstraint, update.TargetVersion)
		return widenedConstraint, true

	default:
		return update.TargetVersion, true
	}
}

// constraintAllowsVersion checks if a version constraint allows a specific version.
// This is a simplified implementation - full implementation would use a proper semver library.
func constraintAllowsVersion(constraint, version string) bool {
	constraint = strings.TrimSpace(constraint)
	version = strings.TrimPrefix(strings.TrimSpace(version), "v")

	// Handle caret constraint (^)
	if strings.HasPrefix(constraint, "^") {
		baseVersion := strings.TrimPrefix(constraint, "^")
		baseVersion = strings.TrimPrefix(baseVersion, "v")

		baseParts := strings.Split(baseVersion, ".")
		versionParts := strings.Split(version, ".")

		// Caret allows changes that do not modify the left-most non-zero digit
		if len(baseParts) > 0 && len(versionParts) > 0 {
			baseMajor := parseIntSafe(baseParts[0])
			versionMajor := parseIntSafe(versionParts[0])

			if baseMajor == 0 {
				// For 0.x.x, caret allows patch updates only
				if len(baseParts) > 1 && len(versionParts) > 1 {
					baseMinor := parseIntSafe(baseParts[1])
					versionMinor := parseIntSafe(versionParts[1])
					return versionMajor == 0 && versionMinor == baseMinor
				}
			}

			// Major must match
			return versionMajor == baseMajor
		}
	}

	// Handle tilde constraint (~)
	if strings.HasPrefix(constraint, "~") {
		baseVersion := strings.TrimPrefix(constraint, "~")
		baseVersion = strings.TrimPrefix(baseVersion, "v")
		baseVersion = strings.TrimPrefix(baseVersion, ">") // Handle ~>

		baseParts := strings.Split(baseVersion, ".")
		versionParts := strings.Split(version, ".")

		// Tilde allows patch-level changes
		if len(baseParts) >= 2 && len(versionParts) >= 2 {
			baseMajor := parseIntSafe(baseParts[0])
			baseMinor := parseIntSafe(baseParts[1])
			versionMajor := parseIntSafe(versionParts[0])
			versionMinor := parseIntSafe(versionParts[1])

			return versionMajor == baseMajor && versionMinor == baseMinor
		}
	}

	// Handle >= constraint
	if strings.HasPrefix(constraint, ">=") {
		minVersion := strings.TrimSpace(strings.TrimPrefix(constraint, ">="))
		return compareVersions(version, minVersion) >= 0
	}

	// Handle exact version
	return strings.TrimPrefix(constraint, "v") == version
}

// widenConstraint creates a wider constraint that includes both old and new versions.
func widenConstraint(currentConstraint, newVersion string) string {
	// Extract the base version from the constraint
	baseVersion := currentConstraint
	prefix := ""

	switch {
	case strings.HasPrefix(currentConstraint, "^"):
		prefix = "^"
		baseVersion = strings.TrimPrefix(currentConstraint, "^")
	case strings.HasPrefix(currentConstraint, "~"):
		prefix = "~"
		baseVersion = strings.TrimPrefix(currentConstraint, "~")
	case strings.HasPrefix(currentConstraint, ">="):
		// For >= constraints, return as-is (already wide)
		return currentConstraint
	}

	// Parse versions
	baseVersion = strings.TrimPrefix(baseVersion, "v")
	newVersion = strings.TrimPrefix(newVersion, "v")

	baseParts := strings.Split(baseVersion, ".")
	newParts := strings.Split(newVersion, ".")

	// Determine the widest constraint
	if len(baseParts) > 0 && len(newParts) > 0 {
		baseMajor := parseIntSafe(baseParts[0])
		newMajor := parseIntSafe(newParts[0])

		if newMajor > baseMajor {
			// Major version increased - use >= with original base
			return ">=" + baseVersion
		}
	}

	// Return caret constraint with original base (allows minor/patch updates)
	return prefix + baseVersion
}

// ShouldUpdateManifest returns whether the manifest should be updated based on versioning strategy.
func (f *UpdateFilter) ShouldUpdateManifest() bool {
	if f.policy == nil || f.policy.VersioningStrategy == "" {
		return true
	}
	return !strings.EqualFold(f.policy.VersioningStrategy, "lockfile-only")
}

// FormatCommitMessage creates a commit message based on configuration.
func (f *UpdateFilter) FormatCommitMessage(updates []Update, manifestPath string) string {
	if f.policy == nil || f.policy.CommitMessage == nil {
		return f.defaultCommitMessage(updates, manifestPath)
	}

	cm := f.policy.CommitMessage
	var prefix string

	// Determine if this is a dev dependency update
	isDev := false
	for i := range updates {
		depType := normalizeDependencyType(updates[i].Dependency.Type)
		if depType == "development" {
			isDev = true
			break
		}
	}

	// Choose prefix
	if isDev && cm.PrefixDevelopment != "" {
		prefix = cm.PrefixDevelopment
	} else if cm.Prefix != "" {
		prefix = cm.Prefix
	}

	// Build message
	var msg strings.Builder

	if prefix != "" {
		msg.WriteString(prefix)
		msg.WriteString(": ")
	}

	// Add scope if requested
	if cm.IncludeScope && prefix == "" {
		if isDev {
			msg.WriteString("deps(dev): ")
		} else {
			msg.WriteString("deps: ")
		}
	}

	// Add update description
	if len(updates) == 1 {
		u := updates[0]
		msg.WriteString("update ")
		msg.WriteString(u.Dependency.Name)
		msg.WriteString(" from ")
		msg.WriteString(u.Dependency.CurrentVersion)
		msg.WriteString(" to ")
		msg.WriteString(u.TargetVersion)
	} else {
		msg.WriteString("update ")
		msg.WriteString(strconv.Itoa(len(updates)))
		msg.WriteString(" dependencies")

		// Check if all updates are in the same group
		groupName := ""
		sameGroup := true
		for i := range updates {
			if groupName == "" {
				groupName = updates[i].Group
			} else if updates[i].Group != groupName {
				sameGroup = false
				break
			}
		}

		if sameGroup && groupName != "" {
			msg.WriteString(" in ")
			msg.WriteString(groupName)
			msg.WriteString(" group")
		}
	}

	// Add manifest path for context
	msg.WriteString(" in ")
	msg.WriteString(filepath.Base(manifestPath))

	return msg.String()
}

// defaultCommitMessage generates a default commit message.
func (f *UpdateFilter) defaultCommitMessage(updates []Update, manifestPath string) string {
	var msg strings.Builder

	if len(updates) == 1 {
		u := updates[0]
		msg.WriteString("chore(deps): update ")
		msg.WriteString(u.Dependency.Name)
		msg.WriteString(" from ")
		msg.WriteString(u.Dependency.CurrentVersion)
		msg.WriteString(" to ")
		msg.WriteString(u.TargetVersion)
	} else {
		msg.WriteString("chore(deps): update ")
		msg.WriteString(strconv.Itoa(len(updates)))
		msg.WriteString(" dependencies in ")
		msg.WriteString(filepath.Base(manifestPath))
	}

	return msg.String()
}

// GetLabels returns the configured PR labels.
func (f *UpdateFilter) GetLabels() []string {
	if f.policy == nil || len(f.policy.Labels) == 0 {
		return []string{"dependencies", "automated"}
	}
	return f.policy.Labels
}

// GetAssignees returns the configured PR assignees.
func (f *UpdateFilter) GetAssignees() []string {
	if f.policy == nil {
		return nil
	}
	return f.policy.Assignees
}

// GetReviewers returns the configured PR reviewers.
func (f *UpdateFilter) GetReviewers() []string {
	if f.policy == nil {
		return nil
	}
	return f.policy.Reviewers
}

// GetOpenPullRequestsLimit returns the max open PRs limit.
func (f *UpdateFilter) GetOpenPullRequestsLimit() int {
	if f.policy == nil || f.policy.OpenPullRequestsLimit == 0 {
		return 5 // Default
	}
	return f.policy.OpenPullRequestsLimit
}

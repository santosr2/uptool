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

// Package resolve provides semantic version resolution and selection logic.
// It implements version constraint checking, policy-based version selection,
// and semantic version impact calculation (patch/minor/major).
//
// The package supports the following constraint syntaxes:
//   - Terraform-style: ~> 5.0, >= 1.0, = 1.2.3
//   - npm/Helm-style: ^1.2.3, ~1.2.3, >=1.0.0
//   - Exact versions: 1.2.3
//
// Policy precedence is applied in this order:
//  1. uptool.yaml policy (highest)
//  2. CLI flags
//  3. Manifest constraints (lowest)
package resolve

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"

	"github.com/santosr2/uptool/internal/engine"
)

// ConstraintType represents the type of version constraint.
type ConstraintType string

const (
	// ConstraintExact matches only the exact version (e.g., "1.2.3" or "= 1.2.3").
	ConstraintExact ConstraintType = "exact"

	// ConstraintPessimistic allows patch updates within a minor version (e.g., "~> 5.0" allows 5.0.x to 5.x.x).
	// In Terraform: ~> 5.0 allows >= 5.0, < 6.0
	// In npm/Helm: ~5.0 allows >= 5.0.0, < 5.1.0
	ConstraintPessimistic ConstraintType = "pessimistic"

	// ConstraintCaret allows minor and patch updates (e.g., "^1.2.3" allows 1.x.x).
	// Commonly used in npm: ^1.2.3 allows >= 1.2.3, < 2.0.0
	ConstraintCaret ConstraintType = "caret"

	// ConstraintMinimum allows any version >= the specified version (e.g., ">= 1.0").
	ConstraintMinimum ConstraintType = "minimum"

	// ConstraintRange allows versions within a range (e.g., ">= 1.0, < 2.0").
	ConstraintRange ConstraintType = "range"

	// updateLevelMajor is the default update level that allows all updates.
	updateLevelMajor = "major"
)

// ParsedConstraint represents a parsed version constraint.
type ParsedConstraint struct {
	// Original is the original constraint string as found in the manifest.
	Original string

	// Type indicates the constraint type (exact, pessimistic, caret, minimum, range).
	Type ConstraintType

	// BaseVersion is the version number extracted from the constraint (without prefix).
	BaseVersion string

	// Constraint is the parsed semver constraint for validation.
	// May be nil if the constraint couldn't be parsed as semver.
	Constraint *semver.Constraints

	// MaxAllowedImpact indicates the maximum update impact this constraint allows.
	// For example, ~> 5.0 allows "minor" (5.x), ^1.2.3 allows "minor" (1.x).
	MaxAllowedImpact engine.Impact
}

// ParseConstraint parses a version constraint string and returns a structured representation.
// It supports various constraint syntaxes from different ecosystems (Terraform, npm, Helm).
func ParseConstraint(constraint string) *ParsedConstraint {
	constraint = strings.TrimSpace(constraint)
	if constraint == "" {
		return &ParsedConstraint{
			Original:         constraint,
			Type:             ConstraintExact,
			MaxAllowedImpact: engine.ImpactMajor,
		}
	}

	result := &ParsedConstraint{
		Original: constraint,
	}

	// Detect constraint type and extract base version
	switch {
	case strings.HasPrefix(constraint, "~>"):
		// Terraform pessimistic constraint: ~> 5.0 allows >= 5.0, < 6.0
		result.Type = ConstraintPessimistic
		result.BaseVersion = strings.TrimSpace(strings.TrimPrefix(constraint, "~>"))
		result.MaxAllowedImpact = computePessimisticImpact(result.BaseVersion)
		result.Constraint = buildPessimisticConstraint(result.BaseVersion)

	case strings.HasPrefix(constraint, "^"):
		// Caret constraint: ^1.2.3 allows >= 1.2.3, < 2.0.0
		result.Type = ConstraintCaret
		result.BaseVersion = strings.TrimPrefix(constraint, "^")
		result.MaxAllowedImpact = engine.ImpactMinor // Allows minor and patch
		result.Constraint = buildCaretConstraint(result.BaseVersion)

	case strings.HasPrefix(constraint, "~") && !strings.HasPrefix(constraint, "~>"):
		// Tilde constraint: ~1.2.3 allows >= 1.2.3, < 1.3.0
		result.Type = ConstraintPessimistic
		result.BaseVersion = strings.TrimPrefix(constraint, "~")
		result.MaxAllowedImpact = engine.ImpactPatch // Only allows patch
		result.Constraint = buildTildeConstraint(result.BaseVersion)

	case strings.HasPrefix(constraint, ">="):
		// Minimum constraint: >= 1.0 allows any version >= 1.0
		result.Type = ConstraintMinimum
		result.BaseVersion = strings.TrimSpace(strings.TrimPrefix(constraint, ">="))
		result.MaxAllowedImpact = engine.ImpactMajor            // Allows all updates
		result.Constraint, _ = semver.NewConstraint(constraint) //nolint:errcheck // nil constraint is handled

	case strings.HasPrefix(constraint, ">"):
		// Greater than constraint
		result.Type = ConstraintMinimum
		result.BaseVersion = strings.TrimSpace(strings.TrimPrefix(constraint, ">"))
		result.MaxAllowedImpact = engine.ImpactMajor
		result.Constraint, _ = semver.NewConstraint(constraint) //nolint:errcheck // nil constraint is handled

	case strings.HasPrefix(constraint, "="):
		// Exact constraint: = 1.2.3
		result.Type = ConstraintExact
		result.BaseVersion = strings.TrimSpace(strings.TrimPrefix(constraint, "="))
		result.MaxAllowedImpact = engine.ImpactNone                            // No updates allowed
		result.Constraint, _ = semver.NewConstraint("= " + result.BaseVersion) //nolint:errcheck // nil constraint is handled

	default:
		// Exact version or unparsable
		result.Type = ConstraintExact
		result.BaseVersion = constraint
		result.MaxAllowedImpact = engine.ImpactNone                    // Exact versions don't allow updates
		result.Constraint, _ = semver.NewConstraint("= " + constraint) //nolint:errcheck // nil constraint is handled
	}

	return result
}

// computePessimisticImpact determines the max impact for Terraform's ~> constraint.
// ~> 5.0 (2 parts) allows minor updates (5.x)
// ~> 5.0.0 (3 parts) allows only patch updates (5.0.x)
func computePessimisticImpact(baseVersion string) engine.Impact {
	parts := strings.Split(baseVersion, ".")
	if len(parts) >= 3 {
		// ~> 5.0.0 only allows patch updates
		return engine.ImpactPatch
	}
	// ~> 5.0 allows minor updates
	return engine.ImpactMinor
}

// buildPessimisticConstraint builds a semver constraint for Terraform's ~> operator.
func buildPessimisticConstraint(baseVersion string) *semver.Constraints {
	parts := strings.Split(baseVersion, ".")

	var constraintStr string
	switch {
	case len(parts) >= 3:
		// ~> 5.0.0 means >= 5.0.0, < 5.1.0
		major := parts[0]
		minor := parts[1]
		constraintStr = fmt.Sprintf(">= %s, < %s.%d.0", baseVersion, major, mustParseInt(minor)+1)
	case len(parts) == 2:
		// ~> 5.0 means >= 5.0, < 6.0
		major := parts[0]
		constraintStr = fmt.Sprintf(">= %s.0, < %d.0.0", baseVersion, mustParseInt(major)+1)
	default:
		// ~> 5 means >= 5.0.0, < 6.0.0
		constraintStr = fmt.Sprintf(">= %s.0.0, < %d.0.0", baseVersion, mustParseInt(baseVersion)+1)
	}

	c, _ := semver.NewConstraint(constraintStr) //nolint:errcheck // nil is acceptable
	return c
}

// buildCaretConstraint builds a semver constraint for npm's ^ operator.
func buildCaretConstraint(baseVersion string) *semver.Constraints {
	// ^1.2.3 means >= 1.2.3, < 2.0.0
	// ^0.2.3 means >= 0.2.3, < 0.3.0 (special case for 0.x)
	// ^0.0.3 means >= 0.0.3, < 0.0.4 (special case for 0.0.x)
	c, _ := semver.NewConstraint("^" + baseVersion) //nolint:errcheck // nil is acceptable
	return c
}

// buildTildeConstraint builds a semver constraint for npm's ~ operator.
func buildTildeConstraint(baseVersion string) *semver.Constraints {
	// ~1.2.3 means >= 1.2.3, < 1.3.0
	c, _ := semver.NewConstraint("~" + baseVersion) //nolint:errcheck // nil is acceptable
	return c
}

// mustParseInt parses a string as int, returning 0 on failure.
func mustParseInt(s string) int {
	var result int
	_, _ = fmt.Sscanf(s, "%d", &result) //nolint:errcheck // best effort parsing
	return result
}

// Allows checks if the constraint allows updating to the given version.
func (pc *ParsedConstraint) Allows(targetVersion string) bool {
	if pc == nil || pc.Constraint == nil {
		return true // No constraint means allow all
	}

	v, err := normalizeAndParse(targetVersion)
	if err != nil {
		return false
	}

	return pc.Constraint.Check(v)
}

// AllowsImpact checks if the constraint allows updates of the given impact level.
func (pc *ParsedConstraint) AllowsImpact(impact engine.Impact) bool {
	if pc == nil {
		return true
	}

	impactOrder := map[engine.Impact]int{
		engine.ImpactNone:  0,
		engine.ImpactPatch: 1,
		engine.ImpactMinor: 2,
		engine.ImpactMajor: 3,
	}

	return impactOrder[impact] <= impactOrder[pc.MaxAllowedImpact]
}

// SelectVersionWithContext chooses the best version applying the full precedence chain:
// 1. CLI flags (highest precedence - if planCtx.CLIFlags.UpdateLevel is set, it overrides everything)
// 2. uptool.yaml policy (if policy.Update is set)
// 3. Manifest constraints (respects ~>, ^, >= etc.)
//
// The constraint parameter should be the original constraint string from the manifest
// (e.g., "~> 5.0", "^1.2.3", ">= 1.0"). If empty, only policy filtering is applied.
func SelectVersionWithContext(
	currentVersion string,
	constraint string,
	availableVersions []string,
	planCtx *engine.PlanContext,
) (string, engine.Impact, error) {
	if len(availableVersions) == 0 {
		return "", engine.ImpactNone, fmt.Errorf("no available versions")
	}

	// Check if version is pinned - if so, don't allow any updates
	if planCtx != nil && planCtx.Policy != nil && planCtx.Policy.Pin {
		return "", engine.ImpactNone, nil // No update needed (pinned)
	}

	// Parse current version (strip constraint prefix for comparison)
	currentClean := stripConstraintPrefix(currentVersion)
	current, err := normalizeAndParse(currentClean)
	if err != nil {
		return "", engine.ImpactNone, fmt.Errorf("parse current version %q: %w", currentVersion, err)
	}

	// Parse constraint if provided and should be respected
	// Constraints are only respected when there's no higher-precedence policy (CLI flags or uptool.yaml)
	var parsedConstraint *ParsedConstraint
	shouldRespectConstraint := true
	if planCtx != nil {
		// If CLI flags or policy from uptool.yaml exists, ignore constraints
		if planCtx.CLIFlags != nil && planCtx.CLIFlags.UpdateLevel != "" {
			shouldRespectConstraint = false
		} else if planCtx.Policy != nil && planCtx.Policy.Update != "" {
			shouldRespectConstraint = false
		}
	}

	if constraint != "" && shouldRespectConstraint {
		parsedConstraint = ParseConstraint(constraint)
	}

	// Get effective policy settings
	updateLevel := updateLevelMajor // Default: allow all
	allowPrerelease := false
	if planCtx != nil {
		updateLevel = planCtx.EffectiveUpdateLevel()
		allowPrerelease = planCtx.EffectiveAllowPrerelease()
	}

	// Parse and filter available versions (pre-allocate for performance)
	candidates := make([]*semver.Version, 0, len(availableVersions))
	for _, v := range availableVersions {
		parsed, err := normalizeAndParse(v)
		if err != nil {
			continue // skip invalid versions
		}

		// Filter out prereleases if not allowed
		if parsed.Prerelease() != "" && !allowPrerelease {
			continue
		}

		// Only consider versions newer than current
		if !parsed.GreaterThan(current) {
			continue
		}

		// Check constraint if we should respect it
		if parsedConstraint != nil && parsedConstraint.Constraint != nil {
			if !parsedConstraint.Constraint.Check(parsed) {
				continue // Version doesn't satisfy constraint
			}
		}

		candidates = append(candidates, parsed)
	}

	if len(candidates) == 0 {
		return "", engine.ImpactNone, nil // no updates available
	}

	// Sort candidates (newest first)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].GreaterThan(candidates[j])
	})

	// Filter by policy update level
	for _, candidate := range candidates {
		impact := determineImpact(current, candidate)

		// Check if policy allows this impact level
		if !isPolicyAllowed(updateLevel, impact) {
			continue
		}

		return candidate.Original(), impact, nil
	}

	return "", engine.ImpactNone, nil // no updates match policy
}

// stripConstraintPrefix removes constraint prefixes like ~>, ^, >=, etc.
func stripConstraintPrefix(version string) string {
	version = strings.TrimSpace(version)
	version = strings.TrimPrefix(version, "~>")
	version = strings.TrimPrefix(version, ">=")
	version = strings.TrimPrefix(version, "<=")
	version = strings.TrimPrefix(version, ">")
	version = strings.TrimPrefix(version, "<")
	version = strings.TrimPrefix(version, "=")
	version = strings.TrimPrefix(version, "^")
	version = strings.TrimPrefix(version, "~")
	return strings.TrimSpace(version)
}

// isPolicyAllowed checks if the given impact level is allowed by the policy.
func isPolicyAllowed(updateLevel string, impact engine.Impact) bool {
	switch updateLevel {
	case "major":
		return true
	case "minor":
		return impact == engine.ImpactMinor || impact == engine.ImpactPatch
	case "patch":
		return impact == engine.ImpactPatch
	case "none":
		return false
	default:
		return true // Unknown policy, allow all
	}
}

// SelectVersion chooses the best version from a list based on policy.
// It filters versions by update strategy and prerelease policy, then returns the latest.
//
// Deprecated: Use SelectVersionWithContext for proper policy precedence support.
func SelectVersion(currentVersion string, availableVersions []string, policy engine.IntegrationPolicy) (string, engine.Impact, error) {
	if len(availableVersions) == 0 {
		return "", engine.ImpactNone, fmt.Errorf("no available versions")
	}

	// Parse current version
	current, err := normalizeAndParse(currentVersion)
	if err != nil {
		return "", engine.ImpactNone, fmt.Errorf("parse current version %q: %w", currentVersion, err)
	}

	// Parse and filter available versions
	var candidates []*semver.Version
	for _, v := range availableVersions {
		parsed, err := normalizeAndParse(v)
		if err != nil {
			continue // skip invalid versions
		}

		// Filter out prereleases if not allowed
		if parsed.Prerelease() != "" && !policy.AllowPrerelease {
			continue
		}

		// Only consider versions newer than current
		if parsed.GreaterThan(current) {
			candidates = append(candidates, parsed)
		}
	}

	if len(candidates) == 0 {
		return "", engine.ImpactNone, nil // no updates available
	}

	// Sort candidates (newest first)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].GreaterThan(candidates[j])
	})

	// Filter by update policy
	for _, candidate := range candidates {
		impact := determineImpact(current, candidate)

		allowed := false
		switch policy.Update {
		case "major":
			allowed = true
		case "minor":
			allowed = impact == engine.ImpactMinor || impact == engine.ImpactPatch
		case "patch":
			allowed = impact == engine.ImpactPatch
		case "none":
			allowed = false
		}

		if allowed {
			return candidate.Original(), impact, nil
		}
	}

	return "", engine.ImpactNone, nil // no updates match policy
}

// normalizeAndParse attempts to parse a version string with lenient normalization.
// It handles versions with or without 'v' prefix and various formats.
func normalizeAndParse(version string) (*semver.Version, error) {
	// Try as-is first
	if v, err := semver.NewVersion(version); err == nil {
		return v, nil
	}

	// Try with 'v' prefix
	if !strings.HasPrefix(version, "v") {
		if v, err := semver.NewVersion("v" + version); err == nil {
			return v, nil
		}
	}

	// Try without 'v' prefix
	if strings.HasPrefix(version, "v") {
		if v, err := semver.NewVersion(strings.TrimPrefix(version, "v")); err == nil {
			return v, nil
		}
	}

	return nil, fmt.Errorf("invalid version: %s", version)
}

// determineImpact calculates the semver impact between two versions.
func determineImpact(current, newVer *semver.Version) engine.Impact {
	if newVer.Major() > current.Major() {
		return engine.ImpactMajor
	}
	if newVer.Minor() > current.Minor() {
		return engine.ImpactMinor
	}
	if newVer.Patch() > current.Patch() {
		return engine.ImpactPatch
	}
	// Prerelease or metadata change
	return engine.ImpactPatch
}

// IsValidSemver checks if a string is a valid semver version.
func IsValidSemver(version string) bool {
	_, err := normalizeAndParse(version)
	return err == nil
}

// CompareVersions returns -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2.
func CompareVersions(v1, v2 string) (int, error) {
	ver1, err := normalizeAndParse(v1)
	if err != nil {
		return 0, fmt.Errorf("parse v1 %q: %w", v1, err)
	}

	ver2, err := normalizeAndParse(v2)
	if err != nil {
		return 0, fmt.Errorf("parse v2 %q: %w", v2, err)
	}

	return ver1.Compare(ver2), nil
}

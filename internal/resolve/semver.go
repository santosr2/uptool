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
package resolve

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"

	"github.com/santosr2/uptool/internal/engine"
)

// SelectVersion chooses the best version from a list based on policy.
// It filters versions by update strategy and prerelease policy, then returns the latest.
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

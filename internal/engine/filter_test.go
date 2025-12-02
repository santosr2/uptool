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
	"testing"
	"time"
)

func TestMatchGlob(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		str     string
		want    bool
	}{
		{"exact match", "lodash", "lodash", true},
		{"exact no match", "lodash", "express", false},
		{"wildcard suffix", "express*", "express", true},
		{"wildcard suffix match", "express*", "express-validator", true},
		{"wildcard suffix no match", "express*", "body-parser", false},
		{"wildcard prefix", "*-parser", "body-parser", true},
		{"wildcard prefix no match", "*-parser", "express", false},
		{"wildcard both", "*lodash*", "lodash", true},
		{"wildcard both match", "*lodash*", "lodash-es", true},
		{"wildcard middle", "express*validator", "express-validator", true},
		{"scoped package", "@types/*", "@types/node", true},
		{"scoped package no match", "@types/*", "@babel/core", false},
		{"all wildcard", "*", "anything", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchGlob(tt.pattern, tt.str)
			if got != tt.want {
				t.Errorf("matchGlob(%q, %q) = %v, want %v", tt.pattern, tt.str, got, tt.want)
			}
		})
	}
}

func TestMatchVersion(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		version string
		want    bool
	}{
		{"exact match", "4.17.21", "4.17.21", true},
		{"exact no match", "4.17.21", "4.17.20", false},
		{"x pattern major", "4.x", "4.17.21", true},
		{"x pattern major no match", "4.x", "5.0.0", false},
		{"x pattern minor", "4.17.x", "4.17.21", true},
		{"x pattern minor no match", "4.17.x", "4.18.0", false},
		{"v prefix", "4.17.21", "v4.17.21", true},
		{"gte", ">= 2.0.0", "2.0.0", true},
		{"gte higher", ">= 2.0.0", "3.0.0", true},
		{"gte lower", ">= 2.0.0", "1.9.9", false},
		{"gt", "> 2.0.0", "2.0.1", true},
		{"gt equal", "> 2.0.0", "2.0.0", false},
		{"lte", "<= 2.0.0", "2.0.0", true},
		{"lte lower", "<= 2.0.0", "1.9.9", true},
		{"lte higher", "<= 2.0.0", "2.0.1", false},
		{"lt", "< 2.0.0", "1.9.9", true},
		{"lt equal", "< 2.0.0", "2.0.0", false},
		{"equal prefix", "= 2.0.0", "2.0.0", true},
		{"equal prefix no match", "= 2.0.0", "2.0.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchVersion(tt.pattern, tt.version)
			if got != tt.want {
				t.Errorf("matchVersion(%q, %q) = %v, want %v", tt.pattern, tt.version, got, tt.want)
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name string
		v1   string
		v2   string
		want int
	}{
		{"equal", "1.0.0", "1.0.0", 0},
		{"v1 less major", "1.0.0", "2.0.0", -1},
		{"v1 greater major", "2.0.0", "1.0.0", 1},
		{"v1 less minor", "1.0.0", "1.1.0", -1},
		{"v1 greater minor", "1.1.0", "1.0.0", 1},
		{"v1 less patch", "1.0.0", "1.0.1", -1},
		{"v1 greater patch", "1.0.1", "1.0.0", 1},
		{"v prefix", "v1.0.0", "1.0.0", 0},
		{"different lengths", "1.0", "1.0.0", 0},
		{"prerelease stripped", "1.0.0-beta", "1.0.0", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareVersions(tt.v1, tt.v2)
			if got != tt.want {
				t.Errorf("compareVersions(%q, %q) = %v, want %v", tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}

func TestNormalizeDependencyType(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"production", "production"},
		{"prod", "production"},
		{"direct", "production"},
		{"development", "development"},
		{"dev", "development"},
		{"devDependencies", "development"},
		{"peer", "peer"},
		{"peerDependencies", "peer"},
		{"optional", "optional"},
		{"optionalDependencies", "optional"},
		{"indirect", "indirect"},
		{"all", "all"},
		{"*", "all"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeDependencyType(tt.input)
			if got != tt.want {
				t.Errorf("normalizeDependencyType(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeUpdateType(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"major", "major"},
		{"MAJOR", "major"},
		{"minor", "minor"},
		{"patch", "patch"},
		{"version-update:semver-major", "major"},
		{"version-update:semver-minor", "minor"},
		{"version-update:semver-patch", "patch"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeUpdateType(tt.input)
			if got != tt.want {
				t.Errorf("normalizeUpdateType(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestUpdateFilter_FilterUpdates_AllowRules(t *testing.T) {
	policy := &IntegrationPolicy{
		Allow: []DependencyRule{
			{DependencyName: "express*"},
			{DependencyType: "production"},
		},
	}
	filter := NewUpdateFilter(policy)

	updates := []Update{
		{
			Dependency:    Dependency{Name: "express", Type: "production"},
			TargetVersion: "5.0.0",
			Impact:        "major",
		},
		{
			Dependency:    Dependency{Name: "express-validator", Type: "production"},
			TargetVersion: "7.0.0",
			Impact:        "major",
		},
		{
			Dependency:    Dependency{Name: "lodash", Type: "production"},
			TargetVersion: "5.0.0",
			Impact:        "major",
		},
		{
			Dependency:    Dependency{Name: "jest", Type: "development"},
			TargetVersion: "30.0.0",
			Impact:        "major",
		},
	}

	filtered, reasons := filter.FilterUpdates(updates, nil)

	// express, express-validator, lodash should pass (match either pattern or type)
	// jest should be filtered (doesn't match pattern and wrong type)
	if len(filtered) != 3 {
		t.Errorf("expected 3 updates after allow filter, got %d", len(filtered))
	}

	if _, ok := reasons["jest"]; !ok {
		t.Error("expected jest to be filtered with reason")
	}
}

func TestUpdateFilter_FilterUpdates_IgnoreRules(t *testing.T) {
	policy := &IntegrationPolicy{
		Ignore: []IgnoreRule{
			{DependencyName: "lodash", Versions: []string{"4.x"}},
			{DependencyName: "*", UpdateTypes: []string{"major"}},
		},
	}
	filter := NewUpdateFilter(policy)

	updates := []Update{
		{
			Dependency:    Dependency{Name: "lodash", Type: "production"},
			TargetVersion: "4.17.21",
			Impact:        "patch",
		},
		{
			Dependency:    Dependency{Name: "express", Type: "production"},
			TargetVersion: "5.0.0",
			Impact:        "major",
		},
		{
			Dependency:    Dependency{Name: "body-parser", Type: "production"},
			TargetVersion: "2.0.0",
			Impact:        "major",
		},
		{
			Dependency:    Dependency{Name: "cors", Type: "production"},
			TargetVersion: "2.8.6",
			Impact:        "minor",
		},
	}

	filtered, reasons := filter.FilterUpdates(updates, nil)

	// lodash 4.17.21 should be ignored (version 4.x)
	// express 5.0.0 should be ignored (major update)
	// body-parser 2.0.0 should be ignored (major update)
	// cors 2.8.6 should pass (minor update, not in ignore list)
	if len(filtered) != 1 {
		t.Errorf("expected 1 update after ignore filter, got %d", len(filtered))
	}

	if filtered[0].Dependency.Name != "cors" || filtered[0].TargetVersion != "2.8.6" {
		t.Errorf("expected cors 2.8.6 to pass filter, got %s %s", filtered[0].Dependency.Name, filtered[0].TargetVersion)
	}

	if len(reasons) != 3 {
		t.Errorf("expected 3 filtered reasons, got %d", len(reasons))
	}
}

func TestUpdateFilter_FilterUpdates_Cooldown(t *testing.T) {
	policy := &IntegrationPolicy{
		Cooldown: &CooldownConfig{
			DefaultDays:     3,
			SemverMajorDays: 7,
			SemverMinorDays: 3,
			SemverPatchDays: 1,
		},
	}
	filter := NewUpdateFilter(policy)

	now := time.Now()
	releaseTimestamps := map[string]time.Time{
		"express@5.0.0":      now.AddDate(0, 0, -2), // 2 days ago (major needs 7)
		"express@4.19.0":     now.AddDate(0, 0, -5), // 5 days ago (minor needs 3) - passed
		"lodash@4.17.22":     now.AddDate(0, 0, -1), // 1 day ago (patch needs 1) - passed
		"body-parser@1.20.0": now.AddDate(0, 0, 0),  // today (patch needs 1) - not passed
	}

	updates := []Update{
		{
			Dependency:    Dependency{Name: "express"},
			TargetVersion: "5.0.0",
			Impact:        "major",
		},
		{
			Dependency:    Dependency{Name: "express"},
			TargetVersion: "4.19.0",
			Impact:        "minor",
		},
		{
			Dependency:    Dependency{Name: "lodash"},
			TargetVersion: "4.17.22",
			Impact:        "patch",
		},
		{
			Dependency:    Dependency{Name: "body-parser"},
			TargetVersion: "1.20.0",
			Impact:        "patch",
		},
	}

	filtered, reasons := filter.FilterUpdates(updates, releaseTimestamps)

	// express 5.0.0 should be filtered (major, only 2 days old, needs 7)
	// express 4.19.0 should pass (minor, 5 days old, needs 3)
	// lodash 4.17.22 should pass (patch, 1 day old, needs 1)
	// body-parser 1.20.0 should be filtered (patch, 0 days old, needs 1)
	if len(filtered) != 2 {
		t.Errorf("expected 2 updates after cooldown filter, got %d", len(filtered))
	}

	if _, ok := reasons["express"]; !ok {
		t.Error("expected express 5.0.0 to be filtered for cooldown")
	}

	if _, ok := reasons["body-parser"]; !ok {
		t.Error("expected body-parser to be filtered for cooldown")
	}
}

func TestUpdateFilter_FilterUpdates_CooldownExclude(t *testing.T) {
	policy := &IntegrationPolicy{
		Cooldown: &CooldownConfig{
			DefaultDays: 7,
			Exclude:     []string{"security-*"},
		},
	}
	filter := NewUpdateFilter(policy)

	now := time.Now()
	releaseTimestamps := map[string]time.Time{
		"security-fix@1.0.0": now, // today
		"normal-pkg@1.0.0":   now, // today
	}

	updates := []Update{
		{
			Dependency:    Dependency{Name: "security-fix"},
			TargetVersion: "1.0.0",
			Impact:        "patch",
		},
		{
			Dependency:    Dependency{Name: "normal-pkg"},
			TargetVersion: "1.0.0",
			Impact:        "patch",
		},
	}

	filtered, _ := filter.FilterUpdates(updates, releaseTimestamps)

	// security-fix should pass (excluded from cooldown)
	// normal-pkg should be filtered (in cooldown)
	if len(filtered) != 1 {
		t.Errorf("expected 1 update after cooldown exclude, got %d", len(filtered))
	}

	if filtered[0].Dependency.Name != "security-fix" {
		t.Error("expected security-fix to pass cooldown exclude")
	}
}

func TestUpdateFilter_GroupUpdates(t *testing.T) {
	policy := &IntegrationPolicy{
		Groups: map[string]*DependencyGroup{
			"express-deps": {
				Patterns:    []string{"express*"},
				UpdateTypes: []string{"minor", "patch"},
			},
			"types": {
				Patterns:       []string{"@types/*"},
				DependencyType: "development",
			},
		},
	}
	filter := NewUpdateFilter(policy)

	updates := []Update{
		{
			Dependency:    Dependency{Name: "express", Type: "production"},
			TargetVersion: "4.19.0",
			Impact:        "minor",
		},
		{
			Dependency:    Dependency{Name: "express-validator", Type: "production"},
			TargetVersion: "7.0.1",
			Impact:        "patch",
		},
		{
			Dependency:    Dependency{Name: "express", Type: "production"},
			TargetVersion: "5.0.0",
			Impact:        "major", // Won't match express-deps (only minor/patch)
		},
		{
			Dependency:    Dependency{Name: "@types/node", Type: "development"},
			TargetVersion: "20.0.0",
			Impact:        "major",
		},
		{
			Dependency:    Dependency{Name: "lodash", Type: "production"},
			TargetVersion: "5.0.0",
			Impact:        "major",
		},
	}

	grouped, ungrouped := filter.GroupUpdates(updates)

	if len(grouped) != 2 {
		t.Errorf("expected 2 groups, got %d", len(grouped))
	}

	if len(grouped["express-deps"]) != 2 {
		t.Errorf("expected 2 updates in express-deps group, got %d", len(grouped["express-deps"]))
	}

	if len(grouped["types"]) != 1 {
		t.Errorf("expected 1 update in types group, got %d", len(grouped["types"]))
	}

	// express 5.0.0 and lodash should be ungrouped
	if len(ungrouped) != 2 {
		t.Errorf("expected 2 ungrouped updates, got %d", len(ungrouped))
	}
}

func TestUpdateFilter_GroupUpdates_ExcludePatterns(t *testing.T) {
	policy := &IntegrationPolicy{
		Groups: map[string]*DependencyGroup{
			"all-deps": {
				Patterns:        []string{"*"},
				ExcludePatterns: []string{"lodash*"},
			},
		},
	}
	filter := NewUpdateFilter(policy)

	updates := []Update{
		{Dependency: Dependency{Name: "express"}, TargetVersion: "5.0.0", Impact: "major"},
		{Dependency: Dependency{Name: "lodash"}, TargetVersion: "5.0.0", Impact: "major"},
		{Dependency: Dependency{Name: "lodash-es"}, TargetVersion: "5.0.0", Impact: "major"},
	}

	grouped, ungrouped := filter.GroupUpdates(updates)

	if len(grouped["all-deps"]) != 1 {
		t.Errorf("expected 1 update in all-deps group, got %d", len(grouped["all-deps"]))
	}

	if len(ungrouped) != 2 {
		t.Errorf("expected 2 ungrouped updates (lodash excluded), got %d", len(ungrouped))
	}
}

func TestUpdateFilter_FormatCommitMessage(t *testing.T) {
	tests := []struct {
		policy   *IntegrationPolicy
		name     string
		manifest string
		want     string
		updates  []Update
	}{
		{
			name:   "default message single update",
			policy: nil,
			updates: []Update{
				{
					Dependency:    Dependency{Name: "express", CurrentVersion: "4.18.0"},
					TargetVersion: "4.19.0",
				},
			},
			manifest: "package.json",
			want:     "chore(deps): update express from 4.18.0 to 4.19.0",
		},
		{
			name:   "default message multiple updates",
			policy: nil,
			updates: []Update{
				{Dependency: Dependency{Name: "express"}},
				{Dependency: Dependency{Name: "lodash"}},
			},
			manifest: "package.json",
			want:     "chore(deps): update 2 dependencies in package.json",
		},
		{
			name: "custom prefix",
			policy: &IntegrationPolicy{
				CommitMessage: &CommitMessageConfig{
					Prefix: "deps",
				},
			},
			updates: []Update{
				{
					Dependency:    Dependency{Name: "express", CurrentVersion: "4.18.0"},
					TargetVersion: "4.19.0",
				},
			},
			manifest: "package.json",
			want:     "deps: update express from 4.18.0 to 4.19.0 in package.json",
		},
		{
			name: "dev prefix for dev deps",
			policy: &IntegrationPolicy{
				CommitMessage: &CommitMessageConfig{
					Prefix:            "deps",
					PrefixDevelopment: "deps(dev)",
				},
			},
			updates: []Update{
				{
					Dependency:    Dependency{Name: "jest", Type: "development", CurrentVersion: "29.0.0"},
					TargetVersion: "30.0.0",
				},
			},
			manifest: "package.json",
			want:     "deps(dev): update jest from 29.0.0 to 30.0.0 in package.json",
		},
		{
			name: "include scope",
			policy: &IntegrationPolicy{
				CommitMessage: &CommitMessageConfig{
					IncludeScope: true,
				},
			},
			updates: []Update{
				{
					Dependency:    Dependency{Name: "express", Type: "production", CurrentVersion: "4.18.0"},
					TargetVersion: "4.19.0",
				},
			},
			manifest: "package.json",
			want:     "deps: update express from 4.18.0 to 4.19.0 in package.json",
		},
		{
			name: "grouped updates",
			policy: &IntegrationPolicy{
				CommitMessage: &CommitMessageConfig{
					Prefix: "deps",
				},
			},
			updates: []Update{
				{Dependency: Dependency{Name: "express"}, Group: "express-deps"},
				{Dependency: Dependency{Name: "express-validator"}, Group: "express-deps"},
			},
			manifest: "package.json",
			want:     "deps: update 2 dependencies in express-deps group in package.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewUpdateFilter(tt.policy)
			got := filter.FormatCommitMessage(tt.updates, tt.manifest)
			if got != tt.want {
				t.Errorf("FormatCommitMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUpdateFilter_ApplyVersioningStrategy(t *testing.T) {
	tests := []struct {
		name        string
		strategy    string
		update      Update
		constraint  string
		wantVersion string
		wantApply   bool
	}{
		{
			name:        "auto strategy",
			strategy:    "auto",
			update:      Update{TargetVersion: "2.0.0"},
			constraint:  "^1.0.0",
			wantVersion: "2.0.0",
			wantApply:   true,
		},
		{
			name:        "lockfile-only",
			strategy:    "lockfile-only",
			update:      Update{TargetVersion: "2.0.0"},
			constraint:  "^1.0.0",
			wantVersion: "2.0.0",
			wantApply:   false, // Don't update manifest
		},
		{
			name:        "increase",
			strategy:    "increase",
			update:      Update{TargetVersion: "2.0.0"},
			constraint:  "^1.0.0",
			wantVersion: "2.0.0",
			wantApply:   true,
		},
		{
			name:        "increase-if-necessary - needs increase",
			strategy:    "increase-if-necessary",
			update:      Update{TargetVersion: "2.0.0"},
			constraint:  "^1.0.0",
			wantVersion: "2.0.0",
			wantApply:   true, // Caret doesn't allow major bump
		},
		{
			name:        "increase-if-necessary - no increase needed",
			strategy:    "increase-if-necessary",
			update:      Update{TargetVersion: "1.5.0"},
			constraint:  "^1.0.0",
			wantVersion: "1.5.0",
			wantApply:   false, // Caret allows 1.5.0
		},
		{
			name:        "widen",
			strategy:    "widen",
			update:      Update{TargetVersion: "2.0.0"},
			constraint:  "^1.0.0",
			wantVersion: ">=1.0.0",
			wantApply:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := &IntegrationPolicy{
				VersioningStrategy: tt.strategy,
			}
			filter := NewUpdateFilter(policy)
			gotVersion, gotApply := filter.ApplyVersioningStrategy(tt.update, tt.constraint)

			if gotVersion != tt.wantVersion {
				t.Errorf("ApplyVersioningStrategy() version = %q, want %q", gotVersion, tt.wantVersion)
			}
			if gotApply != tt.wantApply {
				t.Errorf("ApplyVersioningStrategy() apply = %v, want %v", gotApply, tt.wantApply)
			}
		})
	}
}

func TestUpdateFilter_GetLabels(t *testing.T) {
	tests := []struct {
		name   string
		policy *IntegrationPolicy
		want   []string
	}{
		{
			name:   "nil policy",
			policy: nil,
			want:   []string{"dependencies", "automated"},
		},
		{
			name:   "empty labels",
			policy: &IntegrationPolicy{},
			want:   []string{"dependencies", "automated"},
		},
		{
			name: "custom labels",
			policy: &IntegrationPolicy{
				Labels: []string{"deps", "npm", "automerge"},
			},
			want: []string{"deps", "npm", "automerge"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewUpdateFilter(tt.policy)
			got := filter.GetLabels()
			if len(got) != len(tt.want) {
				t.Errorf("GetLabels() = %v, want %v", got, tt.want)
				return
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("GetLabels()[%d] = %q, want %q", i, v, tt.want[i])
				}
			}
		})
	}
}

func TestUpdateFilter_GetOpenPullRequestsLimit(t *testing.T) {
	tests := []struct {
		policy *IntegrationPolicy
		name   string
		want   int
	}{
		{
			name:   "nil policy",
			policy: nil,
			want:   5,
		},
		{
			name:   "zero limit (use default)",
			policy: &IntegrationPolicy{OpenPullRequestsLimit: 0},
			want:   5,
		},
		{
			name:   "custom limit",
			policy: &IntegrationPolicy{OpenPullRequestsLimit: 10},
			want:   10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewUpdateFilter(tt.policy)
			got := filter.GetOpenPullRequestsLimit()
			if got != tt.want {
				t.Errorf("GetOpenPullRequestsLimit() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestConstraintAllowsVersion(t *testing.T) {
	tests := []struct {
		name       string
		constraint string
		version    string
		want       bool
	}{
		{"caret allows minor", "^1.0.0", "1.5.0", true},
		{"caret allows patch", "^1.0.0", "1.0.5", true},
		{"caret blocks major", "^1.0.0", "2.0.0", false},
		{"tilde allows patch", "~1.0.0", "1.0.5", true},
		{"tilde blocks minor", "~1.0.0", "1.1.0", false},
		{"gte allows higher", ">=1.0.0", "2.0.0", true},
		{"exact match", "1.0.0", "1.0.0", true},
		{"exact no match", "1.0.0", "1.0.1", false},
		{"caret with 0.x", "^0.1.0", "0.1.5", true},
		{"caret with 0.x blocks minor", "^0.1.0", "0.2.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := constraintAllowsVersion(tt.constraint, tt.version)
			if got != tt.want {
				t.Errorf("constraintAllowsVersion(%q, %q) = %v, want %v", tt.constraint, tt.version, got, tt.want)
			}
		})
	}
}

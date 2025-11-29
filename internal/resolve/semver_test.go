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

package resolve

import (
	"testing"

	"github.com/santosr2/uptool/internal/engine"
)

func TestSelectVersion(t *testing.T) {
	tests := []struct {
		policy            engine.IntegrationPolicy
		name              string
		currentVersion    string
		wantVersion       string
		wantImpact        engine.Impact
		availableVersions []string
		wantErr           bool
	}{
		{
			name:              "patch update allowed",
			currentVersion:    "1.0.0",
			availableVersions: []string{"1.0.0", "1.0.1", "1.1.0", "2.0.0"},
			policy:            engine.IntegrationPolicy{Update: "patch"},
			wantVersion:       "1.0.1",
			wantImpact:        engine.ImpactPatch,
			wantErr:           false,
		},
		{
			name:              "minor update allowed",
			currentVersion:    "1.0.0",
			availableVersions: []string{"1.0.0", "1.0.1", "1.1.0", "2.0.0"},
			policy:            engine.IntegrationPolicy{Update: "minor"},
			wantVersion:       "1.1.0",
			wantImpact:        engine.ImpactMinor,
			wantErr:           false,
		},
		{
			name:              "major update allowed",
			currentVersion:    "1.0.0",
			availableVersions: []string{"1.0.0", "1.0.1", "1.1.0", "2.0.0"},
			policy:            engine.IntegrationPolicy{Update: "major"},
			wantVersion:       "2.0.0",
			wantImpact:        engine.ImpactMajor,
			wantErr:           false,
		},
		{
			name:              "no updates policy",
			currentVersion:    "1.0.0",
			availableVersions: []string{"1.0.0", "1.0.1", "1.1.0", "2.0.0"},
			policy:            engine.IntegrationPolicy{Update: "none"},
			wantVersion:       "",
			wantImpact:        engine.ImpactNone,
			wantErr:           false,
		},
		{
			name:              "prerelease filtered out",
			currentVersion:    "1.0.0",
			availableVersions: []string{"1.0.0", "1.0.1", "1.1.0-beta.1", "2.0.0-rc.1"},
			policy:            engine.IntegrationPolicy{Update: "minor", AllowPrerelease: false},
			wantVersion:       "1.0.1",
			wantImpact:        engine.ImpactPatch,
			wantErr:           false,
		},
		{
			name:              "prerelease allowed",
			currentVersion:    "1.0.0",
			availableVersions: []string{"1.0.0", "1.0.1", "1.1.0-beta.1"},
			policy:            engine.IntegrationPolicy{Update: "minor", AllowPrerelease: true},
			wantVersion:       "1.1.0-beta.1",
			wantImpact:        engine.ImpactMinor,
			wantErr:           false,
		},
		{
			name:              "no newer versions",
			currentVersion:    "2.0.0",
			availableVersions: []string{"1.0.0", "1.5.0", "2.0.0"},
			policy:            engine.IntegrationPolicy{Update: "major"},
			wantVersion:       "",
			wantImpact:        engine.ImpactNone,
			wantErr:           false,
		},
		{
			name:              "versions with v prefix",
			currentVersion:    "v1.0.0",
			availableVersions: []string{"v1.0.0", "v1.0.1", "v1.1.0"},
			policy:            engine.IntegrationPolicy{Update: "minor"},
			wantVersion:       "v1.1.0",
			wantImpact:        engine.ImpactMinor,
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVersion, gotImpact, err := SelectVersion(tt.currentVersion, tt.availableVersions, tt.policy)

			if (err != nil) != tt.wantErr {
				t.Errorf("SelectVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if gotVersion != tt.wantVersion {
				t.Errorf("SelectVersion() version = %v, want %v", gotVersion, tt.wantVersion)
			}

			if gotImpact != tt.wantImpact {
				t.Errorf("SelectVersion() impact = %v, want %v", gotImpact, tt.wantImpact)
			}
		})
	}
}

func TestIsValidSemver(t *testing.T) {
	tests := []struct {
		version string
		want    bool
	}{
		{"1.0.0", true},
		{"v1.0.0", true},
		{"1.2.3", true},
		{"1.2.3-beta.1", true},
		{"1.2.3+build.123", true},
		{"invalid", false},
		{"1.2", true}, // semver library accepts this (treats as 1.2.0)
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			if got := IsValidSemver(tt.version); got != tt.want {
				t.Errorf("IsValidSemver(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		v1      string
		v2      string
		want    int
		wantErr bool
	}{
		{"1.0.0", "1.0.0", 0, false},
		{"1.0.0", "1.0.1", -1, false},
		{"1.0.1", "1.0.0", 1, false},
		{"1.1.0", "1.0.0", 1, false},
		{"2.0.0", "1.9.9", 1, false},
		{"v1.0.0", "v1.0.1", -1, false},
		{"invalid", "1.0.0", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.v1+"_vs_"+tt.v2, func(t *testing.T) {
			got, err := CompareVersions(tt.v1, tt.v2)

			if (err != nil) != tt.wantErr {
				t.Errorf("CompareVersions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("CompareVersions(%q, %q) = %v, want %v", tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}

// TestParseConstraint tests the parsing of various constraint syntaxes.
func TestParseConstraint(t *testing.T) {
	tests := []struct {
		name             string
		constraint       string
		wantType         ConstraintType
		wantBaseVersion  string
		wantMaxImpact    engine.Impact
		allowsVersion    string
		disallowsVersion string
		shouldAllowVer   bool
	}{
		{
			name:             "terraform pessimistic ~> 5.0 (two parts)",
			constraint:       "~> 5.0",
			wantType:         ConstraintPessimistic,
			wantBaseVersion:  "5.0",
			wantMaxImpact:    engine.ImpactMinor,
			allowsVersion:    "5.9.0",
			shouldAllowVer:   true,
			disallowsVersion: "6.0.0",
		},
		{
			name:             "terraform pessimistic ~> 5.0.0 (three parts)",
			constraint:       "~> 5.0.0",
			wantType:         ConstraintPessimistic,
			wantBaseVersion:  "5.0.0",
			wantMaxImpact:    engine.ImpactPatch,
			allowsVersion:    "5.0.9",
			shouldAllowVer:   true,
			disallowsVersion: "5.1.0",
		},
		{
			name:             "npm caret ^1.2.3",
			constraint:       "^1.2.3",
			wantType:         ConstraintCaret,
			wantBaseVersion:  "1.2.3",
			wantMaxImpact:    engine.ImpactMinor,
			allowsVersion:    "1.9.0",
			shouldAllowVer:   true,
			disallowsVersion: "2.0.0",
		},
		{
			name:             "npm tilde ~1.2.3",
			constraint:       "~1.2.3",
			wantType:         ConstraintPessimistic,
			wantBaseVersion:  "1.2.3",
			wantMaxImpact:    engine.ImpactPatch,
			allowsVersion:    "1.2.9",
			shouldAllowVer:   true,
			disallowsVersion: "1.3.0",
		},
		{
			name:             "minimum >= 1.0",
			constraint:       ">= 1.0",
			wantType:         ConstraintMinimum,
			wantBaseVersion:  "1.0",
			wantMaxImpact:    engine.ImpactMajor,
			allowsVersion:    "2.0.0",
			shouldAllowVer:   true,
			disallowsVersion: "0.9.0",
		},
		{
			name:             "exact = 1.2.3",
			constraint:       "= 1.2.3",
			wantType:         ConstraintExact,
			wantBaseVersion:  "1.2.3",
			wantMaxImpact:    engine.ImpactNone,
			allowsVersion:    "1.2.3",
			shouldAllowVer:   true,
			disallowsVersion: "1.2.4",
		},
		{
			name:             "exact version (no operator)",
			constraint:       "1.2.3",
			wantType:         ConstraintExact,
			wantBaseVersion:  "1.2.3",
			wantMaxImpact:    engine.ImpactNone,
			allowsVersion:    "1.2.3",
			shouldAllowVer:   true,
			disallowsVersion: "1.2.4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := ParseConstraint(tt.constraint)

			if pc.Type != tt.wantType {
				t.Errorf("ParseConstraint(%q).Type = %v, want %v", tt.constraint, pc.Type, tt.wantType)
			}

			if pc.BaseVersion != tt.wantBaseVersion {
				t.Errorf("ParseConstraint(%q).BaseVersion = %q, want %q", tt.constraint, pc.BaseVersion, tt.wantBaseVersion)
			}

			if pc.MaxAllowedImpact != tt.wantMaxImpact {
				t.Errorf("ParseConstraint(%q).MaxAllowedImpact = %v, want %v", tt.constraint, pc.MaxAllowedImpact, tt.wantMaxImpact)
			}

			if tt.allowsVersion != "" && pc.Constraint != nil {
				if got := pc.Allows(tt.allowsVersion); got != tt.shouldAllowVer {
					t.Errorf("ParseConstraint(%q).Allows(%q) = %v, want %v", tt.constraint, tt.allowsVersion, got, tt.shouldAllowVer)
				}
			}

			if tt.disallowsVersion != "" && pc.Constraint != nil {
				if pc.Allows(tt.disallowsVersion) {
					t.Errorf("ParseConstraint(%q).Allows(%q) = true, want false", tt.constraint, tt.disallowsVersion)
				}
			}
		})
	}
}

// TestSelectVersionWithContext tests the policy precedence logic:
// uptool.yaml policy > CLI flags > manifest constraints.
func TestSelectVersionWithContext(t *testing.T) {
	availableVersions := []string{"5.0.0", "5.0.1", "5.1.0", "5.2.0", "6.0.0", "6.1.0"}

	tests := []struct {
		name           string
		currentVersion string
		constraint     string
		planCtx        *engine.PlanContext
		wantVersion    string
		wantImpact     engine.Impact
		wantErr        bool
	}{
		{
			name:           "no context - respects constraint ~> 5.0",
			currentVersion: "~> 5.0",
			constraint:     "~> 5.0",
			planCtx:        nil,
			wantVersion:    "5.2.0", // Latest within 5.x
			wantImpact:     engine.ImpactMinor,
		},
		{
			name:           "policy minor overrides constraint that would allow major",
			currentVersion: ">= 5.0",
			constraint:     ">= 5.0",
			planCtx: engine.NewPlanContext().WithPolicy(&engine.IntegrationPolicy{
				Update: "minor",
			}),
			wantVersion: "5.2.0", // Policy limits to minor, so no 6.x
			wantImpact:  engine.ImpactMinor,
		},
		{
			name:           "policy major allows all updates",
			currentVersion: "~> 5.0",
			constraint:     "~> 5.0",
			planCtx: engine.NewPlanContext().WithPolicy(&engine.IntegrationPolicy{
				Update: "major",
			}),
			wantVersion: "6.1.0", // Policy overrides constraint, allows major update
			wantImpact:  engine.ImpactMajor,
		},
		{
			name:           "policy patch only allows patch updates",
			currentVersion: "5.0.0",
			constraint:     "~> 5.0",
			planCtx: engine.NewPlanContext().WithPolicy(&engine.IntegrationPolicy{
				Update: "patch",
			}),
			wantVersion: "5.0.1", // Only patch allowed by policy
			wantImpact:  engine.ImpactPatch,
		},
		{
			name:           "policy none blocks all updates",
			currentVersion: "5.0.0",
			constraint:     "~> 5.0",
			planCtx: engine.NewPlanContext().WithPolicy(&engine.IntegrationPolicy{
				Update: "none",
			}),
			wantVersion: "", // No updates allowed
			wantImpact:  engine.ImpactNone,
		},
		{
			name:           "CLI flags override constraints when no policy",
			currentVersion: ">= 5.0",
			constraint:     ">= 5.0", // Would allow 6.x
			planCtx: engine.NewPlanContext().WithCLIFlags(&engine.CLIFlags{
				UpdateLevel: "minor",
			}),
			wantVersion: "5.2.0", // CLI flag limits to minor
			wantImpact:  engine.ImpactMinor,
		},
		{
			name:           "CLI flags take precedence over policy",
			currentVersion: "5.0.0",
			constraint:     ">= 5.0",
			planCtx: engine.NewPlanContext().WithPolicy(&engine.IntegrationPolicy{
				Update: "patch",
			}).WithCLIFlags(&engine.CLIFlags{
				UpdateLevel: "major", // CLI flags override policy
			}),
			wantVersion: "6.1.0", // CLI flags (major) take precedence
			wantImpact:  engine.ImpactMajor,
		},
		{
			name:           "empty constraint uses policy only",
			currentVersion: "5.0.0",
			constraint:     "",
			planCtx: engine.NewPlanContext().WithPolicy(&engine.IntegrationPolicy{
				Update: "minor",
			}),
			wantVersion: "5.2.0", // No constraint, policy allows minor
			wantImpact:  engine.ImpactMinor,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVersion, gotImpact, err := SelectVersionWithContext(
				tt.currentVersion,
				tt.constraint,
				availableVersions,
				tt.planCtx,
			)

			if (err != nil) != tt.wantErr {
				t.Errorf("SelectVersionWithContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if gotVersion != tt.wantVersion {
				t.Errorf("SelectVersionWithContext() version = %q, want %q", gotVersion, tt.wantVersion)
			}

			if gotImpact != tt.wantImpact {
				t.Errorf("SelectVersionWithContext() impact = %v, want %v", gotImpact, tt.wantImpact)
			}
		})
	}
}

// TestPlanContextPrecedence tests the EffectiveUpdateLevel method.
func TestPlanContextPrecedence(t *testing.T) {
	tests := []struct {
		name      string
		planCtx   *engine.PlanContext
		wantLevel string
	}{
		{
			name:      "nil context returns major",
			planCtx:   nil,
			wantLevel: "major",
		},
		{
			name:      "empty context returns major",
			planCtx:   engine.NewPlanContext(),
			wantLevel: "major",
		},
		{
			name: "CLI takes precedence over policy",
			planCtx: engine.NewPlanContext().WithPolicy(&engine.IntegrationPolicy{
				Update: "patch",
			}).WithCLIFlags(&engine.CLIFlags{
				UpdateLevel: "major",
			}),
			wantLevel: "major", // CLI wins
		},
		{
			name: "CLI used when no policy",
			planCtx: engine.NewPlanContext().WithCLIFlags(&engine.CLIFlags{
				UpdateLevel: "minor",
			}),
			wantLevel: "minor",
		},
		{
			name: "policy alone",
			planCtx: engine.NewPlanContext().WithPolicy(&engine.IntegrationPolicy{
				Update: "none",
			}),
			wantLevel: "none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.planCtx.EffectiveUpdateLevel()
			if got != tt.wantLevel {
				t.Errorf("EffectiveUpdateLevel() = %q, want %q", got, tt.wantLevel)
			}
		})
	}
}

func TestSelectVersionWithContext_Pin(t *testing.T) {
	tests := []struct {
		name              string
		currentVersion    string
		wantVersion       string
		wantImpact        engine.Impact
		availableVersions []string
		pin               bool
	}{
		{
			name:              "pin prevents updates",
			currentVersion:    "1.0.0",
			availableVersions: []string{"1.0.0", "1.1.0", "2.0.0"},
			pin:               true,
			wantVersion:       "",
			wantImpact:        engine.ImpactNone,
		},
		{
			name:              "no pin allows updates",
			currentVersion:    "1.0.0",
			availableVersions: []string{"1.0.0", "1.1.0", "2.0.0"},
			pin:               false,
			wantVersion:       "2.0.0",
			wantImpact:        engine.ImpactMajor,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			planCtx := engine.NewPlanContext().WithPolicy(&engine.IntegrationPolicy{
				Pin:    tt.pin,
				Update: "major",
			})

			gotVersion, gotImpact, err := SelectVersionWithContext(
				tt.currentVersion,
				"",
				tt.availableVersions,
				planCtx,
			)
			if err != nil {
				t.Fatalf("SelectVersionWithContext() error = %v", err)
			}

			if gotVersion != tt.wantVersion {
				t.Errorf("SelectVersionWithContext() version = %q, want %q", gotVersion, tt.wantVersion)
			}

			if gotImpact != tt.wantImpact {
				t.Errorf("SelectVersionWithContext() impact = %q, want %q", gotImpact, tt.wantImpact)
			}
		})
	}
}

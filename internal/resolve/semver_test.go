package resolve

import (
	"testing"

	"github.com/santosr2/uptool/internal/engine"
)

func TestSelectVersion(t *testing.T) {
	tests := []struct {
		name              string
		currentVersion    string
		availableVersions []string
		policy            engine.IntegrationPolicy
		wantVersion       string
		wantImpact        engine.Impact
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

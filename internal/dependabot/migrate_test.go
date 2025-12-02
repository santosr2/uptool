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
	"os"
	"path/filepath"
	"testing"
)

const (
	testNpm     = "npm"
	testGomod   = "gomod"
	testWeekly  = "weekly"
	testDaily   = "daily"
	testActions = "actions"
)

func TestMigrateToUptool(t *testing.T) {
	config := &Config{
		Version: 2,
		Updates: []UpdateConfig{
			{
				PackageEcosystem: testNpm,
				Directory:        "/",
				Schedule: Schedule{
					Interval: testWeekly,
					Day:      "monday",
				},
				VersioningStrategy: "auto",
			},
			{
				PackageEcosystem: "github-actions",
				Directory:        "/",
				Schedule: Schedule{
					Interval: testDaily,
				},
			},
		},
	}

	uptoolConfig := config.MigrateToUptool()

	if uptoolConfig.Version != 1 {
		t.Errorf("Version = %d, want 1", uptoolConfig.Version)
	}

	if len(uptoolConfig.Integrations) != 2 {
		t.Fatalf("len(Integrations) = %d, want 2", len(uptoolConfig.Integrations))
	}

	// Check npm integration
	npmInteg := uptoolConfig.Integrations[0]
	if npmInteg.ID != testNpm {
		t.Errorf("npm integration ID = %q, want %q", npmInteg.ID, testNpm)
	}
	if !npmInteg.Enabled {
		t.Error("npm integration should be enabled")
	}
	if npmInteg.Policy.Cadence != testWeekly {
		t.Errorf("npm policy cadence = %q, want %q", npmInteg.Policy.Cadence, testWeekly)
	}

	// Check actions integration
	actionsInteg := uptoolConfig.Integrations[1]
	if actionsInteg.ID != testActions {
		t.Errorf("actions integration ID = %q, want %q", actionsInteg.ID, testActions)
	}
	if actionsInteg.Policy.Cadence != testDaily {
		t.Errorf("actions policy cadence = %q, want %q", actionsInteg.Policy.Cadence, testDaily)
	}
}

func TestMigrateToUptool_VersioningStrategy(t *testing.T) {
	tests := []struct {
		name       string
		strategy   string
		wantUpdate string
	}{
		{"auto", "auto", "minor"},
		{"increase", "increase", "major"},
		{"increase-if-necessary", "increase-if-necessary", "minor"},
		{"lockfile-only", "lockfile-only", "none"},
		{"widen", "widen", "major"},
		{"empty", "", "minor"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Version: 2,
				Updates: []UpdateConfig{
					{
						PackageEcosystem:   "npm",
						Directory:          "/",
						Schedule:           Schedule{Interval: "weekly"},
						VersioningStrategy: tt.strategy,
					},
				},
			}

			uptoolConfig := config.MigrateToUptool()

			if uptoolConfig.Integrations[0].Policy.Update != tt.wantUpdate {
				t.Errorf("update policy = %q, want %q", uptoolConfig.Integrations[0].Policy.Update, tt.wantUpdate)
			}
		})
	}
}

func TestMigrateToUptool_ScheduleToCadence(t *testing.T) {
	tests := []struct {
		name        string
		interval    string
		wantCadence string
	}{
		{"daily", "daily", "daily"},
		{"weekly", "weekly", "weekly"},
		{"monthly", "monthly", "monthly"},
		{"quarterly", "quarterly", "monthly"},
		{"semiannually", "semiannually", "monthly"},
		{"yearly", "yearly", "monthly"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Version: 2,
				Updates: []UpdateConfig{
					{
						PackageEcosystem: "npm",
						Directory:        "/",
						Schedule:         Schedule{Interval: tt.interval},
					},
				},
			}

			uptoolConfig := config.MigrateToUptool()

			if uptoolConfig.Integrations[0].Policy.Cadence != tt.wantCadence {
				t.Errorf("cadence = %q, want %q", uptoolConfig.Integrations[0].Policy.Cadence, tt.wantCadence)
			}
		})
	}
}

func TestMigrateToUptool_FilePatterns(t *testing.T) {
	config := &Config{
		Version: 2,
		Updates: []UpdateConfig{
			{
				PackageEcosystem: "npm",
				Directories: []string{
					"/apps/frontend",
					"/apps/backend",
				},
				Schedule:     Schedule{Interval: "weekly"},
				ExcludePaths: []string{"**/node_modules/**"},
			},
		},
	}

	uptoolConfig := config.MigrateToUptool()

	integ := uptoolConfig.Integrations[0]
	if integ.Match == nil {
		t.Fatal("Match config should not be nil")
	}

	if len(integ.Match.Files) != 2 {
		t.Errorf("len(Match.Files) = %d, want 2", len(integ.Match.Files))
	}

	if len(integ.Match.Exclude) != 1 {
		t.Errorf("len(Match.Exclude) = %d, want 1", len(integ.Match.Exclude))
	}

	if integ.Match.Exclude[0] != "**/node_modules/**" {
		t.Errorf("Match.Exclude[0] = %q, want %q", integ.Match.Exclude[0], "**/node_modules/**")
	}
}

func TestMigrateWithReport(t *testing.T) {
	config := &Config{
		Version: 2,
		Registries: map[string]Registry{
			"npm-private": {
				Type: "npm-registry",
				URL:  "https://npm.example.com",
			},
		},
		Updates: []UpdateConfig{
			{
				PackageEcosystem: "npm",
				Directory:        "/",
				Schedule:         Schedule{Interval: "weekly"},
				Groups: map[string]Group{
					"production": {
						Patterns: []string{"*"},
					},
				},
				Allow: []AllowRule{
					{DependencyName: "express"},
				},
				Ignore: []IgnoreRule{
					{DependencyName: "lodash"},
				},
				Cooldown: &Cooldown{
					DefaultDays: 3,
				},
				CommitMessage: &CommitMessage{
					Prefix: "deps",
				},
				Labels:    []string{"dependencies"},
				Assignees: []string{"user1"},
				Reviewers: []string{"user2"},
				Vendor:    true,
			},
		},
	}

	uptoolConfig, report := config.MigrateWithReport("test/dependabot.yml")

	// Check config was created
	if uptoolConfig == nil {
		t.Fatal("uptoolConfig should not be nil")
	}

	// Check report fields
	if report.SourceFile != "test/dependabot.yml" {
		t.Errorf("SourceFile = %q, want %q", report.SourceFile, "test/dependabot.yml")
	}

	if report.IntegrationsCreated != 1 {
		t.Errorf("IntegrationsCreated = %d, want 1", report.IntegrationsCreated)
	}

	if len(report.EcosystemsMigrated) != 1 {
		t.Errorf("len(EcosystemsMigrated) = %d, want 1", len(report.EcosystemsMigrated))
	}

	if report.EcosystemsMigrated[0] != testNpm {
		t.Errorf("EcosystemsMigrated[0] = %q, want %q", report.EcosystemsMigrated[0], testNpm)
	}

	// Check unsupported features are reported
	if len(report.UnsupportedFeatures) == 0 {
		t.Error("UnsupportedFeatures should contain items")
	}

	// Should include registry warning
	hasRegistryWarning := false
	for _, feature := range report.UnsupportedFeatures {
		if feature == "private registries (require manual configuration)" {
			hasRegistryWarning = true
			break
		}
	}
	if !hasRegistryWarning {
		t.Error("UnsupportedFeatures should include registry warning")
	}

	// Check warnings
	if len(report.Warnings) == 0 {
		t.Error("Warnings should contain items")
	}
}

func TestMigrateWithReport_UnsupportedEcosystem(t *testing.T) {
	config := &Config{
		Version: 2,
		Updates: []UpdateConfig{
			{
				PackageEcosystem: "custom-unknown-ecosystem",
				Directory:        "/",
				Schedule:         Schedule{Interval: "weekly"},
			},
		},
	}

	_, report := config.MigrateWithReport("test/dependabot.yml")

	hasEcosystemWarning := false
	for _, warning := range report.Warnings {
		if warning == "custom-unknown-ecosystem: ecosystem may not be fully supported yet" {
			hasEcosystemWarning = true
			break
		}
	}

	if !hasEcosystemWarning {
		t.Error("Warnings should include ecosystem support warning")
	}
}

func TestMigrateWithReport_MultiEcosystemGroups(t *testing.T) {
	config := &Config{
		Version: 2,
		MultiEcosystemGroups: map[string]MultiEcosystemGroup{
			"all-deps": {
				Schedule: Schedule{Interval: "weekly"},
			},
		},
		Updates: []UpdateConfig{
			{
				PackageEcosystem: "npm",
				Directory:        "/",
				Schedule:         Schedule{Interval: "weekly"},
			},
		},
	}

	_, report := config.MigrateWithReport("test/dependabot.yml")

	hasMultiEcosystemWarning := false
	for _, feature := range report.UnsupportedFeatures {
		if feature == "multi-ecosystem groups" {
			hasMultiEcosystemWarning = true
			break
		}
	}

	if !hasMultiEcosystemWarning {
		t.Error("UnsupportedFeatures should include multi-ecosystem groups warning")
	}
}

func TestMigrateToUptool_AllEcosystems(t *testing.T) {
	ecosystems := []struct {
		ecosystem   string
		integration string
	}{
		{"npm", "npm"},
		{"github-actions", "actions"},
		{"gomod", "gomod"},
		{"docker", "docker"},
		{"docker-compose", "docker"},
		{"helm", "helm"},
		{"terraform", "terraform"},
		{"pip", "pip"},
		{"bundler", "bundler"},
		{"cargo", "cargo"},
		{"composer", "composer"},
		{"maven", "maven"},
		{"gradle", "gradle"},
		{"nuget", "nuget"},
		{"mix", "hex"},
		{"pub", "pub"},
		{"swift", "swift"},
	}

	for _, tt := range ecosystems {
		t.Run(tt.ecosystem, func(t *testing.T) {
			config := &Config{
				Version: 2,
				Updates: []UpdateConfig{
					{
						PackageEcosystem: tt.ecosystem,
						Directory:        "/",
						Schedule:         Schedule{Interval: "weekly"},
					},
				},
			}

			uptoolConfig := config.MigrateToUptool()

			if uptoolConfig.Integrations[0].ID != tt.integration {
				t.Errorf("integration ID = %q, want %q", uptoolConfig.Integrations[0].ID, tt.integration)
			}
		})
	}
}

func TestMigrateToUptool_Integration(t *testing.T) {
	// Test a full migration from file to uptool config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "dependabot.yml")

	configContent := `version: 2
updates:
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
      day: "monday"
    open-pull-requests-limit: 5
    labels:
      - "dependencies"
      - "go"
    groups:
      go-dependencies:
        patterns:
          - "*"
        update-types:
          - "minor"
          - "patch"

  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "weekly"
    labels:
      - "dependencies"
      - "github-actions"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	depConfig, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	uptoolConfig, report := depConfig.MigrateWithReport(configPath)

	// Verify migration was successful
	if uptoolConfig.Version != 1 {
		t.Errorf("Version = %d, want 1", uptoolConfig.Version)
	}

	if len(uptoolConfig.Integrations) != 2 {
		t.Fatalf("len(Integrations) = %d, want 2", len(uptoolConfig.Integrations))
	}

	// Check gomod integration
	gomodInteg := uptoolConfig.Integrations[0]
	if gomodInteg.ID != testGomod {
		t.Errorf("gomod integration ID = %q, want %q", gomodInteg.ID, testGomod)
	}
	if !gomodInteg.Enabled {
		t.Error("gomod integration should be enabled")
	}
	if gomodInteg.Policy.Cadence != testWeekly {
		t.Errorf("gomod cadence = %q, want %q", gomodInteg.Policy.Cadence, testWeekly)
	}

	// Check actions integration
	actionsInteg := uptoolConfig.Integrations[1]
	if actionsInteg.ID != testActions {
		t.Errorf("actions integration ID = %q, want %q", actionsInteg.ID, testActions)
	}

	// Check report
	if report.IntegrationsCreated != 2 {
		t.Errorf("IntegrationsCreated = %d, want 2", report.IntegrationsCreated)
	}

	if len(report.EcosystemsMigrated) != 2 {
		t.Errorf("len(EcosystemsMigrated) = %d, want 2", len(report.EcosystemsMigrated))
	}
}

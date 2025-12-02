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

func TestLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "dependabot.yml")

	configContent := `version: 2
updates:
  - package-ecosystem: "npm"
    directory: "/"
    schedule:
      interval: "weekly"
      day: "monday"
    open-pull-requests-limit: 5
    labels:
      - "dependencies"
      - "npm"
    commit-message:
      prefix: "deps"
      include: "scope"
    reviewers:
      - "santosr2"
    groups:
      npm-dependencies:
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

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if config.Version != 2 {
		t.Errorf("Version = %d, want 2", config.Version)
	}

	if len(config.Updates) != 2 {
		t.Fatalf("len(Updates) = %d, want 2", len(config.Updates))
	}

	// Check npm config
	npmUpdate := config.Updates[0]
	if npmUpdate.PackageEcosystem != "npm" {
		t.Errorf("Updates[0].PackageEcosystem = %q, want %q", npmUpdate.PackageEcosystem, "npm")
	}

	if npmUpdate.Directory != "/" {
		t.Errorf("Updates[0].Directory = %q, want %q", npmUpdate.Directory, "/")
	}

	if npmUpdate.Schedule.Interval != "weekly" {
		t.Errorf("Updates[0].Schedule.Interval = %q, want %q", npmUpdate.Schedule.Interval, "weekly")
	}

	if npmUpdate.Schedule.Day != "monday" {
		t.Errorf("Updates[0].Schedule.Day = %q, want %q", npmUpdate.Schedule.Day, "monday")
	}

	if npmUpdate.OpenPullRequestsLimit != 5 {
		t.Errorf("Updates[0].OpenPullRequestsLimit = %d, want 5", npmUpdate.OpenPullRequestsLimit)
	}

	if len(npmUpdate.Labels) != 2 {
		t.Errorf("len(Updates[0].Labels) = %d, want 2", len(npmUpdate.Labels))
	}

	// Check commit message
	if npmUpdate.CommitMessage == nil {
		t.Error("Updates[0].CommitMessage is nil")
	} else {
		if npmUpdate.CommitMessage.Prefix != "deps" {
			t.Errorf("CommitMessage.Prefix = %q, want %q", npmUpdate.CommitMessage.Prefix, "deps")
		}
		if npmUpdate.CommitMessage.Include != "scope" {
			t.Errorf("CommitMessage.Include = %q, want %q", npmUpdate.CommitMessage.Include, "scope")
		}
	}

	// Check groups
	if len(npmUpdate.Groups) != 1 {
		t.Errorf("len(Updates[0].Groups) = %d, want 1", len(npmUpdate.Groups))
	}
	if group, ok := npmUpdate.Groups["npm-dependencies"]; !ok {
		t.Error("Groups['npm-dependencies'] not found")
	} else {
		if len(group.Patterns) != 1 || group.Patterns[0] != "*" {
			t.Errorf("Group.Patterns = %v, want [*]", group.Patterns)
		}
		if len(group.UpdateTypes) != 2 {
			t.Errorf("len(Group.UpdateTypes) = %d, want 2", len(group.UpdateTypes))
		}
	}

	// Check github-actions config
	actionsUpdate := config.Updates[1]
	if actionsUpdate.PackageEcosystem != "github-actions" {
		t.Errorf("Updates[1].PackageEcosystem = %q, want %q", actionsUpdate.PackageEcosystem, "github-actions")
	}
}

func TestLoadConfig_InvalidVersion(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "dependabot.yml")

	configContent := `version: 1
updates:
  - package-ecosystem: "npm"
    directory: "/"
    schedule:
      interval: "weekly"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("LoadConfig() should error for version != 2")
	}
}

func TestLoadConfig_MissingPackageEcosystem(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "dependabot.yml")

	configContent := `version: 2
updates:
  - directory: "/"
    schedule:
      interval: "weekly"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("LoadConfig() should error for missing package-ecosystem")
	}
}

func TestLoadConfig_MissingDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "dependabot.yml")

	configContent := `version: 2
updates:
  - package-ecosystem: "npm"
    schedule:
      interval: "weekly"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("LoadConfig() should error for missing directory")
	}
}

func TestLoadConfig_InvalidScheduleInterval(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "dependabot.yml")

	configContent := `version: 2
updates:
  - package-ecosystem: "npm"
    directory: "/"
    schedule:
      interval: "invalid"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("LoadConfig() should error for invalid schedule interval")
	}
}

func TestLoadConfig_InvalidScheduleDay(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "dependabot.yml")

	configContent := `version: 2
updates:
  - package-ecosystem: "npm"
    directory: "/"
    schedule:
      interval: "weekly"
      day: "invalid-day"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("LoadConfig() should error for invalid schedule day")
	}
}

func TestLoadConfig_CronSchedule(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "dependabot.yml")

	configContent := `version: 2
updates:
  - package-ecosystem: "npm"
    directory: "/"
    schedule:
      interval: "cron"
      cronjob: "0 9 * * 1"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if config.Updates[0].Schedule.Interval != "cron" {
		t.Errorf("Schedule.Interval = %q, want %q", config.Updates[0].Schedule.Interval, "cron")
	}

	if config.Updates[0].Schedule.Cronjob != "0 9 * * 1" {
		t.Errorf("Schedule.Cronjob = %q, want %q", config.Updates[0].Schedule.Cronjob, "0 9 * * 1")
	}
}

func TestLoadConfig_CronMissingCronjob(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "dependabot.yml")

	configContent := `version: 2
updates:
  - package-ecosystem: "npm"
    directory: "/"
    schedule:
      interval: "cron"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("LoadConfig() should error when interval is 'cron' but cronjob is missing")
	}
}

func TestLoadConfig_InvalidVersioningStrategy(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "dependabot.yml")

	configContent := `version: 2
updates:
  - package-ecosystem: "npm"
    directory: "/"
    schedule:
      interval: "weekly"
    versioning-strategy: "invalid"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("LoadConfig() should error for invalid versioning-strategy")
	}
}

func TestLoadConfig_ValidVersioningStrategies(t *testing.T) {
	strategies := []string{"auto", "increase", "increase-if-necessary", "lockfile-only", "widen"}

	for _, strategy := range strategies {
		t.Run(strategy, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "dependabot.yml")

			configContent := `version: 2
updates:
  - package-ecosystem: "npm"
    directory: "/"
    schedule:
      interval: "weekly"
    versioning-strategy: "` + strategy + `"
`

			if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
				t.Fatalf("failed to create test config: %v", err)
			}

			config, err := LoadConfig(configPath)
			if err != nil {
				t.Errorf("LoadConfig() error for valid strategy %q: %v", strategy, err)
			}

			if config.Updates[0].VersioningStrategy != strategy {
				t.Errorf("VersioningStrategy = %q, want %q", config.Updates[0].VersioningStrategy, strategy)
			}
		})
	}
}

func TestLoadConfig_InvalidOpenPullRequestsLimit(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "dependabot.yml")

	configContent := `version: 2
updates:
  - package-ecosystem: "npm"
    directory: "/"
    schedule:
      interval: "weekly"
    open-pull-requests-limit: 15
`

	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("LoadConfig() should error for open-pull-requests-limit > 10")
	}
}

func TestLoadConfig_Directories(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "dependabot.yml")

	configContent := `version: 2
updates:
  - package-ecosystem: "npm"
    directories:
      - "/apps/frontend"
      - "/apps/backend"
      - "/libs/*"
    schedule:
      interval: "weekly"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if len(config.Updates[0].Directories) != 3 {
		t.Errorf("len(Directories) = %d, want 3", len(config.Updates[0].Directories))
	}

	dirs := config.Updates[0].GetDirectories()
	if len(dirs) != 3 {
		t.Errorf("GetDirectories() returned %d items, want 3", len(dirs))
	}
}

func TestLoadConfig_AllowIgnoreRules(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "dependabot.yml")

	configContent := `version: 2
updates:
  - package-ecosystem: "npm"
    directory: "/"
    schedule:
      interval: "weekly"
    allow:
      - dependency-name: "express"
      - dependency-type: "production"
    ignore:
      - dependency-name: "lodash"
        versions:
          - "4.x"
      - dependency-name: "*"
        update-types:
          - "version-update:semver-major"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	update := config.Updates[0]

	if len(update.Allow) != 2 {
		t.Errorf("len(Allow) = %d, want 2", len(update.Allow))
	}

	if update.Allow[0].DependencyName != "express" {
		t.Errorf("Allow[0].DependencyName = %q, want %q", update.Allow[0].DependencyName, "express")
	}

	if update.Allow[1].DependencyType != "production" {
		t.Errorf("Allow[1].DependencyType = %q, want %q", update.Allow[1].DependencyType, "production")
	}

	if len(update.Ignore) != 2 {
		t.Errorf("len(Ignore) = %d, want 2", len(update.Ignore))
	}

	if update.Ignore[0].DependencyName != "lodash" {
		t.Errorf("Ignore[0].DependencyName = %q, want %q", update.Ignore[0].DependencyName, "lodash")
	}

	if len(update.Ignore[0].Versions) != 1 || update.Ignore[0].Versions[0] != "4.x" {
		t.Errorf("Ignore[0].Versions = %v, want [4.x]", update.Ignore[0].Versions)
	}
}

func TestLoadConfig_Cooldown(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "dependabot.yml")

	configContent := `version: 2
updates:
  - package-ecosystem: "npm"
    directory: "/"
    schedule:
      interval: "weekly"
    cooldown:
      default-days: 3
      semver-major-days: 7
      semver-minor-days: 3
      semver-patch-days: 1
      include:
        - "express*"
      exclude:
        - "lodash"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	cooldown := config.Updates[0].Cooldown
	if cooldown == nil {
		t.Fatal("Cooldown is nil")
	}

	if cooldown.DefaultDays != 3 {
		t.Errorf("Cooldown.DefaultDays = %d, want 3", cooldown.DefaultDays)
	}

	if cooldown.SemverMajorDays != 7 {
		t.Errorf("Cooldown.SemverMajorDays = %d, want 7", cooldown.SemverMajorDays)
	}

	if cooldown.SemverMinorDays != 3 {
		t.Errorf("Cooldown.SemverMinorDays = %d, want 3", cooldown.SemverMinorDays)
	}

	if cooldown.SemverPatchDays != 1 {
		t.Errorf("Cooldown.SemverPatchDays = %d, want 1", cooldown.SemverPatchDays)
	}

	if len(cooldown.Include) != 1 || cooldown.Include[0] != "express*" {
		t.Errorf("Cooldown.Include = %v, want [express*]", cooldown.Include)
	}

	if len(cooldown.Exclude) != 1 || cooldown.Exclude[0] != "lodash" {
		t.Errorf("Cooldown.Exclude = %v, want [lodash]", cooldown.Exclude)
	}
}

func TestLoadConfig_Registries(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "dependabot.yml")

	configContent := `version: 2
registries:
  npm-private:
    type: npm-registry
    url: https://npm.example.com
    token: "${{ secrets.NPM_TOKEN }}"
  docker-private:
    type: docker-registry
    url: https://registry.example.com
    username: "${{ secrets.DOCKER_USER }}"
    password: "${{ secrets.DOCKER_PASS }}"
updates:
  - package-ecosystem: "npm"
    directory: "/"
    schedule:
      interval: "weekly"
    registries:
      - npm-private
`

	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if len(config.Registries) != 2 {
		t.Errorf("len(Registries) = %d, want 2", len(config.Registries))
	}

	npmReg, ok := config.Registries["npm-private"]
	if !ok {
		t.Fatal("Registries['npm-private'] not found")
	}

	if npmReg.Type != "npm-registry" {
		t.Errorf("npm-private.Type = %q, want %q", npmReg.Type, "npm-registry")
	}

	if npmReg.URL != "https://npm.example.com" {
		t.Errorf("npm-private.URL = %q, want %q", npmReg.URL, "https://npm.example.com")
	}
}

func TestGetIntegrationID(t *testing.T) {
	tests := []struct {
		ecosystem string
		expected  string
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
		{"unknown-ecosystem", "unknown-ecosystem"},
	}

	for _, tt := range tests {
		t.Run(tt.ecosystem, func(t *testing.T) {
			got := GetIntegrationID(tt.ecosystem)
			if got != tt.expected {
				t.Errorf("GetIntegrationID(%q) = %q, want %q", tt.ecosystem, got, tt.expected)
			}
		})
	}
}

func TestUpdateConfig_GetFilePatterns(t *testing.T) {
	tests := []struct {
		name      string
		wantFirst string
		update    UpdateConfig
		wantCount int
	}{
		{
			name: "npm root directory",
			update: UpdateConfig{
				PackageEcosystem: "npm",
				Directory:        "/",
			},
			wantCount: 1,
			wantFirst: "package.json",
		},
		{
			name: "npm subdirectory",
			update: UpdateConfig{
				PackageEcosystem: "npm",
				Directory:        "/apps/frontend",
			},
			wantCount: 1,
			wantFirst: "apps/frontend/package.json",
		},
		{
			name: "github-actions",
			update: UpdateConfig{
				PackageEcosystem: "github-actions",
				Directory:        "/",
			},
			wantCount: 1,
			wantFirst: ".github/workflows/*.yml",
		},
		{
			name: "terraform",
			update: UpdateConfig{
				PackageEcosystem: "terraform",
				Directory:        "/infra",
			},
			wantCount: 1,
			wantFirst: "infra/*.tf",
		},
		{
			name: "multiple directories",
			update: UpdateConfig{
				PackageEcosystem: "npm",
				Directories: []string{
					"/apps/frontend",
					"/apps/backend",
				},
			},
			wantCount: 2,
			wantFirst: "apps/frontend/package.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patterns := tt.update.GetFilePatterns()

			if len(patterns) != tt.wantCount {
				t.Errorf("GetFilePatterns() returned %d patterns, want %d", len(patterns), tt.wantCount)
			}

			if len(patterns) > 0 && patterns[0] != tt.wantFirst {
				t.Errorf("GetFilePatterns()[0] = %q, want %q", patterns[0], tt.wantFirst)
			}
		})
	}
}

func TestLoadConfig_NoUpdates(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "dependabot.yml")

	configContent := `version: 2
updates: []
`

	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("LoadConfig() should error when updates is empty")
	}
}

func TestLoadConfig_NonExistentFile(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/dependabot.yml")
	if err == nil {
		t.Error("LoadConfig() should error for non-existent file")
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "dependabot.yml")

	if err := os.WriteFile(configPath, []byte("invalid: [yaml"), 0o644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("LoadConfig() should error for invalid YAML")
	}
}

// testScheduleField is a helper to test schedule field parsing.
func testScheduleField(
	t *testing.T,
	values []string,
	configTemplate string,
	fieldName string,
	getField func(*Config) string,
) {
	t.Helper()
	for _, value := range values {
		t.Run(value, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "dependabot.yml")
			configContent := configTemplate + value + `"
`
			if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
				t.Fatalf("failed to create test config: %v", err)
			}
			config, err := LoadConfig(configPath)
			if err != nil {
				t.Errorf("LoadConfig() error for valid %s %q: %v", fieldName, value, err)
			}
			if got := getField(config); got != value {
				t.Errorf("Schedule.%s = %q, want %q", fieldName, got, value)
			}
		})
	}
}

func TestLoadConfig_AllScheduleIntervals(t *testing.T) {
	intervals := []string{"daily", "weekly", "monthly", "quarterly", "semiannually", "yearly"}
	template := `version: 2
updates:
  - package-ecosystem: "npm"
    directory: "/"
    schedule:
      interval: "`
	testScheduleField(t, intervals, template, "Interval", func(c *Config) string {
		return c.Updates[0].Schedule.Interval
	})
}

func TestLoadConfig_AllWeekDays(t *testing.T) {
	days := []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}
	template := `version: 2
updates:
  - package-ecosystem: "npm"
    directory: "/"
    schedule:
      interval: "weekly"
      day: "`
	testScheduleField(t, days, template, "Day", func(c *Config) string {
		return c.Updates[0].Schedule.Day
	})
}

func TestLoadConfig_CompleteExample(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "dependabot.yml")

	// This is a comprehensive example matching the project's actual dependabot.yml format
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
      - "dependabot"
    commit-message:
      prefix: "deps"
      prefix-development: "deps(dev)"
      include: "scope"
    reviewers:
      - "santosr2"
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
      day: "monday"
    open-pull-requests-limit: 3
    labels:
      - "dependencies"
      - "github-actions"
      - "dependabot"
    commit-message:
      prefix: "ci"
      include: "scope"
    reviewers:
      - "santosr2"
    groups:
      github-actions:
        patterns:
          - "*"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if len(config.Updates) != 2 {
		t.Fatalf("len(Updates) = %d, want 2", len(config.Updates))
	}

	// Validate gomod config
	gomodUpdate := config.Updates[0]
	if gomodUpdate.PackageEcosystem != "gomod" {
		t.Errorf("gomod update ecosystem = %q, want %q", gomodUpdate.PackageEcosystem, "gomod")
	}
	if gomodUpdate.CommitMessage.PrefixDevelopment != "deps(dev)" {
		t.Errorf("CommitMessage.PrefixDevelopment = %q, want %q", gomodUpdate.CommitMessage.PrefixDevelopment, "deps(dev)")
	}

	// Validate github-actions config
	actionsUpdate := config.Updates[1]
	if actionsUpdate.PackageEcosystem != "github-actions" {
		t.Errorf("actions update ecosystem = %q, want %q", actionsUpdate.PackageEcosystem, "github-actions")
	}
}

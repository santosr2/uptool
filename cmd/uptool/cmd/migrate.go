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

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/santosr2/uptool/internal/dependabot"
)

var (
	migrateOutputFlag string
	migrateDryRunFlag bool
	migrateForceFlag  bool
	migrateSourceFlag string

	migrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Migrate from Dependabot to uptool configuration",
		Long: `Migrate an existing dependabot.yml configuration to uptool.yaml format.

This command reads your Dependabot configuration and converts it to the equivalent
uptool configuration, preserving all supported settings including:

  - Package ecosystems (npm, github-actions, docker, helm, terraform, etc.)
  - Schedule settings (interval, day, time, timezone, cron)
  - Dependency groups
  - Allow and ignore rules
  - Cooldown settings
  - Commit message customization
  - PR labels, assignees, and reviewers
  - Open pull requests limit
  - Versioning strategy

Example:
  # Auto-detect dependabot.yml and create uptool.yaml
  uptool migrate

  # Specify source and output files
  uptool migrate --source .github/dependabot.yml --output uptool.yaml

  # Preview migration without writing files
  uptool migrate --dry-run

Note: Some Dependabot features may require manual adjustment after migration.
The command will report any features that couldn't be fully converted.`,
		RunE: runMigrate,
	}
)

func init() {
	migrateCmd.Flags().StringVarP(&migrateSourceFlag, "source", "s", "", "path to dependabot.yml (default: auto-detect)")
	migrateCmd.Flags().StringVarP(&migrateOutputFlag, "output", "o", "uptool.yaml", "output path for uptool.yaml")
	migrateCmd.Flags().BoolVar(&migrateDryRunFlag, "dry-run", false, "preview migration without writing files")
	migrateCmd.Flags().BoolVarP(&migrateForceFlag, "force", "f", false, "overwrite existing uptool.yaml")

	rootCmd.AddCommand(migrateCmd)
}

func runMigrate(cmd *cobra.Command, args []string) error {
	// Find source file
	sourcePath := migrateSourceFlag
	if sourcePath == "" {
		// Auto-detect dependabot.yml location
		candidates := []string{
			".github/dependabot.yml",
			".github/dependabot.yaml",
			"dependabot.yml",
			"dependabot.yaml",
		}
		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				sourcePath = candidate
				break
			}
		}
		if sourcePath == "" {
			return fmt.Errorf("no dependabot.yml found; specify with --source flag")
		}
	}

	// Check if source exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source file not found: %s", sourcePath)
	}

	fmt.Printf("Reading Dependabot configuration from: %s\n", sourcePath)

	// Load dependabot configuration
	depConfig, err := dependabot.LoadConfig(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to load dependabot config: %w", err)
	}

	// Convert to uptool configuration
	uptoolConfig, report := depConfig.MigrateWithReport(sourcePath)

	// Print migration report
	printMigrationReport(report)

	// Marshal to YAML
	yamlData, err := yaml.Marshal(uptoolConfig)
	if err != nil {
		return fmt.Errorf("failed to generate uptool config: %w", err)
	}

	// Add header comment
	header := `# uptool configuration
# Migrated from: ` + sourcePath + `
# See https://github.com/santosr2/uptool for documentation

`
	output := header + string(yamlData)

	if migrateDryRunFlag {
		fmt.Println("\n--- Generated uptool.yaml (dry-run) ---")
		fmt.Println(output)
		return nil
	}

	// Check if output exists
	if !migrateForceFlag {
		if _, err := os.Stat(migrateOutputFlag); err == nil {
			return fmt.Errorf("output file %s already exists; use --force to overwrite", migrateOutputFlag)
		}
	}

	// Ensure output directory exists
	outDir := filepath.Dir(migrateOutputFlag)
	if outDir != "" && outDir != "." {
		if err := os.MkdirAll(outDir, 0o750); err != nil { // #nosec G301 -- directory needs to be accessible
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Write output file
	if err := os.WriteFile(migrateOutputFlag, []byte(output), 0o600); err != nil { // #nosec G306 -- config file needs secure permissions
		return fmt.Errorf("failed to write uptool config: %w", err)
	}

	fmt.Printf("\nMigration complete! Written to: %s\n", migrateOutputFlag)

	if len(report.UnsupportedFeatures) > 0 || len(report.Warnings) > 0 {
		fmt.Println("\nPlease review the generated configuration and adjust as needed.")
	}

	return nil
}

func printMigrationReport(report *dependabot.MigrationReport) {
	fmt.Println("\n=== Migration Report ===")
	fmt.Printf("Source: %s\n", report.SourceFile)
	fmt.Printf("Integrations created: %d\n", report.IntegrationsCreated)

	if len(report.EcosystemsMigrated) > 0 {
		fmt.Println("\nMigrated ecosystems:")
		for _, eco := range report.EcosystemsMigrated {
			integrationID := dependabot.GetIntegrationID(eco)
			if integrationID != eco {
				fmt.Printf("  - %s -> %s\n", eco, integrationID)
			} else {
				fmt.Printf("  - %s\n", eco)
			}
		}
	}

	if len(report.UnsupportedFeatures) > 0 {
		fmt.Println("\nUnsupported features (manual configuration may be required):")
		for _, feature := range report.UnsupportedFeatures {
			fmt.Printf("  - %s\n", feature)
		}
	}

	if len(report.Warnings) > 0 {
		fmt.Println("\nWarnings:")
		for _, warning := range report.Warnings {
			fmt.Printf("  - %s\n", warning)
		}
	}
}

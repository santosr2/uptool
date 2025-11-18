package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	updateDryRun  bool
	updateDiff    bool
	updateOnly    string
	updateExclude string
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Apply updates to manifests",
	Long: `Apply dependency updates to manifest files.

This command scans for manifests, generates an update plan, and applies
the updates by rewriting manifest files with new dependency versions.
Formatting and structure are preserved.`,
	Example: `  # Update all dependencies
  uptool update

  # Show what would be updated (dry run)
  uptool update --dry-run

  # Update with diffs
  uptool update --diff

  # Update only npm dependencies
  uptool update --only npm

  # Update everything except terraform
  uptool update --exclude terraform`,
	RunE: runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().BoolVar(&updateDryRun, "dry-run", false, "show changes without applying")
	updateCmd.Flags().BoolVar(&updateDiff, "diff", false, "show diffs of changes")
	updateCmd.Flags().StringVar(&updateOnly, "only", "", "comma-separated integrations to include")
	updateCmd.Flags().StringVar(&updateExclude, "exclude", "", "comma-separated integrations to exclude")

	// Add shell completion for flags
	_ = updateCmd.RegisterFlagCompletionFunc("only", completeIntegrations)
	_ = updateCmd.RegisterFlagCompletionFunc("exclude", completeIntegrations)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	eng := setupEngine()
	ctx := context.Background()

	repoRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	onlyList, excludeList := parseFilters(updateOnly, updateExclude)

	// Scan
	scanResult, err := eng.Scan(ctx, repoRoot, onlyList, excludeList)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	if len(scanResult.Manifests) == 0 {
		fmt.Println("No manifests found.")
		return nil
	}

	// Plan
	planResult, err := eng.Plan(ctx, scanResult.Manifests)
	if err != nil {
		return fmt.Errorf("plan failed: %w", err)
	}

	if len(planResult.Plans) == 0 {
		fmt.Println("No updates available.")
		return nil
	}

	// Show plan
	fmt.Printf("Found %d manifests with updates:\n\n", len(planResult.Plans))
	if err := outputPlanTable(planResult); err != nil {
		return err
	}

	if updateDryRun {
		fmt.Println("\nDry-run mode: no changes applied.")
		return nil
	}

	// Apply
	fmt.Println("\nApplying updates...")
	updateResult, err := eng.Update(ctx, planResult.Plans, false)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	// Show results
	fmt.Println("\n=== Update Results ===")
	for _, result := range updateResult.Results {
		fmt.Printf("\n%s:\n", result.Manifest.Path)
		fmt.Printf("  Applied: %d\n", result.Applied)
		if result.Failed > 0 {
			fmt.Printf("  Failed: %d\n", result.Failed)
		}

		if updateDiff && result.ManifestDiff != "" {
			fmt.Printf("\nDiff:\n%s\n", result.ManifestDiff)
		}
	}

	return nil
}

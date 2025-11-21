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
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/santosr2/uptool/internal/engine"
)

var (
	planFormat  string
	planOut     string
	planOnly    string
	planExclude string
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Generate update plans",
	Long: `Generate update plans showing available dependency updates.

This command scans for manifests and queries registries to determine which
dependencies have newer versions available. The plan shows what would be
updated without making any changes.`,
	Example: `  # Generate plan with table output
  uptool plan

  # Generate plan as JSON
  uptool plan --format json

  # Save plan to file
  uptool plan --out plan.json

  # Plan only npm dependencies
  uptool plan --only npm`,
	RunE: runPlan,
}

func init() {
	rootCmd.AddCommand(planCmd)

	planCmd.Flags().StringVarP(&planFormat, "format", "f", "table", "output format: table, json")
	planCmd.Flags().StringVarP(&planOut, "out", "o", "", "write plan to file")
	planCmd.Flags().StringVar(&planOnly, "only", "", "comma-separated integrations to include")
	planCmd.Flags().StringVar(&planExclude, "exclude", "", "comma-separated integrations to exclude")

	// Add shell completion for flags
	if err := planCmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"table", "json"}, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		// This is a non-critical error during CLI initialization
		fmt.Fprintf(os.Stderr, "Warning: failed to register shell completion: %v\n", err)
	}

	if err := planCmd.RegisterFlagCompletionFunc("only", completeIntegrations); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to register shell completion: %v\n", err)
	}
	if err := planCmd.RegisterFlagCompletionFunc("exclude", completeIntegrations); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to register shell completion: %v\n", err)
	}
	if err := planCmd.RegisterFlagCompletionFunc("out", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveDefault // File completion
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to register shell completion: %v\n", err)
	}
}

func runPlan(cmd *cobra.Command, args []string) error {
	eng := setupEngine()
	ctx := context.Background()

	repoRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	onlyList, excludeList := parseFilters(planOnly, planExclude)

	// First scan
	scanResult, err := eng.Scan(ctx, repoRoot, onlyList, excludeList)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	// Then plan
	planResult, err := eng.Plan(ctx, scanResult.Manifests)
	if err != nil {
		return fmt.Errorf("plan failed: %w", err)
	}

	// Write to file if requested
	if planOut != "" {
		data, err := json.MarshalIndent(planResult, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal plan: %w", err)
		}
		if err := os.WriteFile(planOut, data, 0o600); err != nil {
			return fmt.Errorf("write plan file: %w", err)
		}
		fmt.Printf("Plan written to %s\n", planOut)
	}

	switch planFormat {
	case "json":
		return outputJSON(planResult)
	case "table":
		return outputPlanTable(planResult)
	default:
		return fmt.Errorf("unsupported format: %s", planFormat)
	}
}

func outputPlanTable(result *engine.PlanResult) error {
	if len(result.Plans) == 0 {
		fmt.Println("No updates available.")
		return nil
	}

	for _, plan := range result.Plans {
		fmt.Printf("\n%s (%s):\n", plan.Manifest.Path, plan.Manifest.Type)
		fmt.Printf("%-40s %-15s %-15s %-10s\n", "Package", "Current", "Target", "Impact")
		fmt.Println(strings.Repeat("-", 80))

		for i := range plan.Updates {
			update := &plan.Updates[i]
			pkg := update.Dependency.Name
			if len(pkg) > 40 {
				pkg = pkg[:37] + "..."
			}
			fmt.Printf("%-40s %-15s %-15s %-10s\n",
				pkg,
				update.Dependency.CurrentVersion,
				update.TargetVersion,
				update.Impact)
		}
	}

	totalUpdates := 0
	for _, plan := range result.Plans {
		totalUpdates += len(plan.Updates)
	}

	fmt.Printf("\nTotal: %d updates across %d manifests\n", totalUpdates, len(result.Plans))

	if len(result.Errors) > 0 {
		fmt.Printf("\nErrors:\n")
		for _, e := range result.Errors {
			fmt.Printf("  - %s\n", e)
		}
	}

	return nil
}

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
	scanFormat  string
	scanOnly    string
	scanExclude string
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Discover dependency manifests",
	Long: `Scan the repository for dependency manifests across all supported ecosystems.

This command walks the repository tree and detects manifest files like:
  - package.json (npm)
  - Chart.yaml (Helm)
  - .pre-commit-config.yaml (pre-commit)
  - mise.toml, .mise.toml (mise)
  - .tool-versions (asdf)
  - .tflint.hcl (tflint)
  - main.tf, *.tf (Terraform)

Results can be output in table or JSON format.`,
	Example: `  # Scan all manifests
  uptool scan

  # Scan with JSON output
  uptool scan --format json

  # Scan only npm and helm
  uptool scan --only npm,helm

  # Scan everything except terraform
  uptool scan --exclude terraform`,
	RunE: runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().StringVarP(&scanFormat, "format", "f", "table", "output format: table, json")
	scanCmd.Flags().StringVar(&scanOnly, "only", "", "comma-separated integrations to include")
	scanCmd.Flags().StringVar(&scanExclude, "exclude", "", "comma-separated integrations to exclude")

	// Add shell completion for flags
	if err := scanCmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"table", "json"}, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		// This is a non-critical error during CLI initialization
		fmt.Fprintf(os.Stderr, "Warning: failed to register shell completion: %v\n", err)
	}

	if err := scanCmd.RegisterFlagCompletionFunc("only", completeIntegrations); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to register shell completion: %v\n", err)
	}
	if err := scanCmd.RegisterFlagCompletionFunc("exclude", completeIntegrations); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to register shell completion: %v\n", err)
	}
}

func runScan(cmd *cobra.Command, args []string) error {
	eng := setupEngine()
	ctx := context.Background()

	repoRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	onlyList, excludeList := parseFilters(scanOnly, scanExclude)

	result, err := eng.Scan(ctx, repoRoot, onlyList, excludeList)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	switch scanFormat {
	case "json":
		return outputJSON(result)
	case "table":
		return outputScanTable(result)
	default:
		return fmt.Errorf("unsupported format: %s", scanFormat)
	}
}

func outputScanTable(result *engine.ScanResult) error {
	if len(result.Manifests) == 0 {
		fmt.Println("No manifests found.")
		return nil
	}

	fmt.Printf("%-20s %-50s %-10s\n", "Type", "Path", "Dependencies")
	fmt.Println(strings.Repeat("-", 80))

	for _, m := range result.Manifests {
		path := m.Path
		if len(path) > 50 {
			path = "..." + path[len(path)-47:]
		}
		fmt.Printf("%-20s %-50s %-10d\n", m.Type, path, len(m.Dependencies))
	}

	fmt.Printf("\nTotal: %d manifests\n", len(result.Manifests))

	if len(result.Errors) > 0 {
		fmt.Printf("\nErrors:\n")
		for _, e := range result.Errors {
			fmt.Printf("  - %s\n", e)
		}
	}

	return nil
}

func outputJSON(v interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

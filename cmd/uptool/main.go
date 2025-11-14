// uptool is a manifest-first dependency updater for multiple ecosystems.
// It scans repositories for dependency manifest files (package.json, Chart.yaml, .pre-commit-config.yaml, etc.),
// checks for available updates, and rewrites manifests with new versions while preserving formatting.
//
// Usage:
//
//	uptool scan              Discover manifests and dependencies
//	uptool plan              Show available updates
//	uptool update            Apply updates to manifests
//	uptool version           Show version information
//	uptool help              Show usage information
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/santosr2/uptool/internal/engine"
	"github.com/santosr2/uptool/internal/integrations"
	_ "github.com/santosr2/uptool/internal/integrations/all"
	"github.com/santosr2/uptool/internal/policy"
)

const version = "0.5.0"

var (
	quietFlag   bool
	verboseFlag bool
	logLevel    slog.Level = slog.LevelWarn // Default to WARN for CLI
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		printUsage()
		return nil
	}

	command := os.Args[1]

	// Parse global flags
	parseGlobalFlags()

	switch command {
	case "scan":
		return runScan()
	case "plan":
		return runPlan()
	case "update":
		return runUpdate()
	case "version":
		fmt.Printf("uptool version %s\n", version)
		return nil
	case "help", "-h", "--help":
		printUsage()
		return nil
	default:
		printUsage()
		return fmt.Errorf("unknown command: %s", command)
	}
}

func parseGlobalFlags() {
	for i, arg := range os.Args {
		if arg == "-q" || arg == "--quiet" {
			quietFlag = true
			logLevel = slog.LevelError
			// Remove flag from os.Args
			os.Args = append(os.Args[:i], os.Args[i+1:]...)
			break
		} else if arg == "-v" || arg == "--verbose" {
			verboseFlag = true
			logLevel = slog.LevelDebug
			// Remove flag from os.Args
			os.Args = append(os.Args[:i], os.Args[i+1:]...)
			break
		}
	}
}

func printUsage() {
	fmt.Printf(`uptool - Universal Manifest-First Dependency Updater v%s

Usage:
  uptool [global-options] <command> [options]

Global Options:
  -q, --quiet        Suppress informational output (errors only)
  -v, --verbose      Enable verbose debug output

Commands:
  scan      Discover dependency manifests
  plan      Generate update plans
  update    Apply updates to manifests
  version   Show version
  help      Show this help

Scan Options:
  -format string     Output format: table, json (default: table)
  -only string       Comma-separated integrations to include
  -exclude string    Comma-separated integrations to exclude

Plan Options:
  -format string     Output format: table, json (default: table)
  -out string        Write plan to file
  -only string       Comma-separated integrations to include
  -exclude string    Comma-separated integrations to exclude

Update Options:
  -dry-run           Show what would be updated without applying
  -diff              Show diffs of changes
  -only string       Comma-separated integrations to include
  -exclude string    Comma-separated integrations to exclude

Examples:
  uptool scan
  uptool --quiet plan --format=json
  uptool --verbose plan --only=npm
  uptool update --dry-run --diff
  uptool update --exclude=npm

Supported Integrations:
  asdf        - .tool-versions, mise.toml (runtime versions) [EXPERIMENTAL]
  helm        - Chart.yaml (Kubernetes Helm charts)
  npm         - package.json (JavaScript/TypeScript)
  precommit   - .pre-commit-config.yaml (uses native autoupdate)
  terraform   - *.tf files (Terraform modules)
  tflint      - .tflint.hcl (TFLint plugins)

`, version)
}

func setupEngine() *engine.Engine {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))

	eng := engine.NewEngine(logger)

	// Load configuration if available
	var cfg *policy.Config
	configPath := filepath.Join(".", "uptool.yaml")
	if _, err := os.Stat(configPath); err == nil {
		var loadErr error
		cfg, loadErr = policy.LoadConfig(configPath)
		if loadErr != nil {
			logger.Warn("failed to load config, using defaults", "error", loadErr)
			cfg = nil
		} else {
			logger.Debug("loaded configuration", "path", configPath)
		}
	}

	// Get all registered integrations from the global registry
	allIntegrations := integrations.GetAll()

	// Register integrations based on config
	if cfg != nil {
		// Use config to determine which integrations to enable
		enabledMap := make(map[string]bool)
		for _, ic := range cfg.Integrations {
			if ic.Enabled {
				enabledMap[ic.ID] = true
			}
		}

		for id, integration := range allIntegrations {
			if enabledMap[id] {
				eng.Register(integration)
				logger.Debug("registered integration from config", "id", id)
			} else {
				logger.Debug("skipped integration (disabled in config)", "id", id)
			}
		}
	} else {
		// No config file - register all integrations
		for _, integration := range allIntegrations {
			eng.Register(integration)
		}
	}

	return eng
}

func parseFilters(only, exclude string) ([]string, []string) {
	var onlyList, excludeList []string

	if only != "" {
		onlyList = strings.Split(only, ",")
		for i := range onlyList {
			onlyList[i] = strings.TrimSpace(onlyList[i])
		}
	}

	if exclude != "" {
		excludeList = strings.Split(exclude, ",")
		for i := range excludeList {
			excludeList[i] = strings.TrimSpace(excludeList[i])
		}
	}

	return onlyList, excludeList
}

func runScan() error {
	fs := flag.NewFlagSet("scan", flag.ExitOnError)
	format := fs.String("format", "table", "Output format: table, json")
	only := fs.String("only", "", "Comma-separated integrations to include")
	exclude := fs.String("exclude", "", "Comma-separated integrations to exclude")

	if err := fs.Parse(os.Args[2:]); err != nil {
		return err
	}

	eng := setupEngine()
	ctx := context.Background()

	repoRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	onlyList, excludeList := parseFilters(*only, *exclude)

	result, err := eng.Scan(ctx, repoRoot, onlyList, excludeList)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	switch *format {
	case "json":
		return outputJSON(result)
	case "table":
		return outputScanTable(result)
	default:
		return fmt.Errorf("unsupported format: %s", *format)
	}
}

func runPlan() error {
	fs := flag.NewFlagSet("plan", flag.ExitOnError)
	format := fs.String("format", "table", "Output format: table, json")
	outFile := fs.String("out", "", "Write plan to file")
	only := fs.String("only", "", "Comma-separated integrations to include")
	exclude := fs.String("exclude", "", "Comma-separated integrations to exclude")

	if err := fs.Parse(os.Args[2:]); err != nil {
		return err
	}

	eng := setupEngine()
	ctx := context.Background()

	repoRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	onlyList, excludeList := parseFilters(*only, *exclude)

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
	if *outFile != "" {
		data, err := json.MarshalIndent(planResult, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal plan: %w", err)
		}
		if err := os.WriteFile(*outFile, data, 0600); err != nil {
			return fmt.Errorf("write plan file: %w", err)
		}
		fmt.Printf("Plan written to %s\n", *outFile)
	}

	switch *format {
	case "json":
		return outputJSON(planResult)
	case "table":
		return outputPlanTable(planResult)
	default:
		return fmt.Errorf("unsupported format: %s", *format)
	}
}

func runUpdate() error {
	fs := flag.NewFlagSet("update", flag.ExitOnError)
	dryRun := fs.Bool("dry-run", false, "Show changes without applying")
	showDiff := fs.Bool("diff", false, "Show diffs")
	only := fs.String("only", "", "Comma-separated integrations to include")
	exclude := fs.String("exclude", "", "Comma-separated integrations to exclude")

	if err := fs.Parse(os.Args[2:]); err != nil {
		return err
	}

	eng := setupEngine()
	ctx := context.Background()

	repoRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	onlyList, excludeList := parseFilters(*only, *exclude)

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

	if *dryRun {
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

		if *showDiff && result.ManifestDiff != "" {
			fmt.Printf("\nDiff:\n%s\n", result.ManifestDiff)
		}
	}

	return nil
}

func outputJSON(v interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
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

func outputPlanTable(result *engine.PlanResult) error {
	if len(result.Plans) == 0 {
		fmt.Println("No updates available.")
		return nil
	}

	for _, plan := range result.Plans {
		fmt.Printf("\n%s (%s):\n", plan.Manifest.Path, plan.Manifest.Type)
		fmt.Printf("%-40s %-15s %-15s %-10s\n", "Package", "Current", "Target", "Impact")
		fmt.Println(strings.Repeat("-", 80))

		for _, update := range plan.Updates {
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

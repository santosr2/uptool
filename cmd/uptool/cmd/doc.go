// Package cmd implements the command-line interface for uptool.
// It provides commands for scanning, planning, and updating dependency manifests
// across multiple ecosystems (npm, Helm, Terraform, pre-commit, asdf, mise, tflint).
//
// The CLI is built using Cobra and provides the following commands:
//
//   - scan: Discover all manifests in a repository
//   - plan: Generate an update plan showing available dependency updates
//   - update: Apply updates to manifest files
//   - list: List all supported integrations and their status
//   - completion: Generate shell completion scripts
//
// Global flags available across all commands:
//
//   - -v, --verbose: Enable verbose debug output
//   - -q, --quiet: Suppress informational output (errors only)
//
// Example usage:
//
//	# Scan repository for manifests
//	uptool scan
//
//	# Generate update plan
//	uptool plan
//
//	# Apply updates (dry-run first)
//	uptool update --dry-run --diff
//	uptool update
//
//	# Update only specific integrations
//	uptool update --only=npm,helm
//
// See individual command documentation for detailed usage and options.
package cmd

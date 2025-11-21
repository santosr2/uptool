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

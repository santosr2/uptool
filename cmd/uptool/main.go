// uptool is a manifest-first dependency updater for multiple ecosystems.
// It scans repositories for dependency manifest files (package.json, Chart.yaml, .pre-commit-config.yaml, etc.),
// checks for available updates, and rewrites manifests with new versions while preserving formatting.
package main

import (
	"fmt"
	"os"

	"github.com/santosr2/uptool/cmd/uptool/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

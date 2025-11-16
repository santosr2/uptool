package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/santosr2/uptool/internal/version"
	"github.com/spf13/cobra"
)

var (
	quietFlag   bool
	verboseFlag bool
	logLevel    = slog.LevelWarn

	rootCmd = &cobra.Command{
		Use:   "uptool",
		Short: "Universal Manifest-First Dependency Updater",
		Long: `uptool is a manifest-first dependency updater for multiple ecosystems.
It scans repositories for dependency manifest files (package.json, Chart.yaml,
.pre-commit-config.yaml, etc.), checks for available updates, and rewrites
manifests with new versions while preserving formatting.`,
		Version: version.Get(),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Set log level based on flags
			if quietFlag {
				logLevel = slog.LevelError
			} else if verboseFlag {
				logLevel = slog.LevelDebug
			}
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
)

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&quietFlag, "quiet", "q", false, "suppress informational output (errors only)")
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "enable verbose debug output")
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// GetLogLevel returns the current log level based on flags
func GetLogLevel() slog.Level {
	return logLevel
}

// printError prints error to stderr with proper formatting
func printError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}

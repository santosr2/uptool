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
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/santosr2/uptool/internal/version"
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

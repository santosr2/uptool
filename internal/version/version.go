// Package version provides version information for uptool.
// The version is embedded from the VERSION file at the repository root.
package version

import (
	_ "embed"
	"strings"
)

//go:embed VERSION
var versionFile string

// version holds the current uptool version, read from the embedded VERSION file.
// Can be overridden at build time using -ldflags "-X internal/version.version=X.Y.Z"
var version = strings.TrimSpace(versionFile)

// Get returns the current uptool version.
func Get() string {
	if version == "" {
		return "dev"
	}
	return version
}

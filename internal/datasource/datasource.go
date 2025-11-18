// Package datasource provides a unified interface for querying package registries.
// This abstraction allows multiple integrations to share the same registry client.
package datasource

import (
	"context"
	"fmt"
	"sync"
)

// Datasource represents a package registry or version source.
type Datasource interface {
	// Name returns the datasource identifier (e.g., "npm", "pypi", "github-releases")
	Name() string

	// GetLatestVersion returns the latest stable version for a package.
	GetLatestVersion(ctx context.Context, pkg string) (string, error)

	// GetVersions returns all available versions for a package.
	GetVersions(ctx context.Context, pkg string) ([]string, error)

	// GetPackageInfo returns detailed information about a package.
	GetPackageInfo(ctx context.Context, pkg string) (*PackageInfo, error)
}

// PackageInfo contains metadata about a package.
type PackageInfo struct {
	Name        string
	Description string
	Homepage    string
	Repository  string
	Versions    []VersionInfo
}

// VersionInfo contains metadata about a specific version.
type VersionInfo struct {
	Version      string
	PublishedAt  string
	IsPrerelease bool
	Deprecated   bool
}

var (
	datasources = make(map[string]Datasource)
	mu          sync.RWMutex
)

// Register adds a datasource to the global registry.
func Register(ds Datasource) {
	mu.Lock()
	defer mu.Unlock()

	name := ds.Name()
	if _, exists := datasources[name]; exists {
		panic(fmt.Sprintf("datasource already registered: %s", name))
	}

	datasources[name] = ds
}

// Get returns a datasource by name.
func Get(name string) (Datasource, error) {
	mu.RLock()
	defer mu.RUnlock()

	ds, ok := datasources[name]
	if !ok {
		return nil, fmt.Errorf("datasource %q not found", name)
	}

	return ds, nil
}

// List returns all registered datasource names.
func List() []string {
	mu.RLock()
	defer mu.RUnlock()

	names := make([]string, 0, len(datasources))
	for name := range datasources {
		names = append(names, name)
	}

	return names
}

package datasource

import (
	"context"

	"github.com/santosr2/uptool/internal/registry"
)

func init() {
	Register(NewNPMDatasource())
}

// NPMDatasource implements the Datasource interface for the npm registry.
type NPMDatasource struct {
	client *registry.NPMClient
}

// NewNPMDatasource creates a new npm datasource.
func NewNPMDatasource() *NPMDatasource {
	return &NPMDatasource{
		client: registry.NewNPMClient(),
	}
}

// Name returns the datasource identifier.
func (d *NPMDatasource) Name() string {
	return "npm"
}

// GetLatestVersion returns the latest stable version for an npm package.
func (d *NPMDatasource) GetLatestVersion(ctx context.Context, pkg string) (string, error) {
	return d.client.GetLatestVersion(ctx, pkg)
}

// GetVersions returns all available versions for an npm package.
func (d *NPMDatasource) GetVersions(ctx context.Context, pkg string) ([]string, error) {
	info, err := d.client.GetPackageInfo(ctx, pkg)
	if err != nil {
		return nil, err
	}

	// info.Versions is a map[string]map[string]interface{}
	versions := make([]string, 0, len(info.Versions))
	for version := range info.Versions {
		versions = append(versions, version)
	}

	return versions, nil
}

// GetPackageInfo returns detailed information about an npm package.
func (d *NPMDatasource) GetPackageInfo(ctx context.Context, pkg string) (*PackageInfo, error) {
	info, err := d.client.GetPackageInfo(ctx, pkg)
	if err != nil {
		return nil, err
	}

	// Convert registry.PackageInfo to datasource.PackageInfo
	versions := make([]VersionInfo, 0, len(info.Versions))
	for version, versionData := range info.Versions {
		publishDate := ""
		if info.Time != nil {
			if t, ok := info.Time[version]; ok {
				publishDate = t
			}
		}

		deprecated := false
		if versionData != nil {
			if dep, ok := versionData["deprecated"]; ok && dep != nil {
				deprecated = true
			}
		}

		versions = append(versions, VersionInfo{
			Version:     version,
			PublishedAt: publishDate,
			Deprecated:  deprecated,
		})
	}

	return &PackageInfo{
		Name:        info.Name,
		Description: "", // Not exposed in registry.PackageInfo
		Homepage:    "", // Not exposed in registry.PackageInfo
		Repository:  "", // Not exposed in registry.PackageInfo
		Versions:    versions,
	}, nil
}

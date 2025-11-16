package datasource

import (
	"context"
	"strings"

	"github.com/santosr2/uptool/internal/registry"
)

func init() {
	Register(NewHelmDatasource())
}

// HelmDatasource implements the Datasource interface for Helm chart repositories.
type HelmDatasource struct {
	client *registry.HelmClient
}

// NewHelmDatasource creates a new Helm datasource.
func NewHelmDatasource() *HelmDatasource {
	return &HelmDatasource{
		client: registry.NewHelmClient(),
	}
}

// Name returns the datasource identifier.
func (d *HelmDatasource) Name() string {
	return "helm"
}

// GetLatestVersion returns the latest stable version for a Helm chart.
func (d *HelmDatasource) GetLatestVersion(ctx context.Context, pkg string) (string, error) {
	// pkg format: "repository_url|chart_name"
	parts := strings.Split(pkg, "|")
	if len(parts) != 2 {
		return "", nil
	}

	return d.client.GetLatestChartVersion(ctx, parts[0], parts[1])
}

// GetVersions returns all available versions for a Helm chart.
func (d *HelmDatasource) GetVersions(ctx context.Context, pkg string) ([]string, error) {
	parts := strings.Split(pkg, "|")
	if len(parts) != 2 {
		return nil, nil
	}

	return d.client.GetChartVersions(ctx, parts[0], parts[1])
}

// GetPackageInfo returns detailed information about a Helm chart.
func (d *HelmDatasource) GetPackageInfo(ctx context.Context, pkg string) (*PackageInfo, error) {
	parts := strings.Split(pkg, "|")
	if len(parts) != 2 {
		return nil, nil
	}

	versions, err := d.client.GetChartVersionDetails(ctx, parts[0], parts[1])
	if err != nil {
		return nil, err
	}

	versionInfos := make([]VersionInfo, len(versions))
	for i, v := range versions {
		versionInfos[i] = VersionInfo{
			Version:     v.Version,
			PublishedAt: v.Created.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return &PackageInfo{
		Name:     parts[1],
		Versions: versionInfos,
	}, nil
}

package datasource

import (
	"context"

	"github.com/santosr2/uptool/internal/registry"
)

func init() {
	Register(NewTerraformDatasource())
}

// TerraformDatasource implements the Datasource interface for the Terraform Registry.
type TerraformDatasource struct {
	client *registry.TerraformClient
}

// NewTerraformDatasource creates a new Terraform datasource.
func NewTerraformDatasource() *TerraformDatasource {
	return &TerraformDatasource{
		client: registry.NewTerraformClient(),
	}
}

// Name returns the datasource identifier.
func (d *TerraformDatasource) Name() string {
	return "terraform"
}

// GetLatestVersion returns the latest stable version for a Terraform module or provider.
func (d *TerraformDatasource) GetLatestVersion(ctx context.Context, pkg string) (string, error) {
	// pkg format: "namespace/name/provider" for modules
	// or "namespace/name" for providers
	return d.client.GetLatestModuleVersion(ctx, pkg)
}

// GetVersions returns all available versions for a Terraform module or provider.
func (d *TerraformDatasource) GetVersions(ctx context.Context, pkg string) ([]string, error) {
	versions, err := d.client.GetModuleVersions(ctx, pkg)
	if err != nil {
		return nil, err
	}

	result := make([]string, len(versions))
	for i, v := range versions {
		result[i] = v.Version
	}

	return result, nil
}

// GetPackageInfo returns detailed information about a Terraform module or provider.
func (d *TerraformDatasource) GetPackageInfo(ctx context.Context, pkg string) (*PackageInfo, error) {
	versions, err := d.client.GetModuleVersions(ctx, pkg)
	if err != nil {
		return nil, err
	}

	versionInfos := make([]VersionInfo, len(versions))
	for i, v := range versions {
		versionInfos[i] = VersionInfo{
			Version: v.Version,
		}
	}

	return &PackageInfo{
		Name:     pkg,
		Versions: versionInfos,
	}, nil
}

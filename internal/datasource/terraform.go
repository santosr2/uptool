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

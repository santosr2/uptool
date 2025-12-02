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
	Register(NewGoDatasource())
}

// GoDatasource implements the Datasource interface for the Go module proxy.
type GoDatasource struct {
	client *registry.GoClient
}

// NewGoDatasource creates a new Go module datasource.
func NewGoDatasource() *GoDatasource {
	return &GoDatasource{
		client: registry.NewGoClient(),
	}
}

// Name returns the datasource identifier.
func (d *GoDatasource) Name() string {
	return "go"
}

// GetLatestVersion returns the latest stable version for a Go module.
func (d *GoDatasource) GetLatestVersion(ctx context.Context, pkg string) (string, error) {
	return d.client.GetLatestVersion(ctx, pkg)
}

// GetVersions returns all available versions for a Go module.
func (d *GoDatasource) GetVersions(ctx context.Context, pkg string) ([]string, error) {
	return d.client.GetVersions(ctx, pkg)
}

// GetPackageInfo returns detailed information about a Go module.
func (d *GoDatasource) GetPackageInfo(ctx context.Context, pkg string) (*PackageInfo, error) {
	versions, err := d.client.GetVersions(ctx, pkg)
	if err != nil {
		return nil, err
	}

	versionInfos := make([]VersionInfo, 0, len(versions))
	for _, v := range versions {
		info, err := d.client.GetModuleInfo(ctx, pkg, v)

		vi := VersionInfo{
			Version: v,
		}

		if err == nil && info != nil {
			vi.PublishedAt = info.Time.Format("2006-01-02T15:04:05Z")
		}

		// Check if it's a prerelease (contains -alpha, -beta, -rc, etc.)
		vi.IsPrerelease = isGoPrerelease(v)

		versionInfos = append(versionInfos, vi)
	}

	return &PackageInfo{
		Name:        pkg,
		Description: "",
		Homepage:    "https://pkg.go.dev/" + pkg,
		Repository:  "https://" + pkg,
		Versions:    versionInfos,
	}, nil
}

// isGoPrerelease checks if a Go module version is a prerelease.
func isGoPrerelease(version string) bool {
	// Go modules use standard semver prerelease format: v1.2.3-alpha, v1.2.3-beta.1, v1.2.3-rc.1
	for _, pre := range []string{"-alpha", "-beta", "-rc", "-pre", "-dev"} {
		if contains(version, pre) {
			return true
		}
	}
	return false
}

// contains checks if a string contains a substring (case-insensitive).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || s != "" && containsAt(s, substr, 0))
}

// containsAt recursively checks if s contains substr starting from index i.
func containsAt(s, substr string, i int) bool {
	if i+len(substr) > len(s) {
		return false
	}
	if matchesAt(s, substr, i) {
		return true
	}
	return containsAt(s, substr, i+1)
}

// matchesAt checks if s matches substr at position i.
func matchesAt(s, substr string, i int) bool {
	for j := 0; j < len(substr); j++ {
		sc := s[i+j]
		subc := substr[j]
		// Case-insensitive comparison
		if sc != subc && sc != subc+32 && sc != subc-32 {
			return false
		}
	}
	return true
}

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
	"os"
	"strings"

	"github.com/santosr2/uptool/internal/registry"
)

func init() {
	Register(NewGitHubDatasource())
}

// GitHubDatasource implements the Datasource interface for GitHub Releases.
type GitHubDatasource struct {
	client *registry.GitHubClient
}

// NewGitHubDatasource creates a new GitHub datasource.
func NewGitHubDatasource() *GitHubDatasource {
	token := os.Getenv("GITHUB_TOKEN")
	return &GitHubDatasource{
		client: registry.NewGitHubClient(token),
	}
}

// Name returns the datasource identifier.
func (d *GitHubDatasource) Name() string {
	return "github-releases"
}

// GetLatestVersion returns the latest stable release for a GitHub repository.
func (d *GitHubDatasource) GetLatestVersion(ctx context.Context, pkg string) (string, error) {
	// pkg format: "owner/repo"
	parts := strings.Split(pkg, "/")
	if len(parts) != 2 {
		return "", nil
	}

	version, err := d.client.GetLatestRelease(ctx, parts[0], parts[1])
	if err != nil {
		return "", err
	}

	return version, nil
}

// GetVersions returns all available releases for a GitHub repository.
func (d *GitHubDatasource) GetVersions(ctx context.Context, pkg string) ([]string, error) {
	// pkg format: "owner/repo"
	parts := strings.Split(pkg, "/")
	if len(parts) != 2 {
		return nil, nil
	}

	releases, err := d.client.GetAllReleases(ctx, parts[0], parts[1])
	if err != nil {
		return nil, err
	}

	versions := make([]string, 0, len(releases))
	for _, rel := range releases {
		// Skip drafts and prereleases
		if rel.Draft || rel.Prerelease {
			continue
		}

		// Strip 'v' prefix if present
		version := strings.TrimPrefix(rel.TagName, "v")
		versions = append(versions, version)
	}

	return versions, nil
}

// GetPackageInfo returns detailed information about a GitHub repository's releases.
func (d *GitHubDatasource) GetPackageInfo(ctx context.Context, pkg string) (*PackageInfo, error) {
	// pkg format: "owner/repo"
	parts := strings.Split(pkg, "/")
	if len(parts) != 2 {
		return nil, nil
	}

	releases, err := d.client.GetAllReleases(ctx, parts[0], parts[1])
	if err != nil {
		return nil, err
	}

	versions := make([]VersionInfo, 0, len(releases))
	for _, rel := range releases {
		version := strings.TrimPrefix(rel.TagName, "v")

		versions = append(versions, VersionInfo{
			Version:      version,
			PublishedAt:  rel.PublishedAt,
			IsPrerelease: rel.Prerelease,
			Deprecated:   false, // GitHub doesn't track deprecated releases
		})
	}

	// Build repository URL
	repoURL := "https://github.com/" + pkg

	return &PackageInfo{
		Name:       pkg,
		Repository: repoURL,
		Homepage:   repoURL,
		Versions:   versions,
	}, nil
}

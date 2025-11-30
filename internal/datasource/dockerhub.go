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

//nolint:govet,errcheck // JSON field order is intentional for readability
package datasource

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"
)

func init() {
	Register(NewDockerHubDatasource())
}

// DockerHubDatasource implements the Datasource interface for Docker Hub.
type DockerHubDatasource struct {
	client  *http.Client
	baseURL string
}

// NewDockerHubDatasource creates a new Docker Hub datasource.
func NewDockerHubDatasource() *DockerHubDatasource {
	return &DockerHubDatasource{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://hub.docker.com/v2",
	}
}

// Name returns the datasource identifier.
func (d *DockerHubDatasource) Name() string {
	return "docker-hub"
}

// dockerHubTagsResponse represents the Docker Hub API response for tags.
type dockerHubTagsResponse struct {
	Next     string         `json:"next"`
	Previous string         `json:"previous"`
	Results  []dockerHubTag `json:"results"`
	Count    int            `json:"count"`
}

// dockerHubTag represents a single tag from Docker Hub.
type dockerHubTag struct {
	Name        string    `json:"name"`
	LastUpdated time.Time `json:"last_updated"`
	FullSize    int64     `json:"full_size"`
}

// semverPattern matches semantic version tags
var semverPattern = regexp.MustCompile(`^v?(\d+)(?:\.(\d+))?(?:\.(\d+))?(?:[-.]?(?:alpha|beta|rc|pre)[-.]?\d*)?$`)

// GetLatestVersion returns the latest stable version for a Docker image.
func (d *DockerHubDatasource) GetLatestVersion(ctx context.Context, pkg string) (string, error) {
	versions, err := d.GetVersions(ctx, pkg)
	if err != nil {
		return "", err
	}

	if len(versions) == 0 {
		return "", fmt.Errorf("no versions found for %s", pkg)
	}

	// Return the first (latest) version
	return versions[0], nil
}

// GetVersions returns all available tags for a Docker image.
func (d *DockerHubDatasource) GetVersions(ctx context.Context, pkg string) ([]string, error) {
	// Normalize image name
	namespace, repo := normalizeImageName(pkg)

	url := fmt.Sprintf("%s/repositories/%s/%s/tags?page_size=100", d.baseURL, namespace, repo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			_ = closeErr // Ignore close error
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("docker Hub API returned status %d", resp.StatusCode)
	}

	var tagsResp dockerHubTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		return nil, err
	}

	// Filter and sort tags
	versions := make([]string, 0, len(tagsResp.Results))
	for _, tag := range tagsResp.Results {
		// Skip non-semver tags like "latest", "alpine", "slim"
		if !isSemverTag(tag.Name) {
			continue
		}
		versions = append(versions, tag.Name)
	}

	// Sort versions in descending order (newest first)
	sort.Slice(versions, func(i, j int) bool {
		return compareVersions(versions[i], versions[j]) > 0
	})

	return versions, nil
}

// GetPackageInfo returns detailed information about a Docker image.
func (d *DockerHubDatasource) GetPackageInfo(ctx context.Context, pkg string) (*PackageInfo, error) {
	versions, err := d.GetVersions(ctx, pkg)
	if err != nil {
		return nil, err
	}

	namespace, repo := normalizeImageName(pkg)

	versionInfos := make([]VersionInfo, 0, len(versions))
	for _, v := range versions {
		versionInfos = append(versionInfos, VersionInfo{
			Version:      v,
			IsPrerelease: isPrerelease(v),
		})
	}

	return &PackageInfo{
		Name:       pkg,
		Repository: fmt.Sprintf("https://hub.docker.com/r/%s/%s", namespace, repo),
		Homepage:   fmt.Sprintf("https://hub.docker.com/r/%s/%s", namespace, repo),
		Versions:   versionInfos,
	}, nil
}

// normalizeImageName normalizes a Docker image name to namespace/repo format.
func normalizeImageName(image string) (namespace, repo string) {
	// Handle official images (no namespace)
	if !strings.Contains(image, "/") {
		return "library", image
	}

	// Handle registry prefixes (e.g., gcr.io/project/image)
	parts := strings.Split(image, "/")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}

	// For images with registry prefix, use the last two parts
	if len(parts) >= 2 {
		return parts[len(parts)-2], parts[len(parts)-1]
	}

	return "library", image
}

// isSemverTag checks if a tag looks like a semantic version.
func isSemverTag(tag string) bool {
	// Skip common non-version tags
	nonVersionTags := []string{"latest", "edge", "nightly", "master", "main", "develop", "dev"}
	for _, nvt := range nonVersionTags {
		if tag == nvt {
			return false
		}
	}

	// Check if it matches semver pattern
	return semverPattern.MatchString(tag)
}

// isPrerelease checks if a version is a prerelease.
func isPrerelease(version string) bool {
	lower := strings.ToLower(version)
	return strings.Contains(lower, "alpha") ||
		strings.Contains(lower, "beta") ||
		strings.Contains(lower, "rc") ||
		strings.Contains(lower, "pre")
}

// compareVersions compares two version strings.
// Returns positive if v1 > v2, negative if v1 < v2, 0 if equal.
func compareVersions(v1, v2 string) int {
	// Strip common prefixes
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(parts1) {
			// Extract numeric part
			_, _ = fmt.Sscanf(parts1[i], "%d", &n1)
		}
		if i < len(parts2) {
			_, _ = fmt.Sscanf(parts2[i], "%d", &n2)
		}

		if n1 != n2 {
			return n1 - n2
		}
	}

	return 0
}

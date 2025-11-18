// Package registry provides HTTP clients for querying package registries and release APIs.
// It includes clients for npm Registry, Terraform Registry, GitHub Releases, and Helm repositories,
// enabling version lookups and constraint-based version resolution.
package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Masterminds/semver/v3"
)

const npmRegistryURL = "https://registry.npmjs.org"

// NPMClient queries the npm registry for package information.
type NPMClient struct {
	client  *http.Client
	baseURL string
}

// NewNPMClient creates a new npm registry client.
func NewNPMClient() *NPMClient {
	return &NPMClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: npmRegistryURL,
	}
}

// PackageInfo contains npm package metadata.
type PackageInfo struct {
	Versions map[string]map[string]interface{} `json:"versions"`
	DistTags map[string]string                 `json:"dist-tags"`
	Time     map[string]string                 `json:"time"`
	Name     string                            `json:"name"`
}

// GetLatestVersion fetches the latest version for a package.
func (c *NPMClient) GetLatestVersion(ctx context.Context, packageName string) (string, error) {
	info, err := c.GetPackageInfo(ctx, packageName)
	if err != nil {
		return "", err
	}

	// Return the latest dist-tag
	if latest, ok := info.DistTags["latest"]; ok {
		return latest, nil
	}

	return "", fmt.Errorf("no latest version found for %s", packageName)
}

// GetPackageInfo fetches full package information from npm registry.
func (c *NPMClient) GetPackageInfo(ctx context.Context, packageName string) (*PackageInfo, error) {
	url := fmt.Sprintf("%s/%s", c.baseURL, packageName)

	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch package info: %w", err)
	}
	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // HTTP cleanup best effort

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("package not found: %s", packageName)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var info PackageInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &info, nil
}

// FindBestVersion finds the best version matching a constraint.
func (c *NPMClient) FindBestVersion(ctx context.Context, packageName, constraint string, allowPrerelease bool) (string, error) {
	info, err := c.GetPackageInfo(ctx, packageName)
	if err != nil {
		return "", err
	}

	// Parse constraint
	constraintObj, err := semver.NewConstraint(constraint)
	if err != nil {
		// If constraint parsing fails, return latest
		return c.GetLatestVersion(ctx, packageName)
	}

	// Collect all versions
	var versions []*semver.Version
	for v := range info.Versions {
		parsed, err := semver.NewVersion(v)
		if err != nil {
			continue
		}

		// Filter prereleases
		if parsed.Prerelease() != "" && !allowPrerelease {
			continue
		}

		if constraintObj.Check(parsed) {
			versions = append(versions, parsed)
		}
	}

	if len(versions) == 0 {
		return "", fmt.Errorf("no versions match constraint: %s", constraint)
	}

	// Sort and return the highest
	var best *semver.Version
	for _, v := range versions {
		if best == nil || v.GreaterThan(best) {
			best = v
		}
	}

	return best.Original(), nil
}

// GetVersions returns all available versions for a package.
func (c *NPMClient) GetVersions(ctx context.Context, packageName string) ([]string, error) {
	info, err := c.GetPackageInfo(ctx, packageName)
	if err != nil {
		return nil, err
	}

	versions := make([]string, 0, len(info.Versions))
	for v := range info.Versions {
		versions = append(versions, v)
	}

	return versions, nil
}

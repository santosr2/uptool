package registry

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"gopkg.in/yaml.v3"
)

// HelmClient queries Helm chart repositories.
type HelmClient struct {
	client *http.Client
}

// NewHelmClient creates a new Helm chart repository client.
func NewHelmClient() *HelmClient {
	return &HelmClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ChartIndex represents the index.yaml structure from a Helm repository.
type ChartIndex struct {
	Entries    map[string][]ChartIndexEntry `yaml:"entries"`
	APIVersion string                       `yaml:"apiVersion"`
}

// ChartIndexEntry represents a single chart version entry.
type ChartIndexEntry struct {
	Created     time.Time `yaml:"created"`
	Name        string    `yaml:"name"`
	Version     string    `yaml:"version"`
	AppVersion  string    `yaml:"appVersion"`
	Description string    `yaml:"description"`
}

// GetLatestChartVersion fetches the latest version for a chart from a repository.
// repository: the base URL of the chart repository (e.g., "https://charts.bitnami.com/bitnami")
// chartName: the name of the chart (e.g., "postgresql")
func (c *HelmClient) GetLatestChartVersion(ctx context.Context, repository, chartName string) (string, error) {
	// Fetch index.yaml from repository
	indexURL := strings.TrimSuffix(repository, "/") + "/index.yaml"

	req, err := http.NewRequestWithContext(ctx, "GET", indexURL, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/x-yaml")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch chart index: %w", err)
	}
	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // HTTP cleanup best effort

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("chart repository not found: %s", repository)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	var index ChartIndex
	err = yaml.Unmarshal(body, &index)
	if err != nil {
		return "", fmt.Errorf("parse index.yaml: %w", err)
	}

	// Find chart entries
	entries, ok := index.Entries[chartName]
	if !ok || len(entries) == 0 {
		return "", fmt.Errorf("chart not found in repository: %s", chartName)
	}

	// Find the latest non-prerelease version
	var latest *semver.Version
	for _, entry := range entries {
		v, err := semver.NewVersion(entry.Version)
		if err != nil {
			continue
		}

		// Skip prereleases
		if v.Prerelease() != "" {
			continue
		}

		if latest == nil || v.GreaterThan(latest) {
			latest = v
		}
	}

	if latest == nil {
		return "", fmt.Errorf("no stable versions found for chart: %s", chartName)
	}

	return latest.Original(), nil
}

// FindBestChartVersion finds the best chart version matching a constraint.
func (c *HelmClient) FindBestChartVersion(ctx context.Context, repository, chartName, constraint string) (string, error) {
	// Fetch index.yaml from repository
	indexURL := strings.TrimSuffix(repository, "/") + "/index.yaml"

	req, err := http.NewRequestWithContext(ctx, "GET", indexURL, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/x-yaml")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch chart index: %w", err)
	}
	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // HTTP cleanup best effort

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	var index ChartIndex
	err = yaml.Unmarshal(body, &index)
	if err != nil {
		return "", fmt.Errorf("parse index.yaml: %w", err)
	}

	// Find chart entries
	entries, ok := index.Entries[chartName]
	if !ok || len(entries) == 0 {
		return "", fmt.Errorf("chart not found in repository: %s", chartName)
	}

	// Parse constraint
	constraintObj, err := semver.NewConstraint(constraint)
	if err != nil {
		// If constraint parsing fails, return latest
		return c.GetLatestChartVersion(ctx, repository, chartName)
	}

	// Find matching versions
	var versions []*semver.Version
	for _, entry := range entries {
		v, err := semver.NewVersion(entry.Version)
		if err != nil {
			continue
		}

		// Skip prereleases
		if v.Prerelease() != "" {
			continue
		}

		if constraintObj.Check(v) {
			versions = append(versions, v)
		}
	}

	if len(versions) == 0 {
		return "", fmt.Errorf("no versions match constraint: %s", constraint)
	}

	// Find the highest version
	var best *semver.Version
	for _, v := range versions {
		if best == nil || v.GreaterThan(best) {
			best = v
		}
	}

	return best.Original(), nil
}

// GetChartVersionDetails returns all available versions with metadata for a chart from a repository.
func (c *HelmClient) GetChartVersionDetails(ctx context.Context, repository, chartName string) ([]ChartIndexEntry, error) {
	// Fetch index.yaml from repository
	indexURL := strings.TrimSuffix(repository, "/") + "/index.yaml"

	req, err := http.NewRequestWithContext(ctx, "GET", indexURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/x-yaml")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch chart index: %w", err)
	}
	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // HTTP cleanup best effort

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var index ChartIndex
	if err := yaml.Unmarshal(body, &index); err != nil {
		return nil, fmt.Errorf("parse index.yaml: %w", err)
	}

	// Find chart entries
	entries, ok := index.Entries[chartName]
	if !ok {
		return nil, fmt.Errorf("chart not found in repository: %s", chartName)
	}

	return entries, nil
}

// GetChartVersions returns all available versions for a chart.
func (c *HelmClient) GetChartVersions(ctx context.Context, repository, chartName string) ([]string, error) {
	// Fetch index.yaml from repository
	indexURL := strings.TrimSuffix(repository, "/") + "/index.yaml"

	req, err := http.NewRequestWithContext(ctx, "GET", indexURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch chart index: %w", err)
	}
	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // HTTP cleanup best effort

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var index ChartIndex
	if err := yaml.Unmarshal(body, &index); err != nil {
		return nil, fmt.Errorf("parse index.yaml: %w", err)
	}

	// Find chart entries
	entries, ok := index.Entries[chartName]
	if !ok {
		return nil, fmt.Errorf("chart not found in repository: %s", chartName)
	}

	versions := make([]string, 0, len(entries))
	for _, entry := range entries {
		versions = append(versions, entry.Version)
	}

	return versions, nil
}

// IsOCIRepository checks if a repository URL is an OCI registry.
func IsOCIRepository(repository string) bool {
	return strings.HasPrefix(repository, "oci://")
}

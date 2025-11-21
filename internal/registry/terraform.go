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

package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
)

const terraformRegistryURL = "https://registry.terraform.io"

// TerraformClient queries the Terraform Registry API.
type TerraformClient struct {
	client  *http.Client
	baseURL string
}

// NewTerraformClient creates a new Terraform Registry client.
func NewTerraformClient() *TerraformClient {
	return &TerraformClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: terraformRegistryURL,
	}
}

// ProviderVersions represents the response from /v1/providers/{namespace}/{type}/versions.
type ProviderVersions struct {
	Versions []ProviderVersion `json:"versions"`
}

// ProviderVersion represents a single provider version.
type ProviderVersion struct {
	Version   string   `json:"version"`
	Platforms []string `json:"platforms"`
}

// ModuleVersions represents the response from /v1/modules/{namespace}/{name}/{provider}/versions.
type ModuleVersions struct {
	Modules []Module `json:"modules"`
}

// Module represents a module with its versions.
type Module struct {
	Source   string          `json:"source"`
	Versions []ModuleVersion `json:"versions"`
}

// ModuleVersion represents a single module version.
type ModuleVersion struct {
	Version string `json:"version"`
}

// GetLatestProviderVersion fetches the latest version for a provider.
// source format: "namespace/name" (e.g., "hashicorp/aws")
func (c *TerraformClient) GetLatestProviderVersion(ctx context.Context, source string) (string, error) {
	parts := strings.Split(source, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid provider source format: %s", source)
	}

	namespace := parts[0]
	name := parts[1]

	url := fmt.Sprintf("%s/v1/providers/%s/%s/versions", c.baseURL, namespace, name)

	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch provider versions: %w", err)
	}
	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // HTTP cleanup best effort

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("provider not found: %s", source)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	var providerVersions ProviderVersions
	err = json.Unmarshal(body, &providerVersions)
	if err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	if len(providerVersions.Versions) == 0 {
		return "", fmt.Errorf("no versions found for provider: %s", source)
	}

	// Find the latest non-prerelease version
	var latest *semver.Version
	for _, pv := range providerVersions.Versions {
		v, err := semver.NewVersion(pv.Version)
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
		return "", fmt.Errorf("no stable versions found for provider: %s", source)
	}

	return latest.Original(), nil
}

// GetLatestModuleVersion fetches the latest version for a module.
// source format: "namespace/name/provider" (e.g., "terraform-aws-modules/vpc/aws")
func (c *TerraformClient) GetLatestModuleVersion(ctx context.Context, source string) (string, error) {
	parts := strings.Split(source, "/")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid module source format: %s", source)
	}

	namespace := parts[0]
	name := parts[1]
	provider := parts[2]

	url := fmt.Sprintf("%s/v1/modules/%s/%s/%s/versions", c.baseURL, namespace, name, provider)

	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch module versions: %w", err)
	}
	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // HTTP cleanup best effort

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("module not found: %s", source)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	var moduleVersions ModuleVersions
	if err := json.Unmarshal(body, &moduleVersions); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	if len(moduleVersions.Modules) == 0 || len(moduleVersions.Modules[0].Versions) == 0 {
		return "", fmt.Errorf("no versions found for module: %s", source)
	}

	// Find the latest non-prerelease version
	var latest *semver.Version
	for _, mv := range moduleVersions.Modules[0].Versions {
		v, err := semver.NewVersion(mv.Version)
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
		return "", fmt.Errorf("no stable versions found for module: %s", source)
	}

	return latest.Original(), nil
}

// GetModuleVersions returns all available versions for a module.
// source format: "namespace/name/provider" (e.g., "terraform-aws-modules/vpc/aws")
func (c *TerraformClient) GetModuleVersions(ctx context.Context, source string) ([]ModuleVersion, error) {
	parts := strings.Split(source, "/")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid module source format: %s", source)
	}

	namespace := parts[0]
	name := parts[1]
	provider := parts[2]

	url := fmt.Sprintf("%s/v1/modules/%s/%s/%s/versions", c.baseURL, namespace, name, provider)

	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch module versions: %w", err)
	}
	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // HTTP cleanup best effort

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("module not found: %s", source)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var moduleVersions ModuleVersions
	if err := json.Unmarshal(body, &moduleVersions); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if len(moduleVersions.Modules) == 0 {
		return nil, fmt.Errorf("no modules found: %s", source)
	}

	return moduleVersions.Modules[0].Versions, nil
}

// FindBestProviderVersion finds the best provider version matching a constraint.
func (c *TerraformClient) FindBestProviderVersion(ctx context.Context, source, constraint string) (string, error) {
	parts := strings.Split(source, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid provider source format: %s", source)
	}

	namespace := parts[0]
	name := parts[1]

	url := fmt.Sprintf("%s/v1/providers/%s/%s/versions", c.baseURL, namespace, name)

	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch provider versions: %w", err)
	}
	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // HTTP cleanup best effort

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	var providerVersions ProviderVersions
	err = json.Unmarshal(body, &providerVersions)
	if err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	// Parse constraint
	constraintObj, err := semver.NewConstraint(constraint)
	if err != nil {
		// If constraint parsing fails, return latest
		return c.GetLatestProviderVersion(ctx, source)
	}

	// Find matching versions
	var versions []*semver.Version
	for _, pv := range providerVersions.Versions {
		v, err := semver.NewVersion(pv.Version)
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

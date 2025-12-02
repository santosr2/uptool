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
	"net/url"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
)

const goProxyURL = "https://proxy.golang.org"

// GoClient queries the Go module proxy for version information.
type GoClient struct {
	client  *http.Client
	baseURL string
}

// GoModuleInfo represents the JSON response from the Go module proxy @latest endpoint.
type GoModuleInfo struct {
	Time    time.Time `json:"Time"`
	Version string    `json:"Version"`
}

// NewGoClient creates a new Go module proxy client.
func NewGoClient() *GoClient {
	return &GoClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: goProxyURL,
	}
}

// GetLatestVersion fetches the latest version for a Go module.
// It queries the @latest endpoint which returns the highest semver version.
func (c *GoClient) GetLatestVersion(ctx context.Context, modulePath string) (string, error) {
	// URL encode the module path (required for modules with slashes)
	encodedPath := escapeModulePath(modulePath)
	reqURL := fmt.Sprintf("%s/%s/@latest", c.baseURL, encodedPath)

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch module info: %w", err)
	}
	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // HTTP cleanup best effort

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone {
		return "", fmt.Errorf("module not found: %s", modulePath)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	var info GoModuleInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	return info.Version, nil
}

// GetVersions returns all available versions for a Go module.
// It queries the @v/list endpoint which returns newline-separated versions.
func (c *GoClient) GetVersions(ctx context.Context, modulePath string) ([]string, error) {
	encodedPath := escapeModulePath(modulePath)
	reqURL := fmt.Sprintf("%s/%s/@v/list", c.baseURL, encodedPath)

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch version list: %w", err)
	}
	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // HTTP cleanup best effort

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone {
		return nil, fmt.Errorf("module not found: %s", modulePath)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	// Response is newline-separated versions
	lines := strings.Split(strings.TrimSpace(string(body)), "\n")

	versions := make([]string, 0, len(lines))
	for _, line := range lines {
		if v := strings.TrimSpace(line); v != "" {
			versions = append(versions, v)
		}
	}

	return versions, nil
}

// GetModuleInfo fetches detailed information about a specific version of a module.
func (c *GoClient) GetModuleInfo(ctx context.Context, modulePath, version string) (*GoModuleInfo, error) {
	encodedPath := escapeModulePath(modulePath)
	reqURL := fmt.Sprintf("%s/%s/@v/%s.info", c.baseURL, encodedPath, version)

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch module info: %w", err)
	}
	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // HTTP cleanup best effort

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("version not found: %s@%s", modulePath, version)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var info GoModuleInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &info, nil
}

// FindBestVersion finds the best version matching criteria.
func (c *GoClient) FindBestVersion(ctx context.Context, modulePath string, allowPrerelease bool) (string, error) {
	versions, err := c.GetVersions(ctx, modulePath)
	if err != nil {
		return "", err
	}

	if len(versions) == 0 {
		return "", fmt.Errorf("no versions found for %s", modulePath)
	}

	// Parse and filter versions
	semverVersions := make([]*semver.Version, 0, len(versions))
	for _, v := range versions {
		parsed, err := semver.NewVersion(v)
		if err != nil {
			continue
		}

		// Filter prereleases
		if parsed.Prerelease() != "" && !allowPrerelease {
			continue
		}

		semverVersions = append(semverVersions, parsed)
	}

	if len(semverVersions) == 0 {
		// Fall back to latest from API
		return c.GetLatestVersion(ctx, modulePath)
	}

	// Find highest version
	var best *semver.Version
	for _, v := range semverVersions {
		if best == nil || v.GreaterThan(best) {
			best = v
		}
	}

	return best.Original(), nil
}

// escapeModulePath properly encodes a module path for the Go module proxy.
// The Go module proxy uses case-encoding where uppercase letters are escaped
// with an exclamation mark followed by the lowercase letter.
func escapeModulePath(path string) string {
	var builder strings.Builder
	for _, r := range path {
		if r >= 'A' && r <= 'Z' {
			builder.WriteByte('!')
			builder.WriteRune(r + 32) // Convert to lowercase
		} else {
			builder.WriteRune(r)
		}
	}
	// Also URL encode for safety
	return url.PathEscape(builder.String())
}

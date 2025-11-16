package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const pypiURL = "https://pypi.org/pypi"

// PyPIClient queries the PyPI JSON API for package information.
type PyPIClient struct {
	client  *http.Client
	baseURL string
}

// NewPyPIClient creates a new PyPI client.
func NewPyPIClient() *PyPIClient {
	return &PyPIClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: pypiURL,
	}
}

// PyPIResponse represents the PyPI JSON API response.
type PyPIResponse struct {
	Info struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"info"`
	Releases map[string][]struct {
		Yanked bool `json:"yanked"`
	} `json:"releases"`
}

// GetLatestVersion fetches the latest stable version for a package from PyPI.
func (c *PyPIClient) GetLatestVersion(ctx context.Context, packageName string) (string, error) {
	// Construct URL: https://pypi.org/pypi/{package}/json
	url := fmt.Sprintf("%s/%s/json", c.baseURL, packageName)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	// Set User-Agent
	req.Header.Set("User-Agent", "uptool-python-plugin/1.0")

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("querying PyPI: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("PyPI returned status %d for package %s", resp.StatusCode, packageName)
	}

	// Parse response
	var pypiResp PyPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&pypiResp); err != nil {
		return "", fmt.Errorf("parsing PyPI response: %w", err)
	}

	// Return latest version from info.version
	// PyPI's JSON API returns the latest stable version in info.version
	return pypiResp.Info.Version, nil
}

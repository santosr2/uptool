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

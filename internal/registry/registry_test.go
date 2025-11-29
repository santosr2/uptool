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

//nolint:dupl,govet // Test files use similar table-driven patterns; field alignment not critical for tests
package registry

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

// =============================================================================
// NPM Client Tests
// =============================================================================

func TestNPMClient_GetLatestVersion_Mock(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		response    PackageInfo
		statusCode  int
		wantVersion string
		wantErr     bool
	}{
		{
			name:        "successful latest version",
			packageName: "lodash",
			response: PackageInfo{
				Name: "lodash",
				DistTags: map[string]string{
					"latest": "4.17.21",
				},
				Versions: map[string]map[string]interface{}{
					"4.17.21": {},
					"4.17.20": {},
				},
			},
			statusCode:  http.StatusOK,
			wantVersion: "4.17.21",
			wantErr:     false,
		},
		{
			name:        "package not found",
			packageName: "nonexistent-package-xyz",
			statusCode:  http.StatusNotFound,
			wantErr:     true,
		},
		{
			name:        "no latest tag",
			packageName: "test-pkg",
			response: PackageInfo{
				Name:     "test-pkg",
				DistTags: map[string]string{},
				Versions: map[string]map[string]interface{}{
					"1.0.0": {},
				},
			},
			statusCode: http.StatusOK,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == http.StatusOK {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()

			client := &NPMClient{
				client:  &http.Client{Timeout: 5 * time.Second},
				baseURL: server.URL,
			}

			ctx := context.Background()
			version, err := client.GetLatestVersion(ctx, tt.packageName)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetLatestVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && version != tt.wantVersion {
				t.Errorf("GetLatestVersion() = %v, want %v", version, tt.wantVersion)
			}
		})
	}
}

func TestNPMClient_GetPackageInfo_Mock(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		response    PackageInfo
		statusCode  int
		wantErr     bool
	}{
		{
			name:        "successful package info",
			packageName: "express",
			response: PackageInfo{
				Name: "express",
				DistTags: map[string]string{
					"latest": "4.18.2",
				},
				Versions: map[string]map[string]interface{}{
					"4.18.2": {"name": "express"},
					"4.18.1": {"name": "express"},
				},
				Time: map[string]string{
					"created":  "2010-12-29T19:38:25.450Z",
					"modified": "2023-03-20T15:00:00.000Z",
				},
			},
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:        "package not found",
			packageName: "nonexistent",
			statusCode:  http.StatusNotFound,
			wantErr:     true,
		},
		{
			name:        "server error",
			packageName: "test",
			statusCode:  http.StatusInternalServerError,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == http.StatusOK {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()

			client := &NPMClient{
				client:  &http.Client{Timeout: 5 * time.Second},
				baseURL: server.URL,
			}

			ctx := context.Background()
			info, err := client.GetPackageInfo(ctx, tt.packageName)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetPackageInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if info.Name != tt.response.Name {
					t.Errorf("GetPackageInfo() name = %v, want %v", info.Name, tt.response.Name)
				}
				if len(info.Versions) != len(tt.response.Versions) {
					t.Errorf("GetPackageInfo() versions count = %v, want %v", len(info.Versions), len(tt.response.Versions))
				}
			}
		})
	}
}

func TestNPMClient_FindBestVersion(t *testing.T) {
	tests := []struct {
		name            string
		packageName     string
		constraint      string
		allowPrerelease bool
		response        PackageInfo
		wantVersion     string
		wantErr         bool
	}{
		{
			name:        "find best version with constraint",
			packageName: "lodash",
			constraint:  ">=4.0.0 <5.0.0",
			response: PackageInfo{
				Name: "lodash",
				DistTags: map[string]string{
					"latest": "4.17.21",
				},
				Versions: map[string]map[string]interface{}{
					"4.17.21": {},
					"4.17.20": {},
					"4.16.0":  {},
					"3.10.0":  {},
				},
			},
			wantVersion: "4.17.21",
			wantErr:     false,
		},
		{
			name:        "find best version - minor only",
			packageName: "lodash",
			constraint:  "~4.16.0",
			response: PackageInfo{
				Name: "lodash",
				DistTags: map[string]string{
					"latest": "4.17.21",
				},
				Versions: map[string]map[string]interface{}{
					"4.17.21": {},
					"4.16.6":  {},
					"4.16.5":  {},
					"4.16.0":  {},
				},
			},
			wantVersion: "4.16.6",
			wantErr:     false,
		},
		{
			name:            "exclude prereleases",
			packageName:     "test",
			constraint:      ">=1.0.0",
			allowPrerelease: false,
			response: PackageInfo{
				Name: "test",
				DistTags: map[string]string{
					"latest": "1.0.0",
				},
				Versions: map[string]map[string]interface{}{
					"2.0.0-beta.1": {},
					"1.0.0":        {},
					"0.9.0":        {},
				},
			},
			wantVersion: "1.0.0",
			wantErr:     false,
		},
		{
			name:            "include prereleases with prerelease constraint",
			packageName:     "test",
			constraint:      ">=1.0.0-0", // -0 allows prereleases
			allowPrerelease: true,
			response: PackageInfo{
				Name: "test",
				DistTags: map[string]string{
					"latest": "1.0.0",
				},
				Versions: map[string]map[string]interface{}{
					"1.1.0-beta.1": {},
					"1.0.0":        {},
					"0.9.0":        {},
				},
			},
			wantVersion: "1.1.0-beta.1",
			wantErr:     false,
		},
		{
			name:        "no matching versions",
			packageName: "test",
			constraint:  ">=5.0.0",
			response: PackageInfo{
				Name: "test",
				DistTags: map[string]string{
					"latest": "1.0.0",
				},
				Versions: map[string]map[string]interface{}{
					"1.0.0": {},
					"0.9.0": {},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client := &NPMClient{
				client:  &http.Client{Timeout: 5 * time.Second},
				baseURL: server.URL,
			}

			ctx := context.Background()
			version, err := client.FindBestVersion(ctx, tt.packageName, tt.constraint, tt.allowPrerelease)

			if (err != nil) != tt.wantErr {
				t.Errorf("FindBestVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && version != tt.wantVersion {
				t.Errorf("FindBestVersion() = %v, want %v", version, tt.wantVersion)
			}
		})
	}
}

func TestNPMClient_GetVersions(t *testing.T) {
	response := PackageInfo{
		Name: "test",
		Versions: map[string]map[string]interface{}{
			"1.0.0": {},
			"1.1.0": {},
			"2.0.0": {},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &NPMClient{
		client:  &http.Client{Timeout: 5 * time.Second},
		baseURL: server.URL,
	}

	ctx := context.Background()
	versions, err := client.GetVersions(ctx, "test")
	if err != nil {
		t.Fatalf("GetVersions() error = %v", err)
	}

	if len(versions) != 3 {
		t.Errorf("GetVersions() returned %d versions, want 3", len(versions))
	}
}

func TestNewNPMClient(t *testing.T) {
	client := NewNPMClient()

	if client == nil {
		t.Fatal("NewNPMClient() returned nil")
	}

	if client.baseURL != npmRegistryURL {
		t.Errorf("NewNPMClient() baseURL = %v, want %v", client.baseURL, npmRegistryURL)
	}

	if client.client == nil {
		t.Error("NewNPMClient() http client is nil")
	}
}

// =============================================================================
// GitHub Client Tests
// =============================================================================

func TestGitHubClient_GetLatestRelease(t *testing.T) {
	tests := []struct {
		name        string
		owner       string
		repo        string
		response    Release
		statusCode  int
		wantVersion string
		wantErr     bool
	}{
		{
			name:  "successful latest release",
			owner: "hashicorp",
			repo:  "terraform",
			response: Release{
				TagName:    "v1.5.0",
				Name:       "Terraform v1.5.0",
				Prerelease: false,
				Draft:      false,
			},
			statusCode:  http.StatusOK,
			wantVersion: "1.5.0",
			wantErr:     false,
		},
		{
			name:       "repo not found",
			owner:      "nonexistent",
			repo:       "repo",
			statusCode: http.StatusNotFound,
			wantErr:    true,
		},
		{
			name:  "release without v prefix",
			owner: "test",
			repo:  "repo",
			response: Release{
				TagName:    "2.0.0",
				Prerelease: false,
			},
			statusCode:  http.StatusOK,
			wantVersion: "2.0.0",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == http.StatusOK {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()

			client := &GitHubClient{
				client:  &http.Client{Timeout: 5 * time.Second},
				baseURL: server.URL,
			}

			ctx := context.Background()
			version, err := client.GetLatestRelease(ctx, tt.owner, tt.repo)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetLatestRelease() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && version != tt.wantVersion {
				t.Errorf("GetLatestRelease() = %v, want %v", version, tt.wantVersion)
			}
		})
	}
}

func TestGitHubClient_GetAllReleases(t *testing.T) {
	releases := []Release{
		{TagName: "v1.5.0", Prerelease: false, Draft: false},
		{TagName: "v1.4.0", Prerelease: false, Draft: false},
		{TagName: "v1.5.0-rc1", Prerelease: true, Draft: false},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(releases)
	}))
	defer server.Close()

	client := &GitHubClient{
		client:  &http.Client{Timeout: 5 * time.Second},
		baseURL: server.URL,
	}

	ctx := context.Background()
	result, err := client.GetAllReleases(ctx, "owner", "repo")
	if err != nil {
		t.Fatalf("GetAllReleases() error = %v", err)
	}

	if len(result) != 3 {
		t.Errorf("GetAllReleases() returned %d releases, want 3", len(result))
	}
}

func TestGitHubClient_FindBestRelease(t *testing.T) {
	releases := []Release{
		{TagName: "v2.0.0", Prerelease: false, Draft: false},
		{TagName: "v1.9.0", Prerelease: false, Draft: false},
		{TagName: "v1.8.0", Prerelease: false, Draft: false},
		{TagName: "v2.1.0-beta", Prerelease: true, Draft: false},
		{TagName: "v1.10.0", Prerelease: false, Draft: true}, // draft
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(releases)
	}))
	defer server.Close()

	client := &GitHubClient{
		client:  &http.Client{Timeout: 5 * time.Second},
		baseURL: server.URL,
	}

	tests := []struct {
		name            string
		constraint      string
		allowPrerelease bool
		wantVersion     string
		wantErr         bool
	}{
		{
			name:            "find best version >=1.0.0 <2.0.0",
			constraint:      ">=1.0.0 <2.0.0",
			allowPrerelease: false,
			wantVersion:     "1.9.0",
			wantErr:         false,
		},
		{
			name:            "find best version >=2.0.0",
			constraint:      ">=2.0.0",
			allowPrerelease: false,
			wantVersion:     "2.0.0",
			wantErr:         false,
		},
		{
			name:            "include prereleases",
			constraint:      ">=2.0.0",
			allowPrerelease: true,
			wantVersion:     "2.0.0", // prereleases are only picked when they match the constraint exactly
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			version, err := client.FindBestRelease(ctx, "owner", "repo", tt.constraint, tt.allowPrerelease)

			if (err != nil) != tt.wantErr {
				t.Errorf("FindBestRelease() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && version != tt.wantVersion {
				t.Errorf("FindBestRelease() = %v, want %v", version, tt.wantVersion)
			}
		})
	}
}

func TestGitHubClient_WithToken(t *testing.T) {
	var receivedAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(Release{TagName: "v1.0.0"})
	}))
	defer server.Close()

	client := &GitHubClient{
		client:  &http.Client{Timeout: 5 * time.Second},
		baseURL: server.URL,
		token:   "test-token",
	}

	ctx := context.Background()
	_, _ = client.GetLatestRelease(ctx, "owner", "repo")

	if receivedAuth != "Bearer test-token" {
		t.Errorf("Authorization header = %v, want 'Bearer test-token'", receivedAuth)
	}
}

func TestNewGitHubClient(t *testing.T) {
	client := NewGitHubClient("test-token")

	if client == nil {
		t.Fatal("NewGitHubClient() returned nil")
	}

	if client.token != "test-token" {
		t.Errorf("NewGitHubClient() token = %v, want 'test-token'", client.token)
	}

	if client.baseURL != githubAPIURL {
		t.Errorf("NewGitHubClient() baseURL = %v, want %v", client.baseURL, githubAPIURL)
	}
}

func TestParseGitHubURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{
			name:      "full https URL",
			url:       "https://github.com/hashicorp/terraform",
			wantOwner: "hashicorp",
			wantRepo:  "terraform",
			wantErr:   false,
		},
		{
			name:      "http URL",
			url:       "http://github.com/owner/repo",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "without protocol",
			url:       "github.com/owner/repo",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "short format",
			url:       "owner/repo",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "with .git suffix",
			url:       "https://github.com/owner/repo.git",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:    "invalid URL - no repo",
			url:     "owner",
			wantErr: true,
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := ParseGitHubURL(tt.url)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseGitHubURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if owner != tt.wantOwner {
					t.Errorf("ParseGitHubURL() owner = %v, want %v", owner, tt.wantOwner)
				}
				if repo != tt.wantRepo {
					t.Errorf("ParseGitHubURL() repo = %v, want %v", repo, tt.wantRepo)
				}
			}
		})
	}
}

// =============================================================================
// Helm Client Tests
// =============================================================================

func TestHelmClient_GetLatestChartVersion(t *testing.T) {
	tests := []struct {
		name        string
		chartName   string
		index       ChartIndex
		statusCode  int
		wantVersion string
		wantErr     bool
	}{
		{
			name:      "successful latest version",
			chartName: "nginx",
			index: ChartIndex{
				APIVersion: "v1",
				Entries: map[string][]ChartIndexEntry{
					"nginx": {
						{Name: "nginx", Version: "15.0.0"},
						{Name: "nginx", Version: "14.2.0"},
						{Name: "nginx", Version: "14.1.0"},
					},
				},
			},
			statusCode:  http.StatusOK,
			wantVersion: "15.0.0",
			wantErr:     false,
		},
		{
			name:      "skip prereleases",
			chartName: "test",
			index: ChartIndex{
				Entries: map[string][]ChartIndexEntry{
					"test": {
						{Name: "test", Version: "2.0.0-beta.1"},
						{Name: "test", Version: "1.5.0"},
						{Name: "test", Version: "1.4.0"},
					},
				},
			},
			statusCode:  http.StatusOK,
			wantVersion: "1.5.0",
			wantErr:     false,
		},
		{
			name:       "chart not found",
			chartName:  "nonexistent",
			index:      ChartIndex{Entries: map[string][]ChartIndexEntry{}},
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "repository not found",
			chartName:  "test",
			statusCode: http.StatusNotFound,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == http.StatusOK {
					data, _ := yaml.Marshal(tt.index)
					_, _ = w.Write(data)
				}
			}))
			defer server.Close()

			client := NewHelmClient()

			ctx := context.Background()
			version, err := client.GetLatestChartVersion(ctx, server.URL, tt.chartName)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetLatestChartVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && version != tt.wantVersion {
				t.Errorf("GetLatestChartVersion() = %v, want %v", version, tt.wantVersion)
			}
		})
	}
}

func TestHelmClient_FindBestChartVersion(t *testing.T) {
	index := ChartIndex{
		Entries: map[string][]ChartIndexEntry{
			"nginx": {
				{Name: "nginx", Version: "15.0.0"},
				{Name: "nginx", Version: "14.2.0"},
				{Name: "nginx", Version: "14.1.0"},
				{Name: "nginx", Version: "13.0.0"},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		data, _ := yaml.Marshal(index)
		_, _ = w.Write(data)
	}))
	defer server.Close()

	client := NewHelmClient()

	tests := []struct {
		name        string
		constraint  string
		wantVersion string
		wantErr     bool
	}{
		{
			name:        "find best version >=14.0.0 <15.0.0",
			constraint:  ">=14.0.0 <15.0.0",
			wantVersion: "14.2.0",
			wantErr:     false,
		},
		{
			name:        "find best version >=15.0.0",
			constraint:  ">=15.0.0",
			wantVersion: "15.0.0",
			wantErr:     false,
		},
		{
			name:       "no matching versions",
			constraint: ">=20.0.0",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			version, err := client.FindBestChartVersion(ctx, server.URL, "nginx", tt.constraint)

			if (err != nil) != tt.wantErr {
				t.Errorf("FindBestChartVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && version != tt.wantVersion {
				t.Errorf("FindBestChartVersion() = %v, want %v", version, tt.wantVersion)
			}
		})
	}
}

func TestHelmClient_GetChartVersions(t *testing.T) {
	index := ChartIndex{
		Entries: map[string][]ChartIndexEntry{
			"nginx": {
				{Name: "nginx", Version: "15.0.0"},
				{Name: "nginx", Version: "14.2.0"},
				{Name: "nginx", Version: "14.1.0"},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		data, _ := yaml.Marshal(index)
		_, _ = w.Write(data)
	}))
	defer server.Close()

	client := NewHelmClient()
	ctx := context.Background()

	versions, err := client.GetChartVersions(ctx, server.URL, "nginx")
	if err != nil {
		t.Fatalf("GetChartVersions() error = %v", err)
	}

	if len(versions) != 3 {
		t.Errorf("GetChartVersions() returned %d versions, want 3", len(versions))
	}
}

func TestHelmClient_GetChartVersionDetails(t *testing.T) {
	index := ChartIndex{
		Entries: map[string][]ChartIndexEntry{
			"nginx": {
				{Name: "nginx", Version: "15.0.0", AppVersion: "1.25.0", Description: "nginx chart"},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		data, _ := yaml.Marshal(index)
		_, _ = w.Write(data)
	}))
	defer server.Close()

	client := NewHelmClient()
	ctx := context.Background()

	entries, err := client.GetChartVersionDetails(ctx, server.URL, "nginx")
	if err != nil {
		t.Fatalf("GetChartVersionDetails() error = %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("GetChartVersionDetails() returned %d entries, want 1", len(entries))
	}

	if entries[0].AppVersion != "1.25.0" {
		t.Errorf("GetChartVersionDetails() appVersion = %v, want '1.25.0'", entries[0].AppVersion)
	}
}

func TestIsOCIRepository(t *testing.T) {
	tests := []struct {
		name       string
		repository string
		want       bool
	}{
		{
			name:       "OCI repository",
			repository: "oci://registry.example.com/charts",
			want:       true,
		},
		{
			name:       "HTTPS repository",
			repository: "https://charts.bitnami.com/bitnami",
			want:       false,
		},
		{
			name:       "HTTP repository",
			repository: "http://charts.example.com",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsOCIRepository(tt.repository); got != tt.want {
				t.Errorf("IsOCIRepository() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewHelmClient(t *testing.T) {
	client := NewHelmClient()

	if client == nil {
		t.Fatal("NewHelmClient() returned nil")
	}

	if client.client == nil {
		t.Error("NewHelmClient() http client is nil")
	}
}

// =============================================================================
// Terraform Client Tests
// =============================================================================

func TestTerraformClient_GetLatestProviderVersion(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		response    ProviderVersions
		statusCode  int
		wantVersion string
		wantErr     bool
	}{
		{
			name:   "successful latest version",
			source: "hashicorp/aws",
			response: ProviderVersions{
				Versions: []ProviderVersion{
					{Version: "5.0.0"},
					{Version: "4.67.0"},
					{Version: "4.66.0"},
				},
			},
			statusCode:  http.StatusOK,
			wantVersion: "5.0.0",
			wantErr:     false,
		},
		{
			name:   "skip prereleases",
			source: "hashicorp/aws",
			response: ProviderVersions{
				Versions: []ProviderVersion{
					{Version: "6.0.0-beta1"},
					{Version: "5.0.0"},
					{Version: "4.67.0"},
				},
			},
			statusCode:  http.StatusOK,
			wantVersion: "5.0.0",
			wantErr:     false,
		},
		{
			name:       "provider not found",
			source:     "nonexistent/provider",
			statusCode: http.StatusNotFound,
			wantErr:    true,
		},
		{
			name:    "invalid source format",
			source:  "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == http.StatusOK {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()

			client := &TerraformClient{
				client:  &http.Client{Timeout: 5 * time.Second},
				baseURL: server.URL,
			}

			ctx := context.Background()
			version, err := client.GetLatestProviderVersion(ctx, tt.source)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetLatestProviderVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && version != tt.wantVersion {
				t.Errorf("GetLatestProviderVersion() = %v, want %v", version, tt.wantVersion)
			}
		})
	}
}

func TestTerraformClient_GetLatestModuleVersion(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		response    ModuleVersions
		statusCode  int
		wantVersion string
		wantErr     bool
	}{
		{
			name:   "successful latest version",
			source: "terraform-aws-modules/vpc/aws",
			response: ModuleVersions{
				Modules: []Module{
					{
						Source: "terraform-aws-modules/vpc/aws",
						Versions: []ModuleVersion{
							{Version: "5.0.0"},
							{Version: "4.0.0"},
							{Version: "3.0.0"},
						},
					},
				},
			},
			statusCode:  http.StatusOK,
			wantVersion: "5.0.0",
			wantErr:     false,
		},
		{
			name:       "module not found",
			source:     "nonexistent/module/aws",
			statusCode: http.StatusNotFound,
			wantErr:    true,
		},
		{
			name:    "invalid source format",
			source:  "invalid/format",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == http.StatusOK {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()

			client := &TerraformClient{
				client:  &http.Client{Timeout: 5 * time.Second},
				baseURL: server.URL,
			}

			ctx := context.Background()
			version, err := client.GetLatestModuleVersion(ctx, tt.source)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetLatestModuleVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && version != tt.wantVersion {
				t.Errorf("GetLatestModuleVersion() = %v, want %v", version, tt.wantVersion)
			}
		})
	}
}

func TestTerraformClient_GetModuleVersions(t *testing.T) {
	response := ModuleVersions{
		Modules: []Module{
			{
				Source: "terraform-aws-modules/vpc/aws",
				Versions: []ModuleVersion{
					{Version: "5.0.0"},
					{Version: "4.0.0"},
					{Version: "3.0.0"},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &TerraformClient{
		client:  &http.Client{Timeout: 5 * time.Second},
		baseURL: server.URL,
	}

	ctx := context.Background()
	versions, err := client.GetModuleVersions(ctx, "terraform-aws-modules/vpc/aws")
	if err != nil {
		t.Fatalf("GetModuleVersions() error = %v", err)
	}

	if len(versions) != 3 {
		t.Errorf("GetModuleVersions() returned %d versions, want 3", len(versions))
	}
}

func TestTerraformClient_FindBestProviderVersion(t *testing.T) {
	response := ProviderVersions{
		Versions: []ProviderVersion{
			{Version: "5.0.0"},
			{Version: "4.67.0"},
			{Version: "4.66.0"},
			{Version: "3.0.0"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &TerraformClient{
		client:  &http.Client{Timeout: 5 * time.Second},
		baseURL: server.URL,
	}

	tests := []struct {
		name        string
		constraint  string
		wantVersion string
		wantErr     bool
	}{
		{
			name:        "find best version >=4.0.0 <5.0.0",
			constraint:  ">=4.0.0 <5.0.0",
			wantVersion: "4.67.0",
			wantErr:     false,
		},
		{
			name:        "find best version >=5.0.0",
			constraint:  ">=5.0.0",
			wantVersion: "5.0.0",
			wantErr:     false,
		},
		{
			name:       "no matching versions",
			constraint: ">=6.0.0",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			version, err := client.FindBestProviderVersion(ctx, "hashicorp/aws", tt.constraint)

			if (err != nil) != tt.wantErr {
				t.Errorf("FindBestProviderVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && version != tt.wantVersion {
				t.Errorf("FindBestProviderVersion() = %v, want %v", version, tt.wantVersion)
			}
		})
	}
}

func TestNewTerraformClient(t *testing.T) {
	client := NewTerraformClient()

	if client == nil {
		t.Fatal("NewTerraformClient() returned nil")
	}

	if client.baseURL != terraformRegistryURL {
		t.Errorf("NewTerraformClient() baseURL = %v, want %v", client.baseURL, terraformRegistryURL)
	}

	if client.client == nil {
		t.Error("NewTerraformClient() http client is nil")
	}
}

// =============================================================================
// Error Handling Tests
// =============================================================================

func TestNPMClient_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := &NPMClient{
		client:  &http.Client{Timeout: 5 * time.Second},
		baseURL: server.URL,
	}

	ctx := context.Background()
	_, err := client.GetPackageInfo(ctx, "test")

	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestGitHubClient_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := &GitHubClient{
		client:  &http.Client{Timeout: 5 * time.Second},
		baseURL: server.URL,
	}

	ctx := context.Background()
	_, err := client.GetLatestRelease(ctx, "owner", "repo")

	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestHelmClient_InvalidYAML(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid: yaml: content: ["))
	}))
	defer server.Close()

	client := NewHelmClient()

	ctx := context.Background()
	_, err := client.GetLatestChartVersion(ctx, server.URL, "test")

	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}

func TestTerraformClient_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := &TerraformClient{
		client:  &http.Client{Timeout: 5 * time.Second},
		baseURL: server.URL,
	}

	ctx := context.Background()
	_, err := client.GetLatestProviderVersion(ctx, "hashicorp/aws")

	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

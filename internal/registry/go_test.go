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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	testVersion091      = "v0.9.1"
	testVersionListPath = "/github.com/pkg/errors/@v/list"
	testModulePath      = "github.com/pkg/errors"
)

func TestNewGoClient(t *testing.T) {
	client := NewGoClient()
	if client == nil {
		t.Fatal("NewGoClient() returned nil")
	}
	if client.client == nil {
		t.Error("NewGoClient() http client is nil")
	}
	if client.baseURL != goProxyURL {
		t.Errorf("NewGoClient() baseURL = %q, want %q", client.baseURL, goProxyURL)
	}
}

func TestGoClient_GetLatestVersion(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/github.com/pkg/errors/@latest" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}

			response := GoModuleInfo{
				Version: "v0.9.1",
				Time:    time.Now(),
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := &GoClient{
			client:  server.Client(),
			baseURL: server.URL,
		}

		version, err := client.GetLatestVersion(context.Background(), testModulePath)
		if err != nil {
			t.Fatalf("GetLatestVersion() error = %v", err)
		}
		if version != testVersion091 {
			t.Errorf("GetLatestVersion() = %q, want %q", version, testVersion091)
		}
	})

	t.Run("module not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client := &GoClient{
			client:  server.Client(),
			baseURL: server.URL,
		}

		_, err := client.GetLatestVersion(context.Background(), "github.com/nonexistent/pkg")
		if err == nil {
			t.Error("GetLatestVersion() expected error for not found, got nil")
		}
	})

	t.Run("gone (410) response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusGone)
		}))
		defer server.Close()

		client := &GoClient{
			client:  server.Client(),
			baseURL: server.URL,
		}

		_, err := client.GetLatestVersion(context.Background(), "github.com/deprecated/pkg")
		if err == nil {
			t.Error("GetLatestVersion() expected error for gone status, got nil")
		}
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := &GoClient{
			client:  server.Client(),
			baseURL: server.URL,
		}

		_, err := client.GetLatestVersion(context.Background(), "github.com/pkg/errors")
		if err == nil {
			t.Error("GetLatestVersion() expected error for server error, got nil")
		}
	})

	t.Run("invalid JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		client := &GoClient{
			client:  server.Client(),
			baseURL: server.URL,
		}

		_, err := client.GetLatestVersion(context.Background(), "github.com/pkg/errors")
		if err == nil {
			t.Error("GetLatestVersion() expected error for invalid JSON, got nil")
		}
	})
}

func TestGoClient_GetVersions(t *testing.T) {
	// Helper to test GetVersions with different responses
	testGetVersionsCount := func(t *testing.T, response, modulePath string, wantCount int) {
		t.Helper()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(response))
		}))
		defer server.Close()

		client := &GoClient{client: server.Client(), baseURL: server.URL}
		versions, err := client.GetVersions(context.Background(), modulePath)
		if err != nil {
			t.Fatalf("GetVersions() error = %v", err)
		}
		if len(versions) != wantCount {
			t.Errorf("GetVersions() count = %d, want %d", len(versions), wantCount)
		}
	}

	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != testVersionListPath {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			_, _ = w.Write([]byte("v0.8.0\nv0.8.1\nv0.9.0\nv0.9.1\n"))
		}))
		defer server.Close()

		client := &GoClient{client: server.Client(), baseURL: server.URL}
		versions, err := client.GetVersions(context.Background(), testModulePath)
		if err != nil {
			t.Fatalf("GetVersions() error = %v", err)
		}
		expected := []string{"v0.8.0", "v0.8.1", "v0.9.0", testVersion091}
		if len(versions) != len(expected) {
			t.Fatalf("GetVersions() count = %d, want %d", len(versions), len(expected))
		}
		for i, v := range expected {
			if versions[i] != v {
				t.Errorf("GetVersions()[%d] = %q, want %q", i, versions[i], v)
			}
		}
	})

	t.Run("empty version list", func(t *testing.T) {
		testGetVersionsCount(t, "", "github.com/new/pkg", 0)
	})

	t.Run("handles whitespace", func(t *testing.T) {
		testGetVersionsCount(t, "  v0.9.0  \n  v0.9.1  \n\n", testModulePath, 2)
	})

	t.Run("module not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client := &GoClient{client: server.Client(), baseURL: server.URL}
		_, err := client.GetVersions(context.Background(), "github.com/nonexistent/pkg")
		if err == nil {
			t.Error("GetVersions() expected error for not found, got nil")
		}
	})
}

func TestGoClient_GetModuleInfo(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		expectedTime := time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/github.com/pkg/errors/@v/v0.9.1.info" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}

			response := GoModuleInfo{
				Version: "v0.9.1",
				Time:    expectedTime,
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := &GoClient{
			client:  server.Client(),
			baseURL: server.URL,
		}

		info, err := client.GetModuleInfo(context.Background(), "github.com/pkg/errors", "v0.9.1")
		if err != nil {
			t.Fatalf("GetModuleInfo() error = %v", err)
		}
		if info.Version != "v0.9.1" {
			t.Errorf("GetModuleInfo() version = %q, want %q", info.Version, "v0.9.1")
		}
		if !info.Time.Equal(expectedTime) {
			t.Errorf("GetModuleInfo() time = %v, want %v", info.Time, expectedTime)
		}
	})

	t.Run("version not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client := &GoClient{
			client:  server.Client(),
			baseURL: server.URL,
		}

		_, err := client.GetModuleInfo(context.Background(), "github.com/pkg/errors", "v99.99.99")
		if err == nil {
			t.Error("GetModuleInfo() expected error for not found version, got nil")
		}
	})
}

func TestGoClient_FindBestVersion(t *testing.T) {
	// Helper to test FindBestVersion with different version lists
	testFindBest := func(t *testing.T, versionList string, allowPrerelease bool, wantVersion, errMsg string) {
		t.Helper()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == testVersionListPath {
				_, _ = w.Write([]byte(versionList))
			}
		}))
		defer server.Close()

		client := &GoClient{client: server.Client(), baseURL: server.URL}
		version, err := client.FindBestVersion(context.Background(), testModulePath, allowPrerelease)
		if err != nil {
			t.Fatalf("FindBestVersion() error = %v", err)
		}
		if version != wantVersion {
			t.Errorf("FindBestVersion() = %q, want %q (%s)", version, wantVersion, errMsg)
		}
	}

	t.Run("finds highest stable version", func(t *testing.T) {
		testFindBest(t, "v0.8.0\nv0.9.0\nv0.9.1\nv1.0.0-alpha\n", false, testVersion091, "should skip prerelease")
	})

	t.Run("includes prerelease when allowed", func(t *testing.T) {
		testFindBest(t, "v0.9.1\nv1.0.0-alpha\nv1.0.0-beta\n", true, "v1.0.0-beta", "should include prerelease")
	})

	t.Run("handles empty version list by falling back to latest", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case testVersionListPath:
				// All versions are invalid semver
				_, _ = w.Write([]byte("not-semver\nalso-invalid\n"))
			case "/github.com/pkg/errors/@latest":
				response := GoModuleInfo{Version: testVersion091}
				_ = json.NewEncoder(w).Encode(response)
			}
		}))
		defer server.Close()

		client := &GoClient{client: server.Client(), baseURL: server.URL}
		version, err := client.FindBestVersion(context.Background(), testModulePath, false)
		if err != nil {
			t.Fatalf("FindBestVersion() error = %v", err)
		}
		if version != testVersion091 {
			t.Errorf("FindBestVersion() = %q, want %q (should fallback to @latest)", version, testVersion091)
		}
	})
}

func TestEscapeModulePath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "lowercase path",
			input: "github.com/pkg/errors",
			want:  "github.com%2Fpkg%2Ferrors",
		},
		{
			name:  "mixed case path",
			input: "github.com/Azure/azure-sdk-for-go",
			want:  "github.com%2F%21azure%2Fazure-sdk-for-go",
		},
		{
			name:  "multiple uppercase",
			input: "github.com/BurntSushi/toml",
			want:  "github.com%2F%21burnt%21sushi%2Ftoml",
		},
		{
			name:  "all lowercase",
			input: "golang.org/x/text",
			want:  "golang.org%2Fx%2Ftext",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeModulePath(tt.input)
			if got != tt.want {
				t.Errorf("escapeModulePath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

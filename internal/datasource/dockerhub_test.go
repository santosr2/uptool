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
	"testing"
)

func TestNewDockerHubDatasource(t *testing.T) {
	ds := NewDockerHubDatasource()
	if ds == nil {
		t.Fatal("NewDockerHubDatasource() returned nil")
	}
	if ds.client == nil {
		t.Error("NewDockerHubDatasource() client is nil")
	}
	if ds.baseURL == "" {
		t.Error("NewDockerHubDatasource() baseURL is empty")
	}
}

func TestDockerHubDatasource_Name(t *testing.T) {
	ds := NewDockerHubDatasource()
	if got := ds.Name(); got != "docker-hub" {
		t.Errorf("Name() = %q, want %q", got, "docker-hub")
	}
}

func TestNormalizeImageName(t *testing.T) {
	tests := []struct {
		name              string
		image             string
		expectedNamespace string
		expectedRepo      string
	}{
		{"official image", "nginx", "library", "nginx"},
		{"user image", "myuser/myrepo", "myuser", "myrepo"},
		{"org image", "myorg/myapp", "myorg", "myapp"},
		{"registry prefix", "gcr.io/project/image", "project", "image"},
		{"three parts", "quay.io/org/app", "org", "app"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			namespace, repo := normalizeImageName(tt.image)
			if namespace != tt.expectedNamespace {
				t.Errorf("normalizeImageName(%q) namespace = %q, want %q", tt.image, namespace, tt.expectedNamespace)
			}
			if repo != tt.expectedRepo {
				t.Errorf("normalizeImageName(%q) repo = %q, want %q", tt.image, repo, tt.expectedRepo)
			}
		})
	}
}

func TestIsSemverTag(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		expected bool
	}{
		{"simple semver", "1.2.3", true},
		{"with v prefix", "v1.2.3", true},
		{"major only", "1", true},
		{"major.minor", "1.2", true},
		{"with alpha", "1.2.3-alpha", true},
		{"with beta", "1.2.3-beta.1", true},
		{"with rc", "1.2.3-rc1", true},
		{"latest", "latest", false},
		{"edge", "edge", false},
		{"nightly", "nightly", false},
		{"master", "master", false},
		{"main", "main", false},
		{"develop", "develop", false},
		{"dev", "dev", false},
		{"alpine suffix", "alpine", false},
		{"slim suffix", "slim", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSemverTag(tt.tag)
			if got != tt.expected {
				t.Errorf("isSemverTag(%q) = %v, want %v", tt.tag, got, tt.expected)
			}
		})
	}
}

func TestIsPrerelease(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected bool
	}{
		{"stable version", "1.2.3", false},
		{"alpha version", "1.2.3-alpha", true},
		{"Alpha uppercase", "1.2.3-Alpha.1", true},
		{"beta version", "1.2.3-beta", true},
		{"rc version", "1.2.3-rc1", true},
		{"pre version", "1.2.3-pre", true},
		{"with numbers", "1.2.3-alpha.1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPrerelease(tt.version)
			if got != tt.expected {
				t.Errorf("isPrerelease(%q) = %v, want %v", tt.version, got, tt.expected)
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int
	}{
		{"equal versions", "1.2.3", "1.2.3", 0},
		{"v1 greater major", "2.0.0", "1.0.0", 1},
		{"v2 greater major", "1.0.0", "2.0.0", -1},
		{"v1 greater minor", "1.3.0", "1.2.0", 1},
		{"v2 greater minor", "1.2.0", "1.3.0", -1},
		{"v1 greater patch", "1.2.4", "1.2.3", 1},
		{"v2 greater patch", "1.2.3", "1.2.4", -1},
		{"with v prefix", "v1.2.3", "v1.2.3", 0},
		{"mixed prefix", "v1.2.3", "1.2.3", 0},
		{"different lengths", "1.2", "1.2.0", 0},
		{"longer version", "1.2.3", "1.2", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareVersions(tt.v1, tt.v2)
			if (tt.expected > 0 && got <= 0) || (tt.expected < 0 && got >= 0) || (tt.expected == 0 && got != 0) {
				t.Errorf("compareVersions(%q, %q) = %d, want sign of %d", tt.v1, tt.v2, got, tt.expected)
			}
		})
	}
}

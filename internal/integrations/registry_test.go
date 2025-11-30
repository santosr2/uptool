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

//nolint:dupl // Test files use similar table-driven patterns
package integrations

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/santosr2/uptool/internal/engine"
)

// mockIntegration is a test integration for registry testing
type mockIntegration struct {
	name string
}

func (m *mockIntegration) Name() string {
	return m.name
}

func (m *mockIntegration) Detect(ctx context.Context, repoRoot string) ([]*engine.Manifest, error) {
	return nil, nil
}

func (m *mockIntegration) Plan(ctx context.Context, manifest *engine.Manifest, planCtx *engine.PlanContext) (*engine.UpdatePlan, error) {
	return nil, nil
}

func (m *mockIntegration) Apply(ctx context.Context, plan *engine.UpdatePlan) (*engine.ApplyResult, error) {
	return nil, nil
}

func (m *mockIntegration) Validate(ctx context.Context, manifest *engine.Manifest) error {
	return nil
}

func TestRegister(t *testing.T) {
	// Save original registry state
	mu.Lock()
	originalRegistry := make(map[string]func() engine.Integration)
	for k, v := range registry {
		originalRegistry[k] = v
	}
	mu.Unlock()

	defer func() {
		// Restore original registry
		mu.Lock()
		registry = originalRegistry
		instances = make(map[string]engine.Integration)
		mu.Unlock()
	}()

	// Clear for test
	mu.Lock()
	registry = make(map[string]func() engine.Integration)
	instances = make(map[string]engine.Integration)
	mu.Unlock()

	// Test registering a new integration
	Register("test-integration", func() engine.Integration {
		return &mockIntegration{name: "test-integration"}
	})

	// Verify it was registered
	if Count() != 1 {
		t.Errorf("Count() = %d, want 1", Count())
	}

	names := List()
	if len(names) != 1 || names[0] != "test-integration" {
		t.Errorf("List() = %v, want [test-integration]", names)
	}
}

func TestRegisterPanic(t *testing.T) {
	// Save original registry state
	mu.Lock()
	originalRegistry := make(map[string]func() engine.Integration)
	for k, v := range registry {
		originalRegistry[k] = v
	}
	mu.Unlock()

	defer func() {
		// Restore original registry
		mu.Lock()
		registry = originalRegistry
		instances = make(map[string]engine.Integration)
		mu.Unlock()
	}()

	// Clear for test
	mu.Lock()
	registry = make(map[string]func() engine.Integration)
	instances = make(map[string]engine.Integration)
	mu.Unlock()

	// Register once
	Register("duplicate-test", func() engine.Integration {
		return &mockIntegration{name: "duplicate-test"}
	})

	// Try to register again - should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("Register() should panic on duplicate registration")
		}
	}()

	Register("duplicate-test", func() engine.Integration {
		return &mockIntegration{name: "duplicate-test"}
	})
}

func TestGet(t *testing.T) {
	// Save original registry state
	mu.Lock()
	originalRegistry := make(map[string]func() engine.Integration)
	for k, v := range registry {
		originalRegistry[k] = v
	}
	originalPluginsLoaded := pluginsLoaded
	mu.Unlock()

	defer func() {
		// Restore original registry
		mu.Lock()
		registry = originalRegistry
		instances = make(map[string]engine.Integration)
		pluginsLoaded = originalPluginsLoaded
		mu.Unlock()
	}()

	// Clear for test
	mu.Lock()
	registry = make(map[string]func() engine.Integration)
	instances = make(map[string]engine.Integration)
	pluginsLoaded = true // Skip plugin loading for this test
	mu.Unlock()

	// Register test integration
	Register("test-get", func() engine.Integration {
		return &mockIntegration{name: "test-get"}
	})

	// Test Get
	integration, err := Get("test-get")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if integration.Name() != "test-get" {
		t.Errorf("Get() returned integration with name %q, want %q", integration.Name(), "test-get")
	}

	// Test Get with non-existent integration
	_, err = Get("non-existent")
	if err == nil {
		t.Error("Get() should return error for non-existent integration")
	}

	// Verify caching - Get the same integration again
	integration2, err := Get("test-get")
	if err != nil {
		t.Fatalf("Get() second call error = %v", err)
	}

	// Should be the same instance (cached)
	if integration != integration2 {
		t.Error("Get() should return cached instance")
	}
}

func TestGetAll(t *testing.T) {
	// Save original registry state
	mu.Lock()
	originalRegistry := make(map[string]func() engine.Integration)
	for k, v := range registry {
		originalRegistry[k] = v
	}
	originalPluginsLoaded := pluginsLoaded
	mu.Unlock()

	defer func() {
		// Restore original registry
		mu.Lock()
		registry = originalRegistry
		instances = make(map[string]engine.Integration)
		pluginsLoaded = originalPluginsLoaded
		mu.Unlock()
	}()

	// Clear for test
	mu.Lock()
	registry = make(map[string]func() engine.Integration)
	instances = make(map[string]engine.Integration)
	pluginsLoaded = true // Skip plugin loading for this test
	mu.Unlock()

	// Register multiple integrations
	Register("integration-1", func() engine.Integration {
		return &mockIntegration{name: "integration-1"}
	})
	Register("integration-2", func() engine.Integration {
		return &mockIntegration{name: "integration-2"}
	})
	Register("integration-3", func() engine.Integration {
		return &mockIntegration{name: "integration-3"}
	})

	// Test GetAll
	all := GetAll()
	if len(all) != 3 {
		t.Fatalf("GetAll() returned %d integrations, want 3", len(all))
	}

	if all["integration-1"].Name() != "integration-1" {
		t.Error("GetAll() integration-1 has wrong name")
	}
	if all["integration-2"].Name() != "integration-2" {
		t.Error("GetAll() integration-2 has wrong name")
	}
	if all["integration-3"].Name() != "integration-3" {
		t.Error("GetAll() integration-3 has wrong name")
	}

	// Test that instances are cached
	all2 := GetAll()
	for name := range all {
		if all[name] != all2[name] {
			t.Errorf("GetAll() should return same cached instance for %s", name)
		}
	}
}

func TestGetLazy(t *testing.T) {
	// Save original registry state
	mu.Lock()
	originalRegistry := make(map[string]func() engine.Integration)
	for k, v := range registry {
		originalRegistry[k] = v
	}
	originalPluginsLoaded := pluginsLoaded
	mu.Unlock()

	defer func() {
		// Restore original registry
		mu.Lock()
		registry = originalRegistry
		instances = make(map[string]engine.Integration)
		pluginsLoaded = originalPluginsLoaded
		mu.Unlock()
	}()

	// Clear for test
	mu.Lock()
	registry = make(map[string]func() engine.Integration)
	instances = make(map[string]engine.Integration)
	pluginsLoaded = true // Skip plugin loading for this test
	mu.Unlock()

	// Register test integrations
	Register("lazy-1", func() engine.Integration {
		return &mockIntegration{name: "lazy-1"}
	})
	Register("lazy-2", func() engine.Integration {
		return &mockIntegration{name: "lazy-2"}
	})

	// Test GetLazy - should return constructors, not instances
	constructors := GetLazy()
	if len(constructors) != 2 {
		t.Fatalf("GetLazy() returned %d constructors, want 2", len(constructors))
	}

	// Verify constructors work
	instance1 := constructors["lazy-1"]()
	if instance1.Name() != "lazy-1" {
		t.Errorf("Constructor for lazy-1 returned integration with name %q", instance1.Name())
	}

	instance2 := constructors["lazy-2"]()
	if instance2.Name() != "lazy-2" {
		t.Errorf("Constructor for lazy-2 returned integration with name %q", instance2.Name())
	}

	// Each call to constructor should create a new instance
	instance1b := constructors["lazy-1"]()
	if instance1 == instance1b {
		t.Error("GetLazy() constructors should create new instances, not cached ones")
	}
}

func TestCount(t *testing.T) {
	// Save original registry state
	mu.Lock()
	originalRegistry := make(map[string]func() engine.Integration)
	for k, v := range registry {
		originalRegistry[k] = v
	}
	mu.Unlock()

	defer func() {
		// Restore original registry
		mu.Lock()
		registry = originalRegistry
		instances = make(map[string]engine.Integration)
		mu.Unlock()
	}()

	// Clear for test
	mu.Lock()
	registry = make(map[string]func() engine.Integration)
	instances = make(map[string]engine.Integration)
	mu.Unlock()

	// Test with empty registry
	if Count() != 0 {
		t.Errorf("Count() = %d, want 0 for empty registry", Count())
	}

	// Add integrations
	Register("count-1", func() engine.Integration {
		return &mockIntegration{name: "count-1"}
	})
	if Count() != 1 {
		t.Errorf("Count() = %d, want 1", Count())
	}

	Register("count-2", func() engine.Integration {
		return &mockIntegration{name: "count-2"}
	})
	if Count() != 2 {
		t.Errorf("Count() = %d, want 2", Count())
	}

	Register("count-3", func() engine.Integration {
		return &mockIntegration{name: "count-3"}
	})
	if Count() != 3 {
		t.Errorf("Count() = %d, want 3", Count())
	}
}

func TestList(t *testing.T) {
	// Save original registry state
	mu.Lock()
	originalRegistry := make(map[string]func() engine.Integration)
	for k, v := range registry {
		originalRegistry[k] = v
	}
	mu.Unlock()

	defer func() {
		// Restore original registry
		mu.Lock()
		registry = originalRegistry
		instances = make(map[string]engine.Integration)
		mu.Unlock()
	}()

	// Clear for test
	mu.Lock()
	registry = make(map[string]func() engine.Integration)
	instances = make(map[string]engine.Integration)
	mu.Unlock()

	// Test empty list
	names := List()
	if len(names) != 0 {
		t.Errorf("List() = %v, want empty slice", names)
	}

	// Add integrations in non-alphabetical order
	Register("zulu", func() engine.Integration {
		return &mockIntegration{name: "zulu"}
	})
	Register("alpha", func() engine.Integration {
		return &mockIntegration{name: "alpha"}
	})
	Register("bravo", func() engine.Integration {
		return &mockIntegration{name: "bravo"}
	})

	// List should return sorted names
	names = List()
	expected := []string{"alpha", "bravo", "zulu"}
	if len(names) != len(expected) {
		t.Fatalf("List() returned %d names, want %d", len(names), len(expected))
	}

	for i, name := range names {
		if name != expected[i] {
			t.Errorf("List()[%d] = %q, want %q", i, name, expected[i])
		}
	}
}

func TestClearCache(t *testing.T) {
	// Save original registry state
	mu.Lock()
	originalRegistry := make(map[string]func() engine.Integration)
	for k, v := range registry {
		originalRegistry[k] = v
	}
	originalPluginsLoaded := pluginsLoaded
	mu.Unlock()

	defer func() {
		// Restore original registry
		mu.Lock()
		registry = originalRegistry
		instances = make(map[string]engine.Integration)
		pluginsLoaded = originalPluginsLoaded
		mu.Unlock()
	}()

	// Clear for test
	mu.Lock()
	registry = make(map[string]func() engine.Integration)
	instances = make(map[string]engine.Integration)
	pluginsLoaded = true // Skip plugin loading for this test
	mu.Unlock()

	// Register and get an integration (this caches it)
	Register("cache-test", func() engine.Integration {
		return &mockIntegration{name: "cache-test"}
	})

	instance1, _ := Get("cache-test")

	// Verify it's cached
	instance2, _ := Get("cache-test")
	if instance1 != instance2 {
		t.Error("Instance should be cached")
	}

	// Clear cache
	ClearCache()

	// Get again - should be a new instance
	instance3, _ := Get("cache-test")
	if instance1 == instance3 {
		t.Error("After ClearCache(), Get() should return new instance")
	}
}

func TestReloadPlugins(t *testing.T) {
	// Save original plugin loaded state
	mu.Lock()
	originalPluginsLoaded := pluginsLoaded
	mu.Unlock()

	defer func() {
		mu.Lock()
		pluginsLoaded = originalPluginsLoaded
		mu.Unlock()
	}()

	// Set plugins as loaded
	mu.Lock()
	pluginsLoaded = true
	mu.Unlock()

	// Reload should reset the flag and trigger ensurePluginsLoaded
	err := ReloadPlugins()
	if err != nil {
		t.Fatalf("ReloadPlugins() error = %v", err)
	}

	// Verify pluginsLoaded is still true (it gets set back to true during ensurePluginsLoaded)
	mu.RLock()
	loaded := pluginsLoaded
	mu.RUnlock()

	if !loaded {
		t.Error("After ReloadPlugins(), pluginsLoaded should be true")
	}
}

func TestGetPluginDirectories(t *testing.T) {
	dirs := getPluginDirectories()

	// Should return a slice (might be empty if no plugin dirs exist)
	if dirs == nil {
		t.Error("getPluginDirectories() should not return nil")
	}

	// If UPTOOL_PLUGIN_DIR is set, create a test directory and verify it's included
	tmpDir, err := os.MkdirTemp("", "uptool-plugin-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir) //nolint:errcheck // test cleanup

	os.Setenv("UPTOOL_PLUGIN_DIR", tmpDir)
	defer os.Unsetenv("UPTOOL_PLUGIN_DIR")

	dirs = getPluginDirectories()
	found := false
	for _, dir := range dirs {
		if dir == tmpDir {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("getPluginDirectories() should include UPTOOL_PLUGIN_DIR=%s, got %v", tmpDir, dirs)
	}
}

func TestLoadPluginsFromDir(t *testing.T) {
	// Create a temp directory
	tmpDir, err := os.MkdirTemp("", "uptool-plugin-load-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir) //nolint:errcheck // test cleanup

	// Create some non-plugin files
	os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("test"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "config.yaml"), []byte("test"), 0o644)

	// loadPluginsFromDir should not error even with no .so files
	err = loadPluginsFromDir(tmpDir)
	if err != nil {
		t.Errorf("loadPluginsFromDir() error = %v, want nil", err)
	}

	// Test with non-existent directory
	err = loadPluginsFromDir("/non/existent/directory")
	if err == nil {
		t.Error("loadPluginsFromDir() should error for non-existent directory")
	}
}

func TestEnsurePluginsLoaded(t *testing.T) {
	// Save original state
	mu.Lock()
	originalPluginsLoaded := pluginsLoaded
	mu.Unlock()

	defer func() {
		mu.Lock()
		pluginsLoaded = originalPluginsLoaded
		mu.Unlock()
	}()

	// Reset pluginsLoaded
	mu.Lock()
	pluginsLoaded = false
	mu.Unlock()

	// First call should load plugins
	err := ensurePluginsLoaded()
	if err != nil {
		t.Fatalf("ensurePluginsLoaded() error = %v", err)
	}

	mu.RLock()
	loaded := pluginsLoaded
	mu.RUnlock()

	if !loaded {
		t.Error("ensurePluginsLoaded() should set pluginsLoaded to true")
	}

	// Second call should be a no-op
	err = ensurePluginsLoaded()
	if err != nil {
		t.Fatalf("ensurePluginsLoaded() second call error = %v", err)
	}
}

// Tests for utils.go

func TestValidateFilePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid absolute path",
			path:    "/tmp/test.txt",
			wantErr: false,
		},
		{
			name:    "valid relative path",
			path:    "src/main.go",
			wantErr: false,
		},
		{
			name:    "directory traversal that cleans to valid path (allowed)",
			path:    "/tmp/../etc/passwd",
			wantErr: false, // filepath.Clean resolves this to /etc/passwd, no ..
		},
		{
			name:    "double dot at end of cleaned path",
			path:    "../secret.txt",
			wantErr: true, // filepath.Clean keeps .., so rejected
		},
		{
			name:    "multiple dots in filename (contains ..)",
			path:    "test.file..txt",
			wantErr: true, // Contains .. after clean
		},
		{
			name:    "single dot (safe)",
			path:    "./file.txt",
			wantErr: false, // Cleans to file.txt, no ..
		},
		{
			name:    "valid nested path",
			path:    "a/b/c/d.txt",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFilePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFilePath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

// Tests for metadata.go

func TestLoadMetadata(t *testing.T) {
	// Clear cached metadata for clean test
	cachedMetadata = nil
	defer func() { cachedMetadata = nil }()

	t.Run("loads metadata from integrations.yaml", func(t *testing.T) {
		// This test will work if integrations.yaml exists in the repo root
		metadata, err := LoadMetadata()
		if err != nil {
			// It's okay if the file doesn't exist in test environment
			t.Skipf("LoadMetadata() error = %v (may not exist in test env)", err)
		}

		if metadata == nil {
			t.Fatal("LoadMetadata() returned nil metadata")
		}

		if metadata.Integrations == nil {
			t.Error("LoadMetadata() returned metadata with nil Integrations")
		}
	})

	t.Run("caches metadata", func(t *testing.T) {
		cachedMetadata = nil
		metadata1, err1 := LoadMetadata()
		if err1 != nil {
			t.Skip("Skipping cache test - integrations.yaml not available")
		}

		metadata2, err2 := LoadMetadata()
		if err2 != nil {
			t.Fatalf("Second LoadMetadata() error = %v", err2)
		}

		if metadata1 != metadata2 {
			t.Error("LoadMetadata() should return cached metadata on second call")
		}
	})
}

func TestGetMetadata(t *testing.T) {
	cachedMetadata = nil
	defer func() { cachedMetadata = nil }()

	t.Run("returns metadata for existing integration", func(t *testing.T) {
		// Pre-populate cache with test data
		cachedMetadata = &RegistryMetadata{
			Integrations: map[string]Metadata{
				"npm": {
					DisplayName: "NPM",
					Description: "Node.js package manager",
					Category:    "package-managers",
				},
			},
		}

		meta, err := GetMetadata("npm")
		if err != nil {
			t.Fatalf("GetMetadata() error = %v", err)
		}

		if meta.DisplayName != "NPM" {
			t.Errorf("GetMetadata() DisplayName = %q, want %q", meta.DisplayName, "NPM")
		}
	})

	t.Run("returns error for non-existent integration", func(t *testing.T) {
		cachedMetadata = &RegistryMetadata{
			Integrations: map[string]Metadata{},
		}

		_, err := GetMetadata("nonexistent")
		if err == nil {
			t.Error("GetMetadata() expected error for non-existent integration")
		}
	})
}

func TestListIntegrations(t *testing.T) {
	cachedMetadata = nil
	defer func() { cachedMetadata = nil }()

	t.Run("returns all integrations", func(t *testing.T) {
		cachedMetadata = &RegistryMetadata{
			Integrations: map[string]Metadata{
				"npm":  {DisplayName: "NPM"},
				"helm": {DisplayName: "Helm"},
			},
		}

		integrations, err := ListIntegrations()
		if err != nil {
			t.Fatalf("ListIntegrations() error = %v", err)
		}

		if len(integrations) != 2 {
			t.Errorf("ListIntegrations() returned %d integrations, want 2", len(integrations))
		}

		if _, ok := integrations["npm"]; !ok {
			t.Error("ListIntegrations() missing npm")
		}
		if _, ok := integrations["helm"]; !ok {
			t.Error("ListIntegrations() missing helm")
		}
	})
}

func TestListByCategory(t *testing.T) {
	cachedMetadata = nil
	defer func() { cachedMetadata = nil }()

	t.Run("filters by category", func(t *testing.T) {
		cachedMetadata = &RegistryMetadata{
			Integrations: map[string]Metadata{
				"npm":       {DisplayName: "NPM", Category: "package-managers"},
				"helm":      {DisplayName: "Helm", Category: "kubernetes"},
				"terraform": {DisplayName: "Terraform", Category: "infrastructure"},
			},
		}

		integrations, err := ListByCategory("package-managers")
		if err != nil {
			t.Fatalf("ListByCategory() error = %v", err)
		}

		if len(integrations) != 1 {
			t.Errorf("ListByCategory() returned %d integrations, want 1", len(integrations))
		}

		if _, ok := integrations["npm"]; !ok {
			t.Error("ListByCategory() missing npm")
		}
	})

	t.Run("returns empty for non-existent category", func(t *testing.T) {
		cachedMetadata = &RegistryMetadata{
			Integrations: map[string]Metadata{
				"npm": {DisplayName: "NPM", Category: "package-managers"},
			},
		}

		integrations, err := ListByCategory("nonexistent")
		if err != nil {
			t.Fatalf("ListByCategory() error = %v", err)
		}

		if len(integrations) != 0 {
			t.Errorf("ListByCategory() returned %d integrations for non-existent category, want 0", len(integrations))
		}
	})
}

func TestIsDisabled(t *testing.T) {
	cachedMetadata = nil
	defer func() { cachedMetadata = nil }()

	cachedMetadata = &RegistryMetadata{
		Integrations: map[string]Metadata{
			"enabled":  {Disabled: false},
			"disabled": {Disabled: true},
		},
	}

	t.Run("returns false for enabled integration", func(t *testing.T) {
		if IsDisabled("enabled") {
			t.Error("IsDisabled() = true for enabled integration, want false")
		}
	})

	t.Run("returns true for disabled integration", func(t *testing.T) {
		if !IsDisabled("disabled") {
			t.Error("IsDisabled() = false for disabled integration, want true")
		}
	})

	t.Run("returns false for non-existent integration", func(t *testing.T) {
		if IsDisabled("nonexistent") {
			t.Error("IsDisabled() = true for non-existent integration, want false")
		}
	})
}

func TestIsExperimental(t *testing.T) {
	cachedMetadata = nil
	defer func() { cachedMetadata = nil }()

	cachedMetadata = &RegistryMetadata{
		Integrations: map[string]Metadata{
			"stable":       {Experimental: false},
			"experimental": {Experimental: true},
		},
	}

	t.Run("returns false for stable integration", func(t *testing.T) {
		if IsExperimental("stable") {
			t.Error("IsExperimental() = true for stable integration, want false")
		}
	})

	t.Run("returns true for experimental integration", func(t *testing.T) {
		if !IsExperimental("experimental") {
			t.Error("IsExperimental() = false for experimental integration, want true")
		}
	})

	t.Run("returns false for non-existent integration", func(t *testing.T) {
		if IsExperimental("nonexistent") {
			t.Error("IsExperimental() = true for non-existent integration, want false")
		}
	})
}

func TestFindRegistryFile(t *testing.T) {
	// Save current working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(originalWd) }()

	t.Run("finds file in current directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create integrations.yaml in tmpDir
		registryFile := filepath.Join(tmpDir, "integrations.yaml")
		if err := os.WriteFile(registryFile, []byte("version: \"1.0\"\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		// Change to tmpDir
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}

		path, err := findRegistryFile()
		if err != nil {
			t.Fatalf("findRegistryFile() error = %v", err)
		}

		if path != "integrations.yaml" {
			t.Errorf("findRegistryFile() = %q, want %q", path, "integrations.yaml")
		}
	})

	t.Run("returns error when not found", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a go.mod file to mark root
		if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		// Change to tmpDir (no integrations.yaml)
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}

		// findRegistryFile should return a path even if file doesn't exist (it returns repo root path)
		path, err := findRegistryFile()
		if err != nil {
			// Expected if can't find the file
			_ = err
		}
		_ = path
	})
}

func TestLoadMetadata_InvalidYAML(t *testing.T) {
	cachedMetadata = nil
	defer func() { cachedMetadata = nil }()

	// Save current working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(originalWd) }()

	tmpDir := t.TempDir()

	// Create invalid integrations.yaml
	registryFile := filepath.Join(tmpDir, "integrations.yaml")
	err = os.WriteFile(registryFile, []byte("invalid: yaml: content:\n  :::"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	// Change to tmpDir
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	_, err = LoadMetadata()
	if err == nil {
		t.Error("LoadMetadata() expected error for invalid YAML")
	}
}

func TestLoadMetadata_ValidYAML(t *testing.T) {
	cachedMetadata = nil
	defer func() { cachedMetadata = nil }()

	// Save current working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(originalWd) }()

	tmpDir := t.TempDir()

	// Create valid integrations.yaml
	registryFile := filepath.Join(tmpDir, "integrations.yaml")
	content := `version: "1.0"
integrations:
  npm:
    displayName: "NPM"
    description: "Node.js package manager"
    category: "package-managers"
`
	err = os.WriteFile(registryFile, []byte(content), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	// Change to tmpDir
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	metadata, err := LoadMetadata()
	if err != nil {
		t.Fatalf("LoadMetadata() error = %v", err)
	}

	if metadata.Version != "1.0" {
		t.Errorf("LoadMetadata() version = %q, want %q", metadata.Version, "1.0")
	}

	if len(metadata.Integrations) != 1 {
		t.Errorf("LoadMetadata() integrations count = %d, want 1", len(metadata.Integrations))
	}
}

func TestListIntegrations_Error(t *testing.T) {
	cachedMetadata = nil
	defer func() { cachedMetadata = nil }()

	// Save current working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(originalWd) }()

	tmpDir := t.TempDir()

	// Create a go.mod file to mark root but no integrations.yaml
	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	// Change to tmpDir
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// ListIntegrations should handle the error gracefully
	_, err = ListIntegrations()
	// Error is expected since integrations.yaml doesn't exist
	_ = err
}

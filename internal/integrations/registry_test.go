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

func (m *mockIntegration) Plan(ctx context.Context, manifest *engine.Manifest) (*engine.UpdatePlan, error) {
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

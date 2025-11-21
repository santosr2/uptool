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

// Package integrations provides a central registry for all integration implementations.
// Integrations can be built-in (compiled into the binary) or external (loaded as plugins).
package integrations

//go:generate go run ../../scripts/gen_integrations.go

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"sort"
	"sync"

	"github.com/santosr2/uptool/internal/engine"
)

var (
	// registry holds all registered integration constructors
	registry = make(map[string]func() engine.Integration)
	// instances holds cached integration instances for lazy loading
	instances = make(map[string]engine.Integration)
	// mu protects registry and instances during access
	mu sync.RWMutex
	// pluginsLoaded tracks whether plugins have been discovered
	pluginsLoaded bool
)

// Register adds an integration constructor to the global registry.
// This is typically called from init() functions in integration packages.
//
// Example:
//
//	func init() {
//	    integrations.Register("npm", New)
//	}
func Register(name string, constructor func() engine.Integration) {
	mu.Lock()
	defer mu.Unlock()

	if _, exists := registry[name]; exists {
		panic("integration already registered: " + name)
	}

	registry[name] = constructor
}

// Get returns a single integration by name, creating it lazily if needed.
// This is more efficient than GetAll() when you only need specific integrations.
func Get(name string) (engine.Integration, error) {
	// Ensure plugins are loaded
	if err := ensurePluginsLoaded(); err != nil {
		return nil, fmt.Errorf("loading plugins: %w", err)
	}

	mu.Lock()
	defer mu.Unlock()

	// Check if already instantiated
	if instance, ok := instances[name]; ok {
		return instance, nil
	}

	// Get constructor
	constructor, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("integration %q not found", name)
	}

	// Create and cache instance
	instance := constructor()
	instances[name] = instance

	return instance, nil
}

// GetAll returns a map of all registered integrations.
// Uses lazy loading - only creates instances for integrations that haven't been created yet.
func GetAll() map[string]engine.Integration {
	// Ensure plugins are loaded
	if err := ensurePluginsLoaded(); err != nil {
		// Log error but continue with built-in integrations
		fmt.Fprintf(os.Stderr, "Warning: failed to load plugins: %v\n", err)
	}

	mu.Lock()
	defer mu.Unlock()

	result := make(map[string]engine.Integration, len(registry))
	for name, constructor := range registry {
		// Use cached instance if available
		if instance, ok := instances[name]; ok {
			result[name] = instance
		} else {
			// Create new instance and cache it
			instance := constructor()
			instances[name] = instance
			result[name] = instance
		}
	}

	return result
}

// GetLazy returns a map of constructors (not instances).
// Use this when you want to defer instantiation until actual use.
func GetLazy() map[string]func() engine.Integration {
	// Ensure plugins are loaded
	if err := ensurePluginsLoaded(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load plugins: %v\n", err)
	}

	mu.RLock()
	defer mu.RUnlock()

	// Return a copy of the registry
	result := make(map[string]func() engine.Integration, len(registry))
	for name, constructor := range registry {
		result[name] = constructor
	}

	return result
}

// List returns a sorted list of all registered integration names.
func List() []string {
	mu.RLock()
	defer mu.RUnlock()

	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

// Count returns the number of registered integrations.
func Count() int {
	mu.RLock()
	defer mu.RUnlock()

	return len(registry)
}

// ensurePluginsLoaded loads plugins from standard locations if not already loaded.
func ensurePluginsLoaded() error {
	mu.Lock()
	if pluginsLoaded {
		mu.Unlock()
		return nil
	}
	pluginsLoaded = true
	mu.Unlock()

	// Find plugin directories
	pluginDirs := getPluginDirectories()

	for _, dir := range pluginDirs {
		if err := loadPluginsFromDir(dir); err != nil {
			// Log but don't fail - continue with other directories
			fmt.Fprintf(os.Stderr, "Warning: error loading plugins from %s: %v\n", dir, err)
		}
	}

	return nil
}

// getPluginDirectories returns a list of directories to search for plugins.
func getPluginDirectories() []string {
	dirs := []string{}

	// 1. Current directory ./plugins
	if _, err := os.Stat("./plugins"); err == nil {
		dirs = append(dirs, "./plugins")
	}

	// 2. User's home directory ~/.uptool/plugins
	if home, err := os.UserHomeDir(); err == nil {
		userPluginDir := filepath.Join(home, ".uptool", "plugins")
		if _, err := os.Stat(userPluginDir); err == nil {
			dirs = append(dirs, userPluginDir)
		}
	}

	// 3. System-wide /usr/local/lib/uptool/plugins (Unix-like systems)
	systemPluginDir := "/usr/local/lib/uptool/plugins"
	if _, err := os.Stat(systemPluginDir); err == nil {
		dirs = append(dirs, systemPluginDir)
	}

	// 4. Environment variable UPTOOL_PLUGIN_DIR
	if envDir := os.Getenv("UPTOOL_PLUGIN_DIR"); envDir != "" {
		if _, err := os.Stat(envDir); err == nil {
			dirs = append(dirs, envDir)
		}
	}

	return dirs
}

// loadPluginsFromDir loads all .so plugin files from a directory.
func loadPluginsFromDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only load .so files (shared objects)
		if filepath.Ext(entry.Name()) != ".so" {
			continue
		}

		pluginPath := filepath.Join(dir, entry.Name())
		if err := loadPlugin(pluginPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load plugin %s: %v\n", pluginPath, err)
			continue
		}
	}

	return nil
}

// loadPlugin loads a single plugin file and registers its integrations.
func loadPlugin(path string) error {
	// Open the plugin
	p, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("opening plugin: %w", err)
	}

	// Look for the Register function
	// Plugin must export a function: func Register(func(string, func() engine.Integration))
	registerSymbol, err := p.Lookup("RegisterWith")
	if err != nil {
		return fmt.Errorf("plugin missing RegisterWith function: %w", err)
	}

	// Call the plugin's RegisterWith function, passing our Register function
	registerFunc, ok := registerSymbol.(func(func(string, func() engine.Integration)))
	if !ok {
		return fmt.Errorf("plugin RegisterWith has wrong signature")
	}

	// Plugin will call our Register function to register its integrations
	registerFunc(Register)

	return nil
}

// ClearCache clears all cached instances, forcing reinitialization on next access.
// Useful for testing or when integrations need to be refreshed.
func ClearCache() {
	mu.Lock()
	defer mu.Unlock()

	instances = make(map[string]engine.Integration)
}

// ReloadPlugins clears the plugin loaded flag and reloads all plugins.
// This allows hot-reloading of plugins without restarting the application.
func ReloadPlugins() error {
	mu.Lock()
	pluginsLoaded = false
	mu.Unlock()

	return ensurePluginsLoaded()
}

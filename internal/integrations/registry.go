// Package integrations provides a central registry for all integration implementations.
package integrations

//go:generate go run ../../scripts/gen_integrations.go

import (
	"sort"
	"sync"

	"github.com/santosr2/uptool/internal/engine"
)

var (
	// registry holds all registered integration constructors
	registry = make(map[string]func() engine.Integration)
	// mu protects registry during initialization
	mu sync.RWMutex
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

// GetAll returns a map of all registered integrations.
// Each call creates fresh instances via the registered constructors.
func GetAll() map[string]engine.Integration {
	mu.RLock()
	defer mu.RUnlock()

	result := make(map[string]engine.Integration, len(registry))
	for name, constructor := range registry {
		result[name] = constructor()
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

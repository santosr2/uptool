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

package guards

import (
	"context"
	"testing"
)

// mockGuard is a test guard implementation.
type mockGuard struct {
	checkError  error
	name        string
	description string
	checkResult bool
}

func (g *mockGuard) Name() string {
	return g.name
}

func (g *mockGuard) Description() string {
	return g.description
}

func (g *mockGuard) Check(_ context.Context, _ *Environment) (bool, error) {
	return g.checkResult, g.checkError
}

func TestRegistry_Register(t *testing.T) {
	registry := &Registry{
		guards: make(map[string]Guard),
	}

	guard := &mockGuard{
		name:        "test-guard",
		description: "Test guard for testing",
	}

	registry.Register(guard)

	if len(registry.guards) != 1 {
		t.Errorf("Registry has %d guards, want 1", len(registry.guards))
	}

	if _, exists := registry.guards["test-guard"]; !exists {
		t.Error("Guard was not registered")
	}
}

func TestRegistry_RegisterDuplicate(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when registering duplicate guard")
		}
	}()

	registry := &Registry{
		guards: make(map[string]Guard),
	}

	guard := &mockGuard{name: "duplicate"}

	registry.Register(guard)
	registry.Register(guard) // Should panic
}

func TestRegistry_Get(t *testing.T) {
	registry := &Registry{
		guards: make(map[string]Guard),
	}

	guard := &mockGuard{
		name:        "test-guard",
		description: "Test guard",
	}

	registry.Register(guard)

	// Test existing guard
	retrieved, exists := registry.Get("test-guard")
	if !exists {
		t.Fatal("Get() returned false for existing guard")
	}

	if retrieved.Name() != "test-guard" {
		t.Errorf("Get() returned guard with name %q, want %q", retrieved.Name(), "test-guard")
	}

	// Test non-existent guard
	_, exists = registry.Get("non-existent")
	if exists {
		t.Error("Get() returned true for non-existent guard")
	}
}

func TestRegistry_List(t *testing.T) {
	registry := &Registry{
		guards: make(map[string]Guard),
	}

	guards := []*mockGuard{
		{name: "guard-1"},
		{name: "guard-2"},
		{name: "guard-3"},
	}

	for _, g := range guards {
		registry.Register(g)
	}

	names := registry.List()

	if len(names) != 3 {
		t.Errorf("List() returned %d names, want 3", len(names))
	}

	// Check all expected names are present
	nameMap := make(map[string]bool)
	for _, name := range names {
		nameMap[name] = true
	}

	for _, g := range guards {
		if !nameMap[g.name] {
			t.Errorf("List() missing guard %q", g.name)
		}
	}
}

func TestRegistry_CheckGuard(t *testing.T) {
	ctx := context.Background()
	env := &Environment{
		GitHubRepo:     "owner/repo",
		GitHubToken:    "token",
		GitHubPRNumber: "123",
	}

	tests := []struct {
		name        string
		guardName   string
		checkResult bool
		wantOk      bool
		wantError   bool
	}{
		{
			name:        "successful check",
			guardName:   "success-guard",
			checkResult: true,
			wantOk:      true,
			wantError:   false,
		},
		{
			name:        "failed check",
			guardName:   "fail-guard",
			checkResult: false,
			wantOk:      false,
			wantError:   false,
		},
		{
			name:        "unknown guard",
			guardName:   "unknown",
			checkResult: false,
			wantOk:      false,
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := &Registry{
				guards: make(map[string]Guard),
			}

			if tt.guardName != "unknown" {
				registry.Register(&mockGuard{
					name:        tt.guardName,
					checkResult: tt.checkResult,
				})
			}

			ok, err := registry.CheckGuard(ctx, tt.guardName, env)

			if (err != nil) != tt.wantError {
				t.Errorf("CheckGuard() error = %v, wantError %v", err, tt.wantError)
			}

			if ok != tt.wantOk {
				t.Errorf("CheckGuard() = %v, want %v", ok, tt.wantOk)
			}
		})
	}
}

func TestGlobalRegistry(t *testing.T) {
	// Note: This test uses the global registry, so it may interact with other tests
	// and built-in guards. We're just testing the basic functionality.

	// Register a unique test guard
	testGuard := &mockGuard{
		name:        "global-test-guard-unique-12345",
		description: "Test guard for global registry",
		checkResult: true,
	}

	Register(testGuard)

	// Test Get
	retrieved, exists := Get("global-test-guard-unique-12345")
	if !exists {
		t.Fatal("Get() failed to retrieve guard from global registry")
	}

	if retrieved.Name() != testGuard.name {
		t.Errorf("Get() returned wrong guard name: got %q, want %q", retrieved.Name(), testGuard.name)
	}

	// Test List (should include our guard and any built-in ones)
	names := List()
	found := false
	for _, name := range names {
		if name == "global-test-guard-unique-12345" {
			found = true
			break
		}
	}

	if !found {
		t.Error("List() did not include our registered guard")
	}

	// Test CheckGuard
	ctx := context.Background()
	env := &Environment{
		GitHubRepo:     "owner/repo",
		GitHubPRNumber: "123",
	}

	ok, err := CheckGuard(ctx, "global-test-guard-unique-12345", env)
	if err != nil {
		t.Fatalf("CheckGuard() error = %v", err)
	}

	if !ok {
		t.Error("CheckGuard() returned false, want true")
	}
}

func TestEnvironmentStruct(t *testing.T) {
	env := &Environment{
		GitHubRepo:     "owner/repo",
		GitHubToken:    "secret-token",
		GitHubPRNumber: "42",
	}

	if env.GitHubRepo != "owner/repo" {
		t.Errorf("GitHubRepo = %q, want %q", env.GitHubRepo, "owner/repo")
	}

	if env.GitHubToken != "secret-token" {
		t.Errorf("GitHubToken = %q, want %q", env.GitHubToken, "secret-token")
	}

	if env.GitHubPRNumber != "42" {
		t.Errorf("GitHubPRNumber = %q, want %q", env.GitHubPRNumber, "42")
	}
}

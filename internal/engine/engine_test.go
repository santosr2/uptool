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

package engine

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// mockIntegration implements Integration for testing
type mockIntegration struct {
	detectError     error
	planError       error
	applyError      error
	validateError   error
	applyResult     *ApplyResult
	name            string
	detectManifests []*Manifest
	planUpdates     []Update
	detectCalls     int
	planCalls       int
	applyCalls      int
	mu              sync.Mutex
}

func (m *mockIntegration) Name() string {
	return m.name
}

func (m *mockIntegration) Detect(ctx context.Context, repoRoot string) ([]*Manifest, error) {
	m.mu.Lock()
	m.detectCalls++
	m.mu.Unlock()

	if m.detectError != nil {
		return nil, m.detectError
	}
	return m.detectManifests, nil
}

func (m *mockIntegration) Plan(ctx context.Context, manifest *Manifest, planCtx *PlanContext) (*UpdatePlan, error) {
	m.mu.Lock()
	m.planCalls++
	m.mu.Unlock()

	if m.planError != nil {
		return nil, m.planError
	}

	return &UpdatePlan{
		Manifest: manifest,
		Updates:  m.planUpdates,
		Strategy: "custom_rewrite",
	}, nil
}

func (m *mockIntegration) Apply(ctx context.Context, plan *UpdatePlan) (*ApplyResult, error) {
	m.mu.Lock()
	m.applyCalls++
	m.mu.Unlock()

	if m.applyError != nil {
		return nil, m.applyError
	}

	if m.applyResult != nil {
		return m.applyResult, nil
	}

	return &ApplyResult{
		Manifest: plan.Manifest,
		Applied:  len(plan.Updates),
		Failed:   0,
	}, nil
}

func (m *mockIntegration) Validate(ctx context.Context, manifest *Manifest) error {
	if m.validateError != nil {
		return m.validateError
	}
	return nil
}

func TestNewEngine(t *testing.T) {
	t.Run("creates engine with default logger", func(t *testing.T) {
		e := NewEngine(nil)
		if e == nil {
			t.Fatal("NewEngine() returned nil")
		}
		if e.logger == nil {
			t.Error("NewEngine() created nil logger")
		}
		if e.integrations == nil {
			t.Error("NewEngine() created nil integrations map")
		}
		if e.concurrency != 4 {
			t.Errorf("NewEngine() concurrency = %d, want 4", e.concurrency)
		}
	})

	t.Run("creates engine with custom logger", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
		e := NewEngine(logger)
		if e == nil {
			t.Fatal("NewEngine() returned nil")
		}
		if e.logger != logger {
			t.Error("NewEngine() did not use provided logger")
		}
	})
}

func TestRegister(t *testing.T) {
	e := NewEngine(nil)

	mock := &mockIntegration{name: "test-integration"}
	e.Register(mock)

	if len(e.integrations) != 1 {
		t.Errorf("Register() integrations count = %d, want 1", len(e.integrations))
	}

	retrieved, ok := e.integrations["test-integration"]
	if !ok {
		t.Fatal("Register() did not add integration to map")
	}
	if retrieved != mock {
		t.Error("Register() stored wrong integration")
	}
}

func TestScan(t *testing.T) {
	ctx := context.Background()

	t.Run("successful scan with single integration", func(t *testing.T) {
		e := NewEngine(nil)
		mock := &mockIntegration{
			name: "npm",
			detectManifests: []*Manifest{
				{Path: "package.json", Type: "npm"},
				{Path: "apps/frontend/package.json", Type: "npm"},
			},
		}
		e.Register(mock)

		result, err := e.Scan(ctx, "/test/repo", nil, nil)
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		if len(result.Manifests) != 2 {
			t.Errorf("Scan() manifests count = %d, want 2", len(result.Manifests))
		}
		if result.RepoRoot != "/test/repo" {
			t.Errorf("Scan() repo_root = %q, want %q", result.RepoRoot, "/test/repo")
		}
		if len(result.Errors) != 0 {
			t.Errorf("Scan() errors = %v, want none", result.Errors)
		}
		if mock.detectCalls != 1 {
			t.Errorf("Scan() detectCalls = %d, want 1", mock.detectCalls)
		}
	})

	t.Run("scan with multiple integrations", func(t *testing.T) {
		e := NewEngine(nil)

		npm := &mockIntegration{
			name: "npm",
			detectManifests: []*Manifest{
				{Path: "package.json", Type: "npm"},
			},
		}
		helm := &mockIntegration{
			name: "helm",
			detectManifests: []*Manifest{
				{Path: "charts/app/Chart.yaml", Type: "helm"},
			},
		}

		e.Register(npm)
		e.Register(helm)

		result, err := e.Scan(ctx, "/test/repo", nil, nil)
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		if len(result.Manifests) != 2 {
			t.Errorf("Scan() manifests count = %d, want 2", len(result.Manifests))
		}
	})

	t.Run("scan with integration error", func(t *testing.T) {
		e := NewEngine(nil)

		failing := &mockIntegration{
			name:        "failing",
			detectError: errors.New("detection failed"),
		}
		working := &mockIntegration{
			name: "working",
			detectManifests: []*Manifest{
				{Path: "test.yaml", Type: "working"},
			},
		}

		e.Register(failing)
		e.Register(working)

		result, err := e.Scan(ctx, "/test/repo", nil, nil)
		if err != nil {
			t.Fatalf("Scan() error = %v, want nil (errors should be in result)", err)
		}

		// Should still get manifests from working integration
		if len(result.Manifests) != 1 {
			t.Errorf("Scan() manifests count = %d, want 1", len(result.Manifests))
		}

		// Should record the error
		if len(result.Errors) != 1 {
			t.Errorf("Scan() errors count = %d, want 1", len(result.Errors))
		} else if !strings.Contains(result.Errors[0], "failing") {
			t.Errorf("Scan() error = %q, want error mentioning 'failing'", result.Errors[0])
		}
	})

	t.Run("scan with only filter", func(t *testing.T) {
		e := NewEngine(nil)

		npm := &mockIntegration{
			name: "npm",
			detectManifests: []*Manifest{
				{Path: "package.json", Type: "npm"},
			},
		}
		helm := &mockIntegration{
			name:            "helm",
			detectManifests: []*Manifest{},
		}

		e.Register(npm)
		e.Register(helm)

		_, err := e.Scan(ctx, "/test/repo", []string{"npm"}, nil)
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		// Only npm should be scanned
		if npm.detectCalls != 1 {
			t.Errorf("npm detectCalls = %d, want 1", npm.detectCalls)
		}
		if helm.detectCalls != 0 {
			t.Errorf("helm detectCalls = %d, want 0", helm.detectCalls)
		}
	})

	t.Run("scan with exclude filter", func(t *testing.T) {
		e := NewEngine(nil)

		npm := &mockIntegration{
			name: "npm",
			detectManifests: []*Manifest{
				{Path: "package.json", Type: "npm"},
			},
		}
		helm := &mockIntegration{
			name:            "helm",
			detectManifests: []*Manifest{},
		}

		e.Register(npm)
		e.Register(helm)

		_, err := e.Scan(ctx, "/test/repo", nil, []string{"helm"})
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		// Only npm should be scanned
		if npm.detectCalls != 1 {
			t.Errorf("npm detectCalls = %d, want 1", npm.detectCalls)
		}
		if helm.detectCalls != 0 {
			t.Errorf("helm detectCalls = %d, want 0", helm.detectCalls)
		}
	})
}

func TestPlan(t *testing.T) {
	ctx := context.Background()

	t.Run("generates plans for manifests", func(t *testing.T) {
		e := NewEngine(nil)

		mock := &mockIntegration{
			name: "npm",
			planUpdates: []Update{
				{
					Dependency:    Dependency{Name: "react", CurrentVersion: "17.0.0"},
					TargetVersion: "18.0.0",
					Impact:        "major",
				},
			},
		}
		e.Register(mock)

		manifests := []*Manifest{
			{Path: "package.json", Type: "npm"},
		}

		result, err := e.Plan(ctx, manifests)
		if err != nil {
			t.Fatalf("Plan() error = %v", err)
		}

		if len(result.Plans) != 1 {
			t.Errorf("Plan() plans count = %d, want 1", len(result.Plans))
		}
		if len(result.Plans[0].Updates) != 1 {
			t.Errorf("Plan() updates count = %d, want 1", len(result.Plans[0].Updates))
		}
	})

	t.Run("includes plans with no updates", func(t *testing.T) {
		e := NewEngine(nil)

		mock := &mockIntegration{
			name:        "npm",
			planUpdates: []Update{}, // No updates
		}
		e.Register(mock)

		manifests := []*Manifest{
			{Path: "package.json", Type: "npm"},
		}

		result, err := e.Plan(ctx, manifests)
		if err != nil {
			t.Fatalf("Plan() error = %v", err)
		}

		// Should include plans even with no updates (for --show-up-to-date flag)
		if len(result.Plans) != 1 {
			t.Errorf("Plan() plans count = %d, want 1", len(result.Plans))
		}

		// But the plan should have 0 updates
		if len(result.Plans[0].Updates) != 0 {
			t.Errorf("Plan() updates count = %d, want 0 (no updates)", len(result.Plans[0].Updates))
		}
	})

	t.Run("handles missing integration", func(t *testing.T) {
		e := NewEngine(nil)

		manifests := []*Manifest{
			{Path: "package.json", Type: "npm"}, // npm not registered
		}

		result, err := e.Plan(ctx, manifests)
		if err != nil {
			t.Fatalf("Plan() error = %v, want nil (errors should be in result)", err)
		}

		if len(result.Errors) != 1 {
			t.Errorf("Plan() errors count = %d, want 1", len(result.Errors))
		} else if !strings.Contains(result.Errors[0], "no integration") {
			t.Errorf("Plan() error = %q, want error about missing integration", result.Errors[0])
		}
	})

	t.Run("handles plan errors", func(t *testing.T) {
		e := NewEngine(nil)

		mock := &mockIntegration{
			name:      "npm",
			planError: errors.New("plan failed"),
		}
		e.Register(mock)

		manifests := []*Manifest{
			{Path: "package.json", Type: "npm"},
		}

		result, err := e.Plan(ctx, manifests)
		if err != nil {
			t.Fatalf("Plan() error = %v, want nil (errors should be in result)", err)
		}

		if len(result.Errors) != 1 {
			t.Errorf("Plan() errors count = %d, want 1", len(result.Errors))
		}
	})

	t.Run("processes multiple manifests concurrently", func(t *testing.T) {
		e := NewEngine(nil)

		mock := &mockIntegration{
			name: "npm",
			planUpdates: []Update{
				{
					Dependency:    Dependency{Name: "react", CurrentVersion: "17.0.0"},
					TargetVersion: "18.0.0",
					Impact:        "major",
				},
			},
		}
		e.Register(mock)

		manifests := []*Manifest{
			{Path: "package.json", Type: "npm"},
			{Path: "apps/frontend/package.json", Type: "npm"},
			{Path: "apps/backend/package.json", Type: "npm"},
		}

		result, err := e.Plan(ctx, manifests)
		if err != nil {
			t.Fatalf("Plan() error = %v", err)
		}

		if len(result.Plans) != 3 {
			t.Errorf("Plan() plans count = %d, want 3", len(result.Plans))
		}
		if mock.planCalls != 3 {
			t.Errorf("Plan() planCalls = %d, want 3", mock.planCalls)
		}
	})
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()

	t.Run("applies updates successfully", func(t *testing.T) {
		e := NewEngine(nil)

		mock := &mockIntegration{
			name: "npm",
		}
		e.Register(mock)

		plans := []*UpdatePlan{
			{
				Manifest: &Manifest{Path: "package.json", Type: "npm"},
				Updates: []Update{
					{
						Dependency:    Dependency{Name: "react", CurrentVersion: "17.0.0"},
						TargetVersion: "18.0.0",
					},
				},
			},
		}

		result, err := e.Update(ctx, plans, false)
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		if len(result.Results) != 1 {
			t.Errorf("Update() results count = %d, want 1", len(result.Results))
		}
		if result.Results[0].Applied != 1 {
			t.Errorf("Update() applied = %d, want 1", result.Results[0].Applied)
		}
		if mock.applyCalls != 1 {
			t.Errorf("Update() applyCalls = %d, want 1", mock.applyCalls)
		}
	})

	t.Run("dry-run mode does not apply changes", func(t *testing.T) {
		e := NewEngine(nil)

		mock := &mockIntegration{
			name: "npm",
		}
		e.Register(mock)

		plans := []*UpdatePlan{
			{
				Manifest: &Manifest{Path: "package.json", Type: "npm"},
				Updates: []Update{
					{
						Dependency:    Dependency{Name: "react", CurrentVersion: "17.0.0"},
						TargetVersion: "18.0.0",
					},
				},
			},
		}

		result, err := e.Update(ctx, plans, true)
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		if result.Results != nil {
			t.Errorf("Update() dry-run results = %v, want nil", result.Results)
		}
		if mock.applyCalls != 0 {
			t.Errorf("Update() dry-run applyCalls = %d, want 0", mock.applyCalls)
		}
	})

	t.Run("handles missing integration", func(t *testing.T) {
		e := NewEngine(nil)

		plans := []*UpdatePlan{
			{
				Manifest: &Manifest{Path: "package.json", Type: "npm"}, // npm not registered
				Updates:  []Update{{Dependency: Dependency{Name: "react"}}},
			},
		}

		result, err := e.Update(ctx, plans, false)
		if err != nil {
			t.Fatalf("Update() error = %v, want nil (errors should be in result)", err)
		}

		if len(result.Errors) != 1 {
			t.Errorf("Update() errors count = %d, want 1", len(result.Errors))
		}
	})

	t.Run("handles apply errors", func(t *testing.T) {
		e := NewEngine(nil)

		mock := &mockIntegration{
			name:       "npm",
			applyError: errors.New("apply failed"),
		}
		e.Register(mock)

		plans := []*UpdatePlan{
			{
				Manifest: &Manifest{Path: "package.json", Type: "npm"},
				Updates:  []Update{{Dependency: Dependency{Name: "react"}}},
			},
		}

		result, err := e.Update(ctx, plans, false)
		if err != nil {
			t.Fatalf("Update() error = %v, want nil (errors should be in result)", err)
		}

		if len(result.Errors) != 1 {
			t.Errorf("Update() errors count = %d, want 1", len(result.Errors))
		}
		if len(result.Results) != 0 {
			t.Errorf("Update() results count = %d, want 0 (failed apply should not add result)", len(result.Results))
		}
	})

	t.Run("processes multiple plans concurrently", func(t *testing.T) {
		e := NewEngine(nil)

		mock := &mockIntegration{
			name: "npm",
		}
		e.Register(mock)

		plans := []*UpdatePlan{
			{
				Manifest: &Manifest{Path: "package.json", Type: "npm"},
				Updates:  []Update{{Dependency: Dependency{Name: "react"}}},
			},
			{
				Manifest: &Manifest{Path: "apps/frontend/package.json", Type: "npm"},
				Updates:  []Update{{Dependency: Dependency{Name: "vue"}}},
			},
			{
				Manifest: &Manifest{Path: "apps/backend/package.json", Type: "npm"},
				Updates:  []Update{{Dependency: Dependency{Name: "express"}}},
			},
		}

		result, err := e.Update(ctx, plans, false)
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		if len(result.Results) != 3 {
			t.Errorf("Update() results count = %d, want 3", len(result.Results))
		}
		if mock.applyCalls != 3 {
			t.Errorf("Update() applyCalls = %d, want 3", mock.applyCalls)
		}
	})
}

func TestFilterIntegrations(t *testing.T) {
	e := NewEngine(nil)

	npm := &mockIntegration{name: "npm"}
	helm := &mockIntegration{name: "helm"}
	terraform := &mockIntegration{name: "terraform"}

	e.Register(npm)
	e.Register(helm)
	e.Register(terraform)

	t.Run("no filters returns all", func(t *testing.T) {
		result := e.filterIntegrations(nil, nil)
		if len(result) != 3 {
			t.Errorf("filterIntegrations() count = %d, want 3", len(result))
		}
	})

	t.Run("only filter includes specified", func(t *testing.T) {
		result := e.filterIntegrations([]string{"npm", "helm"}, nil)
		if len(result) != 2 {
			t.Errorf("filterIntegrations() count = %d, want 2", len(result))
		}
		if _, ok := result["npm"]; !ok {
			t.Error("filterIntegrations() missing npm")
		}
		if _, ok := result["helm"]; !ok {
			t.Error("filterIntegrations() missing helm")
		}
		if _, ok := result["terraform"]; ok {
			t.Error("filterIntegrations() should not include terraform")
		}
	})

	t.Run("exclude filter removes specified", func(t *testing.T) {
		result := e.filterIntegrations(nil, []string{"terraform"})
		if len(result) != 2 {
			t.Errorf("filterIntegrations() count = %d, want 2", len(result))
		}
		if _, ok := result["npm"]; !ok {
			t.Error("filterIntegrations() missing npm")
		}
		if _, ok := result["helm"]; !ok {
			t.Error("filterIntegrations() missing helm")
		}
		if _, ok := result["terraform"]; ok {
			t.Error("filterIntegrations() should not include terraform")
		}
	})

	t.Run("only takes precedence over exclude", func(t *testing.T) {
		result := e.filterIntegrations([]string{"npm"}, []string{"helm"})
		if len(result) != 1 {
			t.Errorf("filterIntegrations() count = %d, want 1", len(result))
		}
		if _, ok := result["npm"]; !ok {
			t.Error("filterIntegrations() missing npm")
		}
	})

	t.Run("handles non-existent integration in only", func(t *testing.T) {
		result := e.filterIntegrations([]string{"nonexistent"}, nil)
		if len(result) != 0 {
			t.Errorf("filterIntegrations() count = %d, want 0", len(result))
		}
	})
}

func TestGetIntegration(t *testing.T) {
	e := NewEngine(nil)

	mock := &mockIntegration{name: "npm"}
	e.Register(mock)

	t.Run("retrieves registered integration", func(t *testing.T) {
		integ, ok := e.GetIntegration("npm")
		if !ok {
			t.Fatal("GetIntegration() ok = false, want true")
		}
		if integ != mock {
			t.Error("GetIntegration() returned wrong integration")
		}
	})

	t.Run("returns false for missing integration", func(t *testing.T) {
		_, ok := e.GetIntegration("nonexistent")
		if ok {
			t.Error("GetIntegration() ok = true, want false")
		}
	})
}

func TestListIntegrations(t *testing.T) {
	e := NewEngine(nil)

	t.Run("returns empty list for no integrations", func(t *testing.T) {
		names := e.ListIntegrations()
		if len(names) != 0 {
			t.Errorf("ListIntegrations() count = %d, want 0", len(names))
		}
	})

	t.Run("returns all integration names", func(t *testing.T) {
		e.Register(&mockIntegration{name: "npm"})
		e.Register(&mockIntegration{name: "helm"})
		e.Register(&mockIntegration{name: "terraform"})

		names := e.ListIntegrations()
		if len(names) != 3 {
			t.Errorf("ListIntegrations() count = %d, want 3", len(names))
		}

		// Check all names are present
		nameMap := make(map[string]bool)
		for _, name := range names {
			nameMap[name] = true
		}

		if !nameMap["npm"] || !nameMap["helm"] || !nameMap["terraform"] {
			t.Errorf("ListIntegrations() names = %v, missing expected names", names)
		}
	})
}

// slowMockIntegration simulates slow operations and tracks concurrency
type slowMockIntegration struct {
	concurrencyTracker *concurrencyTracker
	mockIntegration
	delay time.Duration
}

type concurrencyTracker struct {
	mu      sync.Mutex
	current int
	max     int
}

func (c *concurrencyTracker) enter() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.current++
	if c.current > c.max {
		c.max = c.current
	}
}

func (c *concurrencyTracker) exit() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.current--
}

func (c *concurrencyTracker) getMax() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.max
}

func (s *slowMockIntegration) Detect(ctx context.Context, repoRoot string) ([]*Manifest, error) {
	s.concurrencyTracker.enter()
	defer s.concurrencyTracker.exit()

	time.Sleep(s.delay)
	return s.mockIntegration.Detect(ctx, repoRoot)
}

func TestConcurrency(t *testing.T) {
	ctx := context.Background()

	t.Run("respects concurrency limit in Scan", func(t *testing.T) {
		e := NewEngine(nil)
		e.concurrency = 2

		tracker := &concurrencyTracker{}

		// Create multiple slow integrations
		for i := 0; i < 5; i++ {
			slow := &slowMockIntegration{
				mockIntegration: mockIntegration{
					name:            fmt.Sprintf("integration-%d", i),
					detectManifests: []*Manifest{},
				},
				delay:              20 * time.Millisecond,
				concurrencyTracker: tracker,
			}
			e.Register(slow)
		}

		_, err := e.Scan(ctx, "/test", nil, nil)
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		maxConcurrent := tracker.getMax()
		if maxConcurrent > e.concurrency {
			t.Errorf("Scan() maxConcurrent = %d, want <= %d", maxConcurrent, e.concurrency)
		}
		if maxConcurrent < 1 {
			t.Errorf("Scan() maxConcurrent = %d, want >= 1", maxConcurrent)
		}
	})
}

func TestScanTimestamp(t *testing.T) {
	ctx := context.Background()
	e := NewEngine(nil)

	mock := &mockIntegration{
		name:            "npm",
		detectManifests: []*Manifest{{Path: "package.json", Type: "npm"}},
	}
	e.Register(mock)

	before := time.Now()
	result, err := e.Scan(ctx, "/test", nil, nil)
	after := time.Now()

	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if result.Timestamp.Before(before) || result.Timestamp.After(after) {
		t.Errorf("Scan() timestamp = %v, want between %v and %v", result.Timestamp, before, after)
	}
}

func TestPlanTimestamp(t *testing.T) {
	ctx := context.Background()
	e := NewEngine(nil)

	mock := &mockIntegration{
		name: "npm",
		planUpdates: []Update{
			{Dependency: Dependency{Name: "react"}, TargetVersion: "18.0.0"},
		},
	}
	e.Register(mock)

	manifests := []*Manifest{{Path: "package.json", Type: "npm"}}

	before := time.Now()
	result, err := e.Plan(ctx, manifests)
	after := time.Now()

	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}

	if result.Timestamp.Before(before) || result.Timestamp.After(after) {
		t.Errorf("Plan() timestamp = %v, want between %v and %v", result.Timestamp, before, after)
	}
}

func TestUpdateTimestamp(t *testing.T) {
	ctx := context.Background()
	e := NewEngine(nil)

	mock := &mockIntegration{name: "npm"}
	e.Register(mock)

	plans := []*UpdatePlan{
		{
			Manifest: &Manifest{Path: "package.json", Type: "npm"},
			Updates:  []Update{{Dependency: Dependency{Name: "react"}}},
		},
	}

	before := time.Now()
	result, err := e.Update(ctx, plans, false)
	after := time.Now()

	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if result.Timestamp.Before(before) || result.Timestamp.After(after) {
		t.Errorf("Update() timestamp = %v, want between %v and %v", result.Timestamp, before, after)
	}
}

func TestEngine_SetMatchConfigs(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	e := NewEngine(logger)

	matchConfigs := map[string]*MatchConfig{
		"npm": {
			Files: []string{"package.json", "apps/*/package.json"},
		},
		"terraform": {
			Files: []string{"*.tf", "modules/**/*.tf"},
		},
	}

	e.SetMatchConfigs(matchConfigs)

	if len(e.matchConfigs) != 2 {
		t.Errorf("SetMatchConfigs() stored %d configs, want 2", len(e.matchConfigs))
	}

	if len(e.matchConfigs["npm"].Files) != 2 {
		t.Errorf("SetMatchConfigs() npm has %d patterns, want 2", len(e.matchConfigs["npm"].Files))
	}
}

func TestEngine_FilterManifestsByPattern(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	e := NewEngine(logger)

	manifests := []*Manifest{
		{Path: "package.json", Type: "npm"},
		{Path: "apps/frontend/package.json", Type: "npm"},
		{Path: "apps/backend/package.json", Type: "npm"},
		{Path: "libs/shared/package.json", Type: "npm"},
	}

	tests := []struct {
		name        string
		matchConfig *MatchConfig
		want        int
	}{
		{
			name: "match all",
			matchConfig: &MatchConfig{
				Files: []string{"package.json", "apps/*/package.json", "libs/*/package.json"},
			},
			want: 4,
		},
		{
			name: "match root only",
			matchConfig: &MatchConfig{
				Files: []string{"package.json"},
			},
			want: 1,
		},
		{
			name: "match apps only",
			matchConfig: &MatchConfig{
				Files: []string{"apps/*/package.json"},
			},
			want: 2,
		},
		{
			name: "no matches",
			matchConfig: &MatchConfig{
				Files: []string{"services/*/package.json"},
			},
			want: 0,
		},
		{
			name: "match all but exclude libs",
			matchConfig: &MatchConfig{
				Files:   []string{"package.json", "apps/*/package.json", "libs/*/package.json"},
				Exclude: []string{"libs/*/package.json"},
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := e.filterManifestsByPattern(manifests, tt.matchConfig, "/repo")
			if len(filtered) != tt.want {
				t.Errorf("filterManifestsByPattern() = %d manifests, want %d", len(filtered), tt.want)
			}
		})
	}
}

func TestEngine_ScanWithMatchFiltering(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	e := NewEngine(logger)

	// Create mock integration that returns multiple manifests
	mock := &mockIntegration{
		name: "npm",
		detectManifests: []*Manifest{
			{Path: "package.json", Type: "npm"},
			{Path: "apps/frontend/package.json", Type: "npm"},
			{Path: "apps/backend/package.json", Type: "npm"},
		},
	}

	e.Register(mock)

	// Set match config to only include root package.json
	e.SetMatchConfigs(map[string]*MatchConfig{
		"npm": {
			Files: []string{"package.json"},
		},
	})

	ctx := context.Background()
	result, err := e.Scan(ctx, "/repo", nil, nil)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	// Should only have 1 manifest after filtering
	if len(result.Manifests) != 1 {
		t.Errorf("Scan() with match filtering = %d manifests, want 1", len(result.Manifests))
	}

	if result.Manifests[0].Path != "package.json" {
		t.Errorf("Scan() filtered manifest path = %s, want package.json", result.Manifests[0].Path)
	}
}

// Package engine provides the core orchestration layer for uptool.
// It manages integration registration, manifest scanning, update planning, and update application.
// The Engine coordinates concurrent operations across multiple integrations while handling errors and logging.
package engine

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// Engine orchestrates the scan, plan, and update operations.
type Engine struct {
	integrations map[string]Integration
	logger       *slog.Logger
	concurrency  int
}

// NewEngine creates a new engine with the given integrations.
func NewEngine(logger *slog.Logger) *Engine {
	if logger == nil {
		logger = slog.Default()
	}

	return &Engine{
		integrations: make(map[string]Integration),
		logger:       logger,
		concurrency:  4,
	}
}

// Register adds an integration to the engine.
func (e *Engine) Register(integration Integration) {
	e.integrations[integration.Name()] = integration
	e.logger.Info("registered integration", "name", integration.Name())
}

// Scan discovers all manifests across registered integrations.
func (e *Engine) Scan(ctx context.Context, repoRoot string, only, exclude []string) (*ScanResult, error) {
	e.logger.Info("starting scan", "repo", repoRoot)
	start := time.Now()

	integrations := e.filterIntegrations(only, exclude)

	var (
		mu        sync.Mutex
		manifests []*Manifest
		errors    []string
		wg        sync.WaitGroup
	)

	sem := make(chan struct{}, e.concurrency)

	for name, integration := range integrations {
		wg.Add(1)
		go func(n string, integ Integration) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			found, err := integ.Detect(ctx, repoRoot)
			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", n, err))
				e.logger.Error("detect failed", "integration", n, "error", err)
				return
			}

			manifests = append(manifests, found...)
			e.logger.Info("scan complete", "integration", n, "found", len(found))
		}(name, integration)
	}

	wg.Wait()

	e.logger.Info("scan finished", "duration", time.Since(start), "manifests", len(manifests))

	return &ScanResult{
		Manifests: manifests,
		Timestamp: time.Now(),
		RepoRoot:  repoRoot,
		Errors:    errors,
	}, nil
}

// Plan generates update plans for all manifests.
func (e *Engine) Plan(ctx context.Context, manifests []*Manifest) (*PlanResult, error) {
	e.logger.Info("starting plan", "manifests", len(manifests))
	start := time.Now()

	var (
		mu     sync.Mutex
		plans  []*UpdatePlan
		errors []string
		wg     sync.WaitGroup
	)

	sem := make(chan struct{}, e.concurrency)

	for _, manifest := range manifests {
		wg.Add(1)
		go func(m *Manifest) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			integration, ok := e.integrations[m.Type]
			if !ok {
				mu.Lock()
				errors = append(errors, fmt.Sprintf("no integration for type: %s", m.Type))
				mu.Unlock()
				return
			}

			plan, err := integration.Plan(ctx, m)
			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				errors = append(errors, fmt.Sprintf("%s (%s): %v", m.Path, m.Type, err))
				e.logger.Error("plan failed", "manifest", m.Path, "error", err)
				return
			}

			if len(plan.Updates) > 0 {
				plans = append(plans, plan)
				e.logger.Info("plan created", "manifest", m.Path, "updates", len(plan.Updates))
			}
		}(manifest)
	}

	wg.Wait()

	e.logger.Info("plan finished", "duration", time.Since(start), "plans", len(plans))

	return &PlanResult{
		Plans:     plans,
		Timestamp: time.Now(),
		Errors:    errors,
	}, nil
}

// Update applies update plans.
func (e *Engine) Update(ctx context.Context, plans []*UpdatePlan, dryRun bool) (*UpdateResult, error) {
	e.logger.Info("starting update", "plans", len(plans), "dry_run", dryRun)
	start := time.Now()

	if dryRun {
		e.logger.Info("dry-run mode: no changes will be applied")
		return &UpdateResult{
			Results:   nil,
			Timestamp: time.Now(),
		}, nil
	}

	var (
		mu      sync.Mutex
		results []*ApplyResult
		errors  []string
		wg      sync.WaitGroup
	)

	sem := make(chan struct{}, e.concurrency)

	for _, plan := range plans {
		wg.Add(1)
		go func(p *UpdatePlan) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			integration, ok := e.integrations[p.Manifest.Type]
			if !ok {
				mu.Lock()
				errors = append(errors, fmt.Sprintf("no integration for type: %s", p.Manifest.Type))
				mu.Unlock()
				return
			}

			result, err := integration.Apply(ctx, p)
			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", p.Manifest.Path, err))
				e.logger.Error("apply failed", "manifest", p.Manifest.Path, "error", err)
				return
			}

			results = append(results, result)
			e.logger.Info("apply complete", "manifest", p.Manifest.Path, "applied", result.Applied)
		}(plan)
	}

	wg.Wait()

	e.logger.Info("update finished", "duration", time.Since(start), "results", len(results))

	return &UpdateResult{
		Results:   results,
		Timestamp: time.Now(),
		Errors:    errors,
	}, nil
}

// filterIntegrations returns integrations based on only/exclude filters.
func (e *Engine) filterIntegrations(only, exclude []string) map[string]Integration {
	if len(only) == 0 && len(exclude) == 0 {
		return e.integrations
	}

	result := make(map[string]Integration)

	// If "only" is specified, only include those
	if len(only) > 0 {
		for _, name := range only {
			if integ, ok := e.integrations[name]; ok {
				result[name] = integ
			}
		}
		return result
	}

	// Otherwise, include all except excluded
	excludeMap := make(map[string]bool)
	for _, name := range exclude {
		excludeMap[name] = true
	}

	for name, integ := range e.integrations {
		if !excludeMap[name] {
			result[name] = integ
		}
	}

	return result
}

// GetIntegration retrieves a registered integration by name.
func (e *Engine) GetIntegration(name string) (Integration, bool) {
	integ, ok := e.integrations[name]
	return integ, ok
}

// ListIntegrations returns all registered integration names.
func (e *Engine) ListIntegrations() []string {
	names := make([]string, 0, len(e.integrations))
	for name := range e.integrations {
		names = append(names, name)
	}
	return names
}

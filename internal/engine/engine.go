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

// Package engine provides the core orchestration layer for uptool.
// It manages integration registration, manifest scanning, update planning, and update application.
// The Engine coordinates concurrent operations across multiple integrations while handling errors and logging.
package engine

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"sync"
	"time"
)

// Engine orchestrates the scan, plan, and update operations.
type Engine struct {
	integrations map[string]Integration
	policies     map[string]IntegrationPolicy
	matchConfigs map[string]*MatchConfig // integration -> match configuration (files + exclude)
	logger       *slog.Logger
	cliFlags     *CLIFlags
	concurrency  int
}

// NewEngine creates a new engine with the given integrations.
func NewEngine(logger *slog.Logger) *Engine {
	if logger == nil {
		logger = slog.Default()
	}

	return &Engine{
		integrations: make(map[string]Integration),
		policies:     make(map[string]IntegrationPolicy),
		matchConfigs: make(map[string]*MatchConfig),
		logger:       logger,
		concurrency:  4,
	}
}

// SetPolicies configures integration policies from uptool.yaml.
// These policies have the highest precedence in determining allowed updates.
func (e *Engine) SetPolicies(policies map[string]IntegrationPolicy) {
	e.policies = policies
	e.logger.Debug("set integration policies", "count", len(policies))
}

// SetMatchConfigs configures file pattern matching for integrations.
// Manifests will be filtered to only include those matching the configured patterns
// and exclude those matching exclude patterns.
func (e *Engine) SetMatchConfigs(configs map[string]*MatchConfig) {
	e.matchConfigs = configs
	e.logger.Debug("set match configs", "count", len(configs))
}

// SetCLIFlags configures CLI flag overrides for update behavior.
// These override manifest constraints but not uptool.yaml policies.
func (e *Engine) SetCLIFlags(flags *CLIFlags) {
	e.cliFlags = flags
	if flags != nil {
		e.logger.Debug("set CLI flags", "update_level", flags.UpdateLevel)
	}
}

// getPlanContext creates a PlanContext for a specific integration.
// It combines the integration's policy (if any) with CLI flags.
func (e *Engine) getPlanContext(integrationName string) *PlanContext {
	ctx := NewPlanContext()

	// Set policy if one exists for this integration
	if policy, ok := e.policies[integrationName]; ok {
		ctx = ctx.WithPolicy(&policy)
	}

	// Set CLI flags if provided
	if e.cliFlags != nil {
		ctx = ctx.WithCLIFlags(e.cliFlags)
	}

	return ctx
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

			// Filter manifests by match patterns if configured
			if matchConfig, ok := e.matchConfigs[n]; ok && matchConfig != nil {
				filtered := e.filterManifestsByPattern(found, matchConfig, repoRoot)
				manifests = append(manifests, filtered...)
				e.logger.Info("scan complete", "integration", n, "found", len(found), "filtered", len(filtered))
			} else {
				manifests = append(manifests, found...)
				e.logger.Info("scan complete", "integration", n, "found", len(found))
			}
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
// It applies policy precedence: CLI flags > uptool.yaml > manifest constraints.
// If cadence policies are configured, manifests are filtered based on their last check time.
func (e *Engine) Plan(ctx context.Context, manifests []*Manifest) (*PlanResult, error) {
	e.logger.Info("starting plan", "manifests", len(manifests))
	start := time.Now()

	// Filter manifests by cadence if policies are configured
	// Note: Cadence filtering requires state management which is not yet fully implemented
	// For now, cadence is validated in config but not enforced during execution
	// TODO: Implement state file management for cadence tracking

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

			// Get the plan context with policy and CLI flags for this integration
			planCtx := e.getPlanContext(m.Type)

			e.logger.Debug("planning manifest",
				"manifest", m.Path,
				"integration", m.Type,
				"update_level", planCtx.EffectiveUpdateLevel(),
				"allow_prerelease", planCtx.EffectiveAllowPrerelease(),
			)

			plan, err := integration.Plan(ctx, m, planCtx)
			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				errors = append(errors, fmt.Sprintf("%s (%s): %v", m.Path, m.Type, err))
				e.logger.Error("plan failed", "manifest", m.Path, "error", err)
				return
			}

			// Always include plans, even if they have no updates
			// This allows the output layer to decide whether to show them
			plans = append(plans, plan)
			if len(plan.Updates) > 0 {
				e.logger.Info("plan created", "manifest", m.Path, "updates", len(plan.Updates))
			} else {
				e.logger.Debug("plan created with no updates", "manifest", m.Path)
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

// filterManifestsByPattern filters manifests to only include those matching the given file patterns.
func (e *Engine) filterManifestsByPattern(manifests []*Manifest, matchConfig *MatchConfig, repoRoot string) []*Manifest {
	filtered := make([]*Manifest, 0, len(manifests))

	for _, m := range manifests {
		// Build full path for matching
		fullPath := filepath.Join(repoRoot, m.Path)

		// First, check if path matches any include pattern (if specified)
		includeMatched := len(matchConfig.Files) == 0 // If no files patterns, include all by default
		for _, pattern := range matchConfig.Files {
			if e.matchesPattern(pattern, fullPath, m.Path, repoRoot) {
				includeMatched = true
				break
			}
		}

		// If not included, skip
		if !includeMatched {
			e.logger.Debug("manifest filtered out (no include match)", "path", m.Path, "patterns", matchConfig.Files)
			continue
		}

		// Then, check if path matches any exclude pattern
		excluded := false
		for _, pattern := range matchConfig.Exclude {
			if e.matchesPattern(pattern, fullPath, m.Path, repoRoot) {
				excluded = true
				e.logger.Debug("manifest excluded", "path", m.Path, "pattern", pattern)
				break
			}
		}

		// Only add if included and not excluded
		if !excluded {
			filtered = append(filtered, m)
		}
	}

	return filtered
}

// matchesPattern checks if a file path matches a given glob pattern.
// It tries both absolute and relative path matching.
func (e *Engine) matchesPattern(pattern, fullPath, relativePath, repoRoot string) bool {
	// Support both absolute and relative patterns
	patternPath := pattern
	if !filepath.IsAbs(patternPath) {
		patternPath = filepath.Join(repoRoot, pattern)
	}

	// Try matching against full path
	match, err := filepath.Match(patternPath, fullPath)
	if err != nil {
		e.logger.Debug("pattern match error", "pattern", pattern, "path", fullPath, "error", err)
	} else if match {
		return true
	}

	// Also try pattern matching on the relative path directly
	match, err = filepath.Match(pattern, relativePath)
	if err == nil && match {
		return true
	}

	return false
}

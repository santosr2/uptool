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

package cmd

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	_ "github.com/santosr2/uptool/internal/datasource" // Registers all datasources
	"github.com/santosr2/uptool/internal/engine"
	"github.com/santosr2/uptool/internal/integrations"
	_ "github.com/santosr2/uptool/internal/integrations/all" // Registers all integrations
	"github.com/santosr2/uptool/internal/policy"
)

// setupEngine creates and configures an engine instance.
// It loads the uptool.yaml configuration and sets up integration policies
// for policy-aware version selection (precedence: uptool.yaml > CLI flags > constraints).
func setupEngine() *engine.Engine {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: GetLogLevel(),
	}))

	eng := engine.NewEngine(logger)

	// Load configuration if available
	var cfg *policy.Config

	// Determine config file path (custom path via --config flag or default uptool.yaml)
	configPath := GetConfigPath()
	if configPath == "" {
		// No --config flag specified, use default
		configPath = filepath.Join(".", "uptool.yaml")
	}

	if _, err := os.Stat(configPath); err == nil {
		// Convert to absolute path for secureio
		absPath, absErr := filepath.Abs(configPath)
		if absErr != nil {
			logger.Warn("failed to resolve config path", "path", configPath, "error", absErr)
		} else {
			var loadErr error
			cfg, loadErr = policy.LoadConfig(absPath)
			if loadErr != nil {
				logger.Warn("failed to load config, using defaults", "path", absPath, "error", loadErr)
				cfg = nil
			} else {
				logger.Debug("loaded configuration", "path", absPath)
			}
		}
	} else if GetConfigPath() != "" {
		// Custom config path specified but file doesn't exist - warn user
		logger.Warn("config file not found", "path", configPath)
	}

	// Get all registered integrations from the global registry
	allIntegrations := integrations.GetAll()

	// Register integrations based on config
	if cfg != nil {
		// Build map of integration configs (both enabled and disabled)
		configMap := make(map[string]policy.IntegrationConfig)
		for i := range cfg.Integrations {
			configMap[cfg.Integrations[i].ID] = cfg.Integrations[i]
		}

		for id, integration := range allIntegrations {
			if ic, exists := configMap[id]; exists {
				// Integration is in config - respect enabled flag
				if ic.Enabled {
					eng.Register(integration)
					logger.Debug("registered integration from config", "id", id)
				} else {
					logger.Debug("skipped integration (disabled in config)", "id", id)
				}
			} else {
				// Integration not in config - register with defaults
				eng.Register(integration)
				logger.Debug("registered integration (not in config, using defaults)", "id", id)
			}
		}

		// Set integration policies for policy-aware version selection
		// This implements the precedence: CLI flags > uptool.yaml > constraints
		policies := buildPolicies(cfg)
		if len(policies) > 0 {
			eng.SetPolicies(policies)
			logger.Debug("set integration policies", "count", len(policies))
		}

		// Set match configurations for file pattern filtering
		matchConfigs := cfg.ToMatchConfigMap()
		if len(matchConfigs) > 0 {
			eng.SetMatchConfigs(matchConfigs)
			logger.Debug("set match configs", "count", len(matchConfigs))
		}
	} else {
		// No config file - register all integrations
		for _, integration := range allIntegrations {
			eng.Register(integration)
		}
	}

	return eng
}

// buildPolicies extracts IntegrationPolicy objects from the config.
// It uses the ToPolicyMap method to get policies for all integrations,
// then filters to only include enabled integrations with enabled policies.
//
// When policy.enabled is false, the policy is NOT included in the map,
// causing the integration to use default settings (allow all updates, respect constraints).
func buildPolicies(cfg *policy.Config) map[string]engine.IntegrationPolicy {
	// Get all policies from config
	allPolicies := cfg.ToPolicyMap()

	// Filter to only enabled integrations with enabled policies
	policies := make(map[string]engine.IntegrationPolicy)
	for i := range cfg.Integrations {
		ic := &cfg.Integrations[i]
		if !ic.Enabled {
			continue
		}

		// Check if policy is explicitly disabled
		p := allPolicies[ic.ID]
		if !p.Enabled {
			// Policy disabled - skip adding it (integration will use defaults)
			continue
		}

		// Only add policy if it has settings
		if p.Update != "" || p.AllowPrerelease {
			policies[ic.ID] = p
		}
	}

	return policies
}

// loadPolicyConfig loads the uptool.yaml configuration file.
// Returns nil if the file doesn't exist (not an error).
func loadPolicyConfig() (*policy.Config, error) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: GetLogLevel(),
	}))

	configPath := filepath.Join(".", "uptool.yaml")
	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			return nil, nil // File doesn't exist - not an error
		}
		return nil, err
	}

	// Convert to absolute path for secureio
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, err
	}

	cfg, err := policy.LoadConfig(absPath)
	if err != nil {
		return nil, err
	}

	logger.Debug("loaded configuration", "path", absPath)
	return cfg, nil
}

// parseFilters parses comma-separated filter strings
func parseFilters(only, exclude string) ([]string, []string) {
	var onlyList, excludeList []string

	if only != "" {
		onlyList = strings.Split(only, ",")
		for i := range onlyList {
			onlyList[i] = strings.TrimSpace(onlyList[i])
		}
	}

	if exclude != "" {
		excludeList = strings.Split(exclude, ",")
		for i := range excludeList {
			excludeList[i] = strings.TrimSpace(excludeList[i])
		}
	}

	return onlyList, excludeList
}

// completeIntegrations provides shell completion for integration names
func completeIntegrations(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Get list of available integrations
	registered := integrations.List()
	return registered, cobra.ShellCompDirectiveNoFileComp
}

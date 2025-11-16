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

// setupEngine creates and configures an engine instance
func setupEngine() *engine.Engine {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: GetLogLevel(),
	}))

	eng := engine.NewEngine(logger)

	// Load configuration if available
	var cfg *policy.Config
	configPath := filepath.Join(".", "uptool.yaml")
	if _, err := os.Stat(configPath); err == nil {
		var loadErr error
		cfg, loadErr = policy.LoadConfig(configPath)
		if loadErr != nil {
			logger.Warn("failed to load config, using defaults", "error", loadErr)
			cfg = nil
		} else {
			logger.Debug("loaded configuration", "path", configPath)
		}
	}

	// Get all registered integrations from the global registry
	allIntegrations := integrations.GetAll()

	// Register integrations based on config
	if cfg != nil {
		// Use config to determine which integrations to enable
		enabledMap := make(map[string]bool)
		for _, ic := range cfg.Integrations {
			if ic.Enabled {
				enabledMap[ic.ID] = true
			}
		}

		for id, integration := range allIntegrations {
			if enabledMap[id] {
				eng.Register(integration)
				logger.Debug("registered integration from config", "id", id)
			} else {
				logger.Debug("skipped integration (disabled in config)", "id", id)
			}
		}
	} else {
		// No config file - register all integrations
		for _, integration := range allIntegrations {
			eng.Register(integration)
		}
	}

	return eng
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

package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/santosr2/uptool/internal/integrations"
)

var (
	listCategory     string
	listExperimental bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available integrations",
	Long: `List all available integrations and their status.

Integrations can be filtered by category or include experimental ones.`,
	Example: `  # List all integrations
  uptool list

  # List only package managers
  uptool list --category package-manager

  # Include experimental integrations
  uptool list --experimental`,
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().StringVarP(&listCategory, "category", "c", "", "filter by category")
	listCmd.Flags().BoolVar(&listExperimental, "experimental", false, "include experimental integrations")
}

func runList(cmd *cobra.Command, args []string) error {
	// Load metadata
	meta, err := integrations.LoadMetadata()
	if err != nil {
		return fmt.Errorf("load metadata: %w", err)
	}

	// Get all registered integrations
	registered := integrations.List()
	registeredMap := make(map[string]bool)
	for _, name := range registered {
		registeredMap[name] = true
	}

	// Filter integrations
	displayIntegrations := make([]string, 0, len(registered))
	for id, info := range meta.Integrations {
		// Skip if not registered
		if !registeredMap[id] {
			continue
		}

		// Skip experimental unless flag is set
		if info.Experimental && !listExperimental {
			continue
		}

		// Skip if category doesn't match
		if listCategory != "" && info.Category != listCategory {
			continue
		}

		displayIntegrations = append(displayIntegrations, id)
	}

	if len(displayIntegrations) == 0 {
		fmt.Println("No integrations found matching criteria.")
		return nil
	}

	// Sort for consistent output
	sort := func(a, b string) bool { return a < b }
	for i := 0; i < len(displayIntegrations); i++ {
		for j := i + 1; j < len(displayIntegrations); j++ {
			if !sort(displayIntegrations[i], displayIntegrations[j]) {
				displayIntegrations[i], displayIntegrations[j] = displayIntegrations[j], displayIntegrations[i]
			}
		}
	}

	// Display integrations
	fmt.Printf("%-15s %-20s %-50s %s\n", "ID", "Name", "Description", "Status")
	fmt.Println(strings.Repeat("-", 100))

	for _, id := range displayIntegrations {
		info := meta.Integrations[id]
		status := ""
		if info.Experimental {
			status = "[EXPERIMENTAL]"
		}
		if info.Disabled {
			status = "[DISABLED]"
		}

		name := info.DisplayName
		if len(name) > 20 {
			name = name[:17] + "..."
		}

		desc := info.Description
		if len(desc) > 50 {
			desc = desc[:47] + "..."
		}

		fmt.Printf("%-15s %-20s %-50s %s\n", id, name, desc, status)
	}

	fmt.Printf("\nTotal: %d integrations\n", len(displayIntegrations))

	// Show categories if no filter
	if listCategory == "" {
		categories := make(map[string]int)
		for _, id := range displayIntegrations {
			info := meta.Integrations[id]
			categories[info.Category]++
		}

		if len(categories) > 0 {
			fmt.Println("\nCategories:")
			for cat, count := range categories {
				catInfo := meta.Categories[cat]
				fmt.Printf("  %-20s %d integrations - %s\n", cat, count, catInfo.Description)
			}
		}
	}

	return nil
}

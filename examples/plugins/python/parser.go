package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/santosr2/uptool/internal/engine"
)

// requirementRegex matches pip requirement specifications
// Examples: requests==2.28.0, flask>=2.2.0, pytest[dev]~=7.0
var requirementRegex = regexp.MustCompile(`^([a-zA-Z0-9_-]+)(?:\[[a-zA-Z0-9,_-]+\])?([<>=!~]+)([0-9.]+(?:[a-z0-9.]+)?)`)

// ParseRequirements parses a requirements.txt file and extracts dependencies.
// It handles:
// - Simple version pins (package==1.0.0)
// - Version constraints (package>=1.0.0)
// - Comments (# comment)
// - Blank lines
// - Extras (package[extra]==1.0.0)
func ParseRequirements(content string) ([]*engine.Dependency, error) {
	var deps []*engine.Dependency
	seen := make(map[string]bool)

	lines := strings.Split(content, "\n")
	for lineNum, line := range lines {
		// Trim whitespace
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Skip comments
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Remove inline comments
		if idx := strings.Index(line, "#"); idx > 0 {
			line = strings.TrimSpace(line[:idx])
		}

		// Skip pip install flags (--index-url, -e, etc.)
		if strings.HasPrefix(line, "-") {
			continue
		}

		// Parse requirement
		dep, err := parseRequirement(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum+1, err)
		}

		// Skip duplicates
		if seen[dep.Name] {
			continue
		}
		seen[dep.Name] = true

		deps = append(deps, dep)
	}

	return deps, nil
}

// parseRequirement parses a single requirement line.
func parseRequirement(line string) (*engine.Dependency, error) {
	matches := requirementRegex.FindStringSubmatch(line)
	if matches == nil {
		return nil, fmt.Errorf("invalid requirement format: %s", line)
	}

	name := matches[1]
	operator := matches[2]
	version := matches[3]

	// For update purposes, we only care about == constraints
	// Other constraints (>=, ~=, etc.) will be preserved during updates
	dep := &engine.Dependency{
		Name:           name,
		CurrentVersion: version,
		Constraint:     operator, // Store original constraint
	}

	return dep, nil
}

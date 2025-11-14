// Package rewrite provides utilities for rewriting structured files while preserving formatting.
// It includes functions for YAML manipulation, unified diff generation, and patch creation,
// enabling integrations to update configuration files without destroying formatting or comments.
package rewrite

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

// ReplaceYAMLValue replaces a specific value in a YAML document while preserving formatting.
// path specifies the location (e.g., ["repos", "*", "rev"]) where * matches any element.
// matcher is an optional function to further filter which nodes to update.
func ReplaceYAMLValue(content string, path []string, oldValue, newValue string, matcher func(*yaml.Node) bool) (string, error) {
	var root yaml.Node
	if err := yaml.Unmarshal([]byte(content), &root); err != nil {
		return "", fmt.Errorf("parse YAML: %w", err)
	}

	// Traverse and update
	updated := false
	if err := traverseAndReplace(&root, path, 0, oldValue, newValue, matcher, &updated); err != nil {
		return "", err
	}

	if !updated {
		return "", fmt.Errorf("value %q not found at path %v", oldValue, path)
	}

	// Encode back to YAML
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(&root); err != nil {
		return "", fmt.Errorf("encode YAML: %w", err)
	}

	return buf.String(), nil
}

// traverseAndReplace recursively traverses the YAML tree and replaces values.
func traverseAndReplace(node *yaml.Node, path []string, depth int, oldValue, newValue string, matcher func(*yaml.Node) bool, updated *bool) error {
	if node == nil || depth >= len(path) {
		return nil
	}

	currentKey := path[depth]

	switch node.Kind {
	case yaml.DocumentNode:
		for _, child := range node.Content {
			if err := traverseAndReplace(child, path, depth, oldValue, newValue, matcher, updated); err != nil {
				return err
			}
		}

	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			if currentKey == "*" || keyNode.Value == currentKey {
				if depth == len(path)-1 {
					// We've reached the target depth
					if valueNode.Value == oldValue && (matcher == nil || matcher(node)) {
						valueNode.Value = newValue
						*updated = true
					}
				} else {
					// Continue traversing
					if err := traverseAndReplace(valueNode, path, depth+1, oldValue, newValue, matcher, updated); err != nil {
						return err
					}
				}
			}
		}

	case yaml.SequenceNode:
		if currentKey == "*" || currentKey == "[]" {
			// Match all elements in the sequence
			for _, child := range node.Content {
				if err := traverseAndReplace(child, path, depth+1, oldValue, newValue, matcher, updated); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// UpdateYAMLField updates a specific field in a YAML document.
func UpdateYAMLField(content string, path []string, newValue string) (string, error) {
	var root yaml.Node
	if err := yaml.Unmarshal([]byte(content), &root); err != nil {
		return "", fmt.Errorf("parse YAML: %w", err)
	}

	if err := setYAMLField(&root, path, 0, newValue); err != nil {
		return "", err
	}

	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(&root); err != nil {
		return "", fmt.Errorf("encode YAML: %w", err)
	}

	return buf.String(), nil
}

// setYAMLField sets a field value at the specified path.
func setYAMLField(node *yaml.Node, path []string, depth int, value string) error {
	if node == nil || depth >= len(path) {
		return fmt.Errorf("path not found")
	}

	currentKey := path[depth]

	switch node.Kind {
	case yaml.DocumentNode:
		for _, child := range node.Content {
			if err := setYAMLField(child, path, depth, value); err == nil {
				return nil
			}
		}

	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			if keyNode.Value == currentKey {
				if depth == len(path)-1 {
					valueNode.Value = value
					return nil
				}
				return setYAMLField(valueNode, path, depth+1, value)
			}
		}
	}

	return fmt.Errorf("key %q not found", currentKey)
}

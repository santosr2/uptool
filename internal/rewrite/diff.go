package rewrite

import (
	"fmt"
	"strings"
	"time"

	"github.com/pmezard/go-difflib/difflib"
)

// GenerateUnifiedDiff creates a unified diff between old and new content.
func GenerateUnifiedDiff(filename, oldContent, newContent string) (string, error) {
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(oldContent),
		B:        difflib.SplitLines(newContent),
		FromFile: filename,
		ToFile:   filename,
		Context:  3,
		Eol:      "\n",
	}

	text, err := difflib.GetUnifiedDiffString(diff)
	if err != nil {
		return "", fmt.Errorf("generate diff: %w", err)
	}

	return text, nil
}

// GeneratePatch creates a git-style patch with timestamps.
func GeneratePatch(filename, oldContent, newContent string) (string, error) {
	now := time.Now().Format(time.RFC3339)

	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(oldContent),
		B:        difflib.SplitLines(newContent),
		FromFile: fmt.Sprintf("a/%s", filename),
		ToFile:   fmt.Sprintf("b/%s", filename),
		FromDate: now,
		ToDate:   now,
		Context:  3,
		Eol:      "\n",
	}

	text, err := difflib.GetUnifiedDiffString(diff)
	if err != nil {
		return "", fmt.Errorf("generate patch: %w", err)
	}

	return text, nil
}

// CountChanges returns the number of additions and deletions in a diff.
func CountChanges(diff string) (additions, deletions int) {
	lines := strings.Split(diff, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			additions++
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			deletions++
		}
	}
	return additions, deletions
}

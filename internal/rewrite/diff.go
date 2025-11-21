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

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

package builtin

import (
	"context"
	"encoding/json"
	"os/exec"

	"github.com/santosr2/uptool/internal/policy/guards"
)

// CIGreenGuard checks if all CI checks are passing.
type CIGreenGuard struct{}

func init() {
	guards.Register(&CIGreenGuard{})
}

// Name returns the guard's unique identifier.
func (g *CIGreenGuard) Name() string {
	return "ci-green"
}

// Description returns a human-readable description of the guard.
func (g *CIGreenGuard) Description() string {
	return "Verifies that all CI checks are passing (SUCCESS or SKIPPED)"
}

// Check verifies that all CI checks are passing.
func (g *CIGreenGuard) Check(ctx context.Context, env *guards.Environment) (bool, error) {
	// #nosec G204 -- env.GitHubPRNumber validated to be numeric only before calling guards
	cmd := exec.CommandContext(ctx, "gh", "pr", "checks", env.GitHubPRNumber, "--json", "state")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	// Parse check status
	var checks []struct {
		State string `json:"state"`
	}

	if err := json.Unmarshal(output, &checks); err != nil {
		return false, err
	}

	// All checks must be SUCCESS or SKIPPED
	for _, check := range checks {
		if check.State != "SUCCESS" && check.State != "SKIPPED" {
			return false, nil
		}
	}

	// Must have at least one check
	return len(checks) > 0, nil
}

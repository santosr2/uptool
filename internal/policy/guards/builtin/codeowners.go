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

// CodeownersApproveGuard checks if CODEOWNERS have approved the PR.
type CodeownersApproveGuard struct{}

func init() {
	guards.Register(&CodeownersApproveGuard{})
}

// Name returns the guard's unique identifier.
func (g *CodeownersApproveGuard) Name() string {
	return "codeowners-approve"
}

// Description returns a human-readable description of the guard.
func (g *CodeownersApproveGuard) Description() string {
	return "Verifies that repository codeowners have approved the PR"
}

// Check verifies that CODEOWNERS have approved the PR.
func (g *CodeownersApproveGuard) Check(ctx context.Context, env *guards.Environment) (bool, error) {
	// #nosec G204 -- env.GitHubPRNumber validated to be numeric only before calling guards
	cmd := exec.CommandContext(ctx, "gh", "pr", "view", env.GitHubPRNumber, "--json", "reviews")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	var prData struct {
		Reviews []struct {
			Author struct {
				Login string `json:"login"`
			} `json:"author"`
			State         string `json:"state"`
			AuthorIsOwner bool   `json:"authorAssociation"` // OWNER, MEMBER, etc.
		} `json:"reviews"`
	}

	if err := json.Unmarshal(output, &prData); err != nil {
		return false, err
	}

	// Check if any owner/member has approved
	for _, review := range prData.Reviews {
		if review.State == "APPROVED" && review.AuthorIsOwner {
			return true, nil
		}
	}

	return false, nil
}

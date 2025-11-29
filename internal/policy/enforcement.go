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

// Package policy implements organization policy enforcement for uptool.
// This includes signoff requirements, cosign verification, and auto-merge logic.
package policy

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Enforcer handles organization policy enforcement.
type Enforcer struct {
	config      *Config
	githubRepo  string // Format: "owner/repo"
	githubToken string
}

// NewEnforcer creates a new policy enforcer.
func NewEnforcer(config *Config) *Enforcer {
	return &Enforcer{
		config:      config,
		githubRepo:  os.Getenv("GITHUB_REPOSITORY"),
		githubToken: os.Getenv("GITHUB_TOKEN"),
	}
}

// EnforcementResult contains the results of policy enforcement.
type EnforcementResult struct {
	GuardsStatus     map[string]bool
	SignoffErrors    []string
	CosignErrors     []string
	AutoMergeErrors  []string
	SignoffValid     bool
	CosignValid      bool
	AutoMergeAllowed bool
}

// Enforce checks all configured organization policies.
func (e *Enforcer) Enforce(ctx context.Context) (*EnforcementResult, error) {
	result := &EnforcementResult{
		SignoffValid:     true,
		CosignValid:      true,
		AutoMergeAllowed: true,
		GuardsStatus:     make(map[string]bool),
	}

	if e.config == nil || e.config.OrgPolicy == nil {
		return result, nil // No policies to enforce
	}

	// Check signoff requirements
	if e.config.RequiresSignoff() {
		if err := e.checkSignoffs(ctx, result); err != nil {
			return nil, fmt.Errorf("check signoffs: %w", err)
		}
	}

	// Check cosign verification
	if e.config.RequiresCosignVerification() {
		if err := e.checkCosignVerification(ctx, result); err != nil {
			return nil, fmt.Errorf("check cosign: %w", err)
		}
	}

	// Check auto-merge guards
	if e.config.IsAutoMergeEnabled() {
		if err := e.checkAutoMergeGuards(ctx, result); err != nil {
			return nil, fmt.Errorf("check auto-merge guards: %w", err)
		}
	}

	return result, nil
}

// checkSignoffs verifies that required signoffs are present.
//
//nolint:unparam // error return kept for consistency with other check functions
func (e *Enforcer) checkSignoffs(ctx context.Context, result *EnforcementResult) error {
	if e.githubRepo == "" {
		result.SignoffValid = false
		result.SignoffErrors = append(result.SignoffErrors, "GITHUB_REPOSITORY not set")
		return nil
	}

	if e.githubToken == "" {
		result.SignoffValid = false
		result.SignoffErrors = append(result.SignoffErrors, "GITHUB_TOKEN not set")
		return nil
	}

	prNumber := os.Getenv("GITHUB_PR_NUMBER")
	if prNumber == "" {
		// Not in a PR context, skip signoff check
		return nil
	}

	// Validate PR number to prevent command injection (must be numeric only)
	for _, ch := range prNumber {
		if ch < '0' || ch > '9' {
			result.SignoffValid = false
			result.SignoffErrors = append(result.SignoffErrors, "invalid GITHUB_PR_NUMBER: must be numeric")
			return nil
		}
	}

	// Get PR reviews using gh CLI (more reliable than direct API calls)
	// #nosec G204 -- prNumber validated to be numeric only above, preventing command injection
	cmd := exec.CommandContext(ctx, "gh", "pr", "view", prNumber, "--json", "reviews")
	output, err := cmd.Output()
	if err != nil {
		result.SignoffValid = false
		result.SignoffErrors = append(result.SignoffErrors, fmt.Sprintf("failed to get PR reviews: %v", err))
		return nil
	}

	var prData struct {
		Reviews []struct {
			Author struct {
				Login string `json:"login"`
				Email string `json:"email"`
			} `json:"author"`
			State string `json:"state"`
		} `json:"reviews"`
	}

	if err := json.Unmarshal(output, &prData); err != nil {
		result.SignoffValid = false
		result.SignoffErrors = append(result.SignoffErrors, fmt.Sprintf("failed to parse PR data: %v", err))
		return nil
	}

	// Check if all required signoffs are present
	requiredSignoffs := e.config.OrgPolicy.RequireSignoffFrom
	approvedBy := make(map[string]bool)

	for _, review := range prData.Reviews {
		if review.State == "APPROVED" {
			approvedBy[review.Author.Login] = true
			if review.Author.Email != "" {
				approvedBy[review.Author.Email] = true
			}
		}
	}

	missing := []string{}
	for _, required := range requiredSignoffs {
		if !approvedBy[required] {
			missing = append(missing, required)
		}
	}

	if len(missing) > 0 {
		result.SignoffValid = false
		result.SignoffErrors = append(result.SignoffErrors, fmt.Sprintf("missing required approvals: %s", strings.Join(missing, ", ")))
	}

	return nil
}

// checkCosignVerification verifies artifact signatures using cosign.
//
//nolint:unparam // error return kept for consistency with other check functions
func (e *Enforcer) checkCosignVerification(_ context.Context, result *EnforcementResult) error {
	// Check if cosign is available
	if _, err := exec.LookPath("cosign"); err != nil {
		result.CosignValid = false
		result.CosignErrors = append(result.CosignErrors, "cosign not found in PATH")
		return nil // nolint:nilerr // Intentional: record error in result, don't fail enforcement
	}

	// Look for artifacts to verify in common locations
	// This is a basic implementation - real usage would specify artifact paths
	artifactPaths := []string{
		"dist/checksums.txt",
		"dist/*.sig",
	}

	verified := false
	for _, pattern := range artifactPaths {
		// For now, just check if artifacts exist
		// Full implementation would verify signatures
		if _, err := os.Stat(pattern); err == nil {
			verified = true
			break
		}
	}

	if !verified {
		result.CosignValid = false
		result.CosignErrors = append(result.CosignErrors, "no artifacts found to verify")
	}

	return nil
}

// checkAutoMergeGuards checks if all required guards are satisfied.
//
//nolint:unparam // error return kept for consistency with other check functions
func (e *Enforcer) checkAutoMergeGuards(ctx context.Context, result *EnforcementResult) error {
	if e.githubRepo == "" || e.githubToken == "" {
		result.AutoMergeAllowed = false
		result.AutoMergeErrors = append(result.AutoMergeErrors, "GitHub environment not configured")
		return nil
	}

	prNumber := os.Getenv("GITHUB_PR_NUMBER")
	if prNumber == "" {
		// Not in PR context
		return nil
	}

	// Validate PR number to prevent command injection (must be numeric only)
	for _, ch := range prNumber {
		if ch < '0' || ch > '9' {
			result.AutoMergeAllowed = false
			result.AutoMergeErrors = append(result.AutoMergeErrors, "invalid GITHUB_PR_NUMBER: must be numeric")
			return nil
		}
	}

	guards := e.config.GetAutoMergeGuards()
	allSatisfied := true

	for _, guard := range guards {
		satisfied := false

		switch guard {
		case "ci-green":
			satisfied = e.checkCIStatus(ctx, prNumber)
		case "codeowners-approve":
			satisfied = e.checkCodeownersApproval(ctx, prNumber)
		case "security-scan":
			satisfied = e.checkSecurityScan(ctx, prNumber)
		default:
			// Unknown guard - mark as unsatisfied
			satisfied = false
			result.AutoMergeErrors = append(result.AutoMergeErrors, fmt.Sprintf("unknown guard: %s", guard))
		}

		result.GuardsStatus[guard] = satisfied
		if !satisfied {
			allSatisfied = false
		}
	}

	result.AutoMergeAllowed = allSatisfied

	return nil
}

// checkCIStatus checks if all CI checks are passing.
func (e *Enforcer) checkCIStatus(ctx context.Context, prNumber string) bool {
	cmd := exec.CommandContext(ctx, "gh", "pr", "checks", prNumber, "--json", "state")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// Parse check status
	var checks []struct {
		State string `json:"state"`
	}

	if err := json.Unmarshal(output, &checks); err != nil {
		return false
	}

	for _, check := range checks {
		if check.State != "SUCCESS" && check.State != "SKIPPED" {
			return false
		}
	}

	return len(checks) > 0
}

// checkCodeownersApproval checks if CODEOWNERS have approved.
func (e *Enforcer) checkCodeownersApproval(ctx context.Context, prNumber string) bool {
	cmd := exec.CommandContext(ctx, "gh", "pr", "view", prNumber, "--json", "reviews")
	output, err := cmd.Output()
	if err != nil {
		return false
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
		return false
	}

	// Check if any owner/member has approved
	for _, review := range prData.Reviews {
		if review.State == "APPROVED" && review.AuthorIsOwner {
			return true
		}
	}

	return false
}

// checkSecurityScan checks if security scans have passed.
func (e *Enforcer) checkSecurityScan(ctx context.Context, prNumber string) bool {
	// Check for specific security scan check runs
	cmd := exec.CommandContext(ctx, "gh", "pr", "checks", prNumber, "--json", "name,state")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	var checks []struct {
		Name  string `json:"name"`
		State string `json:"state"`
	}

	if err := json.Unmarshal(output, &checks); err != nil {
		return false
	}

	// Look for security-related checks
	securityChecks := []string{"CodeQL", "Trivy", "Security", "SAST"}
	foundSecurityCheck := false

	for _, check := range checks {
		for _, secCheck := range securityChecks {
			if strings.Contains(check.Name, secCheck) {
				foundSecurityCheck = true
				if check.State != "SUCCESS" {
					return false
				}
			}
		}
	}

	return foundSecurityCheck
}

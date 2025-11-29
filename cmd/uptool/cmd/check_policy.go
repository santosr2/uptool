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
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/santosr2/uptool/internal/policy"
)

var checkPolicyCmd = &cobra.Command{
	Use:   "check-policy",
	Short: "Check organization policy compliance",
	Long: `Check if the current changes comply with organization policies.

This command validates:
- Required signoffs from specified teams/users
- Artifact signature verification (cosign)
- Auto-merge guard requirements (CI status, codeowners, security scans)

Organization policies are configured in uptool.yaml under the org_policy section.

Example:
  uptool check-policy

  # In GitHub Actions
  - name: Check org policy
    run: uptool check-policy
    env:
      GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      GITHUB_PR_NUMBER: ${{ github.event.pull_request.number }}`,
	RunE: runCheckPolicy,
}

func init() {
	rootCmd.AddCommand(checkPolicyCmd)
}

func runCheckPolicy(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Load configuration
	cfg, err := loadPolicyConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if cfg == nil {
		fmt.Println("ℹ️  No uptool.yaml found - skipping policy checks")
		return nil
	}

	if cfg.OrgPolicy == nil {
		fmt.Println("ℹ️  No org_policy configured - skipping policy checks")
		return nil
	}

	fmt.Println("🔍 Checking organization policy compliance...")
	fmt.Println()

	// Create enforcer and run checks
	enforcer := policy.NewEnforcer(cfg)
	result, err := enforcer.Enforce(ctx)
	if err != nil {
		return fmt.Errorf("enforce policy: %w", err)
	}

	// Display results
	allPassed := true

	// Signoff results
	if cfg.RequiresSignoff() {
		if result.SignoffValid {
			fmt.Println("✅ Signoff requirements: PASSED")
		} else {
			fmt.Println("❌ Signoff requirements: FAILED")
			for _, err := range result.SignoffErrors {
				fmt.Printf("   - %s\n", err)
			}
			allPassed = false
		}
		fmt.Println()
	}

	// Cosign results
	if cfg.RequiresCosignVerification() {
		if result.CosignValid {
			fmt.Println("✅ Cosign verification: PASSED")
		} else {
			fmt.Println("❌ Cosign verification: FAILED")
			for _, err := range result.CosignErrors {
				fmt.Printf("   - %s\n", err)
			}
			allPassed = false
		}
		fmt.Println()
	}

	// Auto-merge results
	if cfg.IsAutoMergeEnabled() {
		if result.AutoMergeAllowed {
			fmt.Println("✅ Auto-merge guards: PASSED")
		} else {
			fmt.Println("❌ Auto-merge guards: FAILED")
			for _, err := range result.AutoMergeErrors {
				fmt.Printf("   - %s\n", err)
			}
		}

		if len(result.GuardsStatus) > 0 {
			fmt.Println("\n   Guard status:")
			for guard, status := range result.GuardsStatus {
				statusIcon := "✅"
				if !status {
					statusIcon = "❌"
				}
				fmt.Printf("   %s %s\n", statusIcon, guard)
			}
			allPassed = allPassed && result.AutoMergeAllowed
		}
		fmt.Println()
	}

	// Summary
	if allPassed {
		fmt.Println("🎉 All organization policy checks passed!")
		return nil
	}

	fmt.Println("⚠️  Some organization policy checks failed")
	os.Exit(1)
	return nil
}

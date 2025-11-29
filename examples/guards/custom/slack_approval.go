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

// Package custom provides example custom guard implementations.
// This demonstrates how to create custom guards that extend uptool's auto-merge capabilities.
package custom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/santosr2/uptool/internal/policy/guards"
)

// SlackApprovalGuard checks if the PR has been approved via Slack reaction.
// This is an example custom guard that demonstrates:
// - Using external APIs (Slack)
// - Reading custom environment variables
// - Implementing custom approval workflows
type SlackApprovalGuard struct{}

func init() {
	// Automatically register the guard when this package is imported
	guards.Register(&SlackApprovalGuard{})
}

// Name returns the guard's unique identifier.
func (g *SlackApprovalGuard) Name() string {
	return "slack-approval"
}

// Description returns a human-readable description of the guard.
func (g *SlackApprovalGuard) Description() string {
	return "Checks if the PR has received approval via Slack thumbs-up reaction"
}

// Check verifies that the PR has Slack approval via thumbs-up reaction.
func (g *SlackApprovalGuard) Check(ctx context.Context, env *guards.Environment) (bool, error) {
	// Get Slack configuration from environment
	slackToken := os.Getenv("SLACK_BOT_TOKEN")
	if slackToken == "" {
		return false, fmt.Errorf("SLACK_BOT_TOKEN not set")
	}

	slackChannel := os.Getenv("SLACK_CHANNEL_ID")
	if slackChannel == "" {
		return false, fmt.Errorf("SLACK_CHANNEL_ID not set")
	}

	// Search for Slack messages about this PR
	prURL := fmt.Sprintf("https://github.com/%s/pull/%s", env.GitHubRepo, env.GitHubPRNumber)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Search for messages containing the PR URL
	searchURL := fmt.Sprintf("https://slack.com/api/search.messages?query=%s&count=1", prURL)
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, http.NoBody)
	if err != nil {
		return false, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+slackToken)

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("slack API request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close() //nolint:errcheck // Best effort close
	}()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("slack API error: status %d", resp.StatusCode)
	}

	var searchResult struct {
		Messages struct {
			Matches []struct {
				TS string `json:"ts"`
			} `json:"matches"`
		} `json:"messages"`
		OK bool `json:"ok"` // Message timestamp
	}

	if decodeErr := json.NewDecoder(resp.Body).Decode(&searchResult); decodeErr != nil {
		return false, fmt.Errorf("decode slack response: %w", decodeErr)
	}

	if !searchResult.OK || len(searchResult.Messages.Matches) == 0 {
		// No Slack message found for this PR
		return false, nil
	}

	// Check if the message has thumbs-up reactions
	messageTS := searchResult.Messages.Matches[0].TS
	reactionsURL := fmt.Sprintf("https://slack.com/api/reactions.get?channel=%s&timestamp=%s",
		slackChannel, messageTS)

	req, err = http.NewRequestWithContext(ctx, "GET", reactionsURL, http.NoBody)
	if err != nil {
		return false, fmt.Errorf("create reactions request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+slackToken)

	resp, err = client.Do(req)
	if err != nil {
		return false, fmt.Errorf("slack reactions API request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close() //nolint:errcheck // Best effort close
	}()

	var reactionsResult struct {
		Message struct {
			Reactions []struct {
				Name  string `json:"name"`
				Count int    `json:"count"`
			} `json:"reactions"`
		} `json:"message"`
		OK bool `json:"ok"`
	}

	if decodeErr := json.NewDecoder(resp.Body).Decode(&reactionsResult); decodeErr != nil {
		return false, fmt.Errorf("decode reactions response: %w", decodeErr)
	}

	if !reactionsResult.OK {
		return false, nil
	}

	// Check for thumbs-up (":+1:") reactions
	for _, reaction := range reactionsResult.Message.Reactions {
		if reaction.Name == "+1" && reaction.Count > 0 {
			return true, nil
		}
	}

	return false, nil
}

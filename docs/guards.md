# Guard Plugins

Guard plugins are extensible checks that determine whether a Pull Request meets your organization's requirements for auto-merge. uptool provides built-in guards and supports custom guard plugins for organization-specific workflows.

## Overview

Guards enable you to:

- **Enforce quality gates** before auto-merging dependency updates
- **Integrate with external systems** (Slack, JIRA, custom APIs)
- **Implement custom approval workflows** beyond GitHub's built-in features
- **Create reusable policy checks** across multiple repositories

## Architecture

### Guard Interface

All guards implement a simple interface:

```go
// Guard represents a pluggable auto-merge guard check.
type Guard interface {
    // Name returns the unique identifier for this guard (e.g., "ci-green")
    Name() string

    // Description returns a human-readable description
    Description() string

    // Check executes the guard logic and returns true if satisfied
    Check(ctx context.Context, env *Environment) (bool, error)
}
```

### Environment Context

Guards receive GitHub context via the `Environment` struct:

```go
type Environment struct {
    GitHubRepo     string // Format: "owner/repo"
    GitHubToken    string // GitHub API token
    GitHubPRNumber string // PR number (validated numeric)
}
```

### Registry

Guards self-register via a global registry:

```go
import "github.com/santosr2/uptool/internal/policy/guards"

func init() {
    guards.Register(&MyCustomGuard{})
}
```

## Built-in Guards

uptool includes three production-ready guards:

### ci-green

**File**: `internal/policy/guards/builtin/ci_green.go`

**Purpose**: Verify all CI checks pass (SUCCESS or SKIPPED).

**Implementation**:

```go
type CIGreenGuard struct{}

func (g *CIGreenGuard) Name() string {
    return "ci-green"
}

func (g *CIGreenGuard) Check(ctx context.Context, env *guards.Environment) (bool, error) {
    // Uses `gh pr checks` to verify CI status
    cmd := exec.CommandContext(ctx, "gh", "pr", "checks", env.GitHubPRNumber, "--json", "state")
    // Parses JSON and checks all states are SUCCESS or SKIPPED
}
```

### codeowners-approve

**File**: `internal/policy/guards/builtin/codeowners.go`

**Purpose**: Require approval from repository CODEOWNERS.

**Implementation**:

```go
type CodeownersApproveGuard struct{}

func (g *CodeownersApproveGuard) Check(ctx context.Context, env *guards.Environment) (bool, error) {
    // Uses `gh api` to fetch PR reviews
    // Checks if any reviewer is a CODEOWNER
}
```

### security-scan

**File**: `internal/policy/guards/builtin/security_scan.go`

**Purpose**: Verify security scans (CodeQL, Trivy, SAST) pass.

**Implementation**:

```go
type SecurityScanGuard struct{}

func (g *SecurityScanGuard) Check(ctx context.Context, env *guards.Environment) (bool, error) {
    // Uses `gh api` to fetch workflow runs
    // Filters for security-related workflows
    // Verifies all completed successfully
}
```

## Creating Custom Guards

### Step 1: Implement the Guard Interface

Create a new Go file implementing the `Guard` interface:

```go
// examples/guards/custom/slack_approval.go
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

// SlackApprovalGuard checks if a PR has been approved via Slack thumbs-up reaction.
type SlackApprovalGuard struct{}

// Name returns the guard's unique identifier.
func (g *SlackApprovalGuard) Name() string {
    return "slack-approval"
}

// Description returns a human-readable description of the guard.
func (g *SlackApprovalGuard) Description() string {
    return "Verifies that the PR has been approved via Slack thumbs-up reaction"
}

// Check verifies that the PR announcement in Slack has a thumbs-up reaction.
func (g *SlackApprovalGuard) Check(ctx context.Context, env *guards.Environment) (bool, error) {
    // 1. Get configuration from environment
    slackToken := os.Getenv("SLACK_BOT_TOKEN")
    if slackToken == "" {
        return false, fmt.Errorf("SLACK_BOT_TOKEN not set")
    }

    channelID := os.Getenv("SLACK_CHANNEL_ID")
    if channelID == "" {
        return false, fmt.Errorf("SLACK_CHANNEL_ID not set")
    }

    // 2. Search for message about this PR
    prURL := fmt.Sprintf("https://github.com/%s/pull/%s", env.GitHubRepo, env.GitHubPRNumber)

    client := &http.Client{Timeout: 10 * time.Second}

    // Search messages in channel
    searchURL := fmt.Sprintf("https://slack.com/api/search.messages?query=%s&count=1", prURL)
    req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
    if err != nil {
        return false, fmt.Errorf("creating request: %w", err)
    }

    req.Header.Set("Authorization", "Bearer "+slackToken)
    resp, err := client.Do(req)
    if err != nil {
        return false, fmt.Errorf("searching Slack messages: %w", err)
    }
    defer func() {
        _ = resp.Body.Close() //nolint:errcheck // Best effort close
    }()

    var searchResult struct {
        Messages struct {
            Matches []struct {
                TS      string `json:"ts"`
                Channel struct {
                    ID string `json:"id"`
                } `json:"channel"`
            } `json:"matches"`
        } `json:"messages"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
        return false, fmt.Errorf("decoding search result: %w", err)
    }

    if len(searchResult.Messages.Matches) == 0 {
        return false, fmt.Errorf("no Slack message found for PR %s", env.GitHubPRNumber)
    }

    messageTS := searchResult.Messages.Matches[0].TS

    // 3. Get reactions on the message
    reactionsURL := fmt.Sprintf(
        "https://slack.com/api/reactions.get?channel=%s&timestamp=%s",
        channelID,
        messageTS,
    )

    req, err = http.NewRequestWithContext(ctx, "GET", reactionsURL, nil)
    if err != nil {
        return false, fmt.Errorf("creating reactions request: %w", err)
    }

    req.Header.Set("Authorization", "Bearer "+slackToken)
    resp, err = client.Do(req)
    if err != nil {
        return false, fmt.Errorf("fetching reactions: %w", err)
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
    }

    if err := json.NewDecoder(resp.Body).Decode(&reactionsResult); err != nil {
        return false, fmt.Errorf("decoding reactions: %w", err)
    }

    // 4. Check for thumbs-up reaction
    for _, reaction := range reactionsResult.Message.Reactions {
        if reaction.Name == "+1" && reaction.Count > 0 {
            return true, nil
        }
    }

    return false, nil
}
```

### Step 2: Register the Guard

Register via `init()` function:

```go
func init() {
    guards.Register(&SlackApprovalGuard{})
}
```

**Important**: This `init()` function runs automatically when the package is imported.

### Step 3: Import the Guard Package

There are two approaches to using your custom guard:

#### Option A: Modify uptool Source (Recommended)

Import your guard package in `internal/policy/enforcement.go`:

```go
import (
    _ "github.com/santosr2/uptool/examples/guards/custom" // Custom guards
    _ "github.com/santosr2/uptool/internal/policy/guards/builtin"
)
```

Then rebuild uptool:

```bash
go build -o uptool ./cmd/uptool
```

#### Option B: Go Plugin (Future)

In a future release, uptool will support loading guards as Go plugins (`.so` files):

```bash
# Build your guard as a plugin
go build -buildmode=plugin -o my_guard.so custom/my_guard.go

# Load it via environment variable
export UPTOOL_GUARD_PLUGINS="./my_guard.so"
uptool check-policy
```

**Note**: Plugin support is planned but not yet implemented.

### Step 4: Configure the Guard

Add your guard to `uptool.yaml`:

```yaml
org_policy:
  auto_merge:
    enabled: true
    guards:
      - "ci-green"         # Built-in
      - "slack-approval"   # Your custom guard
```

### Step 5: Set Environment Variables

Custom guards often need configuration via environment variables:

```bash
export SLACK_BOT_TOKEN="xoxb-your-token"
export SLACK_CHANNEL_ID="C01234567"
export GITHUB_TOKEN="ghp_your_token"
export GITHUB_REPOSITORY="owner/repo"
export GITHUB_PR_NUMBER="123"

uptool check-policy
```

## Code Generation

uptool uses code generation to automatically maintain the guard registry.

### How It Works

1. **Generator script**: `scripts/gen_guards.go` scans `internal/policy/guards/builtin/` for guard files
2. **Generated file**: `internal/policy/guards/builtin/all.go` lists all discovered guards
3. **Build integration**: Running `go generate ./internal/policy/guards` regenerates the file

### Generated File Structure

```go
// Code generated by scripts/gen_guards.go. DO NOT EDIT.

// Package builtin registers all built-in auto-merge guards.
// The guards are registered via init() functions in their individual files:
//   - ci_green.go
//   - codeowners.go
//   - security_scan.go
package builtin

// This file intentionally blank - guard registration happens via init() functions
// in individual guard files.
```

### Triggering Generation

Three ways to regenerate:

```bash
# 1. Via go generate
go generate ./internal/policy/guards

# 2. Via generator script
go run scripts/gen_guards.go

# 3. Via mise task (runs both integrations and guards)
mise run generate
```

### Adding New Built-in Guards

1. Create `internal/policy/guards/builtin/my_guard.go`
2. Implement the `Guard` interface
3. Add `init()` function to register the guard
4. Run `go generate ./internal/policy/guards`
5. The guard is now automatically included in `all.go`

**Example**:

```bash
# Create new guard
cat > internal/policy/guards/builtin/jira_linked.go <<'EOF'
package builtin

import (
    "context"
    "github.com/santosr2/uptool/internal/policy/guards"
)

type JiraLinkedGuard struct{}

func init() {
    guards.Register(&JiraLinkedGuard{})
}

func (g *JiraLinkedGuard) Name() string {
    return "jira-linked"
}

func (g *JiraLinkedGuard) Description() string {
    return "Verifies that the PR is linked to a JIRA ticket"
}

func (g *JiraLinkedGuard) Check(ctx context.Context, env *guards.Environment) (bool, error) {
    // Implementation here
    return true, nil
}
EOF

# Regenerate
go generate ./internal/policy/guards

# Verify
git diff internal/policy/guards/builtin/all.go
```

## Best Practices

### 1. Validate Inputs

Always validate environment variables and API responses:

```go
func (g *MyGuard) Check(ctx context.Context, env *guards.Environment) (bool, error) {
    token := os.Getenv("MY_API_TOKEN")
    if token == "" {
        return false, fmt.Errorf("MY_API_TOKEN not set")
    }

    // Validate env.GitHubPRNumber is numeric (already validated by caller, but defensive)
    if env.GitHubPRNumber == "" {
        return false, fmt.Errorf("PR number is empty")
    }

    // Continue with guard logic...
}
```

### 2. Use Context Timeouts

Set reasonable timeouts for external API calls:

```go
func (g *MyGuard) Check(ctx context.Context, env *guards.Environment) (bool, error) {
    client := &http.Client{
        Timeout: 10 * time.Second, // Reasonable timeout
    }

    req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
    // Use ctx to respect upstream cancellation
}
```

### 3. Handle Errors Gracefully

Return meaningful errors to help with debugging:

```go
if resp.StatusCode != http.StatusOK {
    body, _ := io.ReadAll(resp.Body)
    return false, fmt.Errorf(
        "API returned %d: %s",
        resp.StatusCode,
        string(body),
    )
}
```

### 4. Never Log Secrets

Avoid logging tokens, credentials, or sensitive data:

```go
// BAD - logs token
log.Printf("Using token: %s", token)

// GOOD - no secrets logged
log.Printf("Making API request to %s", apiURL)
```

### 5. Write Tests

Create unit tests for your guards:

```go
// examples/guards/custom/slack_approval_test.go
package custom_test

import (
    "context"
    "testing"

    "github.com/santosr2/uptool/examples/guards/custom"
    "github.com/santosr2/uptool/internal/policy/guards"
)

func TestSlackApprovalGuard_Name(t *testing.T) {
    g := &custom.SlackApprovalGuard{}
    if got := g.Name(); got != "slack-approval" {
        t.Errorf("Name() = %q, want %q", got, "slack-approval")
    }
}

func TestSlackApprovalGuard_Check(t *testing.T) {
    // Use environment variables or mocks for testing
    t.Setenv("SLACK_BOT_TOKEN", "test-token")
    t.Setenv("SLACK_CHANNEL_ID", "C123")

    g := &custom.SlackApprovalGuard{}
    env := &guards.Environment{
        GitHubRepo:     "owner/repo",
        GitHubToken:    "test-gh-token",
        GitHubPRNumber: "42",
    }

    // Test with mock HTTP server or skip integration tests
    t.Skip("Integration test - requires Slack API")

    ctx := context.Background()
    satisfied, err := g.Check(ctx, env)
    if err != nil {
        t.Fatalf("Check() error = %v", err)
    }

    t.Logf("Guard satisfied: %v", satisfied)
}
```

### 6. Add Clear Documentation

Document guard behavior in the `Description()` method:

```go
func (g *MyGuard) Description() string {
    return "Verifies that the PR has been approved in the ACME ticketing system " +
           "(requires ACME_API_TOKEN and ACME_PROJECT_ID environment variables)"
}
```

## Troubleshooting

### Guard Not Found

**Error**: `unknown guard: my-custom-guard`

**Cause**: Guard not registered in the registry.

**Solution**:

1. Verify `init()` function calls `guards.Register()`
2. Ensure guard package is imported (with `_` blank import if needed)
3. Check that the guard file is in the correct directory

### Guard Always Returns False

**Symptom**: Guard consistently fails even when conditions should pass.

**Debug steps**:

1. Add debug logging to your `Check()` method
2. Run `uptool check-policy --verbose` to see detailed output
3. Verify environment variables are set correctly
4. Check API responses for unexpected formats

### Permission Errors

**Error**: `403 Forbidden` or `401 Unauthorized`

**Cause**: Missing or incorrect API tokens.

**Solution**:

1. Verify all required environment variables are set
2. Check that API tokens have required permissions
3. Test API access manually with `curl` or similar tools

### Timeout Errors

**Error**: `context deadline exceeded`

**Cause**: Guard taking too long to execute.

**Solution**:

1. Increase HTTP client timeout if reasonable
2. Optimize API calls (batch requests, caching)
3. Consider implementing retry logic with backoff

## Examples

### Minimal Guard

```go
type SimpleGuard struct{}

func init() {
    guards.Register(&SimpleGuard{})
}

func (g *SimpleGuard) Name() string {
    return "simple"
}

func (g *SimpleGuard) Description() string {
    return "Always passes (for testing)"
}

func (g *SimpleGuard) Check(ctx context.Context, env *guards.Environment) (bool, error) {
    return true, nil
}
```

### Guard with External API

See `examples/guards/custom/slack_approval.go` for a complete example.

### Guard with Custom Configuration

```go
type ConfigurableGuard struct {
    Threshold int
}

func init() {
    threshold := 3
    if val := os.Getenv("MY_GUARD_THRESHOLD"); val != "" {
        if parsed, err := strconv.Atoi(val); err == nil {
            threshold = parsed
        }
    }

    guards.Register(&ConfigurableGuard{Threshold: threshold})
}

func (g *ConfigurableGuard) Check(ctx context.Context, env *guards.Environment) (bool, error) {
    // Use g.Threshold in logic
    approvals := getApprovalCount(env)
    return approvals >= g.Threshold, nil
}
```

## Contributing Built-in Guards

If you create a useful custom guard, consider contributing it as a built-in guard:

1. Move guard to `internal/policy/guards/builtin/`
2. Add comprehensive unit tests
3. Update documentation in this file
4. Submit a pull request

**Example PR**:

- File: `internal/policy/guards/builtin/jira_linked.go`
- Tests: `internal/policy/guards/builtin/jira_linked_test.go`
- Docs: Update this file with guard description
- Changelog: Add entry to `CHANGELOG.md`

## Related Documentation

- [Organization Policy](policy.md) - Overview of org_policy configuration
- [Configuration](configuration.md) - Full configuration reference
- [Architecture](architecture.md) - System architecture
- [Guard Examples](https://github.com/santosr2/uptool/blob/{% raw %}{{ extra.uptool_version }}{% endraw %}/examples/guards/) - Complete guard examples

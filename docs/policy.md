# Organization Policy

uptool supports organization-wide policies to enforce governance, security, and compliance requirements for dependency updates. These policies ensure that all dependency updates meet your organization's standards before being merged.

## Overview

Organization policies enable you to:

- **Require signoffs** from team members before merging updates
- **Verify artifact signatures** using Cosign to ensure supply chain security
- **Auto-merge** pull requests when all configured guards pass
- **Enforce custom checks** via extensible guard plugins

## Configuration

Configure organization policies in your `uptool.yaml`:

```yaml
version: 1

org_policy:
  # Require signoff from specific GitHub users
  require_signoff_from:
    - "@security-team"
    - "@platform-leads"

  # Verify artifact signatures with Cosign
  signing:
    cosign_verify:
      enabled: true
      public_key: "cosign.pub"  # Path to public key

  # Auto-merge configuration
  auto_merge:
    enabled: true
    guards:
      - "ci-green"             # All CI checks must pass
      - "codeowners-approve"   # CODEOWNERS must approve
      - "security-scan"        # Security scans must pass
      - "custom-guard"         # Your custom guard plugin
```

## Policy Components

### Require Signoff

Require approval from specific GitHub users or teams before merging dependency updates:

```yaml
org_policy:
  require_signoff_from:
    - "@security-team"
    - "@alice"
    - "@bob"
```

**How it works**: The `check-policy` command verifies that the specified users have approved the PR.

### Signing Verification

Verify artifact signatures to ensure integrity and authenticity:

```yaml
org_policy:
  signing:
    cosign_verify:
      enabled: true
      public_key: "/path/to/cosign.pub"
```

**How it works**: Uses [Cosign](https://github.com/sigstore/cosign) to verify signatures against the specified public key.

### Auto-Merge

Automatically merge pull requests when all guards pass:

```yaml
org_policy:
  auto_merge:
    enabled: true
    guards:
      - "ci-green"
      - "codeowners-approve"
      - "security-scan"
```

**How it works**: The `check-policy` command evaluates all configured guards. If all guards pass, the PR is marked as ready for auto-merge.

## Built-in Guards

uptool provides three built-in guards out-of-the-box:

### ci-green

**Purpose**: Ensure all CI checks pass before merging.

**Checks**:

- All CI workflow runs are in `SUCCESS` or `SKIPPED` state
- No failed or pending checks

**Usage**:

```yaml
org_policy:
  auto_merge:
    guards:
      - "ci-green"
```

**Environment variables required**:

- `GITHUB_TOKEN` - GitHub API token
- `GITHUB_REPOSITORY` - Repository in `owner/repo` format
- `GITHUB_PR_NUMBER` - Pull request number

### codeowners-approve

**Purpose**: Require approval from repository CODEOWNERS.

**Checks**:

- At least one CODEOWNER has approved the PR
- Uses GitHub's PR review API

**Usage**:

```yaml
org_policy:
  auto_merge:
    guards:
      - "codeowners-approve"
```

**Environment variables required**:

- `GITHUB_TOKEN` - GitHub API token
- `GITHUB_REPOSITORY` - Repository in `owner/repo` format
- `GITHUB_PR_NUMBER` - Pull request number

### security-scan

**Purpose**: Verify that security scans (CodeQL, Trivy, SAST) pass.

**Checks**:

- CodeQL analysis completed without critical/high findings
- Container image scans pass (if applicable)
- SAST tools report no security vulnerabilities

**Usage**:

```yaml
org_policy:
  auto_merge:
    guards:
      - "security-scan"
```

**Environment variables required**:

- `GITHUB_TOKEN` - GitHub API token
- `GITHUB_REPOSITORY` - Repository in `owner/repo` format
- `GITHUB_PR_NUMBER` - Pull request number

## Custom Guard Plugins

uptool's guard system is fully extensible via Go plugins. You can create custom guards to implement organization-specific workflows.

See [Guard Plugins Documentation](guards.md) for complete details on creating custom guards.

### Quick Example

```go
package custom

import (
    "context"
    "github.com/santosr2/uptool/internal/policy/guards"
)

type MyCustomGuard struct{}

func init() {
    guards.Register(&MyCustomGuard{})
}

func (g *MyCustomGuard) Name() string {
    return "my-custom-guard"
}

func (g *MyCustomGuard) Description() string {
    return "Checks my custom approval workflow"
}

func (g *MyCustomGuard) Check(ctx context.Context, env *guards.Environment) (bool, error) {
    // Your custom logic here
    return true, nil
}
```

Then reference it in your configuration:

```yaml
org_policy:
  auto_merge:
    guards:
      - "my-custom-guard"
```

## Running Policy Checks

Use the `uptool check-policy` command to validate policies:

```bash
# Check all policies for a PR
uptool check-policy

# Set environment variables for GitHub context
export GITHUB_TOKEN="ghp_your_token"
export GITHUB_REPOSITORY="owner/repo"
export GITHUB_PR_NUMBER="123"

uptool check-policy
```

### GitHub Action Integration

Run policy checks in CI/CD workflows:

```yaml
name: Policy Check

on:
  pull_request:
    types: [opened, synchronize, reopened]

jobs:
  policy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Check policy
        uses: santosr2/uptool@v0
        with:
          command: check-policy
          token: {% raw %}${{ secrets.GITHUB_TOKEN }}{% endraw %}
        env:
          GITHUB_REPOSITORY: {% raw %}${{ github.repository }}{% endraw %}
          GITHUB_PR_NUMBER: {% raw %}${{ github.event.pull_request.number }}{% endraw %}
```

### Output Format

The `check-policy` command outputs a detailed report:

```text
✓ Signoff requirement: PASSED
  Required: @security-team, @alice
  Found: @alice, @bob

✓ Artifact signing: PASSED
  Verified signature with cosign.pub

✓ Auto-merge guards: PASSED
  ✓ ci-green: All CI checks passing
  ✓ codeowners-approve: Approved by @alice (CODEOWNER)
  ✓ security-scan: No security issues found

All policy checks passed.
```

### Exit Codes

- `0` - All policy checks passed
- `1` - One or more policy checks failed

## Best Practices

### 1. Start Simple

Begin with basic guards and add complexity as needed:

```yaml
# Start with just CI checks
org_policy:
  auto_merge:
    enabled: true
    guards:
      - "ci-green"
```

### 2. Layer Security

Combine multiple guards for defense-in-depth:

```yaml
org_policy:
  auto_merge:
    guards:
      - "ci-green"           # Tests pass
      - "security-scan"      # No vulnerabilities
      - "codeowners-approve" # Human review
```

### 3. Test Guards Locally

Before deploying to CI/CD, test guards locally:

```bash
export GITHUB_TOKEN="ghp_your_token"
export GITHUB_REPOSITORY="santosr2/uptool"
export GITHUB_PR_NUMBER="42"

uptool check-policy --verbose
```

### 4. Document Custom Guards

If you create custom guards, document their behavior:

```yaml
# uptool.yaml
org_policy:
  auto_merge:
    guards:
      - "ci-green"
      - "slack-approval"  # Custom guard (see examples/guards/)
```

### 5. Monitor Guard Performance

Track how often guards fail to identify bottlenecks:

```bash
# Check policy with verbose output
uptool check-policy --verbose

# Review which guards most frequently block merges
```

## Troubleshooting

### Guard Not Found

**Error**: `unknown guard: my-custom-guard`

**Solution**: Ensure your guard package is imported and registered via `init()`. See [Guard Plugins](guards.md) for details.

### Guard Always Fails

**Symptom**: A guard consistently returns `false`

**Debug steps**:

1. Run with `--verbose` flag to see detailed error messages
2. Check environment variables are set correctly
3. Verify API tokens have required permissions
4. Review guard implementation for logic errors

### GitHub API Rate Limiting

**Error**: `rate limit exceeded` when checking guards

**Solutions**:

- Use a GitHub token with higher rate limits
- Cache guard results to reduce API calls
- Increase delays between checks

### Permission Denied

**Error**: `403 Forbidden` when accessing GitHub API

**Solutions**:

- Verify `GITHUB_TOKEN` has correct scopes
- For organization repositories, token needs `repo` scope
- For public repositories, `public_repo` scope is sufficient

## Examples

See [examples/uptool.yaml](https://github.com/santosr2/uptool/blob/{% raw %}{{ extra.uptool_version }}{% endraw %}/examples/uptool.yaml) for complete configuration examples.

See [examples/guards/](https://github.com/santosr2/uptool/blob/{% raw %}{{ extra.uptool_version }}{% endraw %}/examples/guards/) for custom guard implementations.

## Related Documentation

- [Guard Plugins](guards.md) - Creating custom guard plugins
- [Configuration](configuration.md) - Full configuration reference
- [GitHub Action Usage](action-usage.md) - Using uptool in CI/CD
- [Architecture](architecture.md) - System architecture overview

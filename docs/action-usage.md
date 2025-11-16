# GitHub Action Usage Guide

This guide covers how to use uptool as a GitHub Action to automate dependency updates in your CI/CD workflows.

## Table of Contents

- [Quick Start](#quick-start)
- [Workflow Examples](#workflow-examples)
- [Inputs Reference](#inputs-reference)
- [Outputs Reference](#outputs-reference)
- [Permissions](#permissions)
- [Common Patterns](#common-patterns)
- [Troubleshooting](#troubleshooting)

## Quick Start

### Basic Weekly Update Workflow

Create `.github/workflows/uptool.yml`:

```yaml
name: Dependency Updates

on:
  schedule:
    - cron: '0 9 * * 1'  # Every Monday at 9 AM UTC
  workflow_dispatch:      # Manual trigger

permissions:
  contents: write
  pull-requests: write

jobs:
  update:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: santosr2/uptool@v0  # Recommended: latest stable
        with:
          command: update
          create-pr: 'true'
          token: ${{ secrets.GITHUB_TOKEN }}
```

**Version Pinning Options**:
- `@v0` - Latest stable v0.x.x (recommended, auto-updates)
- `@v0.1` - Latest v0.1.x patch (mutable)
- `@v0.1.0` - Exact version (immutable, most secure)

This workflow will:

1. Run every Monday at 9 AM UTC
2. Scan for dependency updates
3. Create a pull request with changes

## Workflow Examples

### Scan and Report Only

Generate a scan report without making changes:

```yaml
name: Dependency Scan

on:
  pull_request:
    branches: [main]

jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: santosr2/uptool@v0
        with:
          command: scan
          format: json
        id: scan

      - name: Comment scan results
        if: steps.scan.outputs.updates-available == 'true'
        uses: actions/github-script@v7
        with:
          script: |
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: 'Outdated dependencies detected. Run `uptool plan` to see details.'
            })
```

### Plan with Manual Approval

Show what would update, then wait for manual approval:

```yaml
name: Dependency Update (Manual Approval)

on:
  workflow_dispatch:

permissions:
  contents: write
  pull-requests: write

jobs:
  plan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Generate update plan
        uses: santosr2/uptool@v0
        with:
          command: plan
          format: table
        id: plan

      - name: Show plan
        run: echo "Updates available: ${{ steps.plan.outputs.updates-available }}"

  update:
    needs: plan
    runs-on: ubuntu-latest
    environment: production  # Requires manual approval
    steps:
      - uses: actions/checkout@v4

      - name: Apply updates
        uses: santosr2/uptool@v0
        with:
          command: update
          create-pr: 'true'
          token: ${{ secrets.GITHUB_TOKEN }}
```

### Integration-Specific Updates

Run different integrations on different schedules:

```yaml
name: npm Updates (Daily)

on:
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM UTC

permissions:
  contents: write
  pull-requests: write

jobs:
  npm-update:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: santosr2/uptool@v0
        with:
          command: update
          only: npm
          create-pr: 'true'
          pr-title: 'chore(deps): update npm dependencies'
          pr-branch: 'uptool/npm-updates'
          token: ${{ secrets.GITHUB_TOKEN }}
```

```yaml
name: Terraform Updates (Weekly)

on:
  schedule:
    - cron: '0 9 * * 1'  # Monday at 9 AM UTC

permissions:
  contents: write
  pull-requests: write

jobs:
  terraform-update:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: santosr2/uptool@v0
        with:
          command: update
          only: terraform,tflint
          create-pr: 'true'
          pr-title: 'chore(infra): update Terraform dependencies'
          pr-branch: 'uptool/terraform-updates'
          token: ${{ secrets.GITHUB_TOKEN }}
```

### Dry-Run for Safety

Preview changes without creating PRs:

```yaml
name: Dependency Update Dry-Run

on:
  schedule:
    - cron: '0 9 * * 1'

jobs:
  dry-run:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Check for updates
        uses: santosr2/uptool@v0
        with:
          command: update
          dry-run: 'true'
          diff: 'true'

      - name: Post to Slack
        if: always()
        run: |
          # Send notification about available updates
          echo "Updates available!"
```

### Monorepo Pattern

Update different parts of a monorepo separately:

```yaml
name: Monorepo Updates

on:
  schedule:
    - cron: '0 9 * * 1'

permissions:
  contents: write
  pull-requests: write

jobs:
  frontend-deps:
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: ./frontend
    steps:
      - uses: actions/checkout@v4

      - uses: santosr2/uptool@v0
        with:
          command: update
          only: npm
          create-pr: 'true'
          pr-title: 'chore(frontend): update dependencies'

  backend-deps:
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: ./backend
    steps:
      - uses: actions/checkout@v4

      - uses: santosr2/uptool@v0
        with:
          command: update
          only: npm
          create-pr: 'true'
          pr-title: 'chore(backend): update dependencies'
```

## Inputs Reference

| Input | Description | Default | Required |
|-------|-------------|---------|----------|
| `command` | Command to run: `scan`, `plan`, or `update` | `plan` | No |
| `format` | Output format: `table` or `json` | `table` | No |
| `only` | Comma-separated integrations to include | `''` | No |
| `exclude` | Comma-separated integrations to exclude | `''` | No |
| `dry-run` | Show changes without applying (update only) | `false` | No |
| `diff` | Show diffs of changes (update only) | `true` | No |
| `create-pr` | Create a pull request with updates | `false` | No |
| `pr-title` | Title for the pull request | `chore: update dependencies` | No |
| `pr-branch` | Branch name for the PR | `uptool/dependency-updates` | No |
| `token` | GitHub token for API access and PR creation | `${{ github.token }}` | No |

### Input Details

**`command`**: The uptool command to run
- `scan`: Discover manifests and dependencies
- `plan`: Show available updates
- `update`: Apply updates to manifests

**`only`**: Filter to specific integrations
```yaml
only: npm              # Single integration
only: npm,helm         # Multiple integrations
only: npm,terraform,tflint
```

**`exclude`**: Exclude specific integrations
```yaml
exclude: precommit     # Skip pre-commit
exclude: terraform,tflint
```

**`create-pr`**: Automatically create a pull request
- Requires `command: update`
- Requires appropriate permissions

**`token`**: GitHub authentication
- Use `${{ secrets.GITHUB_TOKEN }}` for auto-scoped token
- Can use PAT for additional permissions

## Outputs Reference

| Output | Description | Type |
|--------|-------------|------|
| `updates-available` | Whether updates were found | `true`/`false` |
| `manifests-updated` | Number of manifests with updates | number |
| `dependencies-updated` | Total dependencies updated | number |

### Using Outputs

```yaml
steps:
  - uses: santosr2/uptool@v0.1 # or v0.1.0 or v0 or commit hash
    id: uptool
    with:
      command: plan

  - name: Check if updates available
    if: steps.uptool.outputs.updates-available == 'true'
    run: |
      echo "Found updates in ${{ steps.uptool.outputs.manifests-updated }} manifests"
      echo "Total dependencies to update: ${{ steps.uptool.outputs.dependencies-updated }}"
```

## Permissions

### Minimum Permissions

For `scan` and `plan` (read-only):
```yaml
permissions:
  contents: read
```

For `update` with `create-pr`:
```yaml
permissions:
  contents: write        # To create commits
  pull-requests: write   # To create PRs
```

### Token Scopes

**`GITHUB_TOKEN` (recommended)**:
- Automatically provided by GitHub Actions
- Scoped to the current repository
- Expires after the workflow run

**Personal Access Token (PAT)**:
- Use when cross-repository access needed
- Required scopes: `repo`, `workflow` (if updating workflows)
- Store as repository secret

## Common Patterns

### Skip CI on Update PRs

Add `[skip ci]` to PR commits:

```yaml
- uses: santosr2/uptool@v0.1 # or v0.1.0 or v0 or commit hash
  with:
    command: update
    create-pr: 'true'
    pr-title: '[skip ci] chore: update dependencies'
```

### Auto-Merge Patch Updates

Automatically merge patch-level updates:

```yaml
jobs:
  update:
    # ... update steps ...

  auto-merge:
    needs: update
    if: needs.update.outputs.updates-available == 'true'
    runs-on: ubuntu-latest
    steps:
      - name: Enable auto-merge for patch updates
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          gh pr merge --auto --squash uptool/dependency-updates
```

### Notify on Failures

Send notifications when updates fail:

```yaml
jobs:
  update:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: santosr2/uptool@v0
        with:
          command: update
          create-pr: 'true'
        continue-on-error: true
        id: uptool

      - name: Notify on failure
        if: failure()
        uses: actions/github-script@v7
        with:
          script: |
            github.rest.issues.create({
              owner: context.repo.owner,
              repo: context.repo.repo,
              title: 'Dependency update failed',
              body: 'The automated dependency update workflow failed. Check the logs.'
            })
```

## Troubleshooting

### PR Not Created

**Problem**: Workflow runs but no PR appears

**Solutions**:
1. Check permissions are set correctly
2. Verify `create-pr: 'true'` (must be string)
3. Ensure token has `pull-requests: write` scope
4. Check if a PR already exists on that branch

### No Updates Found

**Problem**: uptool reports "No updates available" but dependencies are outdated

**Solutions**:
1. Check integration is enabled: `--only=<integration>`
2. Verify manifest files are in expected locations
3. Test locally: `uptool scan` and `uptool plan`
4. Check registry connectivity (npm, Terraform, etc.)

### Authentication Errors

**Problem**: "403 Forbidden" or "401 Unauthorized"

**Solutions**:
1. For GitHub releases: May be rate-limited without auth
2. Use `GITHUB_TOKEN`: `token: ${{ secrets.GITHUB_TOKEN }}`
3. For private registries: Configure credentials separately

### Action Times Out

**Problem**: Workflow exceeds time limits

**Solutions**:
1. Use `--only` to run integrations separately
2. Schedule different integrations at different times
3. Increase timeout: `timeout-minutes: 30`

### Merge Conflicts

**Problem**: Update PR has conflicts

**Solutions**:
1. Close PR and let workflow create a fresh one
2. Update the base branch: `git merge main`
3. Schedule updates when `main` is stable

## Advanced Configuration

### Custom PR Body Template

```yaml
- uses: santosr2/uptool@v0.1 # or v0.1.0 or v0 or commit hash
  with:
    command: update
    create-pr: 'true'

# Note: PR body is currently fixed, but this shows the structure
# Future enhancement could allow custom templates
```

### Integration with Other Tools

**Run tests after update**:
```yaml
- uses: santosr2/uptool@v0.1 # or v0.1.0 or v0 or commit hash
  with:
    command: update

- name: Run tests
  run: npm test

- name: Create PR if tests pass
  if: success()
  uses: peter-evans/create-pull-request@v6
  with:
    title: 'chore: update dependencies (tests passing)'
```

### Matrix Strategy for Multiple Integrations

```yaml
jobs:
  update:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        integration: [npm, helm, terraform]
    steps:
      - uses: actions/checkout@v4

      - uses: santosr2/uptool@v0
        with:
          command: update
          only: ${{ matrix.integration }}
          create-pr: 'true'
          pr-branch: uptool/${{ matrix.integration }}-updates
```

## Best Practices

1. **Pin the action version**: Use `@v0` (recommended), `@v0.1`, or `@v0.1.0` - never `@main`
2. **Test dry-run first**: Validate behavior with `dry-run: 'true'`
3. **Use specific integrations**: Run `--only=npm` for targeted updates
4. **Schedule wisely**: Avoid peak hours and weekends
5. **Review PRs**: Always review before merging
6. **Monitor for failures**: Set up notifications
7. **Keep permissions minimal**: Only grant what's needed

---

Questions? Open a [Discussion](https://github.com/santosr2/uptool/discussions)!

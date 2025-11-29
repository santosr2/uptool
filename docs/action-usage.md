# GitHub Action Usage

Use uptool as a GitHub Action to automate dependency updates.

## Quick Start

```yaml
# .github/workflows/uptool.yml
name: Dependency Updates

on:
  schedule:
    - cron: '0 9 * * 1'  # Monday at 9 AM UTC
  workflow_dispatch:

permissions:
  contents: write
  pull-requests: write

jobs:
  update:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: santosr2/uptool@v0  # Latest stable (recommended)
        with:
          command: update
          create-pr: 'true'
          token: {% raw %}${{  secrets.GITHUB_TOKEN  }}{% endraw %}
```

**Version pinning**:

- `@{{ extra.uptool_version_major }}` - Latest stable (auto-updates)
- `@{{ extra.uptool_version_minor }}` - Latest patch
- `@{{ extra.uptool_version }}` - Exact version (most secure)

## Common Patterns

### Scan Only (No Updates)

```yaml
- uses: santosr2/uptool@v0
  with:
    command: scan
    format: json
```

### Dry-Run Before Applying

```yaml
- uses: santosr2/uptool@v0
  with:
    command: update
    dry-run: 'true'
```

### Integration-Specific Updates

```yaml
- uses: santosr2/uptool@v0
  with:
    command: update
    only: npm,helm
    create-pr: 'true'
```

### Custom Config File

```yaml
- uses: santosr2/uptool@v0
  with:
    command: update
    config: configs/production-uptool.yaml
    create-pr: 'true'
```

### Monorepo Pattern

```yaml
strategy:
  matrix:
    package: [api, web, worker]
steps:
  - uses: actions/checkout@v4
  - name: Update dependencies for {% raw %}${{ matrix.package }}{% endraw %}
    working-directory: packages/{% raw %}${{ matrix.package }}{% endraw %}
    run: |
      uptool update --diff
```

Note: For monorepo patterns, run uptool directly with the `working-directory` step option or use a custom config file per package.

### Create Issue Instead of PR

```yaml
- uses: santosr2/uptool@v0
  with:
    command: plan
    create-issue: 'true'
    issue-title: 'Weekly Dependency Report'
    issue-labels: 'dependencies,review-needed'
    token: {% raw %}${{ secrets.GITHUB_TOKEN }}{% endraw %}
```

This creates or updates an issue with available updates instead of automatically creating a PR.

## Inputs

| Input | Required | Default | Description |
|-------|----------|---------|-------------|
| `command` | No | `plan` | Command: `scan`, `plan`, or `update` |
| `format` | No | `table` | Output format: `table` or `json` |
| `only` | No | `''` | Comma-separated integrations to include |
| `exclude` | No | `''` | Comma-separated integrations to exclude |
| `config` | No | `''` | Path to uptool config file (default: `uptool.yaml`) |
| `dry-run` | No | `false` | Preview without applying (update only) |
| `diff` | No | `true` | Show diffs of changes (update only) |
| `create-pr` | No | `false` | Create pull request with updates |
| `pr-title` | No | `chore: update dependencies` | PR title |
| `pr-branch` | No | `uptool/dependency-updates` | PR branch name |
| `create-issue` | No | `false` | Create issue when updates available |
| `issue-title` | No | `Dependency Updates Available` | Issue title |
| `issue-labels` | No | `dependencies,automated` | Comma-separated issue labels |
| `token` | No | `{% raw %}${{ github.token }}{% endraw %}` | GitHub token |
| `skip-install` | No | `false` | Skip uptool installation (use when already in PATH) |

## Outputs

| Output | Description |
|--------|-------------|
| `updates-available` | `true` if updates found |
| `manifests-updated` | Number of manifests with updates |
| `dependencies-updated` | Number of dependencies updated |

**Usage**:

```yaml
- uses: santosr2/uptool@v0
  id: uptool
  with:
    command: plan

- name: Check results
  if: steps.uptool.outputs.updates-available == 'true'
  run: echo "Found {% raw %}${{ steps.uptool.outputs.dependencies-updated }}{% endraw %} updates across {% raw %}${{ steps.uptool.outputs.manifests-updated }}{% endraw %} manifests"
```

## Permissions

**Minimum required**:

```yaml
permissions:
  contents: write          # To commit changes
  pull-requests: write     # To create PRs
```

**For auto-merge**:

```yaml
permissions:
  contents: write
  pull-requests: write
  checks: read             # To verify checks pass
```

## Advanced Patterns

### Skip CI on Update PRs

```yaml
- uses: santosr2/uptool@v0
  with:
    pr-title: 'chore(deps): update dependencies [skip ci]'
```

### Notify on Failures

```yaml
- uses: santosr2/uptool@v0
  continue-on-error: true
  id: uptool

- name: Notify on failure
  if: failure()
  uses: slackapi/slack-github-action@v1
  with:
    payload: |
      {
        "text": "uptool failed in {% raw %}${{ github.repository }}{% endraw %}"
      }
```

### Custom PR with Additional Steps

```yaml
- uses: santosr2/uptool@v0
  id: uptool
  with:
    command: update
    create-pr: 'true'
    pr-title: 'chore(deps): weekly dependency updates'
    token: {% raw %}${{ secrets.GITHUB_TOKEN }}{% endraw %}

- name: Add comment to PR
  if: steps.uptool.outputs.updates-available == 'true'
  run: |
    echo "Updated {% raw %}${{ steps.uptool.outputs.dependencies-updated }}{% endraw %} dependencies"
```

Note: The action automatically generates a detailed PR body with update information.

### Matrix Strategy for Environments

```yaml
strategy:
  matrix:
    env: [staging, production]
steps:
  - uses: actions/checkout@v4
  - name: Update {% raw %}${{ matrix.env }}{% endraw %} dependencies
    working-directory: environments/{% raw %}${{ matrix.env }}{% endraw %}
    run: uptool update --diff
  - name: Create PR
    uses: peter-evans/create-pull-request@v7
    with:
      branch: uptool/updates-{% raw %}${{ matrix.env }}{% endraw %}
      title: "chore({% raw %}${{ matrix.env }}{% endraw %}): update dependencies"
```

Note: Use the `working-directory` step option with direct uptool commands for environment-specific updates.

## Troubleshooting

### PR Not Created

**Check**:

- Permissions include `contents: write` and `pull-requests: write`
- Token has repo access
- No existing PR with same branch name

### No Updates Found

**Check**:

- Manifest files exist in repository
- Integration enabled in `uptool.yaml`
- Run with `dry-run: 'true'` to see debug output

### Authentication Errors

**For private packages**:

```yaml
- name: Setup npm auth
  run: echo "//registry.npmjs.org/:_authToken={% raw %}${{  secrets.NPM_TOKEN  }}{% endraw %}" > ~/.npmrc

- uses: santosr2/uptool@v0
  env:
    GITHUB_TOKEN: {% raw %}${{  secrets.GITHUB_TOKEN  }}{% endraw %}
```

### Action Times Out

**Increase timeout**:

```yaml
- uses: santosr2/uptool@v0
  timeout-minutes: 15  # Default is 360
```

## Best Practices

1. **Use semantic versioning**: Pin to `@v0` for auto-updates
2. **Run on schedule**: Weekly or daily, avoid high-traffic times
3. **Enable manual trigger**: Add `workflow_dispatch` for testing
4. **Test in staging first**: Use matrix strategy for environments
5. **Review PRs**: Don't blindly auto-merge major updates
6. **Set PR labels**: Use `pr-labels: 'dependencies,automated'`
7. **Configure branch protection**: Require reviews for major updates

## Examples

See [.github/workflows/](https://github.com/santosr2/uptool/blob/{{ extra.uptool_version }}/.github/workflows/) for working examples:

- `dependency-updates.yml` - Weekly automated updates
- `dependency-scan.yml` - PR scan checks

## See Also

- [Quick Start](quickstart.md) - CLI usage
- [Configuration](configuration.md) - `uptool.yaml` reference
- [action.yml](https://github.com/santosr2/uptool/blob/{{ extra.uptool_version }}/action.yml) - Action definition

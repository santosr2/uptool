# Pre-Commit Integration

The pre-commit integration updates pre-commit hook versions in `.pre-commit-config.yaml` files.

## Overview

**Integration ID**: `precommit`

**Manifest Files**: `.pre-commit-config.yaml`

**Update Strategy**: **Native command** - Uses `pre-commit autoupdate`

**Registry**: GitHub Releases (for each hook repository)

**Status**: ✅ Stable

## What Gets Updated

The pre-commit integration updates hook repository revisions:

- `repos[].rev` - Git tag/commit of each hook repository

## Example

**Before** (`.pre-commit-config.yaml`):
```yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.3.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-added-large-files

  - repo: https://github.com/psf/black
    rev: 22.10.0
    hooks:
      - id: black

  - repo: https://github.com/PyCQA/flake8
    rev: 5.0.4
    hooks:
      - id: flake8
```

**After** (uptool update):
```yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v6.0.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-added-large-files

  - repo: https://github.com/psf/black
    rev: 24.10.0
    hooks:
      - id: black

  - repo: https://github.com/PyCQA/flake8
    rev: 7.1.1
    hooks:
      - id: flake8
```

## Why Native Command?

uptool uses `pre-commit autoupdate` instead of custom rewriting because:

1. **Manifest-first**: `pre-commit autoupdate` updates `.pre-commit-config.yaml` directly
2. **Comprehensive**: Handles all edge cases (local hooks, complex revisions, etc.)
3. **Maintained**: pre-commit team maintains update logic
4. **Reliable**: Well-tested by entire pre-commit ecosystem

This aligns with uptool's philosophy: **use native commands when they update manifests**.

## CLI Usage

### Scan for Pre-Commit Configs

```bash
uptool scan --only=precommit
```

Output:
```
Type                 Path                        Dependencies
----------------------------------------------------------------
precommit            .pre-commit-config.yaml     3

Total: 1 manifest
```

### Plan Pre-Commit Updates

```bash
uptool plan --only=precommit
```

Output:
```
.pre-commit-config.yaml (precommit):
Hook                 Current         Target          Impact
--------------------------------------------------------
pre-commit-hooks     v4.3.0          v6.0.0          major
black                22.10.0         24.10.0         major
flake8               5.0.4           7.1.1           major

Total: 3 updates across 1 manifest
```

### Apply Pre-Commit Updates

```bash
# Dry run first
uptool update --only=precommit --dry-run --diff

# Apply updates
uptool update --only=precommit

# No additional steps needed!
# (pre-commit autoupdate already updated the manifest)
```

## Configuration

### Update Policy

```yaml
# uptool.yaml
version: 1

integrations:
  - id: precommit
    enabled: true
    policy:
      update: minor              # none, patch, minor, major
      allow_prerelease: false
```

**Update Levels**:
- `none` - No updates
- `patch` - Only patch updates
- `minor` - Patch + minor updates
- `major` - All updates including major

**Note**: Policy is passed to `pre-commit autoupdate` where supported.

## Hook Types

### Remote Hooks

Standard GitHub-based hooks:

```yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.3.0
    hooks:
      - id: trailing-whitespace
```

### Local Hooks

Local hooks are NOT updated:

```yaml
repos:
  - repo: local
    hooks:
      - id: custom-check
        name: Custom Check
        entry: scripts/custom-check.sh
        language: script
```

### Meta Hooks

Meta hooks using `meta` repository:

```yaml
repos:
  - repo: meta
    hooks:
      - id: check-hooks-apply
      - id: check-useless-excludes
```

Meta hooks have no version and are NOT updated.

## Pre-Commit Installation

uptool does **NOT** install or manage pre-commit itself.

Ensure pre-commit is installed:

```bash
# Install pre-commit
pip install pre-commit
# or
brew install pre-commit
# or
mise install pre-commit

# Install hooks
pre-commit install
```

## Requirements

The pre-commit integration requires:

1. **pre-commit installed**: Must be available in `$PATH`
2. **Valid config**: `.pre-commit-config.yaml` must be valid YAML
3. **Git repository**: pre-commit requires a git repository

### Checking Requirements

```bash
# Check pre-commit is installed
which pre-commit
pre-commit --version

# Validate config
pre-commit validate-config

# Ensure in git repo
git status
```

## GitHub Rate Limiting

`pre-commit autoupdate` queries GitHub API for each hook repository.

### Unauthenticated

- Rate limit: 60 requests/hour
- Usually sufficient for small configs

### Authenticated

Set `GITHUB_TOKEN` or configure git credentials:

```bash
# Method 1: Environment variable
export GITHUB_TOKEN="your_github_token"
uptool update --only=precommit

# Method 2: Git credential helper
git config --global credential.helper store
```

## Limitations

1. **Requires pre-commit installed**: Must have `pre-commit` in PATH
   - Solution: Install pre-commit first

2. **Git repository required**: pre-commit needs git
   - Solution: Ensure you're in a git repository

3. **No offline mode**: Requires internet to check versions
   - Solution: Update when online

4. **Limited policy control**: Pre-commit decides what to update
   - uptool passes preferences where supported

## Troubleshooting

### pre-commit Not Found

**Problem**: "pre-commit command not found"

**Causes**:
1. pre-commit not installed
2. Not in PATH

**Solutions**:
```bash
# Install pre-commit
pip install pre-commit

# Or via package manager
brew install pre-commit       # macOS
apt install pre-commit         # Debian/Ubuntu

# Or via mise
mise install pre-commit

# Verify installation
which pre-commit
pre-commit --version
```

### Invalid Config

**Problem**: "Invalid .pre-commit-config.yaml"

**Causes**:
1. YAML syntax errors
2. Invalid hook configuration

**Solutions**:
```bash
# Validate config
pre-commit validate-config

# Check YAML syntax
yamllint .pre-commit-config.yaml

# Fix errors and retry
```

### Hook Repository Not Found

**Problem**: "Repository not found" error

**Causes**:
1. Hook repository URL incorrect
2. Repository deleted or renamed
3. Network issues

**Solutions**:
```bash
# Verify repository exists
curl -I https://github.com/pre-commit/pre-commit-hooks

# Check URL in config
grep repo .pre-commit-config.yaml

# Update URL if repository moved
```

### GitHub Rate Limit

**Problem**: "API rate limit exceeded"

**Causes**:
1. Too many hooks without authentication
2. Running updates frequently

**Solutions**:
```bash
# Set GitHub token
export GITHUB_TOKEN="ghp_xxxxxxxxxxxx"

# Or configure git credentials
git config --global credential.helper store
git fetch  # Trigger credential prompt

# Check rate limit
curl -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/rate_limit
```

### Hooks Not Installing

**Problem**: After update, `pre-commit run` fails

**Causes**:
1. New hook version incompatible
2. Hook environment not set up

**Solutions**:
```bash
# Clean and reinstall hooks
pre-commit clean
pre-commit install-hooks

# Run specific hook
pre-commit run <hook-id> --all-files

# Check hook logs
pre-commit run --verbose
```

## Best Practices

1. **Test after updating**:
   ```bash
   uptool update --only=precommit
   pre-commit run --all-files
   ```

2. **Review hook changelogs**:
   ```bash
   # Check for breaking changes
   # Visit hook repository's releases page
   ```

3. **Update regularly**:
   ```bash
   # Weekly or monthly
   uptool update --only=precommit
   ```

4. **Pin critical hooks** (if needed):
   ```yaml
   repos:
     - repo: https://github.com/psf/black
       rev: 22.10.0  # Exact version
       hooks:
         - id: black
   ```

5. **Use GitHub token in CI**:
   ```yaml
   # .github/workflows/update.yml
   env:
     GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
   steps:
     - uses: santosr2/uptool@v0
       with:
         command: update
         only: precommit
   ```

6. **Commit lock files if used**:
   ```bash
   git add .pre-commit-config.yaml
   git commit -m "chore: update pre-commit hooks"
   ```

## Pre-Commit Version Compatibility

uptool works with pre-commit >= 2.0.0.

**pre-commit < 2.0.0**:
- Not tested
- May not support all features

**pre-commit >= 2.0.0**:
- ✅ Fully supported
- `autoupdate` command available

**pre-commit >= 3.0.0**:
- ✅ Fully supported
- Improved update logic

## Minimum Pre-Commit Version

Some hooks require specific pre-commit versions:

```yaml
minimum_pre_commit_version: 2.9.0
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.3.0
    hooks:
      - id: check-yaml
```

uptool respects this setting when running `pre-commit autoupdate`.

## See Also

- [Pre-Commit Documentation](https://pre-commit.com/)
- [Pre-Commit Hooks](https://github.com/pre-commit/pre-commit-hooks)
- [Supported Hooks](https://pre-commit.com/hooks.html)
- [Manifest Files Reference](../manifests.md)

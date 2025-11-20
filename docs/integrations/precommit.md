# Pre-Commit Integration

Updates pre-commit hook versions in `.pre-commit-config.yaml` files.

## Overview

**Integration ID**: `precommit`

**Manifest Files**: `.pre-commit-config.yaml`

**Update Strategy**: **Native command** - Uses `pre-commit autoupdate`

**Registry**: GitHub Releases (per hook repository)

**Status**: ✅ Stable

## What Gets Updated

Hook repository revisions in the `repos` list:

- `repos[].rev` - Git tag or commit SHA of each hook repository
- Remote hooks only (local and meta hooks skipped)

## Example

**Before**:

```yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.3.0
    hooks:
      - id: trailing-whitespace
      - id: check-yaml

  - repo: https://github.com/psf/black
    rev: 22.10.0
    hooks:
      - id: black

  - repo: local    # Not updated
    hooks:
      - id: custom-check
```

**After**:

```yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v6.0.0      # Updated
    hooks:
      - id: trailing-whitespace
      - id: check-yaml

  - repo: https://github.com/psf/black
    rev: 24.10.0     # Updated
    hooks:
      - id: black

  - repo: local    # Unchanged
    hooks:
      - id: custom-check
```

## Integration-Specific Behavior

### Why Native Command?

uptool uses `pre-commit autoupdate` instead of custom rewriting because:

1. **Manifest-first**: Updates `.pre-commit-config.yaml` directly ✅
2. **Comprehensive**: Handles all edge cases (local hooks, complex revisions)
3. **Maintained**: pre-commit team owns the update logic
4. **Reliable**: Battle-tested by entire pre-commit ecosystem

This aligns with uptool's philosophy: **use native commands when they update manifests**.

### Hook Types

| Type | Example | Updated? |
|------|---------|----------|
| Remote | `repo: https://github.com/...` | ✅ Yes |
| Local | `repo: local` | ❌ No |
| Meta | `repo: meta` | ❌ No |

### GitHub Rate Limits

`pre-commit autoupdate` queries GitHub API for each hook. Set `GITHUB_TOKEN` for higher limits:

```bash
export GITHUB_TOKEN="your_token"
uptool update --only precommit
```

- Unauthenticated: 60 requests/hour
- Authenticated: 5,000 requests/hour

## Configuration

```yaml
version: 1

integrations:
  - id: precommit
    enabled: true
    policy:
      update: major        # Aggressive for dev tools (safe)
      allow_prerelease: false
```

## Requirements

1. **pre-commit installed**: Must be in `$PATH`
2. **Git repository**: pre-commit requires git
3. **Valid config**: `.pre-commit-config.yaml` must be valid YAML

Install pre-commit:

```bash
pip install pre-commit
# or
brew install pre-commit
# or
mise install pre-commit

# Verify
pre-commit --version
```

## Limitations

1. **Requires pre-commit CLI**: Must have `pre-commit` installed and in PATH.
2. **Git repository required**: pre-commit needs a git repository to operate.
3. **Limited policy control**: pre-commit decides what to update (up tool passes preferences where supported).

## See Also

- [CLI Reference](../cli/commands.md) - `uptool scan --only precommit`
- [Configuration Guide](../configuration.md) - Policy settings
- [pre-commit Documentation](https://pre-commit.com/)
- [Supported Hooks](https://pre-commit.com/hooks.html)

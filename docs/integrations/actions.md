# GitHub Actions Integration

Update GitHub Actions workflow files to use the latest action versions.

## Overview

**Integration ID**: `actions`

**Manifest Files**: `.github/workflows/*.yml`, `.github/workflows/*.yaml`

**Update Strategy**: YAML text rewriting (preserves formatting and comments)

**Registry**: GitHub Releases API

**Status**: ✅ Stable

## What Gets Updated

- `uses:` directives in workflow steps (e.g., `actions/checkout@v4` → `actions/checkout@v4.2.2`)
- Action references with version tags (e.g., `@v4`, `@v4.2.2`)

**Not Updated**:

- SHA-pinned actions (e.g., `@11bd71901bbe5b1630ceea73d27597364c9af683`) - kept for security
- Local actions (e.g., `uses: ./.github/actions/my-action`)
- Docker Hub references (e.g., `uses: docker://alpine:3.8`)

## Example

**Before** (`.github/workflows/ci.yml`):

```yaml
name: CI
on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      - uses: actions/cache@v3
        with:
          path: ~/.npm
          key: {% raw %}${{ runner.os }}-node-${{ hashFiles('**/package-lock.json') }}{% endraw %}
```

**After**:

```yaml
name: CI
on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4.2.2
      - uses: actions/setup-node@v4.1.0
        with:
          node-version: '20'
      - uses: actions/cache@v4.1.2
        with:
          path: ~/.npm
          key: {% raw %}${{ runner.os }}-node-${{ hashFiles('**/package-lock.json') }}{% endraw %}
```

## Integration-Specific Behavior

- **Version tags**: Detects `v1`, `v1.2`, `v1.2.3` style tags and updates to latest matching release
- **SHA preservation**: Actions pinned to full commit SHAs are not updated (security-conscious teams often pin to SHAs)
- **Comment preservation**: YAML comments and formatting are preserved during updates
- **Multi-job support**: Scans all jobs and steps in a workflow file
- **Deduplication**: Same action@version appearing multiple times is only counted once

## Configuration

Example `uptool.yaml` configuration:

```yaml
version: 1

integrations:
  - id: actions
    enabled: true
    policy:
      update: minor        # Only update minor/patch versions
      allow_prerelease: false
```

## Limitations

1. **No SHA-to-tag conversion**: If you pin to SHAs, uptool won't convert them to tags
2. **Registry-only actions**: Only actions available on GitHub are supported (no private registries)
3. **Major version jumps**: Use `update: major` policy carefully - major versions may have breaking changes

## See Also

- [CLI Reference](../cli/commands.md)
- [Configuration Guide](../configuration.md)
- [Docker Integration](docker.md) - For updating Docker images in workflows

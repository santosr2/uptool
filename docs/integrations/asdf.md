# asdf Integration

Updates tool versions in `.tool-versions` files used by the asdf version manager.

## Overview

**Integration ID**: `asdf`

**Manifest Files**: `.tool-versions`

**Update Strategy**: Line-based parsing and rewriting

**Registry**: GitHub Releases (per tool via asdf plugin mapping)

**Status**: ⚠️ Experimental (85% test coverage, version resolution not yet implemented)

## What Gets Updated

Tool version entries in `.tool-versions`:

- Each line format: `tool_name version`
- Comments and formatting preserved
- Multiple versions per tool supported (space-separated)

## Example

**Before**:

```text
# Development tools
go 1.23.0
nodejs 20.10.0
terraform 1.5.0

# Build tools
python 3.11.0
ruby 3.2.0
```

**After**:

```text
# Development tools
go 1.25.0
nodejs 22.12.0
terraform 1.10.5

# Build tools
python 3.13.1
ruby 3.3.6
```

## Integration-Specific Behavior

### File Format

Simple line-based format:

```text
tool_name version [version2 version3...]  # Optional comment
```

uptool updates the first (primary) version for each tool.

### Tool Installation

uptool updates **only** `.tool-versions`. Run `asdf install` after to install new versions:

```bash
uptool update --only asdf
asdf install
```

### GitHub Rate Limits

Each tool queries GitHub Releases. Set `GITHUB_TOKEN` for higher limits:

```bash
export GITHUB_TOKEN="your_token"
uptool update --only asdf
```

- Unauthenticated: 60 requests/hour
- Authenticated: 5,000 requests/hour

## Configuration

```yaml
version: 1

integrations:
  - id: asdf
    enabled: true
    policy:
      update: patch        # Conservative for runtimes
      allow_prerelease: false
```

## Limitations

1. **Experimental status**: Version resolution not yet implemented. `uptool plan` returns empty update lists.
2. **Detection only**: Currently only scans and detects `.tool-versions` files and their dependencies.
3. **No updates yet**: Use native `asdf` commands for updates:
   - `asdf plugin update --all` - Update all plugins
   - `asdf list all <tool>` - Check available versions
   - `asdf install <tool> latest` - Install latest version
4. **Future implementation**: Will query GitHub Releases per tool for version checking.

## See Also

- [CLI Reference](../cli/commands.md) - `uptool scan --only asdf`
- [Configuration Guide](../configuration.md) - Policy settings
- [asdf Documentation](https://asdf-vm.com/)
- [mise Integration](mise.md) - Modern alternative to asdf

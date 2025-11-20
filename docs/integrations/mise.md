# mise Integration

Updates development tool versions in `mise.toml` and `.mise.toml` files.

## Overview

**Integration ID**: `mise`

**Manifest Files**: `mise.toml`, `.mise.toml`

**Update Strategy**: TOML parsing and rewriting

**Registry**: GitHub Releases (per tool)

**Status**: ⚠️ Experimental (86% test coverage, version resolution not yet implemented)

## What Gets Updated

Tool versions in the `[tools]` section:

- String format: `tool = "version"`
- Map format: `tool = { version = "version", ... }`
- Preserves whichever format you use

**Monorepo support**: Each `mise.toml` updated independently.

## Example

**Before**:

```toml
[tools]
go = "1.23"
node = "20"
python = "3.11"
terraform = { version = "1.5.0" }
golangci-lint = "2.6"
```

**After**:

```toml
[tools]
go = "1.25"
node = "22"
python = "3.13"
terraform = { version = "1.10.5" }   # Preserves map format
golangci-lint = "2.7"
```

## Integration-Specific Behavior

### Version Formats

mise supports two TOML formats:

| Format | Example | Use Case |
|--------|---------|----------|
| String | `go = "1.25"` | Simple (recommended) |
| Map | `go = { version = "1.25", path = ".go-version" }` | With additional options |

uptool preserves the format - map stays map, string stays string.

### Partial Versions

mise allows partial version specifications that uptool updates:

```toml
go = "1.23"      # Updates to "1.25" (latest 1.x)
node = "20"      # Updates to "22" (latest)
python = "3.11"  # Updates to "3.13" (latest 3.x)
```

### Tool Installation

uptool updates **only** `mise.toml`. Run `mise install` after updating to install new versions:

```bash
uptool update --only mise
mise install
```

### GitHub Rate Limits

Each tool queries GitHub Releases. Set `GITHUB_TOKEN` for higher limits:

```bash
export GITHUB_TOKEN="your_token"
uptool update --only mise
```

- Unauthenticated: 60 requests/hour
- Authenticated: 5,000 requests/hour

## Configuration

```yaml
version: 1

integrations:
  - id: mise
    enabled: true
    match:
      files:
        - "mise.toml"
        - ".mise.toml"     # Hidden variant
        - "*/mise.toml"    # Nested projects
    policy:
      update: patch        # Conservative for runtimes
      allow_prerelease: false
```

## Limitations

1. **Experimental status**: Version resolution not yet implemented. `uptool plan` returns empty update lists.
2. **Detection only**: Currently only scans and detects `mise.toml`/`.mise.toml` files and their dependencies.
3. **No updates yet**: Use native `mise` commands for updates:
   - `mise upgrade` - Update all tools
   - `mise latest <tool>` - Check latest version
   - `mise install <tool>@latest` - Install latest version
4. **Future implementation**: Will query GitHub Releases per tool for version checking.

## See Also

- [CLI Reference](../cli/commands.md) - `uptool scan --only mise`
- [Configuration Guide](../configuration.md) - Policy settings
- [mise Documentation](https://mise.jdx.dev/)
- [asdf Integration](asdf.md) - Alternative runtime manager

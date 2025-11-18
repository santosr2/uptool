# mise Integration

The mise integration updates tool versions in `mise.toml` or `.mise.toml` files used by the mise version manager (formerly rtx).

## Overview

**Integration ID**: `mise`

**Manifest Files**: `mise.toml`, `.mise.toml`

**Update Strategy**: TOML parsing and rewriting

**Registry**: GitHub Releases (per tool)

**Status**: ✅ Stable

## What Gets Updated

The mise integration updates tool versions in the `[tools]` section:

- Tool version values (both string and map format)
- Preserves comments and formatting

## Example

### String Format

**Before** (`mise.toml`):

```toml
[tools]
go = "1.23"
node = "20"
golangci-lint = "2.6"
terraform = "1.5.0"
python = "3.11"
```

**After** (uptool update):

```toml
[tools]
go = "1.25"
node = "22"
golangci-lint = "2.7"
terraform = "1.10.5"
python = "3.13"
```

### Map Format

**Before** (`mise.toml`):

```toml
[tools]
go = { version = "1.23", path = ".go-version" }
node = { version = "20" }
python = { version = "3.11", virtualenv = ".venv" }
```

**After** (uptool update):

```toml
[tools]
go = { version = "1.25", path = ".go-version" }
node = { version = "22" }
python = { version = "3.13", virtualenv = ".venv" }
```

## Tool Detection

uptool detects tools in the `[tools]` section and maps them to GitHub repositories:

### Supported Tools

Common tools with known GitHub release patterns:

- **go** → golang/go
- **node** / **nodejs** → nodejs/node
- **python** → python/cpython
- **ruby** → ruby/ruby
- **terraform** → hashicorp/terraform
- **kubectl** → kubernetes/kubernetes
- **helm** → helm/helm
- **golangci-lint** → golangci/golangci-lint
- And many more

## CLI Usage

### Scan for mise Configs

```bash
uptool scan --only=mise
```

Output:

```text
Type                 Path                Dependencies
----------------------------------------------------------------
mise                 mise.toml           6
mise                 .mise.toml          4

Total: 2 manifests
```

### Plan mise Updates

```bash
uptool plan --only=mise
```

Output:

```text
mise.toml (mise):
Tool             Current         Target          Impact
--------------------------------------------------------
go               1.23            1.25            minor
node             20              22              major
golangci-lint    2.6             2.7             minor
terraform        1.5.0           1.10.5          minor
python           3.11            3.13            minor

Total: 5 updates across 1 manifest
```

### Apply mise Updates

```bash
# Dry run first
uptool update --only=mise --dry-run --diff

# Apply updates
uptool update --only=mise

# Then install new versions
mise install
```

## Version Formats

mise supports two format styles:

### String Format (Simple)

```toml
[tools]
go = "1.23"
node = "20"
```

Recommended for most use cases.

### Map Format (Advanced)

```toml
[tools]
go = { version = "1.23", path = ".go-version" }
node = { version = "20", prefix = "v" }
python = { version = "3.11", virtualenv = ".venv" }
```

Used when additional options are needed.

**uptool preserves the format** - if map format is used, it stays map format.

## Configuration

### Update Policy

```yaml
# uptool.yaml
version: 1

integrations:
  - id: mise
    enabled: true
    policy:
      update: minor              # none, patch, minor, major
      allow_prerelease: false
```

**Update Levels**:

- `none` - No updates
- `patch` - Only patch updates (1.23.0 → 1.23.1)
- `minor` - Patch + minor updates (1.23.0 → 1.24.0)
- `major` - All updates including major (1.23.0 → 2.0.0)

### Exclude Specific Tools

Pin to exact version (won't be updated based on policy):

```toml
[tools]
go = "1.23"              # Will be updated
node = "20.10.0"         # Exact version (patch specified)
```

## Tool Installation

uptool **does NOT** install tool versions.

After updating `mise.toml`, install new versions:

```bash
# Install all tools at specified versions
mise install

# Install specific tool
mise install node

# Verify installations
mise current
```

## Config File Locations

mise supports multiple config file locations:

### Project Config

```tree
my-project/
├── mise.toml                   # Project config (recommended)
├── .mise.toml                  # Hidden variant
└── subproject/
    └── mise.toml               # Nested project
```

### Global Config

```text
~/.config/mise/config.toml      # Global config
```

uptool scans and updates:

- ✅ `mise.toml` in repository
- ✅ `.mise.toml` in repository
- ❌ Global config (not in repo)

## Per-Directory Configuration

mise supports directory-specific configs:

```tree
monorepo/
├── mise.toml                   # Root tools
├── frontend/
│   └── mise.toml               # Frontend-specific tools
└── backend/
    └── mise.toml               # Backend-specific tools
```

Each `mise.toml` is updated independently.

## Version Specification

mise supports various version formats:

### Exact Version

```toml
[tools]
go = "1.23.0"
```

### Partial Version (mise auto-resolves)

```toml
[tools]
go = "1.23"     # mise installs latest 1.23.x
node = "20"     # mise installs latest 20.x
```

**uptool updates the partial version** (1.23 → 1.25).

### Version Prefix

```toml
[tools]
go = { version = "1.23", prefix = "go" }
```

uptool updates the version field only.

## Legacy Tool Version Files

mise can read from other version files:

```toml
[tools]
node = { version = "20", path = ".nvmrc" }
python = { version = "3.11", path = ".python-version" }
```

**uptool updates the version in mise.toml**, not the referenced file.

## GitHub Rate Limiting

uptool queries GitHub Releases API for each tool.

### Unauthenticated

- Rate limit: 60 requests/hour
- May be insufficient for many tools

### Authenticated

Set `GITHUB_TOKEN` environment variable:

```bash
export GITHUB_TOKEN="your_github_token"
uptool update --only=mise
```

Rate limit: 5000 requests/hour

## Limitations

1. **No version installation**: uptool only updates `mise.toml`
   - Solution: Run `mise install` after

2. **GitHub-based tools only**: Tools must release via GitHub Releases
   - Non-GitHub tools not yet supported

3. **No backend validation**: Doesn't check mise backend availability
   - Solution: Test with `mise install` before committing

4. **No .mise.lock**: mise doesn't use lockfiles
   - Versions are declarative

5. **No plugin version updates**: Only tool versions, not mise plugins
   - mise plugins managed separately

## Troubleshooting

### Tool Not Found

**Problem**: "Tool not found in GitHub"

**Causes**:

1. Tool name doesn't map to known GitHub repository
2. Tool uses non-GitHub backend
3. Tool name misspelled

**Solutions**:

```bash
# Check mise tool name
mise ls-remote <tool>

# Verify tool in mise.toml
cat mise.toml

# Test tool availability
mise install <tool>@latest
```

### GitHub Rate Limit

**Problem**: "API rate limit exceeded"

**Causes**:

1. Many tools without authentication
2. Frequent updates

**Solutions**:

```bash
# Set GitHub token
export GITHUB_TOKEN="ghp_xxxxxxxxxxxx"

# Verify rate limit
curl -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/rate_limit

# Run update
uptool update --only=mise
```

### Version Not Available

**Problem**: uptool suggests version that doesn't exist

**Causes**:

1. Version numbering changed
2. Pre-release versions included

**Solutions**:

```bash
# Check available versions
mise ls-remote <tool>

# Or check GitHub releases
gh release list --repo golang/go

# Exclude pre-releases in policy
# uptool.yaml:
policy:
  allow_prerelease: false
```

### Installation Fails

**Problem**: `mise install` fails after update

**Causes**:

1. New version has different dependencies
2. Backend not available
3. Build requirements missing

**Solutions**:

```bash
# Check mise backend
mise doctor

# Update mise
mise self-update

# Install build dependencies
# (varies by tool and OS)

# Check specific tool
mise install <tool>@<version> --verbose
```

### TOML Parse Error

**Problem**: "Failed to parse mise.toml"

**Causes**:

1. Invalid TOML syntax
2. Unsupported mise.toml features

**Solutions**:

```bash
# Validate TOML
cat mise.toml | mise validate

# Or use TOML linter
taplo format mise.toml --check

# Fix syntax errors
```

## Best Practices

1. **Always install after updating**:

   ```bash
   uptool update --only=mise
   mise install
   git add mise.toml
   git commit -m "chore: update tool versions"
   ```

2. **Test versions before committing**:

   ```bash
   mise install
   mise current  # Verify all tools installed
   # Run your test suite
   ```

3. **Use string format when possible**:

   ```toml
   # Preferred (simpler)
   [tools]
   go = "1.23"

   # Only when needed
   [tools]
   go = { version = "1.23", path = ".go-version" }
   ```

4. **Use conservative policy**:

   ```yaml
   # For production
   policy:
     update: patch

   # For development
   policy:
     update: minor
   ```

5. **Set GitHub token**:

   ```bash
   # In CI/CD
   env:
     GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

   # Locally
   export GITHUB_TOKEN="your_token"
   ```

6. **Document required tools**:

   ```toml
   # mise.toml
   [tools]
   # Development tools
   go = "1.23"
   node = "20"

   # CI/CD tools
   golangci-lint = "2.6"
   ```

## mise Version Compatibility

uptool works with mise >= 2024.1.0.

**mise < 2024.1.0** (rtx era):

- May work but not tested
- TOML format may differ

**mise >= 2024.1.0**:

- ✅ Fully supported
- Standard `mise.toml` format

## File Naming

mise supports both naming conventions:

- `mise.toml` - Visible file (recommended)
- `.mise.toml` - Hidden file (legacy/preference)

**uptool scans both**.

## Comparison with asdf

mise and asdf are similar:

| Feature | mise | asdf |
|---------|------|------|
| Config file | `mise.toml` | `.tool-versions` |
| Format | TOML | Simple text |
| Version syntax | `tool = "version"` | `tool version` |
| Multiple versions | Limited | Yes |
| Performance | Faster (Rust) | Slower (Bash) |

See [asdf integration](asdf.md) for asdf-specific documentation.

## Special Tool Names

Some tools have aliases:

```toml
[tools]
node = "20"      # Also: nodejs
go = "1.23"      # Also: golang
```

uptool recognizes common aliases.

## See Also

- [mise Documentation](https://mise.jdx.dev/)
- [mise Configuration](https://mise.jdx.dev/configuration.html)
- [mise Tool Backends](https://mise.jdx.dev/dev-tools/backends/)
- [Manifest Files Reference](../manifests.md)

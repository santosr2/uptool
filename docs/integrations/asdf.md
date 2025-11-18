# asdf Integration

The asdf integration updates tool versions in `.tool-versions` files used by the asdf version manager.

## Overview

**Integration ID**: `asdf`

**Manifest Files**: `.tool-versions`

**Update Strategy**: Line-based parsing and rewriting

**Registry**: GitHub Releases (per tool via asdf plugin mapping)

**Status**: ✅ Stable

## What Gets Updated

The asdf integration updates tool versions:

- Tool version entries in `.tool-versions` file
- Supports multiple versions per tool (space-separated)

## Example

**Before** (`.tool-versions`):

```text
# Development tools
go 1.23.0
nodejs 20.10.0
terraform 1.5.0

# Build tools
python 3.11.0
ruby 3.2.0
```

**After** (uptool update):

```text
# Development tools
go 1.25.0
nodejs 22.12.0
terraform 1.10.5

# Build tools
python 3.13.1
ruby 3.3.6
```

## Tool Detection

uptool detects tools based on:

1. **Tool name**: Must match an asdf plugin name
2. **Version format**: Semantic version or tool-specific format
3. **Plugin mapping**: Maps tool names to GitHub repositories

### Supported Tools

Common tools with known GitHub release patterns:

- **go** → golang/go
- **nodejs** / **node** → nodejs/node
- **python** → python/cpython
- **ruby** → ruby/ruby
- **terraform** → hashicorp/terraform
- **kubectl** → kubernetes/kubernetes
- **helm** → helm/helm
- And many more via asdf plugins

## CLI Usage

### Scan for .tool-versions

```bash
uptool scan --only=asdf
```

Output:

```text
Type                 Path                Dependencies
----------------------------------------------------------------
asdf                 .tool-versions      5

Total: 1 manifest
```

### Plan asdf Updates

```bash
uptool plan --only=asdf
```

Output:

```text
.tool-versions (asdf):
Tool             Current         Target          Impact
--------------------------------------------------------
go               1.23.0          1.25.0          minor
nodejs           20.10.0         22.12.0         major
terraform        1.5.0           1.10.5          minor
python           3.11.0          3.13.1          minor
ruby             3.2.0           3.3.6           minor

Total: 5 updates across 1 manifest
```

### Apply asdf Updates

```bash
# Dry run first
uptool update --only=asdf --dry-run --diff

# Apply updates
uptool update --only=asdf

# Then install new versions
asdf install
```

## Multiple Versions

asdf supports multiple versions per tool:

**Before**:

```text
nodejs 20.10.0 18.19.0 16.20.0
```

**After** (uptool updates latest in each major):

```text
nodejs 22.12.0 18.20.5 16.20.2
```

uptool updates all specified versions independently.

## Configuration

### Update Policy

```yaml
# uptool.yaml
version: 1

integrations:
  - id: asdf
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

Pin to exact version:

```text
# .tool-versions
go 1.23.0        # Will be updated
nodejs 20.10.0   # Will be updated
python 3.11.0    # To pin, use comment
```

**Note**: asdf doesn't have built-in version pinning syntax. To prevent updates, use specific version numbers and update policy.

## Tool Version Installation

uptool **does NOT** install tool versions.

After updating `.tool-versions`, install new versions:

```bash
# Install all tools at specified versions
asdf install

# Install specific tool
asdf install nodejs

# Set global versions
asdf global nodejs 22.12.0
```

## Per-Project Configuration

asdf supports per-project `.tool-versions`:

```tree
my-project/
├── .tool-versions              # Project-specific versions
├── app/
│   └── .tool-versions          # Nested project versions
└── scripts/
```

Each `.tool-versions` is updated independently.

## Legacy Versions

Some tools use non-semver versioning:

```text
# Traditional versioning
java openjdk-11.0.2
erlang 26.0

# System version
nodejs system
python system
```

uptool handles:

- ✅ Semantic versions (1.2.3)
- ✅ Version prefixes (v1.2.3, openjdk-11.0.2)
- ❌ `system` keyword (not updated)
- ❌ Custom version strings (varies by tool)

## GitHub Rate Limiting

uptool queries GitHub Releases API for each tool.

### Unauthenticated

- Rate limit: 60 requests/hour
- May be insufficient for many tools

### Authenticated

Set `GITHUB_TOKEN` environment variable:

```bash
export GITHUB_TOKEN="your_github_token"
uptool update --only=asdf
```

Rate limit: 5000 requests/hour

## Limitations

1. **No version installation**: uptool only updates `.tool-versions`
   - Solution: Run `asdf install` after

2. **GitHub-based tools only**: Tools must release via GitHub Releases
   - Non-GitHub tools not yet supported

3. **No plugin installation**: Assumes asdf plugins already installed
   - Solution: Install plugins first with `asdf plugin add <tool>`

4. **No version validation**: Doesn't verify version compatibility
   - Solution: Test with `asdf install` before committing

5. **No .tool-versions.lock**: asdf doesn't use lockfiles
   - Versions are declarative

## Troubleshooting

### Tool Not Found

**Problem**: "Tool not found in GitHub"

**Causes**:

1. Tool name doesn't map to known GitHub repository
2. asdf plugin uses non-GitHub source
3. Tool name misspelled

**Solutions**:

```bash
# Check asdf plugin for tool
asdf plugin list
asdf plugin list all | grep <tool>

# Verify tool in .tool-versions
cat .tool-versions

# Add plugin if missing
asdf plugin add <tool>
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
uptool update --only=asdf
```

### Version Not Available

**Problem**: uptool suggests version that doesn't exist

**Causes**:

1. Version numbering changed
2. Pre-release versions included

**Solutions**:

```bash
# Check available versions
asdf list all <tool>

# Or check GitHub releases
gh release list --repo golang/go

# Exclude pre-releases in policy
# uptool.yaml:
policy:
  allow_prerelease: false
```

### Installation Fails

**Problem**: `asdf install` fails after update

**Causes**:

1. New version has different dependencies
2. Plugin not updated
3. Build requirements missing

**Solutions**:

```bash
# Update asdf plugin
asdf plugin update <tool>
asdf plugin update --all

# Check plugin requirements
asdf plugin list
cat ~/.asdf/plugins/<tool>/README.md

# Install build dependencies
# (varies by tool and OS)
```

## Best Practices

1. **Always install after updating**:

   ```bash
   uptool update --only=asdf
   asdf install
   git add .tool-versions
   git commit -m "chore: update tool versions"
   ```

2. **Test versions before committing**:

   ```bash
   asdf install
   asdf current  # Verify all tools installed
   # Run your test suite
   ```

3. **Update plugins regularly**:

   ```bash
   asdf plugin update --all
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

6. **Document required plugins**:

   ```bash
   # In README or docs
   # Required asdf plugins:
   asdf plugin add nodejs
   asdf plugin add python
   asdf plugin add terraform
   ```

## asdf Version Compatibility

uptool works with asdf >= 0.8.0.

**asdf < 0.8.0**:

- Not tested
- `.tool-versions` format may differ

**asdf >= 0.8.0**:

- ✅ Fully supported
- Standard `.tool-versions` format

## Tool Version Format

asdf uses simple format:

```text
<tool-name> <version> [<version> ...]
```

Examples:

```text
nodejs 20.10.0
nodejs 20.10.0 18.19.0
terraform 1.5.0
python 3.11.0 3.10.13
```

Comments (lines starting with `#`) are preserved.

## Comparison with mise

asdf and mise use similar concepts:

| Feature | asdf | mise |
|---------|------|------|
| Config file | `.tool-versions` | `mise.toml` or `.mise.toml` |
| Format | Simple text | TOML |
| Version syntax | `tool version` | `tool = "version"` |
| Multiple versions | Yes (space-separated) | Limited |

See [mise integration](mise.md) for mise-specific documentation.

## See Also

- [asdf Documentation](https://asdf-vm.com/)
- [asdf Plugins](https://github.com/asdf-vm/asdf-plugins)
- [.tool-versions Format](https://asdf-vm.com/manage/configuration.html#tool-versions)
- [Manifest Files Reference](../manifests.md)

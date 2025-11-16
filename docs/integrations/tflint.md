# tflint Integration

The tflint integration updates tflint plugin versions in `.tflint.hcl` configuration files.

## Overview

**Integration ID**: `tflint`

**Manifest Files**: `.tflint.hcl`

**Update Strategy**: HCL parsing and rewriting via `hashicorp/hcl`

**Registry**: GitHub Releases (for each plugin)

**Status**: ✅ Stable

## What Gets Updated

The tflint integration updates plugin versions:

- `plugin` block `version` attributes - Plugin versions from GitHub releases

## Example

**Before** (`.tflint.hcl`):
```hcl
plugin "aws" {
  enabled = true
  version = "0.21.0"
  source  = "github.com/terraform-linters/tflint-ruleset-aws"
}

plugin "azurerm" {
  enabled = true
  version = "0.20.0"
  source  = "github.com/terraform-linters/tflint-ruleset-azurerm"
}

plugin "google" {
  enabled = false
  version = "0.18.0"
  source  = "github.com/terraform-linters/tflint-ruleset-google"
}
```

**After** (uptool update):
```hcl
plugin "aws" {
  enabled = true
  version = "0.44.0"
  source  = "github.com/terraform-linters/tflint-ruleset-aws"
}

plugin "azurerm" {
  enabled = true
  version = "0.28.0"
  source  = "github.com/terraform-linters/tflint-ruleset-azurerm"
}

plugin "google" {
  enabled = false
  version = "0.32.0"
  source  = "github.com/terraform-linters/tflint-ruleset-google"
}
```

## Plugin Sources

### GitHub-Based Plugins

Standard format:

```hcl
plugin "aws" {
  enabled = true
  version = "0.21.0"
  source  = "github.com/terraform-linters/tflint-ruleset-aws"
}
```

uptool queries GitHub Releases API for the plugin repository to find latest versions.

### Official tflint Rulesets

Common plugins:
- `github.com/terraform-linters/tflint-ruleset-aws` - AWS
- `github.com/terraform-linters/tflint-ruleset-azurerm` - Azure
- `github.com/terraform-linters/tflint-ruleset-google` - Google Cloud
- Custom plugins following same pattern

## CLI Usage

### Scan for tflint Configs

```bash
uptool scan --only=tflint
```

Output:
```
Type                 Path                Dependencies
----------------------------------------------------------------
tflint               .tflint.hcl         3
tflint               terraform/.tflint.hcl  2

Total: 2 manifests
```

### Plan tflint Updates

```bash
uptool plan --only=tflint
```

Output:
```
.tflint.hcl (tflint):
Plugin           Current         Target          Impact
--------------------------------------------------------
aws              0.21.0          0.44.0          minor
azurerm          0.20.0          0.28.0          minor
google           0.18.0          0.32.0          minor

Total: 3 updates across 1 manifest
```

### Apply tflint Updates

```bash
# Dry run first
uptool update --only=tflint --dry-run --diff

# Apply updates
uptool update --only=tflint

# Then update tflint plugins
tflint --init
```

## Configuration

### Update Policy

```yaml
# uptool.yaml
version: 1

integrations:
  - id: tflint
    enabled: true
    policy:
      update: minor              # none, patch, minor, major
      allow_prerelease: false
```

**Update Levels**:
- `none` - No updates
- `patch` - Only patch updates (0.21.0 → 0.21.1)
- `minor` - Patch + minor updates (0.21.0 → 0.22.0)
- `major` - All updates including major (0.21.0 → 1.0.0)

### Exclude Specific Plugins

Pin to exact version:

```hcl
plugin "aws" {
  enabled = true
  version = "0.21.0"  # Exact version, won't update
  source  = "github.com/terraform-linters/tflint-ruleset-aws"
}
```

## tflint Plugin Installation

uptool **does NOT** install plugins.

After updating `.tflint.hcl`, install updated plugins:

```bash
# Install/update all plugins
tflint --init

# Or with specific config
tflint --init --config=.tflint.hcl
```

## Multiple Config Files

tflint supports multiple configuration files:

```
project/
├── .tflint.hcl              # Root config, updated
└── modules/
    ├── network/
    │   └── .tflint.hcl      # Module config, updated
    └── database/
        └── .tflint.hcl      # Module config, updated
```

Each `.tflint.hcl` is scanned and updated independently.

## Plugin Versioning

tflint plugins follow semantic versioning (mostly):

- Plugins in 0.x.x may have breaking changes in minor versions
- Plugins >= 1.0.0 follow strict semver

**Recommendation**: Use `update: patch` policy for 0.x.x plugins.

## GitHub Rate Limiting

uptool queries GitHub Releases API for each plugin.

### Unauthenticated

- Rate limit: 60 requests/hour
- May be insufficient for many plugins

### Authenticated

Set `GITHUB_TOKEN` environment variable:

```bash
export GITHUB_TOKEN="your_github_token"
uptool update --only=tflint
```

Rate limit: 5000 requests/hour

## Limitations

1. **No plugin installation**: uptool only updates `.tflint.hcl`
   - Solution: Run `tflint --init` after

2. **GitHub-only plugins**: Only plugins on GitHub are supported
   - Other registries not yet supported

3. **No config validation**: uptool doesn't validate tflint configuration
   - Solution: Run `tflint --init` to validate

4. **Rate limiting**: GitHub API has rate limits
   - Solution: Set `GITHUB_TOKEN` for higher limits

## Troubleshooting

### Plugin Not Found

**Problem**: "Plugin not found on GitHub"

**Causes**:
1. Plugin repository doesn't exist
2. Source URL incorrect
3. Plugin renamed or moved

**Solutions**:
```bash
# Verify plugin exists
curl -I https://api.github.com/repos/terraform-linters/tflint-ruleset-aws/releases

# Check plugin source in .tflint.hcl
grep source .tflint.hcl

# Fix source if incorrect
```

### GitHub Rate Limit

**Problem**: "API rate limit exceeded"

**Causes**:
1. Too many plugin checks without authentication
2. Many plugins in config

**Solutions**:
```bash
# Set GitHub token
export GITHUB_TOKEN="ghp_xxxxxxxxxxxx"

# Verify rate limit
curl -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/rate_limit

# Run update
uptool update --only=tflint
```

### Version Not Available

**Problem**: uptool suggests version that doesn't exist

**Causes**:
1. Plugin releases use different versioning
2. Pre-release versions included

**Solutions**:
```bash
# Check available releases
gh release list --repo terraform-linters/tflint-ruleset-aws

# Set policy to exclude pre-releases
# in uptool.yaml:
policy:
  allow_prerelease: false
```

### Plugin Installation Fails

**Problem**: `tflint --init` fails after update

**Causes**:
1. New version incompatible with tflint version
2. Plugin architecture changed

**Solutions**:
```bash
# Check tflint version
tflint --version

# Upgrade tflint if needed
brew upgrade tflint  # macOS
# or download from GitHub releases

# Check plugin compatibility
# Visit plugin's GitHub page for requirements
```

## Best Practices

1. **Always reinstall plugins**:
   ```bash
   uptool update --only=tflint
   tflint --init
   git add .tflint.hcl .tflint.d/
   git commit -m "chore(tflint): update plugin versions"
   ```

2. **Test after updating**:
   ```bash
   tflint --init
   tflint
   # Verify linting still works
   ```

3. **Review plugin changelogs**:
   ```bash
   # Check for new rules or breaking changes
   # Visit plugin's GitHub releases page
   ```

4. **Use conservative update policy**:
   ```yaml
   # For 0.x.x plugins, use patch only
   policy:
     update: patch
   ```

5. **Set GitHub token**:
   ```bash
   # In CI/CD
   env:
     GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

   # Locally
   export GITHUB_TOKEN="your_token"
   ```

6. **Pin critical plugins**:
   ```hcl
   plugin "aws" {
     enabled = true
     version = "0.21.0"  # Exact version
     source  = "github.com/terraform-linters/tflint-ruleset-aws"
   }
   ```

## tflint Version Compatibility

uptool works with tflint >= 0.40.0 (plugin system v0.1.0).

**tflint < 0.40.0**:
- May work but not tested
- Plugin system may differ

**tflint >= 0.40.0**:
- ✅ Fully supported
- Plugin configuration in HCL

## See Also

- [tflint Documentation](https://github.com/terraform-linters/tflint)
- [tflint Plugins](https://github.com/terraform-linters/tflint/blob/master/docs/user-guide/plugins.md)
- [Available Rulesets](https://github.com/terraform-linters/tflint/blob/master/docs/user-guide/plugins.md#available-plugins)
- [Manifest Files Reference](../manifests.md)

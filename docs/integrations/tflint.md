# tflint Integration

Updates tflint plugin versions in `.tflint.hcl` configuration files.

## Overview

**Integration ID**: `tflint`

**Manifest Files**: `.tflint.hcl`

**Update Strategy**: HCL parsing and rewriting via `hashicorp/hcl`

**Registry**: GitHub Releases (per plugin)

**Status**: ✅ Stable

## What Gets Updated

Plugin versions in `plugin` blocks:

- `plugin` block `version` attributes - GitHub Releases versions

## Example

**Before**:

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
```

**After**:

```hcl
plugin "aws" {
  enabled = true
  version = "0.44.0"   # Updated
  source  = "github.com/terraform-linters/tflint-ruleset-aws"
}

plugin "azurerm" {
  enabled = true
  version = "0.28.0"   # Updated
  source  = "github.com/terraform-linters/tflint-ruleset-azurerm"
}
```

## Integration-Specific Behavior

### Plugin Sources

Updates plugins from GitHub sources:

```hcl
# ✅ Updated - GitHub source
plugin "aws" {
  version = "0.44.0"
  source  = "github.com/terraform-linters/tflint-ruleset-aws"
}

# ❌ Not updated - Custom source
plugin "custom" {
  version = "1.0.0"
  source  = "example.com/custom-plugin"
}
```

### tflint Init Required

uptool updates **only** `.tflint.hcl`. Run `tflint --init` after to install new plugin versions:

```bash
uptool update --only tflint
tflint --init
```

### GitHub Rate Limits

Each plugin queries GitHub Releases. Set `GITHUB_TOKEN` for higher limits:

```bash
export GITHUB_TOKEN="your_token"
uptool update --only tflint
```

- Unauthenticated: 60 requests/hour
- Authenticated: 5,000 requests/hour

## Configuration

```yaml
version: 1

integrations:
  - id: tflint
    enabled: true
    policy:
      update: major        # Aggressive for linters (safe)
      allow_prerelease: false
```

## Limitations

1. **GitHub sources only**: Custom plugin sources not supported.
2. **No plugin installation**: Run `tflint --init` after updating.

## See Also

- [CLI Reference](../cli/commands.md) - `uptool scan --only tflint`
- [Configuration Guide](../configuration.md) - Policy settings
- [TFLint Documentation](https://github.com/terraform-linters/tflint)
- [TFLint Rulesets](https://github.com/terraform-linters/tflint/blob/master/docs/user-guide/plugins.md)

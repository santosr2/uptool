# Terraform Integration

Updates Terraform module versions in `*.tf` files.

## Overview

**Integration ID**: `terraform`

**Manifest Files**: `*.tf`

**Update Strategy**: HCL parsing and rewriting via `hashicorp/hcl`

**Registry**: Terraform Registry API (`https://registry.terraform.io`)

**Status**: ✅ Stable

## What Gets Updated

Module versions in `module` blocks:

- `module` block `version` attributes - Terraform Registry modules

**Not yet supported** (future):
- Provider versions in `required_providers` blocks
- Git-based module source versions

## Example

**Before**:

```hcl
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "3.0.0"

  name = "my-vpc"
  cidr = "10.0.0.0/16"
}

module "security_group" {
  source  = "terraform-aws-modules/security-group/aws"
  version = "~> 4.0"

  name = "my-sg"
}
```

**After**:

```hcl
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "5.13.0"   # Updated

  name = "my-vpc"
  cidr = "10.0.0.0/16"
}

module "security_group" {
  source  = "terraform-aws-modules/security-group/aws"
  version = "~> 5.0"   # Updated (preserves constraint)

  name = "my-sg"
}
```

## Integration-Specific Behavior

### Version Constraint Preservation

uptool preserves version constraint operators:

| Constraint | Meaning | Before | After |
|------------|---------|--------|-------|
| (none) | Exact | `"3.0.0"` | `"5.13.0"` |
| `~>` | Pessimistic | `"~> 4.0"` | `"~> 5.0"` |
| `>=` | Greater or equal | `">= 3.0"` | `">= 5.13"` |

### Terraform Init Required

uptool updates **only** `.tf` files. Run `terraform init` after to update lockfile:

```bash
uptool update --only terraform
terraform init -upgrade
```

### Module Sources

Only Terraform Registry modules updated:

```hcl
# ✅ Updated - Registry module
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "5.13.0"
}

# ❌ Not updated - Git source
module "custom" {
  source = "git::https://github.com/org/repo.git?ref=v1.0.0"
}

# ❌ Not updated - Local path
module "local" {
  source = "./modules/networking"
}
```

## Configuration

```yaml
version: 1

integrations:
  - id: terraform
    enabled: true
    match:
      files:
        - "*.tf"
        - "**/*.tf"              # All subdirectories
    policy:
      update: patch              # Conservative for infrastructure
      allow_prerelease: false
```

## Limitations

1. **Registry modules only**: Local and Git sources not supported.
2. **No provider updates**: `required_providers` versions not yet updated.
3. **No lockfile updates**: Run `terraform init -upgrade` after.

## See Also

- [CLI Reference](../cli/commands.md) - `uptool scan --only terraform`
- [Configuration Guide](../configuration.md) - Policy settings
- [Terraform Registry](https://registry.terraform.io/)
- [Terraform Module Sources](https://www.terraform.io/language/modules/sources)

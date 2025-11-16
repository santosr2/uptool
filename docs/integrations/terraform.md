# Terraform Integration

The Terraform integration updates Terraform module versions in `*.tf` files.

## Overview

**Integration ID**: `terraform`

**Manifest Files**: `*.tf` (any Terraform file)

**Update Strategy**: HCL parsing and rewriting via `hashicorp/hcl`

**Registry**: Terraform Registry API (`https://registry.terraform.io`)

**Status**: ✅ Stable

## What Gets Updated

The Terraform integration currently updates:

- `module` block `version` attributes - Module versions from Terraform Registry

**Future support** (not yet implemented):
- Provider versions in `required_providers` blocks
- Module source versions in git URLs

## Example

**Before** (`main.tf`):
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

  name   = "my-sg"
  vpc_id = module.vpc.vpc_id
}

module "rds" {
  source  = "terraform-aws-modules/rds/aws"
  version = ">= 5.0"

  identifier = "mydb"
}
```

**After** (uptool update):
```hcl
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "5.13.0"

  name = "my-vpc"
  cidr = "10.0.0.0/16"
}

module "security_group" {
  source  = "terraform-aws-modules/security-group/aws"
  version = "~> 5.0"

  name   = "my-sg"
  vpc_id = module.vpc.vpc_id
}

module "rds" {
  source  = "terraform-aws-modules/rds/aws"
  version = ">= 6.0"

  identifier = "mydb"
}
```

## Version Constraints

Terraform supports various version constraint operators:

| Constraint | Meaning | Example Before | Example After |
|------------|---------|----------------|---------------|
| `=` | Exact version | `= 3.0.0` | `= 5.13.0` |
| `~>` | Pessimistic constraint | `~> 4.0` | `~> 5.0` |
| `>=` | Greater than or equal | `>= 5.0` | `>= 6.0` |
| `>` | Greater than | `> 3.0` | `> 5.0` |
| `<=` | Less than or equal | `<= 6.0` | `<= 8.0` |
| `<` | Less than | `< 6.0` | `< 8.0` |
| (none) | Exact version | `3.0.0` | `5.13.0` |

uptool preserves the constraint operator and updates the version number.

## CLI Usage

### Scan for Terraform Modules

```bash
uptool scan --only=terraform
```

Output:
```
Type                 Path                    Dependencies
----------------------------------------------------------------
terraform            main.tf                 3
terraform            modules/network/main.tf  2
terraform            infra/staging/main.tf    5

Total: 3 manifests
```

### Plan Terraform Updates

```bash
uptool plan --only=terraform
```

Output:
```
main.tf (terraform):
Module           Current         Target          Impact
--------------------------------------------------------
vpc              3.0.0           5.13.0          major
security_group   ~> 4.0          ~> 5.0          major
rds              >= 5.0          >= 6.0          major

modules/network/main.tf (terraform):
Module           Current         Target          Impact
--------------------------------------------------------
subnets          2.0.0           2.5.0           minor

Total: 4 updates across 2 manifests
```

### Apply Terraform Updates

```bash
# Dry run first
uptool update --only=terraform --dry-run --diff

# Apply updates
uptool update --only=terraform

# Then upgrade Terraform lockfile
terraform init -upgrade
```

## Module Sources

### Terraform Registry Modules

Standard format for registry modules:

```hcl
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "3.0.0"
}
```

Format: `NAMESPACE/NAME/PROVIDER`

**Supported registries**:
- Public Terraform Registry (default)
- Private Terraform Cloud/Enterprise registries (future)

### Git Sources

**Not yet supported** by uptool:

```hcl
module "consul" {
  source = "git::https://github.com/hashicorp/consul.git?ref=v1.0.0"
}
```

### Local Modules

Not applicable (no version):

```hcl
module "vpc" {
  source = "./modules/vpc"
}
```

## Configuration

### Update Policy

```yaml
# uptool.yaml
version: 1

integrations:
  - id: terraform
    enabled: true
    policy:
      update: minor              # none, patch, minor, major
      allow_prerelease: false
```

**Update Levels**:
- `none` - No updates
- `patch` - Only patch updates (3.0.0 → 3.0.1)
- `minor` - Patch + minor updates (3.0.0 → 3.1.0)
- `major` - All updates including major (3.0.0 → 5.0.0)

### Exclude Specific Modules

Pin to exact version:

```hcl
module "critical" {
  source  = "terraform-aws-modules/rds/aws"
  version = "5.0.0"  # Exact version, won't update
}
```

## Terraform Lock File

uptool **does NOT** update `.terraform.lock.hcl`.

After updating module versions, regenerate the lockfile:

```bash
# Upgrade modules and providers
terraform init -upgrade

# Or just recalculate hashes
terraform providers lock
```

## Monorepo / Multi-Environment

Common structure:

```
terraform/
├── environments/
│   ├── dev/
│   │   └── main.tf          # Updated independently
│   ├── staging/
│   │   └── main.tf          # Updated independently
│   └── prod/
│       └── main.tf          # Updated independently
└── modules/
    ├── network/
    │   └── main.tf          # Updated independently
    └── database/
        └── main.tf          # Updated independently
```

Each `*.tf` file is scanned and updated independently.

## Private Registries

### Terraform Cloud

Configure Terraform CLI:

```bash
# ~/.terraformrc
credentials "app.terraform.io" {
  token = "YOUR_TOKEN"
}
```

uptool respects Terraform's credential configuration.

### Terraform Enterprise

```hcl
module "private" {
  source  = "app.terraform.io/my-org/vpc/aws"
  version = "1.0.0"
}
```

Ensure Terraform is authenticated to your Enterprise instance.

## Limitations

1. **No .terraform.lock.hcl updates**: uptool only updates `*.tf` files
   - Solution: Run `terraform init -upgrade` after

2. **No provider version updates**: Only module versions are updated (for now)
   - Future enhancement

3. **No git source version updates**: Git URLs with refs not supported
   - Future enhancement

4. **No workspace awareness**: All workspaces use same module versions
   - Intentional (workspaces share code)

## Troubleshooting

### Module Not Found

**Problem**: "Module not found in registry"

**Causes**:
1. Module doesn't exist in Terraform Registry
2. Module name misspelled
3. Private module not accessible

**Solutions**:
```bash
# Search registry
terraform search modules vpc aws

# Verify module exists
curl https://registry.terraform.io/v1/modules/terraform-aws-modules/vpc/aws

# Check authentication for private modules
terraform login
```

### Version Constraint Errors

**Problem**: After updating, `terraform init` fails

**Causes**:
1. New version has breaking changes
2. Version constraint too restrictive

**Solutions**:
```bash
# Check module changelog
terraform show -json | jq '.module_calls'

# Test with specific version
terraform init -upgrade

# Review breaking changes in module docs
```

### Lock File Conflicts

**Problem**: `.terraform.lock.hcl` has conflicts after update

**Solution**:
```bash
# Delete lock file and regenerate
rm .terraform.lock.hcl
terraform init -upgrade

# Or update specific providers
terraform providers lock \
  -platform=linux_amd64 \
  -platform=darwin_amd64 \
  -platform=darwin_arm64
```

### Authentication Errors

**Problem**: "403 Forbidden" for private modules

**Causes**:
1. Not logged in to Terraform Cloud/Enterprise
2. Token expired
3. Insufficient permissions

**Solutions**:
```bash
# Login to Terraform Cloud
terraform login

# Verify credentials
cat ~/.terraform.d/credentials.tfrc.json

# Test module access
terraform init
```

## Best Practices

1. **Always regenerate lockfile**:
   ```bash
   uptool update --only=terraform
   terraform init -upgrade
   git add *.tf .terraform.lock.hcl
   git commit -m "chore(terraform): update module versions"
   ```

2. **Test in non-prod first**:
   ```bash
   cd environments/dev
   uptool update --only=terraform
   terraform init -upgrade
   terraform plan
   ```

3. **Review module changelogs**:
   ```bash
   # Check breaking changes
   # Visit module's GitHub releases or CHANGELOG
   ```

4. **Use version constraints wisely**:
   ```hcl
   # Good: Allow patch updates
   version = "~> 5.0"

   # Risky: Allow any major version
   version = ">= 3.0"

   # Safe but manual: Exact version
   version = "5.0.0"
   ```

5. **Separate PRs for major updates**:
   ```bash
   # Minor/patch updates together
   uptool update --only=terraform  # (with policy: minor)

   # Major updates separately
   # Review each module's breaking changes
   ```

6. **Run terraform plan after updating**:
   ```bash
   terraform init -upgrade
   terraform plan
   # Review changes before applying
   ```

## Terraform Version Compatibility

uptool works with Terraform >= 0.13 (HCL2 format).

**Terraform 0.12 and older**:
- Not officially supported
- May work but not tested

**Terraform >= 0.13**:
- ✅ Fully supported
- Uses HCL2 format
- Module registry syntax

**Terraform >= 1.0**:
- ✅ Fully supported
- Lockfile format supported

## See Also

- [Terraform Modules](https://developer.hashicorp.com/terraform/language/modules)
- [Module Sources](https://developer.hashicorp.com/terraform/language/modules/sources)
- [Version Constraints](https://developer.hashicorp.com/terraform/language/expressions/version-constraints)
- [Terraform Registry](https://registry.terraform.io/)
- [Manifest Files Reference](../manifests.md)

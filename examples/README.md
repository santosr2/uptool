# uptool Examples

This directory contains example configuration files for all supported integrations.

## Overview

These examples demonstrate:
- **Manifest file formats** for each integration
- **Version syntax** and constraints
- **Common patterns** and use cases
- **Before/after** update scenarios

## Example Files

### Configuration Files

- [`uptool.yaml`](uptool.yaml) - Complete uptool configuration with all integrations
- [`uptool-minimal.yaml`](uptool-minimal.yaml) - Minimal configuration example
- [`uptool-monorepo.yaml`](uptool-monorepo.yaml) - Monorepo configuration example

### Integration Manifests

| Integration | Example Files | Description |
|-------------|---------------|-------------|
| **npm** | [`package.json`](package.json) | JavaScript/TypeScript dependencies |
| **Helm** | [`Chart.yaml`](Chart.yaml) | Kubernetes Helm chart |
| **Terraform** | [`main.tf`](main.tf) | Terraform modules |
| **TFLint** | [`.tflint.hcl`](.tflint.hcl) | TFLint plugins |
| **pre-commit** | [`.pre-commit-config.yaml`](.pre-commit-config.yaml) | Pre-commit hooks |
| **asdf** | [`.tool-versions`](.tool-versions) | asdf runtime versions |
| **mise** | [`mise.toml`](mise.toml), [`.mise.toml`](.mise.toml) | mise runtime versions |

## Usage

### Copy to Your Project

```bash
# Copy uptool configuration
cp examples/uptool.yaml .

# Copy integration manifests as needed
cp examples/package.json .
cp examples/Chart.yaml .
cp examples/.pre-commit-config.yaml .
```

### Test with uptool

```bash
# Scan example manifests
cd examples
uptool scan

# Preview updates
uptool plan

# Apply updates (dry run)
uptool update --dry-run --diff
```

## Example Scenarios

### Scenario 1: Basic Project

**Files needed**:
- `uptool.yaml` (minimal)
- `package.json` (npm)
- `.pre-commit-config.yaml` (pre-commit)

**Workflow**:
1. Copy configuration files
2. Run `uptool scan` to verify detection
3. Run `uptool plan` to see available updates
4. Run `uptool update` to apply

### Scenario 2: Terraform Infrastructure

**Files needed**:
- `uptool.yaml` (with terraform, tflint)
- `main.tf` (Terraform modules)
- `.tflint.hcl` (TFLint plugins)

**Workflow**:
1. Configure conservative update policy (patch only)
2. Run `uptool plan` weekly
3. Review Terraform module changelogs
4. Apply updates in non-production first

### Scenario 3: Monorepo

**Files needed**:
- `uptool-monorepo.yaml` (multi-path configuration)
- Multiple `package.json` files in subdirectories
- Multiple `Chart.yaml` files for services

**Workflow**:
1. Configure path matching for each integration
2. Run `uptool scan` to verify all manifests detected
3. Use `--only` flag to update specific ecosystems
4. Test each service independently

## Configuration Examples

### Conservative (Production)

```yaml
# uptool.yaml
version: 1

integrations:
  - id: terraform
    enabled: true
    policy:
      update: patch       # Only patch updates
      allow_prerelease: false
      pin: true
```

### Aggressive (Development)

```yaml
# uptool.yaml
version: 1

integrations:
  - id: npm
    enabled: true
    policy:
      update: major       # All updates including major
      allow_prerelease: true
      pin: false
```

### Hybrid (Realistic)

```yaml
# uptool.yaml
version: 1

integrations:
  # Conservative for infrastructure
  - id: terraform
    enabled: true
    policy:
      update: patch

  # Moderate for application code
  - id: npm
    enabled: true
    policy:
      update: minor

  # Aggressive for dev tools
  - id: precommit
    enabled: true
    policy:
      update: major
```

## Version Constraint Examples

### npm (package.json)

```json
{
  "dependencies": {
    "react": "^18.0.0",      // Minor updates (18.x.x)
    "express": "~4.18.0",    // Patch updates (4.18.x)
    "lodash": "4.17.21"      // Exact version
  }
}
```

### Terraform (main.tf)

```hcl
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"         # Pessimistic constraint
}

module "rds" {
  source  = "terraform-aws-modules/rds/aws"
  version = ">= 6.0"         # Greater than or equal
}
```

### mise (mise.toml)

```toml
[tools]
go = "1.23"                  # Partial version (latest 1.23.x)
node = { version = "20", path = ".nvmrc" }  # Map format
```

## Testing Examples

To test these examples:

```bash
# Clone repository
git clone https://github.com/santosr2/uptool
cd uptool/examples

# Run scan
uptool scan

# Expected output:
# Type                 Path                Dependencies
# ----------------------------------------------------------------
# npm                  package.json        5
# helm                 Chart.yaml          3
# terraform            main.tf             2
# tflint               .tflint.hcl         3
# precommit            .pre-commit-config.yaml  4
# asdf                 .tool-versions      6
# mise                 mise.toml           5
#
# Total: 7 manifests

# Run plan (will show available updates)
uptool plan

# Run update (dry run)
uptool update --dry-run --diff
```

## See Also

- [Main Documentation](../docs/README.md)
- [Configuration Reference](../docs/configuration.md)
- [Integration Guides](../docs/integrations/README.md)
- [GitHub Action Usage](../docs/action-usage.md)

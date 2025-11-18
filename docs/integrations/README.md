# Integration Guides

This directory contains detailed guides for each uptool integration.

## Available Integrations

### Package Managers

- **[npm](npm.md)** - JavaScript/Node.js dependencies (`package.json`)
  - Status: ✅ Stable
  - Manifests: `package.json`
  - Registry: npm Registry

### Infrastructure as Code

- **helm** - Kubernetes Helm charts (`Chart.yaml`)
  - Status: ✅ Stable
  - Manifests: `Chart.yaml`
  - Registry: Helm chart repositories

- **terraform** - Terraform modules (`*.tf`)
  - Status: ✅ Stable
  - Manifests: `*.tf` files
  - Registry: Terraform Registry

- **tflint** - Terraform linter plugins (`.tflint.hcl`)
  - Status: ✅ Stable
  - Manifests: `.tflint.hcl`
  - Registry: GitHub Releases

### Development Tools

- **precommit** - Pre-commit hooks (`.pre-commit-config.yaml`)
  - Status: ✅ Stable
  - Manifests: `.pre-commit-config.yaml`
  - Registry: GitHub Releases
  - Note: Uses native `pre-commit autoupdate`

- **asdf** - asdf version manager (`.tool-versions`)
  - Status: ✅ Stable
  - Manifests: `.tool-versions`
  - Registry: GitHub Releases (per tool)

- **mise** - mise version manager (`mise.toml`)
  - Status: ✅ Stable
  - Manifests: `mise.toml`, `.mise.toml`
  - Registry: GitHub Releases (per tool)

## Integration Categories

### By Update Strategy

**Custom Rewriting** (uptool parses and rewrites manifest):

- npm (`package.json`)
- helm (`Chart.yaml`)
- terraform (`*.tf`)
- tflint (`.tflint.hcl`)
- asdf (`.tool-versions`)
- mise (`mise.toml`)

**Native Command** (integration calls native tool):

- precommit (`pre-commit autoupdate`)

### By Registry Type

**HTTP APIs**:

- npm → npm Registry API
- terraform → Terraform Registry API

**Repository Indexes**:

- helm → Helm chart repository `index.yaml`

**GitHub Releases**:

- precommit → Hook repository releases
- tflint → Plugin repository releases
- asdf → Tool repository releases (via plugin mapping)
- mise → Tool repository releases

## Quick Reference

| Integration | Manifest | Update Strategy | Registry |
|-------------|----------|-----------------|----------|
| npm | `package.json` | Custom rewrite | npm API |
| helm | `Chart.yaml` | Custom rewrite | Helm repos |
| terraform | `*.tf` | Custom rewrite | Terraform API |
| tflint | `.tflint.hcl` | Custom rewrite | GitHub |
| precommit | `.pre-commit-config.yaml` | Native command | GitHub |
| asdf | `.tool-versions` | Custom rewrite | GitHub |
| mise | `mise.toml` | Custom rewrite | GitHub |

## Common Patterns

### Scan for Specific Integration

```bash
uptool scan --only=npm
uptool scan --only=terraform,tflint
```

### Update Single Integration

```bash
uptool update --only=helm --diff
```

### Exclude Specific Integrations

```bash
uptool update --exclude=precommit
```

### Integration Priority

When using multiple integrations, they run in parallel:

```bash
uptool update  # All integrations run concurrently
```

To control execution:

```yaml
# uptool.yaml
version: 1

integrations:
  - id: npm
    enabled: true

  - id: terraform
    enabled: false    # Disabled
```

## See Also

- [Manifest Files Reference](../manifests.md)
- [Configuration Guide](../configuration.md)
- [Main README](../overview.md#supported-integrations)

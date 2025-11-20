# Integration Guides

Detailed guides for each uptool integration.

## Available Integrations

| Integration | Manifest | Status | Registry |
|-------------|----------|--------|----------|
| **[npm](npm.md)** | `package.json` | ✅ Stable | npm Registry API |
| **[helm](helm.md)** | `Chart.yaml` | ✅ Stable | Helm chart repositories |
| **[terraform](terraform.md)** | `*.tf` | ✅ Stable | Terraform Registry API |
| **[tflint](tflint.md)** | `.tflint.hcl` | ✅ Stable | GitHub Releases |
| **[precommit](precommit.md)** | `.pre-commit-config.yaml` | ✅ Stable | GitHub Releases |
| **[asdf](asdf.md)** | `.tool-versions` | ⚠️ Experimental | GitHub Releases |
| **[mise](mise.md)** | `mise.toml` | ⚠️ Experimental | GitHub Releases |

## By Category

### Package Managers

- **[npm](npm.md)** - JavaScript/Node.js dependencies

### Infrastructure as Code

- **[helm](helm.md)** - Kubernetes package manager
- **[terraform](terraform.md)** - Terraform modules
- **[tflint](tflint.md)** - Terraform linter plugins

### Development Tools

- **[precommit](precommit.md)** - Pre-commit hooks (uses native `pre-commit autoupdate`)
- **[asdf](asdf.md)** - asdf version manager
- **[mise](mise.md)** - mise version manager (modern asdf alternative)

## Common Patterns

### Scan Specific Integration

```bash
uptool scan --only=npm
uptool scan --only=terraform,tflint
```

### Update Single Integration

```bash
uptool update --only=helm --diff
```

### Exclude Integrations

```bash
uptool update --exclude=precommit,terraform
```

## Configuration

Control integrations via `uptool.yaml`:

```yaml
version: 1

integrations:
  - id: npm
    enabled: true
    policy:
      update: minor

  - id: terraform
    enabled: false
```

See [Configuration Guide](../configuration.md) for complete options.

## See Also

- [CLI Reference](../cli/commands.md) - Command documentation
- [Configuration](../configuration.md) - Policy settings
- [Template](../INTEGRATION_TEMPLATE.md) - Template for new integrations

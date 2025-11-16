# uptool Documentation

Welcome to the uptool documentation. uptool is a universal, manifest-first dependency updater for DevOps tools and ecosystems.

## Quick Links

### Getting Started
- [Main README](../README.md) - Project overview, installation, and quick start
- [Configuration Guide](configuration.md) - How to configure uptool for your project
- [Manifest Files Reference](manifests.md) - Understanding uptool's manifest-first philosophy

### Integration Guides
- [Integration Overview](integrations/README.md) - All supported integrations
- [npm](integrations/npm.md) - JavaScript/TypeScript package.json
- [Helm](integrations/helm.md) - Kubernetes Chart.yaml
- [Terraform](integrations/terraform.md) - Terraform modules (*.tf)
- [TFLint](integrations/tflint.md) - TFLint plugins (.tflint.hcl)
- [pre-commit](integrations/precommit.md) - Pre-commit hooks (.pre-commit-config.yaml)
- [asdf](integrations/asdf.md) - asdf .tool-versions
- [mise](integrations/mise.md) - mise mise.toml/.mise.toml

### Advanced Topics
- [GitHub Action Usage](action-usage.md) - Using uptool in GitHub Actions workflows
- [Plugin Development](plugin-development.md) - Create external plugins for custom integrations
- [Versioning & Releases](versioning.md) - Mutable tags, semantic versioning, and release process
- [GitHub Environments](environments.md) - Approval gates and deployment environments

### Maintainers
- [Patch Release Workflow](patch-release-workflow.md) - Managing security patches and backports for previous versions

### Contributing
- [Contributing Guide](../CONTRIBUTING.md) - How to contribute to uptool
- [Code of Conduct](../CODE_OF_CONDUCT.md) - Community guidelines
- [Governance](../GOVERNANCE.md) - Project governance and decision-making
- [Security Policy](../SECURITY.md) - Reporting vulnerabilities

## Documentation Structure

```
docs/
├── README.md                     # This file - Documentation index
├── action-usage.md               # GitHub Action usage guide
├── configuration.md              # Configuration file reference
├── environments.md               # GitHub Environments setup
├── manifests.md                  # Manifest-first philosophy
├── patch-release-workflow.md     # Patch release and backport workflow
├── versioning.md                 # Versioning and release process
├── integrations/                 # Integration-specific guides
│   ├── README.md                 # Integration overview
│   ├── npm.md                    # npm integration
│   ├── helm.md                   # Helm integration
│   ├── terraform.md              # Terraform integration
│   ├── tflint.md                 # TFLint integration
│   ├── precommit.md              # pre-commit integration
│   ├── asdf.md                   # asdf integration
│   └── mise.md                   # mise integration
└── _templates/                   # GitHub Pages HTML templates
    ├── README.md                 # Template documentation
    └── index.html                # Landing page template
```

## Core Concepts

### Manifest-First Philosophy

uptool follows a **manifest-first** approach:

1. **Manifests are source of truth** - Always update manifest files (package.json, Chart.yaml, etc.)
2. **Lockfiles are derived** - Regenerate lockfiles after manifest updates
3. **Native commands when available** - Use native tools when they update manifests (e.g., `pre-commit autoupdate`)

See [Manifest Files Reference](manifests.md) for detailed explanation.

### Supported Ecosystems

| Ecosystem | Manifest Files | Registry | Documentation |
|-----------|----------------|----------|---------------|
| npm | `package.json` | npm registry | [npm.md](integrations/npm.md) |
| Helm | `Chart.yaml` | Artifact Hub | [helm.md](integrations/helm.md) |
| Terraform | `*.tf` | Terraform Registry | [terraform.md](integrations/terraform.md) |
| TFLint | `.tflint.hcl` | GitHub Releases | [tflint.md](integrations/tflint.md) |
| pre-commit | `.pre-commit-config.yaml` | GitHub | [precommit.md](integrations/precommit.md) |
| asdf | `.tool-versions` | GitHub Releases | [asdf.md](integrations/asdf.md) |
| mise | `mise.toml`, `.mise.toml` | GitHub Releases | [mise.md](integrations/mise.md) |

### Update Strategies

uptool uses different update strategies per integration:

- **Custom rewriting**: npm, Helm, Terraform, TFLint, asdf, mise
- **Native commands**: pre-commit (`pre-commit autoupdate`)

See individual integration guides for details.

## Common Use Cases

### Local Development

```bash
# Scan repository for manifests
uptool scan

# Preview available updates
uptool plan

# Apply updates (dry run first)
uptool update --dry-run --diff
uptool update
```

### GitHub Actions Workflow

```yaml
name: Update Dependencies
on:
  schedule:
    - cron: '0 0 * * 1'  # Weekly on Monday

jobs:
  update:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: santosr2/uptool@v0
        with:
          command: update
          create-pr: true
```

See [GitHub Action Usage](action-usage.md) for complete examples.

### CI/CD Integration

```yaml
# .github/workflows/ci.yml
- name: Check for outdated dependencies
  uses: santosr2/uptool@v0
  with:
    command: plan
    fail-on-updates: true
```

## Configuration Examples

### Basic Configuration

```yaml
# uptool.yaml
version: 1

integrations:
  - id: npm
    enabled: true
    policy:
      update: minor
      allow_prerelease: false
```

### Advanced Configuration

```yaml
# uptool.yaml
version: 1

integrations:
  - id: terraform
    enabled: true
    match:
      files:
        - "*.tf"
        - "modules/**/*.tf"
    policy:
      update: patch        # Conservative for IaC
      allow_prerelease: false
      pin: true
      cadence: monthly
```

See [Configuration Guide](configuration.md) for all options.

## Troubleshooting

### Common Issues

**"No manifests found"**
- Ensure you're in the repository root
- Check integration is enabled in `uptool.yaml`
- Verify manifest files match expected patterns

**"Rate limit exceeded"**
- Set `GITHUB_TOKEN` environment variable
- Use authenticated API access

**"Lockfile out of sync"**
- Regenerate lockfile after manifest updates:
  - npm: `npm install`
  - Helm: `helm dependency update`
  - Terraform: `terraform init -upgrade`

See individual integration guides for specific troubleshooting.

## API Documentation

Generated Go package documentation is available at:
- **GitHub Pages**: https://santosr2.github.io/uptool/api/
- **pkg.go.dev**: https://pkg.go.dev/github.com/santosr2/uptool

## Support

- **GitHub Issues**: https://github.com/santosr2/uptool/issues
- **Discussions**: https://github.com/santosr2/uptool/discussions
- **Security**: See [SECURITY.md](../SECURITY.md) for vulnerability reporting

## License

uptool is licensed under the Apache License 2.0. See [LICENSE](../LICENSE) for details.

---

**Navigation**: [← Back to README](../README.md) | [Integration Guides →](integrations/README.md)

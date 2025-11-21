# uptool

[![Go Version](https://img.shields.io/github/go-mod/go-version/santosr2/uptool)](https://go.dev/)
[![License](https://img.shields.io/github/license/santosr2/uptool)](LICENSE)
[![Latest Release](https://img.shields.io/github/v/release/santosr2/uptool?include_prereleases)](https://github.com/santosr2/uptool/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/santosr2/uptool)](https://goreportcard.com/report/github.com/santosr2/uptool)

[![CI Status](https://github.com/santosr2/uptool/workflows/CI/badge.svg)](https://github.com/santosr2/uptool/actions/workflows/ci.yml)
[![Documentation](https://github.com/santosr2/uptool/workflows/Documentation/badge.svg)](https://github.com/santosr2/uptool/actions/workflows/docs-deploy.yml)
[![CodeQL](https://github.com/santosr2/uptool/workflows/CodeQL/badge.svg)](https://github.com/santosr2/uptool/actions/workflows/codeql.yml)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/santosr2/uptool/badge)](https://scorecard.dev/viewer/?uri=github.com/santosr2/uptool)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/9999/badge)](https://www.bestpractices.dev/projects/9999)
[![codecov](https://codecov.io/gh/santosr2/uptool/branch/main/graph/badge.svg)](https://codecov.io/gh/santosr2/uptool)

[![pre-commit](https://img.shields.io/badge/pre--commit-enabled-brightgreen?logo=pre-commit)](https://github.com/pre-commit/pre-commit)
[![Security Policy](https://img.shields.io/badge/security-policy-blue)](SECURITY.md)
[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg)](https://conventionalcommits.org)

**Universal, Manifest-First Dependency Updater**

`uptool` combines the ecosystem breadth of [Topgrade](https://github.com/topgrade-rs/topgrade), the precision of [Dependabot](https://github.com/dependabot/dependabot-core), and the flexibility of [Renovate](https://github.com/renovatebot/renovate) — but with a **manifest-first philosophy** that works across ANY project toolchain defined in configuration files.

Unlike traditional dependency updaters that focus on lockfiles, uptool updates **manifest files directly** (package.json, Chart.yaml, .tf files, etc.), preserving your intent while keeping dependencies current.

---

## Why uptool?

Modern projects span multiple ecosystems—npm, Helm, Terraform, pre-commit, runtime managers. Each has its own update mechanism. uptool provides a **unified interface** to scan, plan, and update all dependencies from one tool.

### Manifest-First Philosophy

1. Update **manifests** (package.json, Chart.yaml, *.tf) first
2. Use native commands only when they update manifests
3. Then run lockfile updates (npm install, terraform init)

This ensures **declared dependencies** stay current, not just resolved versions.

---

## Features

- **Multi-Ecosystem Support**: npm, Helm, Terraform, tflint, pre-commit, asdf, mise — all in one tool
- **Manifest-First Updates**: Updates configuration files directly, preserving formatting and comments
- **Dual Usage Modes**: Use as a CLI tool locally or as a GitHub Action in CI/CD
- **Intelligent Version Resolution**: Queries upstream registries (npm, Terraform Registry, Helm repos, GitHub Releases)
- **Safe by Default**: Dry-run mode, diff generation, validation
- **Concurrent Execution**: Parallel scanning and planning with worker pools
- **Flexible Filtering**: Run specific integrations with `--only` or exclude with `--exclude`
- **Automated Versioning**: Commit-based semantic versioning with GitHub Actions
- **Clean Integration Interface**: Easy to add support for new ecosystems

---

## Supported Integrations

| Integration | Status | Manifest Files | Update Strategy | Registry |
|-------------|--------|----------------|-----------------|----------|
| **npm** | ✅ Stable | `package.json` | Custom JSON rewriting | npm Registry API |
| **Helm** | ✅ Stable | `Chart.yaml` | YAML rewriting | Helm chart repositories |
| **pre-commit** | ✅ Stable | `.pre-commit-config.yaml` | Native `pre-commit autoupdate` | GitHub Releases |
| **Terraform** | ✅ Stable | `*.tf` | HCL parsing/rewriting | Terraform Registry API |
| **tflint** | ✅ Stable | `.tflint.hcl` | HCL parsing/rewriting | GitHub Releases |
| **asdf** | ⚠️ Experimental | `.tool-versions` | Detection only (updates not implemented) | GitHub Releases (per tool) |
| **mise** | ⚠️ Experimental | `mise.toml`, `.mise.toml` | Detection only (updates not implemented) | GitHub Releases (per tool) |

### Roadmap

- [ ] Python (`pyproject.toml`, `requirements.txt`, `Pipfile`)
- [ ] Go modules (`go.mod`)
- [ ] Docker (`Dockerfile`, `compose.yml`, `docker-compose.yml`)
- [ ] GitHub Actions (workflow `.yml` files)
- [ ] Generic version matcher (custom YAML/TOML/JSON/HCL patterns)

---

## Quick Start

### Installation

#### Docker (Recommended)

```bash
# Pull the latest stable image
docker pull ghcr.io/santosr2/uptool:latest

# Run uptool (use an alias for convenience)
alias uptool='docker run --rm -v "$PWD:/workspace" ghcr.io/santosr2/uptool'

# Verify installation
uptool version
```

#### Pre-built Binaries

```bash
# Download the latest release for your platform
# Linux (AMD64)
curl -LO https://github.com/santosr2/uptool/releases/latest/download/uptool-linux-amd64
chmod +x uptool-linux-amd64
sudo mv uptool-linux-amd64 /usr/local/bin/uptool

# macOS (Apple Silicon)
curl -LO https://github.com/santosr2/uptool/releases/latest/download/uptool-darwin-arm64
chmod +x uptool-darwin-arm64
sudo mv uptool-darwin-arm64 /usr/local/bin/uptool

# Verify installation
uptool version
```

#### Go Install

```bash
# Install from source (requires Go 1.25+)
go install github.com/santosr2/uptool/cmd/uptool@latest
```

#### Build from Source

```bash
git clone https://github.com/santosr2/uptool.git
cd uptool
mise run build  # Binary will be in dist/uptool
```

### CLI Usage

```bash
# 1. Scan your repository for updateable dependencies
$ uptool scan

Type                 Path                                   Dependencies
--------------------------------------------------------------------------------
npm                  package.json                           4
helm                 charts/app/Chart.yaml                  3
terraform            infra/terraform                        2
precommit            .pre-commit-config.yaml                2
mise                 mise.toml                              7

Total: 5 manifests

# 2. Generate an update plan
$ uptool plan

package.json (npm):
Package          Current         Target          Impact
--------------------------------------------------------
express          ^4.18.0         ^4.19.2         minor
lodash           ^4.17.20        ^4.17.21        patch

charts/app/Chart.yaml (helm):
Package          Current         Target          Impact
--------------------------------------------------------
postgresql       12.0.0          18.1.8          major
redis            17.0.0          23.2.12         major

mise.toml (mise):
Tool             Current         Target          Impact
--------------------------------------------------------
go               1.23            1.25            minor
node             20              22              major

Total: 7 updates across 3 manifests

# 3. Preview changes with dry-run
$ uptool update --dry-run --diff

# 4. Apply updates
$ uptool update --diff

# 5. Run specific integrations
$ uptool update --only=npm,helm

# 6. Exclude integrations
$ uptool update --exclude=terraform
```

### GitHub Action Usage

Create `.github/workflows/dependency-updates.yml`:

```yaml
name: Dependency Updates

on:
  schedule:
    - cron: '0 9 * * 1'  # Every Monday at 9 AM UTC
  workflow_dispatch:      # Manual trigger

permissions:
  contents: write
  pull-requests: write

jobs:
  update:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Update dependencies
        uses: santosr2/uptool@v0  # Pin to major version (recommended)
        with:
          command: update
          create-pr: 'true'
          pr-title: 'chore(deps): update dependencies'
          pr-branch: 'uptool/updates-${{ github.run_number }}'
          token: ${{ secrets.GITHUB_TOKEN }}
```

---

## Commands

**Global flags**: `-v/--verbose`, `-q/--quiet`, `--help`

| Command | Purpose | Key Flags |
|---------|---------|-----------|
| `uptool scan` | Discover manifest files | `--only`, `--exclude`, `--format` |
| `uptool plan` | Generate update plan | `--only`, `--exclude`, `--out` |
| `uptool update` | Apply updates | `--dry-run`, `--diff`, `--only` |
| `uptool list` | List integrations | `--category`, `--experimental` |

See [CLI Reference](docs/cli/commands.md) for complete documentation.

---

## GitHub Action

See [docs/action-usage.md](docs/action-usage.md) for comprehensive GitHub Action documentation.

### Version Pinning

uptool supports **mutable tags** for convenient version pinning:

```yaml
# Recommended: Pin to major version (gets latest minor/patch updates)
- uses: santosr2/uptool@v0

# Pin to minor version (gets latest patch updates only)
- uses: santosr2/uptool@v0.1

# Pin to exact version (no automatic updates)
- uses: santosr2/uptool@v0.1.0
```

**How it works**:

- `@v0` - Automatically updated to latest `v0.x.x` release
- `@v0.1` - Automatically updated to latest `v0.1.x` patch
- `@v0.1.0` - Fixed to exact release (never changes)

**Pre-release tags** (for testing):

- `@v0-rc` - Latest release candidate in v0
- `@v0.1-rc` - Latest release candidate in v0.1
- `@v0.1.0-rc.1` - Exact pre-release (immutable)

### Inputs

| Input | Description | Default | Required |
|-------|-------------|---------|----------|
| `command` | Command to run: `scan`, `plan`, or `update` | `plan` | No |
| `format` | Output format: `table` or `json` | `table` | No |
| `only` | Comma-separated integrations to include | `''` | No |
| `exclude` | Comma-separated integrations to exclude | `''` | No |
| `dry-run` | Show changes without applying (update only) | `false` | No |
| `diff` | Show diffs of changes (update only) | `true` | No |
| `create-pr` | Create a pull request with updates | `false` | No |
| `pr-title` | Title for the pull request | `chore: update dependencies` | No |
| `pr-branch` | Branch name for the PR | `uptool/dependency-updates` | No |
| `token` | GitHub token for API access and PR creation | `${{ github.token }}` | No |

### Outputs

| Output | Description |
|--------|-------------|
| `updates-available` | Whether updates were found (`true`/`false`) |
| `manifests-updated` | Number of manifests with updates applied |
| `dependencies-updated` | Total number of dependencies updated |

### Required Permissions

```yaml
permissions:
  contents: write          # To push commits
  pull-requests: write     # To create PRs
```

---

## Integration Details

See [docs/integrations/](docs/integrations/) for detailed guides.

**Quick examples**:

- **npm**: Updates `package.json`, preserves constraints (`^`, `~`)
- **Helm**: Updates `Chart.yaml` dependencies
- **Terraform**: Updates module versions in `*.tf` files
- **tflint**: Updates plugin versions in `.tflint.hcl`
- **pre-commit**: Uses native `pre-commit autoupdate`
- **asdf/mise**: Updates runtime tool versions (experimental)

---

## Architecture

uptool uses a clean integration interface:

```go
type Integration interface {
    Detect(ctx context.Context, repoRoot string) ([]*Manifest, error)
    Plan(ctx context.Context, manifest *Manifest) (*UpdatePlan, error)
    Apply(ctx context.Context, plan *UpdatePlan) (*ApplyResult, error)
}
```

**Workflow**: Detect manifests → Query registries → Update files → Validate

See [docs/architecture.md](docs/architecture.md) for complete system design.

---

## Configuration

Optional `uptool.yaml` in your repository root:

```yaml
version: 1
integrations:
  - id: npm
    enabled: true
    policy:
      update: minor              # none, patch, minor, major
      allow_prerelease: false
```

See [docs/configuration.md](docs/configuration.md) and [`examples/`](examples/) for complete reference.

---

## Version Management

Automated semantic versioning via conventional commits:

- `feat:` → minor bump (0.1.0 → 0.2.0)
- `fix:` → patch bump (0.1.0 → 0.1.1)
- `BREAKING CHANGE:` → major bump (0.1.0 → 1.0.0)

GitHub Actions handle releases with approval gates. See [docs/versioning.md](docs/versioning.md).

---

## Development

**Prerequisites**: Go 1.25+, [mise](https://mise.jdx.dev/)

```bash
# Build
mise run build

# Test
mise run test

# Quality checks
mise run check  # fmt + vet + lint + test
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines and [.vscode/README.md](.vscode/README.md) for VS Code setup.

---

## Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

**Quick start**: Fork → Create branch → Add tests → Run `mise run check` → Open PR

Use conventional commits: `feat:`, `fix:`, `docs:`, `chore:`

---

## Security & Governance

- **Security**: Report vulnerabilities via [GitHub Security Advisories](https://github.com/santosr2/uptool/security/advisories), not public issues
- **Governance**: Trunk-based development. See [GOVERNANCE.md](GOVERNANCE.md)

---

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.

---

## Documentation & Support

**Docs**: [Overview](docs/overview.md) • [Quick Start](docs/quickstart.md) • [Configuration](docs/configuration.md) • [Integrations](docs/integrations/) • [Examples](examples/)

**Community**: [Issues](https://github.com/santosr2/uptool/issues) • [Discussions](https://github.com/santosr2/uptool/discussions) • [Changelog](CHANGELOG.md)

Built with Go, inspired by Topgrade, Dependabot, and Renovate.

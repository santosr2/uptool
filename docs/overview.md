# uptool

[![Go Version](https://img.shields.io/github/go-mod/go-version/santosr2/uptool)](https://go.dev/)
[![License](https://img.shields.io/github/license/santosr2/uptool)](LICENSE.md)
[![Latest Release](https://img.shields.io/github/v/release/santosr2/uptool?include_prereleases)](https://github.com/santosr2/uptool/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/santosr2/uptool)](https://goreportcard.com/report/github.com/santosr2/uptool)

[![CI Status](https://github.com/santosr2/uptool/workflows/CI/badge.svg)](https://github.com/santosr2/uptool/actions/workflows/ci.yml)
[![CodeQL](https://github.com/santosr2/uptool/workflows/CodeQL/badge.svg)](https://github.com/santosr2/uptool/actions/workflows/codeql.yml)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/santosr2/uptool/badge)](https://scorecard.dev/viewer/?uri=github.com/santosr2/uptool)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/9999/badge)](https://www.bestpractices.dev/projects/9999)
[![codecov](https://codecov.io/gh/santosr2/uptool/branch/main/graph/badge.svg)](https://codecov.io/gh/santosr2/uptool)

[![pre-commit](https://img.shields.io/badge/pre--commit-enabled-brightgreen?logo=pre-commit)](https://github.com/pre-commit/pre-commit)
[![Security Policy](https://img.shields.io/badge/security-policy-blue)](https://github.com/santosr2/uptool/blob/{{ extra.uptool_version }}/SECURITY.md)
[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg)](https://conventionalcommits.org)

**Universal, Manifest-First Dependency Updater**

`uptool` combines the ecosystem breadth of [Topgrade](https://github.com/topgrade-rs/topgrade), the precision of [Dependabot](https://github.com/dependabot/dependabot-core), and the flexibility of [Renovate](https://github.com/renovatebot/renovate) — but with a **manifest-first philosophy** that works across ANY project toolchain defined in configuration files.

Unlike traditional dependency updaters that focus on lockfiles, uptool updates **manifest files directly** (package.json, Chart.yaml, .tf files, etc.), preserving your intent while keeping dependencies current.

---

## Why uptool?

**The Problem**: Modern projects use dozens of tools across multiple ecosystems:

- Language dependencies (npm, pip, go modules)
- Infrastructure tools (Terraform, Helm)
- Development tools (pre-commit, tflint)
- Runtime version managers (asdf with `.tool-versions`, mise with `mise.toml`)

Each ecosystem has its own update mechanism. Keeping them all current is tedious and error-prone.

**The Solution**: uptool provides a **unified interface** to scan, plan, and update dependencies across all your manifest files, whether you're managing JavaScript packages, Kubernetes charts, or Terraform modules.

### Manifest-First Philosophy

uptool updates **manifests** (source of truth), not just lockfiles:

1. **Update manifests first**: package.json, Chart.yaml, *.tf files
2. **Use native commands when they update manifests**: `pre-commit autoupdate` updates .pre-commit-config.yaml ✅
3. **Don't use commands that only update lockfiles**: `npm update` only touches package-lock.json ❌
4. **Then optionally run native lockfile updates**: `npm install`, `terraform init`, etc.

This ensures your **declared dependencies** stay current, not just resolved versions.

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
| **asdf** | ✅ Stable | `.tool-versions` | Line-based parsing/rewriting | GitHub Releases (per tool) |
| **mise** | ✅ Stable | `mise.toml`, `.mise.toml` | TOML parsing/rewriting | GitHub Releases (per tool) |

### Roadmap

- [ ] Python (`pyproject.toml`, `requirements.txt`, `Pipfile`)
- [ ] Go modules (`go.mod`)
- [ ] Docker (`Dockerfile`, `compose.yml`, `docker-compose.yml`)
- [ ] GitHub Actions (workflow `.yml` files)
- [ ] Generic version matcher (custom YAML/TOML/JSON/HCL patterns)

---

## Quick Start

### Installation

```bash
# Install from source (requires Go 1.25+)
go install github.com/santosr2/uptool/cmd/uptool@latest

# Or download pre-built binaries from releases
# Linux (AMD64)
curl -LO https://github.com/santosr2/uptool/releases/download/v0.1.0/uptool-linux-amd64
chmod +x uptool-linux-amd64
sudo mv uptool-linux-amd64 /usr/local/bin/uptool

# macOS (Apple Silicon)
curl -LO https://github.com/santosr2/uptool/releases/download/v0.1.0/uptool-darwin-arm64
chmod +x uptool-darwin-arm64
sudo mv uptool-darwin-arm64 /usr/local/bin/uptool

# Or build from source
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

## CLI Reference

### Global Flags

All commands support these global flags:

```bash
-h, --help       Show help for command
-v, --verbose    Enable verbose debug output
-q, --quiet      Suppress informational output (errors only)
--version        Show uptool version
```

### `uptool scan`

Discover manifest files and extract current dependency versions.

```bash
uptool scan                      # Table output
uptool scan --format=json        # JSON output for scripting
uptool scan --only=npm,helm      # Only specific integrations
uptool scan --exclude=terraform  # Exclude integrations
```

**Flags**:

- `--format=FORMAT`: Output format (`table` or `json`)
- `--only=INTEGRATIONS`: Comma-separated list of integrations to run
- `--exclude=INTEGRATIONS`: Comma-separated list to skip

**Output**: List of manifests with dependency counts

### `uptool plan`

Query upstream registries and generate an update plan.

```bash
uptool plan                      # Show available updates
uptool plan --format=json        # JSON output
uptool plan --out=plan.json      # Save to file
uptool plan --only=npm           # Specific integrations
```

**Flags**:

- `--format=FORMAT`: Output format (`table` or `json`)
- `--out=FILE`: Save output to file
- `--only=INTEGRATIONS`: Comma-separated integrations
- `--exclude=INTEGRATIONS`: Comma-separated integrations to skip

**Output**: Update plans showing current → target versions with impact assessment

### `uptool update`

Apply updates to manifest files.

```bash
uptool update --dry-run --diff   # Preview changes without applying
uptool update --diff             # Apply with diff output
uptool update --only=npm,helm    # Update specific ecosystems
uptool update --exclude=precommit
```

**Flags**:

- `--dry-run`: Show what would change without modifying files
- `--diff`: Display unified diffs of changes
- `--only=INTEGRATIONS`: Comma-separated list of integrations to run
- `--exclude=INTEGRATIONS`: Comma-separated list to skip
- `--format=FORMAT`: Output format (`table` or `json`)

### `uptool list`

List all available integrations and their status.

```bash
uptool list                          # List all integrations
uptool list --category package-manager  # Filter by category
uptool list --experimental            # Include experimental integrations
```

**Flags**:

- `--category=CATEGORY`: Filter by category (e.g., `package-manager`, `infrastructure`, `tooling`)
- `--experimental`: Include experimental integrations

**Output**: Table showing integration ID, name, description, and status

---

## GitHub Action

See [./action-usage.md](./action-usage.md) for comprehensive GitHub Action documentation.

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

**For detailed integration guides, see [./integrations/](./integrations/README.md)**.

This section provides a quick overview. For comprehensive documentation including troubleshooting, examples, and best practices, refer to the individual integration guides.

### npm

**Files**: `package.json`
**Strategy**: Custom JSON rewriting with constraint preservation
**Registry**: npm Registry API (`https://registry.npmjs.org`)

Updates all dependency types:

- `dependencies`
- `devDependencies`
- `peerDependencies`
- `optionalDependencies`

Preserves version constraint prefixes (`^`, `~`, `>=`, etc.).

**Example**:

```json
{
  "dependencies": {
    "express": "^4.18.0",  // Updates to "^4.19.2"
    "lodash": "~4.17.20"   // Updates to "~4.17.21"
  }
}
```

### Helm

**Files**: `Chart.yaml`
**Strategy**: YAML rewriting
**Registry**: Helm chart repositories (index.yaml)

Updates chart dependencies while preserving structure and comments.

**Example**:

```yaml
dependencies:
  - name: postgresql
    version: 12.0.0  # Updates to 18.1.8
    repository: https://charts.bitnami.com/bitnami
```

### pre-commit

**Files**: `.pre-commit-config.yaml`
**Strategy**: Native `pre-commit autoupdate` command
**Registry**: GitHub Releases (for hook repositories)

Leverages pre-commit's built-in update mechanism since it updates the manifest directly.

**Example**:

```yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.3.0  # Updates to v6.0.0
```

### Terraform

**Files**: `*.tf`
**Strategy**: HCL parsing and rewriting
**Registry**: Terraform Registry API (`https://registry.terraform.io`)

Updates module versions in `module` blocks. Provider updates coming soon.

**Example**:

```hcl
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "3.0.0"  # Updates to "5.8.1"
}
```

### tflint

**Files**: `.tflint.hcl`
**Strategy**: HCL parsing and rewriting
**Registry**: GitHub Releases (for plugins)

Updates tflint plugin versions.

**Example**:

```hcl
plugin "aws" {
  enabled = true
  version = "0.21.0"  # Updates to "0.44.0"
  source  = "github.com/terraform-linters/tflint-ruleset-aws"
}
```

### asdf

**Files**: `.tool-versions`
**Strategy**: Line-based parsing and rewriting
**Registry**: GitHub Releases (per tool via asdf plugin mapping)

Updates tool versions managed by asdf. Currently supports detection and parsing. Version resolution and updates in active development.

**Example**:

```text
go 1.23.0        # Updates to 1.25.0
nodejs 20.10.0   # Updates to 22.0.0
terraform 1.5.0  # Updates to 1.10.5
```

### mise

**Files**: `mise.toml`, `.mise.toml`
**Strategy**: TOML parsing and rewriting
**Registry**: GitHub Releases (per tool)

Updates tool versions managed by mise (formerly rtx). Supports both string format (`go = "1.23"`) and map format (`go = { version = "1.23" }`).

**Example**:

```toml
[tools]
go = "1.23"              # Updates to "1.25"
node = "20"              # Updates to "22"
golangci-lint = "2.6"    # Updates to "2.7"
```

---

## Architecture

### Project Structure

```tree
uptool/
├── cmd/uptool/              # CLI entry point
│   ├── main.go              # Main entry and integration registration
│   └── cmd/                 # Cobra command handlers
│       ├── root.go          # Root command and global flags
│       ├── scan.go          # Scan command
│       ├── plan.go          # Plan command
│       └── update.go        # Update command
├── internal/
│   ├── version/             # Version embedding
│   │   ├── VERSION          # Single source of truth for version
│   │   └── version.go       # Version package
│   ├── engine/              # Core orchestration
│   │   ├── types.go         # Core types (Manifest, UpdatePlan, etc.)
│   │   └── engine.go        # Scan/Plan/Update orchestration
│   ├── registry/            # Registry API clients
│   │   ├── npm.go           # npm Registry client
│   │   ├── terraform.go     # Terraform Registry client
│   │   ├── github.go        # GitHub Releases client
│   │   └── helm.go          # Helm repository client
│   ├── datasource/          # Version data sources
│   ├── policy/              # Update policy engine
│   ├── rewrite/             # Manifest rewriting
│   ├── resolve/             # Version resolution
│   └── integrations/        # Ecosystem integrations
│       ├── npm/             # npm package.json
│       ├── helm/            # Helm Chart.yaml
│       ├── precommit/       # pre-commit hooks
│       ├── terraform/       # Terraform modules
│       ├── tflint/          # tflint plugins
│       ├── asdf/            # asdf .tool-versions
│       └── mise/            # mise mise.toml
├── .github/
│   ├── workflows/           # GitHub Actions CI/CD
│   │   ├── ci.yml           # Continuous Integration
│   │   ├── pre-release.yml  # Automated pre-release creation
│   │   └── promote-release.yml  # Stable release promotion
│   └── actions/             # Reusable actions
│       └── setup-mise/      # mise setup action
├── testdata/                # Test fixtures for each integration
├── examples/                # Example configuration files
├── ./                    # Detailed documentation
├── action.yml               # GitHub Action definition
├── mise.toml                # Development tool versions
└── README.md                # This file
```

### Integration Interface

All integrations implement the same interface:

```go
type Integration interface {
    Name() string
    Detect(ctx context.Context, repoRoot string) ([]*Manifest, error)
    Plan(ctx context.Context, manifest *Manifest) (*UpdatePlan, error)
    Apply(ctx context.Context, plan *UpdatePlan) (*ApplyResult, error)
    Validate(ctx context.Context, manifest *Manifest) error
}
```

**Workflow**:

1. **Detect**: Scan repository tree for manifest files
2. **Plan**: Query registries for available updates
3. **Apply**: Rewrite manifests with new versions
4. **Validate**: Check syntax and optionally run tool-specific validation

### Concurrent Execution

uptool uses Go's concurrency primitives for performance:

- Concurrent scanning across integrations
- Parallel planning with semaphore-controlled worker pools
- Atomic file updates with diffs

---

## Configuration

uptool supports optional configuration via a `uptool.yaml` file in your repository root. This allows you to control which integrations run and customize update policies per integration.

See [./configuration.md](./configuration.md) for complete configuration reference.

### Quick Configuration Example

Create a `uptool.yaml` file:

```yaml
version: 1

integrations:
  - id: npm
    enabled: true
    policy:
      update: minor              # none, patch, minor, major
      allow_prerelease: false
      pin: true

  - id: helm
    enabled: true
    policy:
      update: minor
      allow_prerelease: false

  - id: terraform
    enabled: true
    policy:
      update: minor
      allow_prerelease: false

  - id: precommit
    enabled: true
    policy:
      update: minor
      allow_prerelease: false

  - id: tflint
    enabled: false             # Disable specific integrations
    policy:
      update: none
```

### Behavior

- If `uptool.yaml` exists, only enabled integrations will run
- If no config file exists, all integrations run by default
- CLI flags (`--only`, `--exclude`) override configuration
- Invalid configuration will log a warning and use defaults

### Example Configuration Files

The [`examples/`](https://github.com/santosr2/uptool/blob/{{ extra.uptool_version }}/examples/) directory contains sample configuration files for various integrations:

- **[`uptool.yaml`](https://github.com/santosr2/uptool/blob/{{ extra.uptool_version }}/examples/uptool.yaml)** - Complete uptool configuration with all integrations
- **[`.tool-versions`](https://github.com/santosr2/uptool/blob/{{ extra.uptool_version }}/examples/.tool-versions)** - asdf runtime version manager configuration
- **[`mise.toml`](https://github.com/santosr2/uptool/blob/{{ extra.uptool_version }}/examples/mise.toml)** - mise tool version manager configuration (string format)
- **[`.mise.toml`](https://github.com/santosr2/uptool/blob/{{ extra.uptool_version }}/examples/.mise.toml)** - mise configuration (hidden file variant with map format)

---

## Version Management

uptool uses **automated semantic versioning** based on conventional commits. This eliminates manual version management and ensures consistent, predictable releases.

### How It Works

1. **Commits determine version bumps**:
   - `feat:` commits → minor version bump (0.1.0 → 0.2.0)
   - `fix:` commits → patch version bump (0.1.0 → 0.1.1)
   - `BREAKING CHANGE:` → major version bump (0.1.0 → 1.0.0)
   - Other types (`chore:`, `docs:`, etc.) → no version bump

2. **Pre-release workflow** (automated with approval):
   - Trigger workflow with pre-release type (rc/beta/alpha)
   - System calculates next version from commits
   - Tests run automatically
   - **Approval gate**: Designated reviewers must approve
   - Creates pre-release (e.g., `v0.2.0-rc.1`)
   - Updates VERSION file across codebase
   - Builds and publishes artifacts

3. **Stable release workflow** (automated with approval):
   - Provide pre-release tag to promote
   - System auto-detects stable version (v0.2.0-rc.1 → v0.2.0)
   - Tests run automatically
   - **Approval gate**: Multiple reviewers must approve
   - Updates VERSION file
   - Promotes artifacts
   - Updates CHANGELOG

### Local Development

```bash
# Show current version
mise run version-show

# Manually bump versions (for local testing only)
mise run version-bump-patch   # 0.1.0 → 0.1.1
mise run version-bump-minor   # 0.1.0 → 0.2.0
mise run version-bump-major   # 0.1.0 → 1.0.0
```

See [./versioning.md](./versioning.md) for complete documentation.

---

## Development

### Prerequisites

- **Go 1.25** or later
- Git
- [mise](https://mise.jdx.dev/) (for task runner and tool management)
- **VS Code** (recommended) - See [.vscode/README.md](https://github.com/santosr2/uptool/blob/{{ extra.uptool_version }}/.vscode/README.md) for setup

### Available Tasks

All development tasks are managed through mise. To see all available tasks:

```bash
mise tasks ls
```

Common tasks:

- `mise run build` - Build the binary
- `mise run test` - Run tests
- `mise run check` - Run all quality checks
- `mise run fmt` - Format code
- `mise run lint` - Run linter
- `mise run clean` - Clean build artifacts

See `mise.toml` for the complete list of tasks.

### Build from Source

```bash
git clone https://github.com/santosr2/uptool.git
cd uptool

# Install tools and build (recommended)
mise install
mise run build

# Built binary will be in dist/uptool
```

### Run Tests

```bash
# Run all tests
mise run test

# Run with coverage
mise run test-coverage

# Run specific package tests
go test ./internal/integrations/npm/...

# Run with race detector
go test -race ./...
```

### Code Quality

```bash
# Format code
mise run fmt

# Run linter
mise run lint

# Run all checks (fmt + vet + complexity + lint + test)
mise run check
```

### Run Locally on This Repository

```bash
# Build
mise run build

# Scan this repository
mise run run-scan

# Plan updates
mise run run-plan

# Dry-run
mise run run-update
```

### Adding a New Integration

See [CONTRIBUTING.md](https://github.com/santosr2/uptool/blob/{{ extra.uptool_version }}/CONTRIBUTING.md) for detailed instructions.

Quick overview:

1. Create `internal/integrations/<name>/<name>.go`
2. Implement the `engine.Integration` interface
3. Add registry client in `internal/registry/<name>.go` if needed
4. Register in `internal/integrations/registry.go`
5. Add test fixtures in `testdata/<name>/`
6. Add integration tests in `internal/integrations/<name>/<name>_test.go`
7. Update documentation (README, ./integrations/<name>.md)
8. Add example configuration in `examples/`

---

## Contributing

We welcome contributions! Please read [CONTRIBUTING.md](https://github.com/santosr2/uptool/blob/{{ extra.uptool_version }}/CONTRIBUTING.md) for:

- Development setup
- Coding standards
- Testing requirements
- Trunk-based workflow
- Conventional commit guidelines
- Version management process

**Quick Start**:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes with tests
4. Run `mise run check` to validate
5. Commit using conventional commits (`git commit -m 'feat: add amazing feature'`)
6. Push (`git push origin feature/amazing-feature`)
7. Open a Pull Request

---

## Security

For security concerns, please see [SECURITY.md](https://github.com/santosr2/uptool/blob/{{ extra.uptool_version }}/SECURITY.md).

**tl;dr**: Report vulnerabilities via GitHub Security Advisories, not public issues.

---

## Governance

See [GOVERNANCE.md](https://github.com/santosr2/uptool/blob/{{ extra.uptool_version }}/GOVERNANCE.md) for:

- Maintainer responsibilities
- Decision-making process
- PR review expectations

**Workflow**: Trunk-based development (no Git Flow). All changes merge directly to `main` after review.

---

## License

This project is licensed under the MIT License. See [LICENSE](https://github.com/santosr2/uptool/blob/{{ extra.uptool_version }}/LICENSE) for details.

---

## Acknowledgments

- Inspired by [Topgrade](https://github.com/topgrade-rs/topgrade), [Dependabot](https://github.com/dependabot/dependabot-core), and [Renovate](https://github.com/renovatebot/renovate)
- Built with ❤️ in Go
- Uses excellent libraries:
  - [semver](https://github.com/Masterminds/semver) for version comparison
  - [yaml.v3](https://gopkg.in/yaml.v3) for YAML parsing
  - [HCL](https://github.com/hashicorp/hcl) for Terraform/tflint parsing
  - [Cobra](https://github.com/spf13/cobra) for CLI interface
  - [go-toml](https://github.com/pelletier/go-toml) for TOML parsing

---

## Support & Community

### Documentation

- **[Documentation Portal](./index.md)** - Complete documentation index
- **[Configuration Guide](./configuration.md)** - Complete `uptool.yaml` reference
- **[Manifest Files Reference](./manifests.md)** - All supported manifest types
- **[Integration Guides](./integrations/README.md)** - Detailed guides for each integration
- **[Examples](https://github.com/santosr2/uptool/blob/{{ extra.uptool_version }}/examples/)** - Example configurations for all integrations
- **[Plugin Development](./plugin-development.md)** - Create external plugins
- **[Version Management](./versioning.md)** - Automated versioning with conventional commits
- **[GitHub Environments](./environments.md)** - Approval gates for releases
- **[GitHub Action Usage](./action-usage.md)** - Using uptool in CI/CD

### Community

- **Issues**: [GitHub Issues](https://github.com/santosr2/uptool/issues)
- **Discussions**: [GitHub Discussions](https://github.com/santosr2/uptool/discussions)
- **Changelog**: [CHANGELOG.md](https://github.com/santosr2/uptool/blob/{{ extra.uptool_version }}/CHANGELOG.md)
- **Contributing**: [CONTRIBUTING.md](https://github.com/santosr2/uptool/blob/{{ extra.uptool_version }}/CONTRIBUTING.md)
- **Security**: [SECURITY.md](https://github.com/santosr2/uptool/blob/{{ extra.uptool_version }}/SECURITY.md)

**Questions?** Open a discussion or reach out to the maintainers.

# Welcome to uptool

**Universal, Manifest-First Dependency Updater**

uptool combines the ecosystem breadth of [Topgrade](https://github.com/topgrade-rs/topgrade), the precision of [Dependabot](https://github.com/dependabot/dependabot-core), and the flexibility of [Renovate](https://github.com/renovatebot/renovate) — but with a **manifest-first philosophy** that works across ANY project toolchain defined in configuration files.

---

## Why uptool?

Modern projects use dozens of tools across multiple ecosystems. uptool helps you manage:

- **Language dependencies**: npm packages
- **Infrastructure tools**: Terraform modules, Helm charts
- **Development tools**: pre-commit hooks, tflint plugins
- **Runtime version managers**: asdf (`.tool-versions`), mise (`mise.toml`)

Each ecosystem has its own update mechanism. Keeping them all current is tedious and error-prone.

**uptool provides a unified interface** to scan, plan, and update dependencies across all your manifest files.

---

## Manifest-First Philosophy

Unlike traditional dependency updaters that focus on lockfiles, uptool updates **manifest files directly** (package.json, Chart.yaml, *.tf files, etc.), preserving your intent while keeping dependencies current.

### The Approach

1. **Update manifests first**: `package.json`, `Chart.yaml`, `*.tf` files
2. **Use native commands when they update manifests**: `pre-commit autoupdate` updates `.pre-commit-config.yaml` ✅
3. **Don't use commands that only update lockfiles**: `npm update` only touches `package-lock.json` ❌
4. **Then optionally run native lockfile updates**: `npm install`, `terraform init`, etc.

This ensures your **declared dependencies** stay current, not just resolved versions.

---

## Key Features

<div class="grid cards" markdown>

- :material-tools: **Multi-Ecosystem Support**

    ---

    npm, Helm, Terraform, tflint, pre-commit, asdf, mise — all in one tool

- :material-file-document: **Manifest-First Updates**

    ---

    Updates configuration files directly, preserving formatting and comments

- :material-cloud-check: **Intelligent Version Resolution**

    ---

    Queries upstream registries (npm, Terraform Registry, Helm repos, GitHub Releases)

- :material-shield-check: **Safe by Default**

    ---

    Dry-run mode, diff generation, validation before applying changes

- :material-timer-sand: **Concurrent Execution**

    ---

    Parallel scanning and planning with worker pools for fast performance

- :octicons-git-branch-16: **GitHub Action Integration**

    ---

    Use as a CLI tool locally or as a GitHub Action in CI/CD pipelines

</div>

---

## Quick Example

```bash
# Scan for outdated dependencies
uptool scan

# Preview available updates
uptool plan

# Apply updates with diff preview
uptool update --diff

# Filter by integration
uptool update --only npm,terraform
```

---

## Supported Integrations

| Integration | Status | Manifest Files | Registry |
|-------------|--------|----------------|----------|
| **npm** | ✅ Stable | `package.json` | npm Registry API |
| **Helm** | ✅ Stable | `Chart.yaml` | Helm chart repositories |
| **pre-commit** | ✅ Stable | `.pre-commit-config.yaml` | GitHub Releases |
| **Terraform** | ✅ Stable | `*.tf` | Terraform Registry API |
| **tflint** | ✅ Stable | `.tflint.hcl` | GitHub Releases |
| **asdf** | ⚠️ Experimental | `.tool-versions` | GitHub Releases (per tool) |
| **mise** | ⚠️ Experimental | `mise.toml`, `.mise.toml` | GitHub Releases (per tool) |

---

## Getting Started

Ready to get started? Choose your path:

<div class="grid cards" markdown>

- :material-download: [**Installation**](installation.md)

    ---

    Install uptool via Go, pre-built binaries, or package managers

- :material-rocket-launch: [**Quick Start**](quickstart.md)

    ---

    Get up and running in 5 minutes with a sample project

- :material-book-open-variant: [**User Guide**](configuration.md)

    ---

    Deep dive into configuration, environments, and advanced usage

- :material-github: [**GitHub Action**](action-usage.md)

    ---

    Automate dependency updates in your CI/CD pipelines

</div>

---

## Community & Support

- **GitHub**: [santosr2/uptool](https://github.com/santosr2/uptool)
- **Issues**: [Report a bug](https://github.com/santosr2/uptool/issues/new)
- **Discussions**: [Ask questions](https://github.com/santosr2/uptool/discussions)
- **Security**: See our [Security Policy](https://github.com/santosr2/uptool/blob/{{ extra.uptool_version }}/SECURITY.md)

---

## Project Status

!!! info Development Status
    uptool is under active development. The current focus is on:

- ✅ Stabilizing core integrations (npm, Helm, Terraform, pre-commit, tflint)
- 🚧 Completing asdf/mise integrations (detection works, updates not yet implemented)
- 🚧 Expanding test coverage (target: >70%)
- 🚧 Adding Python ecosystem support
- 📝 Improving documentation and examples

---

## License

uptool is released under the [MIT License](LICENSE.md).

---

<small>Made with ❤️ by the uptool contributors</small>

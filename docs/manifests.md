# Manifest Files Reference

uptool is a **manifest-first** dependency updater. This document catalogs all supported manifest file types across different ecosystems.

## Philosophy: Manifest-First

uptool updates **manifest files** (source of truth) rather than lockfiles or resolved dependencies:

1. **Manifests declare intent** - They specify which versions you want
2. **Lockfiles are generated** - They record resolved versions
3. **Update manifests first** - Changes propagate to lockfiles via native tools

This ensures your declared dependencies stay current, not just resolved versions.

## Supported Manifest Types

### JavaScript/Node.js (npm)

**Integration**: `npm`

**Manifest Files**:

- `package.json`

**What Gets Updated**:

- `dependencies`
- `devDependencies`
- `peerDependencies`
- `optionalDependencies`

**Update Strategy**:

- Custom JSON rewriting
- Preserves version constraint prefixes (`^`, `~`, `>=`, etc.)
- Preserves formatting and key order

**Example**:

```json
{
  "dependencies": {
    "express": "^4.18.0",     // Updated to "^4.19.2"
    "lodash": "~4.17.20"       // Updated to "~4.17.21"
  },
  "devDependencies": {
    "jest": ">=29.0.0"         // Updated to ">=29.7.0"
  }
}
```

**Registry**: npm Registry API (`https://registry.npmjs.org`)

**Notes**:

- Does NOT update `package-lock.json` directly
- Run `npm install` after updating to regenerate lockfile
- Workspace support: Yes (monorepos with `workspaces` field)

---

### Kubernetes/Helm

**Integration**: `helm`

**Manifest Files**:

- `Chart.yaml`

**What Gets Updated**:

- `dependencies[].version` - Chart dependencies

**Update Strategy**:

- YAML parsing and rewriting
- Preserves comments and formatting

**Example**:

```yaml
apiVersion: v2
name: my-app
dependencies:
  - name: postgresql
    version: 12.0.0           # Updated to 18.1.8
    repository: https://charts.bitnami.com/bitnami
  - name: redis
    version: 17.0.0           # Updated to 23.2.12
    repository: https://charts.bitnami.com/bitnami
```

**Registry**: Helm chart repositories (index.yaml)

**Notes**:

- Does NOT update `Chart.lock`
- Run `helm dependency update` after to regenerate lockfile
- Only updates dependency versions, not chart metadata

---

### Terraform

**Integration**: `terraform`

**Manifest Files**:

- `*.tf` (any Terraform file)
- `main.tf`, `modules.tf`, `providers.tf`, etc.

**What Gets Updated**:

- `module` block `version` attributes
- Module source versions in git URLs (future)
- Provider versions (future)

**Update Strategy**:

- HCL parsing and rewriting via `hashicorp/hcl`
- Preserves comments and formatting

**Example**:

```hcl
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "3.0.0"           # Updated to "5.13.0"
}

module "security_group" {
  source  = "terraform-aws-modules/security-group/aws"
  version = "~> 4.0"          # Updated to "~> 5.0"
}
```

**Registry**: Terraform Registry API (`https://registry.terraform.io`)

**Notes**:

- Does NOT update `.terraform.lock.hcl`
- Run `terraform init -upgrade` after to regenerate lockfile
- Version constraints are preserved

---

### tflint

**Integration**: `tflint`

**Manifest Files**:

- `.tflint.hcl`

**What Gets Updated**:

- `plugin` block `version` attributes

**Update Strategy**:

- HCL parsing and rewriting
- Preserves comments and formatting

**Example**:

```hcl
plugin "aws" {
  enabled = true
  version = "0.21.0"          # Updated to "0.44.0"
  source  = "github.com/terraform-linters/tflint-ruleset-aws"
}

plugin "azurerm" {
  enabled = true
  version = "0.20.0"          # Updated to "0.28.0"
  source  = "github.com/terraform-linters/tflint-ruleset-azurerm"
}
```

**Registry**: GitHub Releases (for plugins)

**Notes**:

- Plugin sources must be valid GitHub repository paths
- Follows semantic versioning

---

### Pre-Commit Hooks

**Integration**: `precommit`

**Manifest Files**:

- `.pre-commit-config.yaml`

**What Gets Updated**:

- `repos[].rev` - Hook repository revisions

**Update Strategy**:

- **Native command**: `pre-commit autoupdate`
- Uses pre-commit's built-in update mechanism
- This is because `pre-commit autoupdate` updates the manifest directly

**Example**:

```yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.3.0               # Updated to v6.0.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer

  - repo: https://github.com/psf/black
    rev: 22.10.0              # Updated to 24.10.0
    hooks:
      - id: black
```

**Registry**: GitHub Releases (for hook repositories)

**Notes**:

- Uses native `pre-commit autoupdate` command
- Respects minimum_pre_commit_version
- Does NOT create `.pre-commit-config.yaml.lock` (pre-commit doesn't use lockfiles)

---

### asdf Version Manager

**Integration**: `asdf`

**Manifest Files**:

- `.tool-versions`

**What Gets Updated**:

- Tool versions (e.g., `go 1.23.0` → `go 1.25.0`)

**Update Strategy**:

- Line-based parsing and rewriting
- Preserves formatting and comments

**Example**:

```text
# Development tools
go 1.23.0                     # Updated to 1.25.0
nodejs 20.10.0                # Updated to 22.12.0
terraform 1.5.0               # Updated to 1.10.5

# Build tools
python 3.11.0                 # Updated to 3.13.1
```

**Registry**: GitHub Releases (per tool via asdf plugin mapping)

**Notes**:

- Does NOT update installed versions
- Run `asdf install` after to install new versions
- Supports multiple versions per tool (space-separated)

---

### mise Version Manager

**Integration**: `mise`

**Manifest Files**:

- `mise.toml`
- `.mise.toml`

**What Gets Updated**:

- `[tools]` section tool versions

**Update Strategy**:

- TOML parsing and rewriting
- Supports both string format and map format
- Preserves comments and formatting

**Example (String Format)**:

```toml
[tools]
go = "1.23"                   # Updated to "1.25"
node = "20"                   # Updated to "22"
golangci-lint = "2.6"         # Updated to "2.7"
terraform = "1.5.0"           # Updated to "1.10.5"
```

**Example (Map Format)**:

```toml
[tools]
go = { version = "1.23" }     # Updated to { version = "1.25" }
node = { version = "20", path = ".nvmrc" }
```

**Registry**: GitHub Releases (per tool)

**Notes**:

- Does NOT install new versions automatically
- Run `mise install` after to install new versions
- Supports both mise.toml and .mise.toml (hidden file)

---

## Manifest Detection

uptool automatically detects manifest files by:

1. **Filename matching**: Exact matches like `package.json`, `Chart.yaml`
2. **Pattern matching**: Glob patterns like `*.tf`, `mise.toml`
3. **Directory walking**: Recursively scans from repository root

### Detection Order

Each integration defines its own detection logic:

```go
// Example: npm integration
func Detect(ctx context.Context, repoRoot string) ([]*Manifest, error) {
    // Look for package.json files
    matches, err := filepath.Glob(filepath.Join(repoRoot, "**/package.json"))
    // ...
}
```

### Ignored Directories

By default, uptool skips:

- `.git/`
- `node_modules/`
- `vendor/`
- `.terraform/`
- `dist/`, `build/`

## Manifest-First Principles

### ✅ DO: Update Manifests

```bash
# Good: Updates package.json (manifest)
uptool update --only=npm

# Then regenerate lockfile
npm install
```

### ❌ DON'T: Rely on Lockfile-Only Tools

```bash
# Bad: npm update only updates package-lock.json
npm update

# package.json still has old versions!
```

### Why Manifest-First?

1. **Intent over resolution**: Manifests declare what you want, lockfiles record what you got
2. **Portability**: Manifests work across environments, lockfiles don't
3. **Auditability**: Changes to manifests are explicit in version control
4. **Consistency**: Everyone gets the same declared versions

## Native Commands vs Custom Rewriting

### When Native Commands Are Used

uptool uses native commands **only when they update the manifest**:

| Integration | Native Command | Reason |
|-------------|---------------|--------|
| `precommit` | `pre-commit autoupdate` | Updates `.pre-commit-config.yaml` directly |

### When Custom Rewriting Is Used

All other integrations use custom parsing/rewriting:

| Integration | Reason |
|-------------|--------|
| `npm` | `npm update` only updates lockfile |
| `helm` | `helm dependency update` only updates Chart.lock |
| `terraform` | `terraform init -upgrade` only updates .terraform.lock.hcl |
| `tflint` | No native update command exists |
| `asdf` | `.tool-versions` is plain text, no native update |
| `mise` | `mise.toml` is TOML, custom parsing needed |

## Configuration

You can configure which manifests to process via `uptool.yaml`:

```yaml
version: 1

integrations:
  - id: npm
    enabled: true
    policy:
      update: minor              # Only minor/patch updates

  - id: helm
    enabled: true
    policy:
      update: minor

  - id: terraform
    enabled: false               # Disable Terraform updates
```

See [configuration.md](configuration.md) for complete reference.

## See Also

- [Configuration Reference](configuration.md) - Configure update policies
- [Troubleshooting Guide](troubleshooting.md) - Common issues and solutions
- [Integration Guides](integrations/README.md) - Detailed integration documentation

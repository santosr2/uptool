# Configuration Reference

uptool can be configured using a `uptool.yaml` file in your repository root. Configuration is **optional** — if no file exists, uptool runs all integrations with sensible defaults.

## Quick Start

Create `uptool.yaml` in your repository root:

```yaml
version: 1

integrations:
  - id: npm
    enabled: true
    policy:
      update: minor
      allow_prerelease: false
```

## Configuration Schema

### Top-Level Structure

```yaml
version: 1                    # Configuration format version (required)

integrations:                 # List of integration configurations
  - id: <integration_id>      # Integration identifier
    enabled: true|false       # Enable/disable this integration
    match:                    # Optional: File matching rules
      files: [...]            # List of file patterns
    policy:                   # Update policy for this integration
      update: none|patch|minor|major
      allow_prerelease: true|false
      pin: true|false
      cadence: daily|weekly|monthly  # Update frequency (optional)

org_policy:                   # Optional: Organization-level policies
  require_signoff_from: [...]  # List of required approvers
  signing:                    # Artifact signing verification
    cosign_verify: true|false
  auto_merge:                 # Automatic PR merging
    enabled: true|false
    guards: [...]             # Required conditions
```

## Configuration Fields

### version

**Type**: `integer` | **Required**: Yes | **Default**: N/A

The configuration format version. Currently only `1` is supported.

### integrations

**Type**: `array` | **Required**: No | **Default**: All integrations enabled

List of integration configurations.

#### id

**Type**: `string` | **Required**: Yes

**Allowed values**: `npm`, `helm`, `terraform`, `tflint`, `precommit`, `asdf`, `mise`

The integration identifier.

#### enabled

**Type**: `boolean` | **Required**: No | **Default**: `true`

Whether this integration should run. CLI flags `--only` and `--exclude` override this setting.

#### match

**Type**: `object` | **Required**: No

File matching rules for this integration.

**match.files** - Array of glob patterns:

```yaml
- id: npm
  match:
    files:
      - "package.json"           # Root package.json
      - "apps/*/package.json"    # Monorepo packages
      - "packages/*/package.json"
```

**Default patterns by integration**:

| Integration | Default Patterns |
|-------------|------------------|
| npm | `package.json` |
| helm | `Chart.yaml`, `*/Chart.yaml`, `charts/*/Chart.yaml` |
| terraform | `*.tf`, `**/*.tf` |
| tflint | `.tflint.hcl` |
| precommit | `.pre-commit-config.yaml` |
| asdf | `.tool-versions` |
| mise | `mise.toml`, `.mise.toml` |

#### policy

**Type**: `object` | **Required**: No

Update policy for this integration.

**policy.update** - Maximum version bump to allow:

| Value | Allows | Example |
|-------|--------|---------|
| `none` | No updates | Scan/plan only |
| `patch` | Patch updates only | 1.2.3 → 1.2.4 |
| `minor` | Patch + minor | 1.2.3 → 1.3.0 |
| `major` | All updates | 1.2.3 → 2.0.0 |

**Default**: `minor`

**policy.allow_prerelease** - Include pre-release versions:

**Type**: `boolean` | **Default**: `false`

When `true`, considers versions like `1.2.3-alpha20250708`, `1.2.3-beta2`, `1.2.3-rc1`.

**policy.pin** - Write exact versions or ranges:

**Type**: `boolean` | **Default**: Depends on integration

| Integration | pin: true | pin: false |
|-------------|-----------|------------|
| npm | `"4.19.2"` | `"^4.19.2"` (preserves constraint) |
| helm | `12.0.0` | `12.0.0` (always pinned) |
| terraform | `"5.8.1"` | `"5.8.1"` (always pinned) |
| mise | `"1.25"` | `"1.25"` (always pinned) |

**policy.cadence** - Update frequency for scheduled runs:

**Type**: `string` | **Default**: None

**Values**: `daily`, `weekly`, `monthly`

Controls how often to check for updates in automated scenarios (primarily for GitHub Actions integration).

### org_policy

**Type**: `object` | **Required**: No

Organization-level policies for governance and automation.

**org_policy.require_signoff_from** - Required approvers:

**Type**: `array of strings` | **Default**: None

List of email addresses or team identifiers that must approve changes.

```yaml
org_policy:
  require_signoff_from:
    - "platform-team@company.com"
    - "security-team@company.com"
```

**org_policy.signing** - Artifact signing verification:

**Type**: `object` | **Default**: None

```yaml
org_policy:
  signing:
    cosign_verify: true  # Verify signatures with Cosign
```

**org_policy.auto_merge** - Automatic PR merging:

**Type**: `object` | **Default**: None

```yaml
org_policy:
  auto_merge:
    enabled: true
    guards:
      - "ci-green"           # All CI checks must pass
      - "codeowners-approve"  # CODEOWNERS must approve
```

## Complete Examples

### Conservative (Production)

For production systems requiring stability:

```yaml
version: 1

integrations:
  # Only patch updates for runtime dependencies
  - id: mise
    policy:
      update: patch

  - id: npm
    policy:
      update: patch

  # Minor updates for infrastructure
  - id: terraform
    policy:
      update: minor

  - id: helm
    policy:
      update: minor
```

### Comprehensive (All Features)

Full configuration with all options:

```yaml
version: 1

integrations:
  - id: npm
    enabled: true
    match:
      files:
        - "package.json"
        - "apps/*/package.json"     # Monorepo support
        - "packages/*/package.json"
    policy:
      update: minor
      allow_prerelease: false
      pin: false  # Use version ranges

  - id: helm
    enabled: true
    policy:
      update: minor
      allow_prerelease: false

  - id: terraform
    enabled: true
    match:
      files:
        - "infrastructure/**/*.tf"
        - "modules/**/*.tf"
    policy:
      update: minor
      allow_prerelease: false

  - id: precommit
    enabled: true
    policy:
      update: major  # Aggressive for dev tools

  - id: tflint
    enabled: true
    policy:
      update: major

  - id: asdf
    enabled: true
    policy:
      update: patch  # Conservative for runtimes

  - id: mise
    enabled: true
    policy:
      update: patch  # Conservative for runtimes
```

## Configuration Precedence

Settings are applied in this order (later overrides earlier):

1. **Integration defaults** (hardcoded in code)
2. **`uptool.yaml` configuration** (if exists)
3. **CLI flags** (`--only`, `--exclude`)

Example:

```yaml
# uptool.yaml
integrations:
  - id: npm
    enabled: false
```

```bash
# CLI overrides config file
uptool update --only=npm  # npm WILL update despite enabled: false
```

## Validation

uptool validates configuration on startup. Invalid values log warnings and use defaults:

```text
WARN: Unknown integration 'unknown_integration' in config, skipping
WARN: Invalid update policy 'invalid_value' for npm, using default 'minor'
```

## Best Practices

1. **Start conservative**: Use `patch` or `minor` for production
2. **Separate policies**: Conservative for runtime, aggressive for dev tools
3. **Explicit paths**: Avoid broad wildcards in monorepos (`apps/*/package.json` not `**/package.json`)
4. **Document decisions**: Add comments explaining policy choices

## Troubleshooting

**Config not loading**: Check `uptool.yaml` exists in root, valid YAML syntax, run with `-v`

**Integration not running**: Verify `enabled: true`, no CLI overrides (`--exclude`), files match pattern

**Policy not applied**: Policy applies to `uptool update`, bypassed with `--only` flag

## See Also

- [CLI Reference](cli/commands.md) - Complete command documentation
- [Integration Guides](integrations/README.md) - Integration-specific details
- [GitHub Action Usage](action-usage.md) - CI/CD configuration
- [Examples](https://github.com/santosr2/uptool/tree/main/examples) - Sample configurations

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

**policy.enabled** - Enable/disable policy enforcement for this integration:

**Type**: `boolean` | **Default**: `true`

When `false`, the integration still runs but uses **default policy settings** instead of configured values:

- `update: major` (allow all updates)
- `allow_prerelease: false` (no pre-releases)
- `pin: <default per integration>`
- Respects manifest constraints (`^`, `~`, `>=`, etc.)

When `true`, the configured policy settings are applied.

**Note**: This is different from `integrations[*].enabled`, which controls whether the integration is registered at all. Use `policy.enabled: false` to temporarily disable policy restrictions without removing the integration.

**Use case**: Temporarily allow all updates for an integration during major version upgrades without modifying other policy settings.

## Policy Precedence

uptool follows a clear precedence order when determining which updates to allow:

### 1. CLI Flags (Highest Priority)

Command-line flags always override configuration file settings:

```bash
# Override uptool.yaml and allow all updates
uptool update --update-level=major

# Allow prereleases regardless of config
uptool update --allow-prerelease
```

### 2. uptool.yaml Integration Policy

Per-integration policies in `uptool.yaml`:

```yaml
integrations:
  - id: npm
    policy:
      update: minor  # Limits npm updates to minor/patch only
```

### 3. Manifest Constraints

Version constraints in manifest files (package.json, Chart.yaml, etc.):

```json
{
  "dependencies": {
    "express": "^4.18.0"  // Constraint: only 4.x versions allowed
  }
}
```

Even if `update: major` is set, uptool respects the `^4.18.0` constraint and won't propose `express@5.0.0`.

### 4. Default Behavior (Lowest Priority)

If no policy or constraints exist, uptool allows all stable updates (equivalent to `update: major`).

### Precedence Example

Given this configuration:

```yaml
integrations:
  - id: npm
    policy:
      update: minor  # Only minor/patch updates
```

And this package.json:

```json
{
  "dependencies": {
    "lodash": "^4.17.20"  // Constraint allows 4.x only
  }
}
```

**Result**:

- ✅ `lodash@4.17.21` - Allowed (patch update, within 4.x constraint)
- ✅ `lodash@4.18.0` - Allowed (minor update, within 4.x constraint)
- ❌ `lodash@5.0.0` - Blocked (major update exceeds policy + constraint)

Running with `--update-level=major`:

- ✅ `lodash@5.0.0` - Now allowed (CLI flag overrides policy, but constraint needs manual update)

## Policy Best Practices

### Conservative (Production)

For production systems requiring stability:

```yaml
integrations:
  # Runtime dependencies: patch updates only
  - id: mise
    policy:
      update: patch
      pin: true

  # Application dependencies: minor updates
  - id: npm
    policy:
      update: minor
      pin: true
      cadence: monthly

  # Infrastructure: patch updates only
  - id: terraform
    policy:
      update: patch
      cadence: monthly

  - id: helm
    policy:
      update: patch
      cadence: monthly
```

### Moderate (Staging)

Balanced approach for staging environments:

```yaml
integrations:
  - id: npm
    policy:
      update: minor  # Allow minor updates
      pin: false     # Preserve version ranges

  - id: terraform
    policy:
      update: minor

  - id: helm
    policy:
      update: minor
```

### Aggressive (Development)

Keep development tools up-to-date:

```yaml
integrations:
  - id: precommit
    policy:
      update: major  # Allow all updates
      cadence: weekly

  - id: npm
    match:
      files: ["package.json"]  # Root only
    policy:
      update: major
      allow_prerelease: false
      cadence: weekly
```

### Monorepo Example

Different policies for different parts of a monorepo:

```yaml
integrations:
  # Production apps: conservative
  - id: npm
    match:
      files: ["apps/*/package.json"]
    policy:
      update: patch
      pin: true

  # Shared libraries: moderate
  - id: npm
    match:
      files: ["packages/*/package.json"]
    policy:
      update: minor
      pin: false

  # Development tools: aggressive
  - id: precommit
    policy:
      update: major
```

**Note**: Multiple configurations for the same integration ID use the **last matching configuration**.

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

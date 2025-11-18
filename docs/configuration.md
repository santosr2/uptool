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

  - id: helm
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
      enabled: true|false
      update: none|patch|minor|major
      allow_prerelease: true|false
      pin: true|false
      cadence: daily|weekly|monthly

org_policy:                   # Optional: Organization-wide settings
  # Advanced settings (future)
```

## version

**Type**: `integer`
**Required**: Yes
**Default**: N/A

The configuration format version. Currently only `1` is supported.

```yaml
version: 1
```

Future versions may introduce breaking changes to the configuration schema.

## integrations

**Type**: `array`
**Required**: No
**Default**: All integrations enabled with default policies

List of integration configurations. Each integration can be individually configured.

### Integration Object

```yaml
- id: npm                     # Integration identifier
  enabled: true               # Enable this integration
  match:                      # File matching (optional)
    files:
      - "package.json"
      - "apps/*/package.json"
  policy:                     # Update policy
    enabled: true
    update: minor
    allow_prerelease: false
    pin: true
    cadence: weekly
```

### id

**Type**: `string`
**Required**: Yes
**Allowed values**: `npm`, `helm`, `terraform`, `tflint`, `precommit`, `asdf`, `mise`

The integration identifier. Must match a registered integration.

**Example**:

```yaml
integrations:
  - id: npm
  - id: helm
  - id: terraform
```

### enabled

**Type**: `boolean`
**Required**: No
**Default**: `true`

Whether this integration should run.

**Example**:

```yaml
# Enable npm updates
- id: npm
  enabled: true

# Disable terraform updates
- id: terraform
  enabled: false
```

**Note**: CLI flags `--only` and `--exclude` override this setting.

### match

**Type**: `object`
**Required**: No
**Default**: Integration-specific defaults

File matching rules for this integration. Overrides integration defaults.

#### match.files

**Type**: `array` of `string`
**Required**: No
**Default**: Integration-specific patterns

Glob patterns for files to include.

**Example**:

```yaml
- id: npm
  match:
    files:
      - "package.json"           # Root package.json
      - "apps/*/package.json"    # Monorepo packages
      - "packages/*/package.json"
```

**Default Patterns by Integration**:

| Integration | Default Patterns |
|-------------|------------------|
| npm | `package.json` |
| helm | `Chart.yaml`, `*/Chart.yaml`, `charts/*/Chart.yaml` |
| terraform | `*.tf`, `**/*.tf` |
| tflint | `.tflint.hcl` |
| precommit | `.pre-commit-config.yaml` |
| asdf | `.tool-versions` |
| mise | `mise.toml`, `.mise.toml` |

### policy

**Type**: `object`
**Required**: No
**Default**: See defaults below

Update policy for this integration.

```yaml
policy:
  enabled: true
  update: minor
  allow_prerelease: false
  pin: true
  cadence: weekly
```

#### policy.enabled

**Type**: `boolean`
**Required**: No
**Default**: `true`

Whether to apply updates for this integration.

```yaml
# Scan and plan, but don't update
- id: terraform
  policy:
    enabled: false
```

#### policy.update

**Type**: `string`
**Required**: No
**Default**: `minor`
**Allowed values**: `none`, `patch`, `minor`, `major`

Maximum version bump to allow.

**Values**:

- **`none`**: No updates (scan/plan only)
- **`patch`**: Only patch updates (1.2.3 → 1.2.4)
- **`minor`**: Patch and minor updates (1.2.3 → 1.3.0)
- **`major`**: All updates including major (1.2.3 → 2.0.0)

**Example**:

```yaml
integrations:
  # Conservative: only patch updates for runtime tools
  - id: mise
    policy:
      update: patch

  # Standard: minor updates for libraries
  - id: npm
    policy:
      update: minor

  # Aggressive: major updates for development tools
  - id: precommit
    policy:
      update: major
```

**Semantic Versioning Rules**:

| Current | Patch Allows | Minor Allows | Major Allows |
|---------|--------------|--------------|--------------|
| 1.2.3 | 1.2.4 | 1.3.0 | 2.0.0 |
| 0.5.2 | 0.5.3 | 0.6.0 | 1.0.0 |
| 2.0.0 | 2.0.1 | 2.1.0 | 3.0.0 |

#### policy.allow_prerelease

**Type**: `boolean`
**Required**: No
**Default**: `false`

Whether to consider pre-release versions (alpha, beta, rc).

**Example**:

```yaml
# Stable versions only
- id: terraform
  policy:
    allow_prerelease: false

# Include pre-releases for testing new features
- id: helm
  policy:
    allow_prerelease: true
```

**Pre-release formats recognized**:

- `1.2.3-alpha.1`
- `1.2.3-beta.2`
- `1.2.3-rc.1`
- `1.2.3-pre`

#### policy.pin

**Type**: `boolean`
**Required**: No
**Default**: Depends on integration

Whether to write exact versions or ranges.

**Example**:

```yaml
# Write exact versions
- id: terraform
  policy:
    pin: true
# Result: version = "5.8.1"

# Write version ranges (where supported)
- id: npm
  policy:
    pin: false
# Result: "express": "^4.19.2"
```

**Integration Support**:

| Integration | pin: true | pin: false |
|-------------|-----------|------------|
| npm | `"4.19.2"` | `"^4.19.2"` (preserves constraint) |
| helm | `12.0.0` | `12.0.0` (always pinned) |
| terraform | `"5.8.1"` | `">= 5.8.1"` (not yet implemented) |
| tflint | `"0.44.0"` | `"0.44.0"` (always pinned) |
| precommit | `v6.0.0` | `v6.0.0` (always pinned) |
| asdf | `1.25.0` | `1.25.0` (always pinned) |
| mise | `"1.25"` | `"1.25"` (always pinned) |

#### policy.cadence

**Type**: `string`
**Required**: No
**Default**: Not enforced
**Allowed values**: `daily`, `weekly`, `monthly`
**Status**: ⚠️ **Planned feature** (not yet implemented)

Recommended update frequency for CI/CD automation.

**Example**:

```yaml
- id: mise
  policy:
    cadence: weekly  # Update runtime tools weekly

- id: npm
  policy:
    cadence: daily   # Update dependencies daily
```

This field is currently for documentation only. Future versions may enforce cadence in GitHub Actions.

## org_policy

**Type**: `object`
**Required**: No
**Default**: None
**Status**: ⚠️ **Planned feature** (not yet implemented)

Organization-wide policy settings for enterprise use.

**Planned features**:

```yaml
org_policy:
  # Require sign-off for certain update types
  require_signoff_from:
    - "platform-team@company.com"
    - "security-team@company.com"

  # Artifact signing verification
  signing:
    cosign_verify: true
    policy_url: "https://company.com/signing-policy.json"

  # Auto-merge settings for GitHub Actions
  auto_merge:
    enabled: true
    guards:
      - "ci-green"
      - "codeowners-approve"
      - "security-scan-pass"
```

## Complete Examples

### Conservative Configuration

For production systems requiring stability:

```yaml
version: 1

integrations:
  # Only patch updates for runtime dependencies
  - id: mise
    enabled: true
    policy:
      update: patch
      allow_prerelease: false

  - id: npm
    enabled: true
    policy:
      update: patch
      allow_prerelease: false

  # Minor updates for infrastructure
  - id: terraform
    enabled: true
    policy:
      update: minor
      allow_prerelease: false

  - id: helm
    enabled: true
    policy:
      update: minor
      allow_prerelease: false

  # Disable dev tools updates
  - id: precommit
    enabled: false

  - id: tflint
    enabled: false
```

### Aggressive Configuration

For development environments or testing:

```yaml
version: 1

integrations:
  - id: npm
    enabled: true
    policy:
      update: major
      allow_prerelease: true

  - id: helm
    enabled: true
    policy:
      update: major
      allow_prerelease: false

  - id: terraform
    enabled: true
    policy:
      update: major
      allow_prerelease: false

  - id: precommit
    enabled: true
    policy:
      update: major
      allow_prerelease: false

  - id: tflint
    enabled: true
    policy:
      update: major
      allow_prerelease: false

  - id: mise
    enabled: true
    policy:
      update: minor  # Still conservative for runtime tools
      allow_prerelease: false
```

### Monorepo Configuration

For monorepos with multiple package.json files:

```yaml
version: 1

integrations:
  - id: npm
    enabled: true
    match:
      files:
        - "package.json"            # Root package.json
        - "apps/*/package.json"     # App packages
        - "packages/*/package.json" # Shared packages
        - "tools/*/package.json"    # Tool packages
    policy:
      update: minor
      allow_prerelease: false
      pin: false  # Use ranges for npm

  - id: terraform
    enabled: true
    match:
      files:
        - "infrastructure/**/*.tf"
        - "modules/**/*.tf"
    policy:
      update: minor
      allow_prerelease: false
```

### Selective Integration Configuration

Only update specific ecosystems:

```yaml
version: 1

integrations:
  # Enable only npm and Helm updates
  - id: npm
    enabled: true
    policy:
      update: minor
      allow_prerelease: false

  - id: helm
    enabled: true
    policy:
      update: minor
      allow_prerelease: false

  # Explicitly disable others
  - id: terraform
    enabled: false

  - id: tflint
    enabled: false

  - id: precommit
    enabled: false

  - id: mise
    enabled: false
```

## Configuration Precedence

Settings are applied in this order (later overrides earlier):

1. **Integration defaults** (hardcoded in code)
2. **`uptool.yaml` configuration** (if exists)
3. **CLI flags** (`--only`, `--exclude`)

**Example**:

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

uptool validates configuration on startup:

**Valid**:

```yaml
version: 1
integrations:
  - id: npm
    enabled: true
    policy:
      update: minor
```

**Invalid** (logs warning, uses defaults):

```yaml
version: 1
integrations:
  - id: unknown_integration  # ❌ Unknown integration
    enabled: true

  - id: npm
    policy:
      update: invalid_value  # ❌ Invalid update value
```

**Validation errors are logged**:

```text
WARN: Unknown integration 'unknown_integration' in config, skipping
WARN: Invalid update policy 'invalid_value' for npm, using default 'minor'
```

## Best Practices

### 1. Start Conservative

Begin with `patch` or `minor` updates:

```yaml
version: 1
integrations:
  - id: npm
    policy:
      update: minor  # Safe default
```

### 2. Separate Runtime from Development Tools

Use different update policies:

```yaml
integrations:
  # Conservative for runtime
  - id: mise
    policy:
      update: patch

  # Aggressive for dev tools
  - id: precommit
    policy:
      update: major
```

### 3. Test with Pre-releases First

Enable pre-releases in development, disable in production:

```yaml
# dev-uptool.yaml
integrations:
  - id: npm
    policy:
      allow_prerelease: true

# prod-uptool.yaml
integrations:
  - id: npm
    policy:
      allow_prerelease: false
```

### 4. Use Explicit File Patterns for Monorepos

Avoid wildcards that match too broadly:

```yaml
# Good: Explicit paths
- id: npm
  match:
    files:
      - "apps/api/package.json"
      - "apps/web/package.json"

# Risky: May match too many
- id: npm
  match:
    files:
      - "**/package.json"  # Might match node_modules!
```

### 5. Document Policy Decisions

Add comments explaining choices:

```yaml
integrations:
  # Only patch updates for Terraform to avoid breaking infrastructure changes
  - id: terraform
    policy:
      update: patch  # Infrastructure stability critical

  # Major updates OK for pre-commit hooks (non-breaking)
  - id: precommit
    policy:
      update: major  # Hooks are isolated, safe to update aggressively
```

## Integration-Specific Notes

### npm

- Preserves constraint prefixes (`^`, `~`) when `pin: false`
- Updates all dependency types (dependencies, devDependencies, peerDependencies, optionalDependencies)
- Respects package-lock.json after updates (run `npm install` to sync)

### Helm

- Updates chart dependencies in `dependencies` array
- Requires repository URL to be specified in Chart.yaml
- Always pins exact versions

### Terraform

- Updates module versions in `module` blocks
- Provider updates coming soon
- Respects HCL syntax and formatting

### precommit

- Uses native `pre-commit autoupdate` command
- Updates hook revisions in `.pre-commit-config.yaml`
- Queries GitHub Releases for latest versions

### tflint

- Updates plugin versions in `plugin` blocks
- Queries GitHub Releases for tflint rulesets
- Preserves HCL formatting

### asdf

- Updates tool versions in `.tool-versions` file
- One tool per line format
- Currently supports detection; version resolution in development

### mise

- Updates tool versions in `mise.toml` or `.mise.toml`
- Supports both string format (`go = "1.25"`) and map format (`go = { version = "1.25" }`)
- Preserves TOML structure and comments

## Troubleshooting

### Configuration Not Loading

**Problem**: uptool ignores config file

**Causes**:

1. File not in repository root
2. Filename typo (must be `uptool.yaml`, not `uptool.yml`)
3. Invalid YAML syntax

**Solution**:

```bash
# Check file exists in root
ls uptool.yaml

# Validate YAML syntax
yamllint uptool.yaml

# Run with verbose output
uptool scan -v
```

### Integration Not Running

**Problem**: Integration doesn't run despite configuration

**Check**:

1. Is `enabled: true`?
2. Are CLI flags overriding? (`--exclude=npm`)
3. Are files matched by pattern?

```yaml
# Enable and verify file patterns
- id: npm
  enabled: true
  match:
    files:
      - "package.json"  # Check this path is correct
```

### Policy Not Applied

**Problem**: Updates bypass policy limits

**Remember**: Policy only affects automatic updates, not manual selection

```bash
# Policy applies
uptool update

# Policy BYPASSED (manual override)
uptool update --only=npm  # Updates regardless of policy.enabled
```

## See Also

- [Integration Details](overview.md#integration-details)
- [CLI Reference](overview.md#cli-reference)
- [GitHub Action Configuration](action-usage.md)
- [Example Configurations](https://github.com/santosr2/uptool/blob/{{ extra.uptool_version }}/examples/uptool.yaml)

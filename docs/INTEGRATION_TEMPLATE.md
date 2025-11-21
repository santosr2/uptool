# [Integration Name] Integration

Brief one-sentence description of what this integration does.

## Overview

**Integration ID**: `integration-id`

**Manifest Files**: `manifest-file.ext`

**Update Strategy**: How it updates (e.g., YAML rewriting, native command, HCL parsing)

**Registry**: Where it queries versions (e.g., npm Registry API, GitHub Releases)

**Status**: ✅ Stable | 🧪 Experimental

## What Gets Updated

Clear bullet list of what parts of the manifest are updated:

- `section1` - Description
- `section2` - Description

## Example

**Before**:

```format
# Example manifest before update
dependency: 1.0.0
```

**After**:

```format
# Example manifest after update
dependency: 1.5.0
```

## Integration-Specific Behavior

Any unique behavior for this integration:

- **Version constraints**: How constraints are handled
- **Format preservation**: What formatting is preserved
- **Special features**: Any unique features

## Configuration

Example `uptool.yaml` configuration:

```yaml
version: 1

integrations:
  - id: integration-id
    enabled: true
    policy:
      update: minor
      allow_prerelease: false
```

## Limitations

*(Optional section - only include if there are notable limitations)*

List any known limitations:

1. **Limitation name**: Description and workaround

## See Also

- [CLI Reference](cli/commands.md)
- [Configuration Guide](configuration.md)
- [Official Documentation](https://santosr2.github.io/uptool)

---

## Template Guidelines

### Length Target

- **Minimum**: 60 lines
- **Target**: 80-120 lines
- **Maximum**: 150 lines

### Required Sections

1. Overview (ID, files, strategy, registry, status)
2. What Gets Updated (bullet list)
3. Example (before/after code blocks)
4. Integration-Specific Behavior

### Optional Sections

- Configuration (if non-standard)
- Limitations (only if notable)

### Sections to AVOID

- ❌ Generic CLI usage (covered in CLI reference)
- ❌ Obvious troubleshooting (e.g., "check internet connection")
- ❌ Redundant registry API details
- ❌ Monorepo support (mention briefly in "What Gets Updated")
- ❌ Version constraint tables (show in example instead)
- ❌ Step-by-step workflows (covered in quickstart)

### Style Guidelines

- Be direct and concise
- Use tables for structured data
- Use code blocks for examples
- Link to other docs instead of duplicating content
- Focus on integration-specific details only

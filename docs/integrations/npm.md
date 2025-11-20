# npm Integration

Updates JavaScript/Node.js dependencies in `package.json` files.

## Overview

**Integration ID**: `npm`

**Manifest Files**: `package.json`

**Update Strategy**: Custom JSON rewriting with constraint preservation

**Registry**: npm Registry API (`https://registry.npmjs.org`)

**Status**: ✅ Stable

## What Gets Updated

All dependency types in `package.json`:

- `dependencies` - Production dependencies
- `devDependencies` - Development dependencies
- `peerDependencies` - Peer dependencies
- `optionalDependencies` - Optional dependencies

**Monorepo support**: Each `package.json` updated independently.

## Example

**Before**:

```json
{
  "dependencies": {
    "express": "^4.18.0",
    "lodash": "~4.17.20",
    "axios": ">=0.27.0"
  },
  "devDependencies": {
    "jest": "^29.0.0"
  }
}
```

**After**:

```json
{
  "dependencies": {
    "express": "^4.19.2",     // Preserves ^ constraint
    "lodash": "~4.17.21",     // Preserves ~ constraint
    "axios": ">=1.7.0"        // Preserves >= constraint
  },
  "devDependencies": {
    "jest": "^29.7.0"
  }
}
```

## Integration-Specific Behavior

### Version Constraint Preservation

uptool preserves version constraint prefixes:

| Constraint | Meaning | Before | After |
|------------|---------|--------|-------|
| `^` | Compatible with | `^4.18.0` | `^4.19.2` |
| `~` | Approximately | `~4.17.20` | `~4.17.21` |
| `>=` | Greater than or equal | `>=0.27.0` | `>=1.7.0` |
| (none) | Exact version | `1.0.0` | `1.5.0` |

### Lockfile Handling

uptool updates **only** `package.json`. Run `npm install` after updating to sync lockfiles:

```bash
uptool update --only npm
npm install
```

### Private Registries

Respects npm configuration from `.npmrc` or `npm config`. Configure authentication separately:

```bash
npm config set registry https://registry.company.com/
npm login --registry=https://registry.company.com/
```

## Configuration

```yaml
version: 1

integrations:
  - id: npm
    enabled: true
    match:
      files:
        - "package.json"
        - "apps/*/package.json"     # Monorepo paths
        - "packages/*/package.json"
    policy:
      update: minor                 # none, patch, minor, major
      allow_prerelease: false
      pin: false                    # false = preserve constraints
```

## Limitations

1. **No lockfile updates**: `package-lock.json` not modified. Run `npm install` after updates.
2. **No peer dependency validation**: Run `npm install` to see peer dependency warnings.

## See Also

- [CLI Reference](../cli/commands.md) - `uptool scan --only npm`, `uptool plan --only npm`
- [Configuration Guide](../configuration.md) - Policy settings
- [npm Registry API](https://github.com/npm/registry/blob/master/docs/REGISTRY-API.md)
- [Semantic Versioning](https://semver.org/)

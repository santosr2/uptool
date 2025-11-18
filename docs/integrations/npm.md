# npm Integration

The npm integration updates JavaScript/Node.js dependencies in `package.json` files.

## Overview

**Integration ID**: `npm`

**Manifest Files**: `package.json`

**Update Strategy**: Custom JSON rewriting

**Registry**: npm Registry API (`https://registry.npmjs.org`)

**Status**: ✅ Stable

## What Gets Updated

The npm integration updates all dependency types in `package.json`:

- `dependencies` - Production dependencies
- `devDependencies` - Development-only dependencies
- `peerDependencies` - Peer dependencies (plugins, extensions)
- `optionalDependencies` - Optional dependencies

## Example

**Before** (`package.json`):

```json
{
  "name": "my-app",
  "version": "1.0.0",
  "dependencies": {
    "express": "^4.18.0",
    "lodash": "~4.17.20",
    "axios": ">=0.27.0"
  },
  "devDependencies": {
    "jest": "^29.0.0",
    "@types/node": "^18.0.0"
  }
}
```

**After** (uptool update):

```json
{
  "name": "my-app",
  "version": "1.0.0",
  "dependencies": {
    "express": "^4.19.2",
    "lodash": "~4.17.21",
    "axios": ">=1.7.0"
  },
  "devDependencies": {
    "jest": "^29.7.0",
    "@types/node": "^22.0.0"
  }
}
```

## Version Constraint Preservation

The npm integration preserves version constraint prefixes:

| Constraint | Meaning | Example Before | Example After |
|------------|---------|----------------|---------------|
| `^` | Compatible with (minor) | `^4.18.0` | `^4.19.2` |
| `~` | Approximately equivalent | `~4.17.20` | `~4.17.21` |
| `>=` | Greater than or equal | `>=0.27.0` | `>=1.7.0` |
| `>` | Greater than | `>0.27.0` | `>1.7.0` |
| `<=` | Less than or equal | `<=1.0.0` | `<=1.0.0` |
| `<` | Less than | `<2.0.0` | `<2.0.0` |
| `=` | Exact version | `=1.0.0` | `=1.5.0` |
| (none) | Exact version | `1.0.0` | `1.5.0` |

**Note**: Only the version number is updated, the constraint operator is preserved.

## CLI Usage

### Scan for npm Dependencies

```bash
uptool scan --only=npm
```

Output:

```text
Type                 Path                Dependencies
----------------------------------------------------------------
npm                  package.json        12
npm                  apps/web/package.json    8
npm                  apps/api/package.json    15

Total: 3 manifests
```

### Plan npm Updates

```bash
uptool plan --only=npm
```

Output:

```text
package.json (npm):
Package          Current         Target          Impact
--------------------------------------------------------
express          ^4.18.0         ^4.19.2         patch
lodash           ~4.17.20        ~4.17.21        patch
axios            >=0.27.0        >=1.7.0         major

apps/web/package.json (npm):
Package          Current         Target          Impact
--------------------------------------------------------
react            ^18.2.0         ^18.3.1         minor
react-dom        ^18.2.0         ^18.3.1         minor

Total: 5 updates across 2 manifests
```

### Apply npm Updates

```bash
# Dry run first
uptool update --only=npm --dry-run --diff

# Apply updates
uptool update --only=npm

# Then regenerate lockfile
npm install
```

## Monorepo Support

The npm integration supports npm workspaces:

```tree
my-monorepo/
├── package.json           # Root workspace
├── package-lock.json
└── packages/
    ├── app/
    │   └── package.json   # Workspace member
    ├── lib/
    │   └── package.json   # Workspace member
    └── utils/
        └── package.json   # Workspace member
```

**Each `package.json` is updated independently**.

After updating, run `npm install` at the root to update the root lockfile.

## Configuration

### Update Policy

Control which updates are applied:

```yaml
# uptool.yaml
version: 1

integrations:
  - id: npm
    enabled: true
    policy:
      update: minor              # none, patch, minor, major
      allow_prerelease: false    # Don't update to pre-release versions
      pin: false                 # Don't pin to exact versions
```

**Update Levels**:

- `none` - No updates
- `patch` - Only patch updates (1.0.0 → 1.0.1)
- `minor` - Patch + minor updates (1.0.0 → 1.1.0)
- `major` - All updates including major (1.0.0 → 2.0.0)

### Exclude Specific Packages

To exclude specific packages, use lockfiles or version pinning:

```json
{
  "dependencies": {
    "old-package": "1.0.0"    // Exact version, won't update
  }
}
```

Or use npm's `overrides` field:

```json
{
  "overrides": {
    "old-package": "1.0.0"
  }
}
```

## Registry API

The npm integration queries the npm Registry API:

**Endpoint**: `https://registry.npmjs.org/{package}`

**Example**:

```bash
curl https://registry.npmjs.org/express
```

Returns:

```json
{
  "name": "express",
  "dist-tags": {
    "latest": "4.19.2"
  },
  "versions": {
    "4.19.2": { ... },
    "4.19.1": { ... },
    ...
  }
}
```

## Private Registries

uptool respects npm's registry configuration:

### npm config

```bash
# Set private registry
npm config set registry https://registry.company.com/

# Or use .npmrc
echo "registry=https://registry.company.com/" > .npmrc
```

### Authentication

uptool does NOT handle npm authentication. Configure npm auth separately:

```bash
# Login to private registry
npm login --registry=https://registry.company.com/

# Or use auth token in .npmrc
echo "//registry.company.com/:_authToken=TOKEN" >> ~/.npmrc
```

## Limitations

1. **No lockfile updates**: uptool only updates `package.json`, not `package-lock.json`
   - Solution: Run `npm install` after updating

2. **No workspace dependency resolution**: Workspaces are updated independently
   - Solution: Run `npm install` at root to resolve cross-workspace dependencies

3. **No peer dependency warnings**: uptool doesn't warn about peer dependency conflicts
   - Solution: Run `npm install` to see peer dependency warnings

4. **No package availability check**: uptool assumes packages exist in the registry
   - Solution: Check `npm install` output for missing packages

## Troubleshooting

### Updates Not Applied

**Problem**: `uptool plan` shows updates but `uptool update` doesn't apply them

**Causes**:

1. Policy restrictions in `uptool.yaml`
2. Version constraints don't allow the update

**Solutions**:

```bash
# Check policy
cat uptool.yaml

# Try with different policy
uptool update --only=npm   # Uses config policy

# Check version constraints
cat package.json
```

### Registry Errors

**Problem**: "Failed to fetch package info from registry"

**Causes**:

1. Network connectivity issues
2. Private registry authentication failed
3. Package doesn't exist

**Solutions**:

```bash
# Test registry connectivity
npm view express

# Check registry configuration
npm config get registry

# Test authentication
npm whoami
```

### Lockfile Out of Sync

**Problem**: After updating, `npm install` reports conflicts

**Solution**:

```bash
# Delete lockfile and reinstall
rm package-lock.json
npm install

# Or use npm ci for clean install
rm -rf node_modules package-lock.json
npm ci
```

## Best Practices

1. **Always regenerate lockfile**:

   ```bash
   uptool update --only=npm
   npm install
   git add package.json package-lock.json
   git commit -m "chore(deps): update npm dependencies"
   ```

2. **Test after updating**:

   ```bash
   npm test
   npm run build
   ```

3. **Review major updates carefully**:

   ```bash
   # Plan first to see impact
   uptool plan --only=npm

   # Check BREAKING CHANGES in changelogs
   ```

4. **Use separate PRs for major updates**:

   ```bash
   # Minor/patch updates together
   uptool update --only=npm  # (with policy: minor)

   # Major updates separately
   # (requires manual policy override)
   ```

5. **Pin critical dependencies**:

   ```json
   {
     "dependencies": {
       "critical-package": "1.2.3"  // Exact version
     }
   }
   ```

## See Also

- [npm Registry API](https://github.com/npm/registry/blob/master/docs/REGISTRY-API.md)
- [npm Workspaces](https://docs.npmjs.com/cli/v8/using-npm/workspaces)
- [Semantic Versioning](https://semver.org/)
- [Manifest Files Reference](../manifests.md)

# Version Management

Automated semantic versioning via conventional commits and GitHub Actions.

## Overview

1. **Conventional commits** determine version bumps
2. **GitHub Actions** calculate and apply versions
3. **`bump-my-version`** updates all files
4. **`internal/version/VERSION`** is the single source of truth

## Conventional Commits

| Type | Bump | Example |
|------|------|---------|
| `feat:` | Minor (0.1.0 → 0.2.0) | `feat: add Python integration` |
| `fix:` | Patch (0.1.0 → 0.1.1) | `fix: handle empty manifests` |
| `feat!:` or `BREAKING CHANGE:` | Major (0.1.0 → 1.0.0) | `feat!: redesign API` |
| `docs:`, `chore:`, `test:` | None | `docs: update README` |

**Format**: `<type>(<scope>): <subject>`

**Breaking changes**: Use `feat!:` or add `BREAKING CHANGE:` footer

## Local Development

```bash
# Show current version
mise run version-show

# Bump for testing only (don't commit to PRs)
mise run version-bump-patch   # 0.1.0 → 0.1.1
mise run version-bump-minor   # 0.1.0 → 0.2.0
mise run version-bump-major   # 0.1.0 → 1.0.0
```

Production releases are automated via GitHub Actions.

## Release Process

### Pre-Release

1. Maintainer triggers pre-release workflow
2. System calculates version from commits
3. Approval gate (designated reviewers)
4. Creates pre-release (e.g., `v0.2.0-rc1`, `v0.2.0-beta3`, `v0.2.0-alpha20250708`)
5. Builds artifacts

### Stable Release

1. Maintainer triggers promote workflow
2. Extracts stable version (`v0.2.0-rc1` → `v0.2.0`)
3. Approval gate (multiple reviewers)
4. Promotes artifacts
5. Updates CHANGELOG

See [environments.md](environments.md) for approval gate setup.

## Version Tags

**Immutable** (never change):

- Stable: `v0.1.0`, `v0.2.0`, `v1.13.0`
- Pre-release: `v0.2.0-rc1`, `v1.0.0-beta3`, `v1.13.0-alpha20250708`

**Mutable** (auto-updated for GitHub Actions):

- `v0` → latest `v0.x.x` stable
- `v0.1` → latest `v0.1.x` patch
- `v0-rc`, `v0-beta`, `v0-alpha` → latest `v0.x.x` pre-release of each type

**Usage**:

```yaml
# Recommended
- uses: santosr2/uptool@v0

# Pin to minor
- uses: santosr2/uptool@v0.1

# Pin to exact version
- uses: santosr2/uptool@v0.1.0
```

## Files Updated Automatically

When versions change, `bump-my-version` updates:

- `internal/version/VERSION` - Source of truth
- `README.md` - Action version examples
- `SECURITY.md` - Supported versions
- `docs/action-usage.md` - Usage examples

## Configuration

**Location**: `.bumpversion.toml`

Defines version format, files to update, and search/replace patterns.

## Workflows

- `.github/workflows/pre-release.yml` - Create RC/beta/alpha releases
- `.github/workflows/promote-release.yml` - Promote to stable
- `.github/workflows/patch-release.yml` - Security/bugfix patches
- `.github/workflows/create-release-branch.yml` - Branch for patches

## Best Practices

1. **Contributors**: Write conventional commits, don't bump versions manually
2. **Maintainers**: Use workflows, never push version tags directly
3. **Testing**: Use local bump commands, don't commit version changes
4. **Pre-releases**: Test thoroughly before promoting to stable

## See Also

- [CONTRIBUTING.md](CONTRIBUTING.md) - Conventional commit guidelines
- [environments.md](environments.md) - Approval gate configuration
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Semantic Versioning](https://semver.org/)

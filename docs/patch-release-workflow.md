# Patch Release Workflow Guide

This guide explains how to manage security patches and bug fixes for previous minor versions of uptool, in accordance with our [Security Policy](https://github.com/santosr2/uptool/blob/{{ extra.uptool_version }}/SECURITY.md).

## Overview

uptool supports **multiple minor versions simultaneously**:

- **Latest minor version** (e.g., 0.2.x): Full support (features, bug fixes, security patches)
- **Previous minor version** (e.g., 0.1.x): Security patches only for 6 months after next minor release
- **Older versions**: No support

## Workflow Files

### 1. Create Release Branch (`create-release-branch.yml`)

**Purpose**: Create a long-lived release branch for a minor version to enable backporting.

**When to use**: After releasing a new minor version (e.g., after v0.2.0 is released, create `release-0.1` for backports).

**How to use**:

1. Go to **Actions** → **Create Release Branch**
2. Click **Run workflow**
3. Enter the minor version (e.g., `0.1`)
4. Click **Run workflow**

**What it does**:

- Creates a `release-X.Y` branch from the latest `vX.Y.Z` tag
- Adds branch protection rules
- Creates a README explaining the branch purpose

**Example**:

```text
Input: 0.1
Result: Creates branch `release-0.1` from tag `v0.1.0`
```

### 2. Patch Release (`patch-release.yml`)

**Purpose**: Create a patch release (e.g., v0.1.1) from a release branch with security or bug fixes.

**When to use**: After cherry-picking fixes to a release branch.

**How to use**:

1. Cherry-pick fixes to the release branch:

   ```bash
   git checkout release-0.1
   git cherry-pick <commit-sha>
   git push origin release-0.1
   ```

2. Go to **Actions** → **Patch Release**
3. Click **Run workflow**
4. Configure:
   - **Release branch**: `release-0.1`
   - **Patch type**: `security` or `bugfix`
   - **Changelog notes** (optional): Additional context
5. Click **Run workflow**

**What it does**:

- Calculates the next patch version (e.g., 0.1.0 → 0.1.1)
- Updates version files
- Builds and signs binaries for all platforms
- Generates SBOM
- Creates a GitHub release
- Updates the mutable minor tag (e.g., `v0.1` → `v0.1.1`)

**Example**:

```text
Branch: release-0.1
Current: v0.1.0
Next: v0.1.1 (security patch)
```

### 3. Coordinate Security Patches (`security-patch.yml`)

**Purpose**: Automate cherry-picking security fixes to multiple release branches.

**When to use**: When a security vulnerability affects multiple versions.

**How to use**:

1. Fix the vulnerability on `main` and note the commit SHA(s)

2. Go to **Actions** → **Coordinate Security Patches**
3. Click **Run workflow**
4. Configure:
   - **Advisory ID**: `GHSA-xxxx-xxxx-xxxx` (if applicable)
   - **Affected versions**: `0.1.x,0.2.x` (comma-separated)
   - **Severity**: `critical`, `high`, `medium`, or `low`
   - **Description**: Brief explanation of the vulnerability
   - **Fix commits**: `abc123,def456` (commit SHAs from main)
5. Click **Run workflow**

**What it does**:

- Identifies which release branches need patches
- Creates patch branches for each affected version
- Cherry-picks the fix commits
- Creates PRs for each release branch
- If cherry-pick fails, creates an issue for manual backporting

**Example**:

```text
Affected: 0.1.x, 0.2.x
Commits: abc123, def456
Result:
  - PR to release-0.1
  - PR to release-0.2
```

## Quick Examples

**Security patch**: Fix on main → Run "Coordinate Security Patches" workflow → Review automated PRs → Run "Patch Release" for each branch → Publish advisory

**Bug fix**: Cherry-pick to release branch → Run "Patch Release" workflow

**New release branch**: Automatically created when promoting to stable (or use "Create Release Branch" workflow)

## Release Branch Management

### Branch Naming

- Format: `release-X.Y`
- Examples: `release-0.1`, `release-1.0`, `release-1.2`

### Branch Protection

Release branches have the same protection as `main`:

- Required status checks
- Required PR reviews
- No force pushes
- No deletions

### Support Timeline

```text
Timeline: Support for release-0.1

0.1.0 released ─────────── 0.2.0 released ──────────── +6 months ──────────>
     │                            │                            │
     ├─ Full support              ├─ Security patches only     ├─ Archive
     │                            │                            │
release-0.1 created               │                       End of support
```

### End of Support

When a release branch reaches end of support:

1. **Announce end of support** (1 month before)
2. **Final patch release** (if needed)
3. **Archive the branch**:

   ```bash
   git tag archive/release-0.1 release-0.1
   git push origin archive/release-0.1
   ```

4. **Update documentation**
5. **Close remaining PRs/issues** for that branch

## Versioning

### Version Tags

- **Immutable tags**: `v0.1.0`, `v0.1.1`, `v0.2.0` (never change)
- **Mutable tags**: `v0.1`, `v0.2`, `v0` (updated with each patch)

### Tag Updates

When creating a patch release (e.g., v0.1.1):

1. Create immutable tag: `v0.1.1`
2. Update mutable minor tag: `v0.1` → `v0.1.1`
3. Do NOT update major tag: `v0` stays at latest minor (e.g., `v0.2.0`)

### GitHub Action Pinning

Users can pin to different levels:

```yaml
# Exact version (most secure, no automatic updates)
- uses: santosr2/uptool@{{ extra.uptool_version }}

# Minor version (gets security patches automatically)
- uses: santosr2/uptool@{{ extra.uptool_version_minor }}

# Major version (gets all updates in v0.x)
- uses: santosr2/uptool@{{ extra.uptool_version_major }}
```

## Troubleshooting

**Cherry-pick conflicts**: Check created issue for manual backport instructions, resolve conflicts manually, run Patch Release workflow

**Missing release branch**: Run "Create Release Branch" workflow first

**Failed workflow**: Check logs, fix issues on release branch, re-run

**Wrong version**: Delete tag, fix version files, re-run workflow

## Best Practices

- Only backport critical security fixes and bugs
- Test patches thoroughly before release
- Announce security updates via advisories and releases
- Follow semantic versioning strictly
- Never change immutable tags or mix features with patches

## FAQ

**When to create release branch?** After releasing a new minor version (automated via Promote workflow)

**Support duration?** 6 months security patches after next minor release

**Backport features?** No, only security fixes and critical bugs

**Cherry-pick conflicts?** Workflow creates an issue with manual instructions

**Update mutable tags?** Automatic via workflow

## Related Documentation

- [Security Policy](https://github.com/santosr2/uptool/blob/{{ extra.uptool_version }}/SECURITY.md) - Support timelines and reporting
- [Contributing Guide](https://github.com/santosr2/uptool/blob/{{ extra.uptool_version }}/CONTRIBUTING.md) - Development workflow
- [Versioning Guide](versioning.md) - Semantic versioning details
- [Release Process](https://github.com/santosr2/uptool/blob/{{ extra.uptool_version }}/CONTRIBUTING.md#release-process) - Main branch releases

## Support

For questions about patch releases:

- **GitHub Discussions**: <https://github.com/santosr2/uptool/discussions>
- **Security Issues**: <https://github.com/santosr2/uptool/security/advisories>

---

**Last Updated**: 2025-01-16
**Maintained By**: uptool maintainers

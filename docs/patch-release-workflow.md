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

## Complete Workflow Examples

### Example 1: Security Vulnerability in Multiple Versions

**Scenario**: Critical security vulnerability affects 0.1.x and 0.2.x

**Steps**:

1. **Fix on main**:

   ```bash
   git checkout main
   # Fix the vulnerability
   git commit -m "fix(security): resolve CVE-2024-XXXXX"
   git push origin main
   # Note the commit SHA: abc123
   ```

2. **Ensure release branches exist**:
   - Check if `release-0.1` and `release-0.2` exist
   - If not, use **Create Release Branch** workflow

3. **Coordinate patches**:
   - Run **Coordinate Security Patches** workflow
   - Affected versions: `0.1.x,0.2.x`
   - Fix commits: `abc123`
   - Severity: `critical`

4. **Review and merge PRs**:
   - Review the automated PRs created for each branch
   - Ensure tests pass
   - Merge PRs

5. **Create patch releases**:
   - Run **Patch Release** workflow for `release-0.1`
   - Run **Patch Release** workflow for `release-0.2`
   - Patch type: `security`

6. **Announce**:
   - Publish security advisory
   - Notify users to upgrade
   - Update affected versions in advisory

### Example 2: Manual Bug Fix to Previous Version

**Scenario**: User reports critical bug in 0.1.x, but main is on 0.2.x

**Steps**:

1. **Reproduce and fix**:

   ```bash
   git checkout main
   # Fix the bug
   git commit -m "fix: critical bug in dependency parsing"
   git push origin main
   # Note the commit SHA: def456
   ```

2. **Cherry-pick to release branch**:

   ```bash
   git checkout release-0.1
   git cherry-pick def456
   git push origin release-0.1
   ```

3. **Create patch release**:
   - Run **Patch Release** workflow
   - Release branch: `release-0.1`
   - Patch type: `bugfix`

### Example 3: Creating a New Release Branch

**Scenario**: Just released v0.2.0, need to support 0.1.x for 6 months

**✨ Automated Approach** (Recommended):

When you run the **Promote to Stable Release** workflow for v0.2.0, it will **automatically** create the `release-0.1` branch. No manual action needed!

**Manual Approach** (if needed):

1. **Verify latest stable tag**:

   ```bash
   git tag -l "v0.1.*" | grep -v '\-' | sort -V | tail -n 1
   # Output: v0.1.0
   ```

2. **Create release branch**:
   - Run **Create Release Branch** workflow
   - Version: `0.1`

3. **Verify branch**:

   ```bash
   git fetch origin
   git checkout release-0.1
   cat .github/RELEASE_BRANCH_README.md
   ```

4. **Set calendar reminder**:
   - 6 months from v0.2.0 release date
   - Archive `release-0.1` when support ends

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
- uses: santosr2/uptool@v0.1.0

# Minor version (gets security patches automatically)
- uses: santosr2/uptool@v0.1

# Major version (gets all updates in v0.x)
- uses: santosr2/uptool@v0
```

## Troubleshooting

### Cherry-pick Conflicts

If **Coordinate Security Patches** fails with conflicts:

1. Check the created issue for manual backport instructions
2. Manually resolve conflicts:

   ```bash
   git checkout release-0.1
   git cherry-pick abc123
   # Resolve conflicts
   git add .
   git cherry-pick --continue
   git push origin release-0.1
   ```

3. Run **Patch Release** workflow

### Missing Release Branch

If you try to patch a version without a release branch:

1. Run **Create Release Branch** first
2. Then proceed with patching

### Failed Patch Release

If **Patch Release** workflow fails:

1. Check workflow logs for errors
2. Common issues:
   - Version files not updated correctly
   - Build failures
   - Test failures
3. Fix issues on the release branch
4. Re-run workflow

### Wrong Version Number

If the patch version is wrong:

1. Delete the tag (if created):

   ```bash
   git tag -d v0.1.1
   git push origin :refs/tags/v0.1.1
   ```

2. Fix version files manually
3. Re-run workflow

## Best Practices

### 1. Security Patches

- ✅ **Do**: Use automated workflows for consistency
- ✅ **Do**: Test patches thoroughly before release
- ✅ **Do**: Announce security updates prominently
- ❌ **Don't**: Delay security patches
- ❌ **Don't**: Mix features with security fixes

### 2. Bug Fixes

- ✅ **Do**: Only backport critical bugs
- ✅ **Do**: Document why the backport is necessary
- ❌ **Don't**: Backport non-critical bugs
- ❌ **Don't**: Introduce new features in patches

### 3. Communication

- ✅ **Do**: Update security advisories with fixed versions
- ✅ **Do**: Notify users via GitHub releases
- ✅ **Do**: Document in CHANGELOG.md
- ❌ **Don't**: Silently release security fixes

### 4. Version Management

- ✅ **Do**: Follow semantic versioning strictly
- ✅ **Do**: Keep mutable tags updated
- ✅ **Do**: Document support timelines
- ❌ **Don't**: Change immutable tags
- ❌ **Don't**: Extend support without announcement

## FAQ

### When should I create a release branch?

After releasing a new minor version. For example:

- Release v0.2.0 → Create `release-0.1` to support 0.1.x

### How long are release branches supported?

Security patches only, for 6 months after the next minor release.

### Can I backport features?

No. Only security fixes and critical bug fixes. Features go to `main` only.

### What if a security fix doesn't apply cleanly?

The **Coordinate Security Patches** workflow will create an issue with manual instructions.

### How do I know which versions need patches?

Check the security advisory for affected versions. Example:

- Affected: < 0.2.3
- Needs patches: 0.1.x (via release-0.1)

### Should I update mutable tags for patches?

The workflow automatically updates the minor tag (e.g., `v0.1`). You don't need to manually update it.

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

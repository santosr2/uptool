# Version Management

uptool uses **automated semantic versioning** based on conventional commits and GitHub Actions workflows. This document explains the complete version management system.

## Overview

Version management in uptool is fully automated:

1. **Conventional commits** determine version bumps
2. **GitHub Actions** calculate and apply versions
3. **`bump-my-version`** updates all files consistently
4. **VERSION file** is the single source of truth

## Version Sources

### Single Source of Truth

The canonical version is stored in `internal/version/VERSION`:

```text
0.1.0
```

This file is:

- Embedded into Go binaries at build time
- Updated automatically by GitHub Actions
- Used by `bump-my-version` to update all documentation

### Version Embedding

The CLI reads the version from the embedded VERSION file:

```go
// internal/version/version.go
package version

import (
    _ "embed"
    "strings"
)

//go:embed VERSION
var versionFile string

var version = strings.TrimSpace(versionFile)

func Get() string {
    if version == "" {
        return "dev"
    }
    return version
}
```

When you build the binary:

```bash
go build ./cmd/uptool
./uptool --version
# Output: uptool version 0.1.0
```

## Conventional Commits

uptool follows the [Conventional Commits](https://www.conventionalcommits.org/) specification to determine version bumps.

### Commit Format

```text
<type>(<scope>): <subject>

[optional body]

[optional footer(s)]
```

### Version Bump Rules

| Commit Type | Example | Version Bump | Example |
|-------------|---------|--------------|---------|
| `feat:` | `feat: add Python integration` | **Minor** | 0.1.0 → 0.2.0 |
| `fix:` | `fix: handle empty manifests` | **Patch** | 0.1.0 → 0.1.1 |
| `BREAKING CHANGE:` | `feat!: redesign API` | **Major** | 0.1.0 → 1.0.0 |
| `chore:`, `docs:`, etc. | `docs: update README` | **None** | No bump |

### Supported Types

- **`feat:`** - New feature (minor bump)
- **`fix:`** - Bug fix (patch bump)
- **`docs:`** - Documentation only (no bump)
- **`style:`** - Formatting, no code change (no bump)
- **`refactor:`** - Code change that neither fixes nor adds (no bump)
- **`perf:`** - Performance improvement (patch bump)
- **`test:`** - Adding/updating tests (no bump)
- **`chore:`** - Maintenance tasks (no bump)
- **`ci:`** - CI/CD changes (no bump)
- **`revert:`** - Revert previous commit (patch bump)

### Breaking Changes

To trigger a major version bump, use either:

**Method 1: Exclamation mark**

```bash
git commit -m "feat!: redesign integration interface"
```

**Method 2: Footer**

```bash
git commit -m "feat: redesign integration interface

BREAKING CHANGE: Integration interface now requires Validate method"
```

### Examples

**Feature Addition (Minor Bump)**:

```bash
git commit -m "feat(integrations): add Python integration

Add support for pyproject.toml, requirements.txt, and Pipfile.
Includes version resolution via PyPI API and TOML rewriting."
```

→ **0.1.0 → 0.2.0**

**Bug Fix (Patch Bump)**:

```bash
git commit -m "fix(npm): handle empty devDependencies

Previously crashed when package.json had missing devDependencies section."
```

→ **0.1.0 → 0.1.1**

**Breaking Change (Major Bump)**:

```bash
git commit -m "feat!: redesign configuration format

BREAKING CHANGE: uptool.yaml now requires version: 2 and uses
different policy structure. See migration guide in docs/."
```

→ **0.1.0 → 1.0.0**

**Documentation (No Bump)**:

```bash
git commit -m "docs: add versioning guide"
```

→ **No version change**

## Automated Release Process

### Pre-Release Workflow

**Location**: `.github/workflows/pre-release.yml`

**Trigger**: Manual dispatch (workflow_dispatch)

**Steps**:

1. **Calculate Version**
   - Uses `mathieudutour/github-tag-action` to analyze commits since last tag
   - Determines semantic version bump (major/minor/patch)
   - Calculates pre-release version with type suffix

2. **Update VERSION Files**
   - Uses `bump-my-version` to update:
     - `internal/version/VERSION`
     - All documentation version references
     - README.md examples
     - SECURITY.md policy
     - docs/action-usage.md

3. **Commit Changes**
   - Commits version updates to `main` branch
   - Commit message: `chore(release): bump version to v0.2.0-rc.1 [skip ci]`

4. **Run Tests**
   - Checks out updated code
   - Runs full test suite
   - Runs `go vet`

5. **Approval Gate** ⚠️
   - Workflow pauses for manual approval
   - Uses GitHub Environment: `pre-release`
   - Requires approval from designated reviewers
   - See [Environment Setup](environments.md) for configuration

6. **Build Artifacts** (after approval)
   - Builds for all platforms (Linux, macOS, Windows)
   - Generates checksums (SHA256)
   - Creates SBOM files (SPDX + CycloneDX)
   - Signs with Cosign (keyless)

7. **Create Pre-Release**
   - Creates Git tag (e.g., `v0.2.0-rc.1`)
   - Generates changelog from commits
   - Uploads all artifacts to GitHub Release
   - Marks as pre-release

### Running a Pre-Release

1. Go to **Actions** → **Pre-Release**
2. Click **Run workflow**
3. Select pre-release type:
   - **rc** - Release candidate (recommended for testing before stable)
   - **beta** - Beta version (earlier testing phase)
   - **alpha** - Alpha version (very early, unstable)
4. Click **Run workflow**
5. **Wait for approval** ⚠️
   - Workflow will pause after tests complete
   - Designated reviewers receive notification
   - Review the changes and approve/reject
   - See [Approving Deployments](environments.md#approving-a-deployment)

**Example**:

- Current version: `v0.1.0`
- Commits since last tag: 2 `feat:` commits, 1 `fix:` commit
- Pre-release type: `rc`
- **Result**: `v0.2.0-rc.1` (feat causes minor bump)

If you run another pre-release without new commits:

- **Result**: `v0.2.0-rc.2` (increments rc number)

### Promote to Stable Workflow

**Location**: `.github/workflows/promote-release.yml`

**Trigger**: Manual dispatch with pre-release tag input

**Steps**:

1. **Validate Pre-Release**
   - Checks that pre-release tag exists
   - Extracts stable version (`v0.2.0-rc.1` → `v0.2.0`)
   - Verifies stable tag doesn't already exist

2. **Update to Stable Version**
   - Uses `bump-my-version` to remove pre-release suffix
   - Updates `internal/version/VERSION` to `0.2.0`
   - Updates all documentation

3. **Commit Changes**
   - Commits version updates to `main` branch
   - Commit message: `chore(release): release v0.2.0 [skip ci]`

4. **Run Final Tests**
   - Runs full test suite on stable version
   - Verifies build works

5. **Approval Gate** ⚠️
   - Workflow pauses for manual approval
   - Uses GitHub Environment: `production`
   - Requires approval from designated reviewers (recommended: 2+)
   - See [Environment Setup](environments.md) for configuration

6. **Promote Release** (after approval)
   - Downloads all artifacts from pre-release
   - Creates Git tag (e.g., `v0.2.0`)
   - Creates stable GitHub Release
   - Marks as latest release
   - Updates CHANGELOG.md

### Running a Stable Release

1. Go to **Actions** → **Promote to Stable Release**
2. Click **Run workflow**
3. Enter pre-release tag (e.g., `v0.2.0-rc.1`)
4. Click **Run workflow**
5. **Wait for approval** ⚠️
   - Workflow will pause after tests complete
   - Designated reviewers receive notification
   - Review the pre-release testing results
   - Approve/reject the promotion to stable
   - See [Approving Deployments](environments.md#approving-a-deployment)

**The system automatically** (after approval):

- Detects stable version (`v0.2.0`)
- Updates VERSION file
- Promotes artifacts
- Updates CHANGELOG
- Creates stable release

## Local Version Management

For local development and testing, you can manually manage versions using Mise tasks targets.

### Show Current Version

```bash
mise run version-show
# Output:
# Current version:
# 0.1.0
```

### Bump Version Locally

**Important**: Only use these for local testing. Production releases are automated via GitHub Actions.

```bash
# Patch bump (0.1.0 → 0.1.1)
mise run version-bump-patch

# Minor bump (0.1.0 → 0.2.0)
mise run version-bump-minor

# Major bump (0.1.0 → 1.0.0)
mise run version-bump-major
```

These commands use `bump-my-version` to update:

- `internal/version/VERSION`
- `README.md`
- `SECURITY.md`
- `docs/action-usage.md`

**After bumping locally**:

```bash
# Rebuild to test
mise run build
./dist/uptool --version

# Commit changes
git add .
git commit -m "chore: bump version to 0.2.0"
```

## File Update Mechanism

### bump-my-version Configuration

**Location**: `.bumpversion.toml`

```toml
[tool.bumpversion]
current_version = "0.1.0"
parse = "(?P<major>\\d+)\\.(?P<minor>\\d+)\\.(?P<patch>\\d+)(\\-(?P<pre_label>rc|beta|alpha)\\.(?P<pre_num>\\d+))?"
serialize = [
    "{major}.{minor}.{patch}-{pre_label}.{pre_num}",
    "{major}.{minor}.{patch}",
]

# Source of truth
[[tool.bumpversion.files]]
filename = "internal/version/VERSION"

# Documentation
[[tool.bumpversion.files]]
filename = "README.md"
search = "@v{current_version}"
replace = "@v{new_version}"

# ... more files ...
```

### Files Updated Automatically

When `bump-my-version` runs, it updates:

1. **internal/version/VERSION** - Single source of truth
2. **README.md** - GitHub Action version examples
3. **SECURITY.md** - Supported version policy
4. **docs/action-usage.md** - Action usage examples

All updates preserve file formatting and context.

## Version Tagging

### Tag Format

uptool uses **semantic versioning** with `v` prefix and supports both **immutable** and **mutable** tags:

#### Immutable Tags (Never Change)

- Stable releases: `v0.1.0`, `v0.2.0`, `v1.0.0`
- Pre-releases: `v0.2.0-rc.1`, `v1.0.0-beta.2`, `v0.3.0-alpha.1`

#### Mutable Tags (Auto-Updated)

For GitHub Actions convenience, mutable tags are automatically maintained:

**Stable Release Tags**:

- `v0` - Always points to latest `v0.x.x` stable release
- `v0.1` - Always points to latest `v0.1.x` patch release
- `v1` - Always points to latest `v1.x.x` stable release

**Pre-Release Tags**:

- `v0-rc` - Always points to latest release candidate in v0
- `v0.1-rc` - Always points to latest release candidate in v0.1
- `v0-beta` - Always points to latest beta in v0
- `v0.1-beta` - Always points to latest beta in v0.1

### Tag Creation

Tags are created automatically by GitHub Actions:

#### Pre-release Workflow

Creates 3 tags:

```bash
# Immutable tag (exact version)
git tag -a "v0.2.0-rc.1" -m "Pre-release v0.2.0-rc.1"
git push origin "v0.2.0-rc.1"

# Mutable tags (force-updated to follow latest)
git tag -fa "v0.2-rc" -m "Latest rc pre-release in v0.2"
git push origin "v0.2-rc" --force

git tag -fa "v0-rc" -m "Latest rc pre-release in v0"
git push origin "v0-rc" --force
```

**Example**: If you release `v0.2.0-rc.1`, then `v0.2.0-rc.2`:

- `v0.2.0-rc.1` - Points to rc.1 (immutable)
- `v0.2.0-rc.2` - Points to rc.2 (immutable)
- `v0.2-rc` - Updates from rc.1 → rc.2 (mutable)
- `v0-rc` - Updates from rc.1 → rc.2 (mutable)

#### Promote Workflow

Creates 3 tags:

```bash
# Immutable tag (exact version)
git tag -a "v0.2.0" -m "Release v0.2.0 (promoted from v0.2.0-rc.1)"
git push origin "v0.2.0"

# Mutable tags (force-updated to follow latest)
git tag -fa "v0.2" -m "Latest stable release in v0.2 (currently v0.2.0)"
git push origin "v0.2" --force

git tag -fa "v0" -m "Latest stable release in v0 (currently v0.2.0)"
git push origin "v0" --force
```

**Example**: If you release `v0.1.0`, then `v0.2.0`, then `v0.2.1`:

- `v0.1.0`, `v0.2.0`, `v0.2.1` - Immutable exact versions
- `v0.1` - Points to v0.1.0 (immutable after v0.2.0 release)
- `v0.2` - Updates from v0.2.0 → v0.2.1 (mutable)
- `v0` - Updates from v0.1.0 → v0.2.0 → v0.2.1 (mutable)

### GitHub Actions Usage

These mutable tags make it easy to reference uptool in workflows:

```yaml
# Recommended: Pin to major version
- uses: santosr2/uptool@v0
  # Gets latest v0.x.x automatically (v0.1.0 → v0.2.0 → v0.2.1)

# Pin to minor version
- uses: santosr2/uptool@v0.2
  # Gets latest v0.2.x patches (v0.2.0 → v0.2.1 → v0.2.2)

# Pin to exact version (most secure)
- uses: santosr2/uptool@v0.2.0
  # Never changes

# Test with pre-releases
- uses: santosr2/uptool@v0-rc
  # Gets latest release candidate
```

### Viewing Tags

```bash
# List all tags
git tag

# List only stable releases (immutable)
git tag -l 'v[0-9]*.[0-9]*.[0-9]*' | grep -v '\-'

# List only pre-releases (immutable)
git tag -l 'v[0-9]*.[0-9]*.[0-9]*-*'

# List mutable stable tags
git tag -l 'v[0-9]*' | grep -v '\.' | grep -v '\-'
git tag -l 'v[0-9]*.[0-9]*' | grep -v '\.' | grep -v '\-'

# List mutable pre-release tags
git tag -l 'v*-rc' 'v*-beta' 'v*-alpha'

# Show tag details
git show v0.1.0
git show v0     # Shows what v0 currently points to
git show v0-rc  # Shows latest rc
```

### Tag Management Rules

**Immutable Tags** (created once, never modified):

- ✅ Exact versions: `v0.1.0`, `v1.2.3`
- ✅ Exact pre-releases: `v0.2.0-rc.1`, `v1.0.0-beta.2`

**Mutable Tags** (force-updated on each release):

- ⚠️ Major version: `v0`, `v1`
- ⚠️ Minor version: `v0.1`, `v1.2`
- ⚠️ Pre-release major: `v0-rc`, `v1-beta`
- ⚠️ Pre-release minor: `v0.1-rc`, `v1.2-alpha`

## CHANGELOG Management

The CHANGELOG is automatically generated from commit messages using `git-cliff`.

### Configuration

**Location**: `git-cliff.toml`

Defines:

- Commit grouping by type (`feat`, `fix`, etc.)
- Link generation for commits and comparisons
- Version extraction from tags
- Output format (Keep a Changelog style)

### Update Process

**Pre-release workflow**: Generates changelog section for release notes

**Promote workflow**: Updates CHANGELOG.md with new version:

```bash
git-cliff --config git-cliff.toml --tag "v0.2.0" --prepend CHANGELOG.md
```

### CHANGELOG Format

```markdown
# Changelog

All notable changes to this project will be documented in this file.

## [0.2.0] - 2025-01-15

### Features
- Add Python integration (#123)
- Support for pyproject.toml (#124)

### Bug Fixes
- Handle empty devDependencies in npm (#125)

## [0.1.0] - 2025-01-10

Initial release
```

## Release Approval Process

uptool uses **GitHub Environments** with approval gates to control releases. This provides an additional security layer and creates an audit trail for all releases.

### Overview

Both release workflows require manual approval before artifacts are created or promoted:

- **Pre-Release Workflow**: Requires approval before building and publishing pre-release artifacts
- **Promote Workflow**: Requires approval before promoting to stable release

### GitHub Environments

Two environments are configured:

#### pre-release Environment

- **Purpose**: Protect pre-release creation
- **Used by**: `.github/workflows/pre-release.yml`
- **Recommended reviewers**: 1+ maintainers
- **Approval required**: Before building artifacts

#### production Environment

- **Purpose**: Protect stable release promotion
- **Used by**: `.github/workflows/promote-release.yml`
- **Recommended reviewers**: 2+ senior maintainers
- **Approval required**: Before promoting to stable

### Setting Up Environments

Environments must be configured by repository administrators:

1. Go to **Settings** → **Environments**
2. Create `pre-release` and `production` environments
3. Configure protection rules:
   - Add required reviewers
   - Restrict to `main` branch only
   - Optional: Add wait timer

**Complete setup instructions**: See [Environment Setup Guide](environments.md)

### Approval Workflow

When a release workflow runs:

1. Workflow triggers and runs initial steps (calculate version, update files, test)
2. **Workflow pauses** at the environment-protected job
3. GitHub sends notification to required reviewers
4. Reviewers see:
   - Version being released
   - Commit SHA
   - Test results
   - Link to pending deployment
5. Reviewers must click **"Review deployments"** and approve/reject
6. If approved: Workflow continues with build/promotion
7. If rejected: Workflow is cancelled

### Reviewing a Release

Before approving a pre-release:

- [ ] Version number is correct based on commits
- [ ] All tests passed
- [ ] No critical issues in commits
- [ ] CHANGELOG will be updated
- [ ] Pre-release type matches intent (rc/beta/alpha)

Before approving a stable release:

- [ ] Pre-release was tested successfully
- [ ] No blocking issues reported
- [ ] Documentation is accurate
- [ ] All artifacts are present in pre-release
- [ ] Multiple reviewers agree (for production)

### Deployment History

GitHub tracks all deployments:

- Go to **Settings** → **Environments** → [Environment name]
- View deployment history showing:
  - Who triggered the workflow
  - Who approved/rejected
  - When deployment occurred
  - Links to workflow runs

This creates a complete audit trail for compliance and security reviews.

### Emergency Releases

For critical security fixes that need immediate release:

1. Ensure multiple reviewers are available
2. Follow normal process but expedite review
3. Document reason in approval comments:

   ```text
   Approved: Critical security fix for CVE-2025-XXXX
   Expedited due to active exploitation
   ```

**Never bypass approvals** - instead, add emergency contacts as reviewers.

## Version Support Policy

See [SECURITY.md](SECURITY.md) for the official support policy.

**General Policy**:

- **Latest minor version**: Full support with security patches
- **Previous minor version**: Security patches only (6 months)
- **Older versions**: No support

**Example** (current version: 0.2.x):

- ✅ `0.2.x` - Fully supported
- ⚠️ `0.1.x` - Security patches only (until July 2025)
- ❌ `< 0.1` - No support

## Troubleshooting

### Version Not Updating

**Problem**: Ran pre-release workflow but version didn't change

**Causes**:

1. No version-bumping commits since last tag
   - Solution: Ensure you have `feat:` or `fix:` commits
2. Only `chore:` or `docs:` commits
   - Solution: These don't trigger version bumps by design

**Check commits**:

```bash
# See commits since last tag
git log $(git describe --tags --abbrev=0)..HEAD --oneline

# Check commit types
git log $(git describe --tags --abbrev=0)..HEAD --pretty=format:"%s"
```

### Build Has Wrong Version

**Problem**: Built binary shows wrong version

**Solution**:

```bash
# Ensure VERSION file is up to date
cat internal/version/VERSION

# Rebuild completely
mise run clean
mise run build

# Verify
./dist/uptool --version
```

### Documentation Out of Sync

**Problem**: README shows different version than binary

**Solution**:

```bash
# Run bump-my-version manually
bump-my-version bump --new-version "0.2.0" patch

# Check git diff
git diff

# Commit if correct
git add .
git commit -m "chore: sync version across files"
```

## Best Practices

### 1. Use Conventional Commits

**Always** follow the conventional commit format:

```bash
# Good
git commit -m "feat(npm): add peer dependencies support"
git commit -m "fix(helm): handle missing Chart.lock"

# Bad
git commit -m "added feature"
git commit -m "bug fix"
```

### 2. Meaningful Commit Messages

Include context in the commit body:

```bash
git commit -m "feat(terraform): add provider version updates

Previously only module versions were updated. This adds support for
updating provider versions in required_providers blocks.

Closes #123"
```

### 3. Test Before Pre-Release

Before creating a pre-release:

```bash
# Run full test suite
mise run check

# Test locally
mise run build
./dist/uptool scan
./dist/uptool plan
```

### 4. Verify Pre-Release Before Promoting

Before promoting to stable:

1. Download and test pre-release artifacts
2. Run integration tests
3. Check documentation accuracy
4. Verify GitHub Action works

### 5. Document Breaking Changes

For major version bumps, provide migration guidance:

```bash
git commit -m "feat!: redesign configuration format

BREAKING CHANGE: Configuration format changed from version 1 to 2.
See docs/migration/v1-to-v2.md for upgrade guide.

Old format:
  integrations:
    - name: npm

New format:
  integrations:
    - id: npm
      enabled: true
"
```

## Reference

### Workflow Inputs

**Pre-Release Workflow**:

- `prerelease_type`: `rc` | `beta` | `alpha`

**Promote Workflow**:

- `pre_release_tag`: Tag to promote (e.g., `v0.2.0-rc.1`)

### Environment Variables

- `VERSION` - Current version (in workflows)
- `NEW_VERSION` - Calculated next version (in workflows)
- `NEW_TAG` - Git tag with `v` prefix (in workflows)

### Commands

```bash
# Show version
mise run version-show

# Bump versions locally
mise run version-bump-patch
mise run version-bump-minor
mise run version-bump-major

# View workflow runs
gh run list --workflow=pre-release.yml
gh run list --workflow=promote-release.yml

# Trigger pre-release
gh workflow run pre-release.yml -f prerelease_type=rc

# Trigger promotion
gh workflow run promote-release.yml -f pre_release_tag=v0.2.0-rc.1
```

## See Also

- [GitHub Environments Setup Guide](environments.md) - Configure approval gates for releases
- [Conventional Commits Specification](https://www.conventionalcommits.org/)
- [Semantic Versioning](https://semver.org/)
- [Keep a Changelog](https://keepachangelog.com/)
- [git-cliff Documentation](https://git-cliff.org/)
- [bump-my-version Documentation](https://github.com/callowayproject/bump-my-version)
- [GitHub Environments Documentation](https://docs.github.com/en/actions/deployment/targeting-different-environments/using-environments-for-deployment)

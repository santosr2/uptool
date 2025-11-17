# Security Policy

## Supported Versions

We release patches for security vulnerabilities for the following versions:

| Version | Supported          | Support Until |
| ------- | ------------------ | ------------- |
| 0.1.x   | :white_check_mark: | Current stable version |
| < 0.1   | :x:                | No support    |

**Support Policy**:

- **Latest minor version** (currently 0.1.x): Full support including features, bug fixes, and security patches
- **Previous minor version**: Security patches only for 6 months after the next minor release
- **Older versions**: No support

**Example** (when 0.2.0 is released):

- ✅ `0.2.x` - Full support
- ⚠️ `0.1.x` - Security patches only (until 6 months after 0.2.0 release)
- ❌ `< 0.1` - No support

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report them via one of these methods:

### Option 1: GitHub Security Advisories (Preferred)

1. Go to <https://github.com/santosr2/uptool/security/advisories>
2. Click "Report a vulnerability"
3. Fill in the details

This method allows secure, private discussion with maintainers.

### Option 2: Email

Email security concerns to: **<security@santosr2.dev>** (if applicable)

Include:

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

## What to Report

### In Scope

Report vulnerabilities in:

- **Code execution**: Arbitrary code execution through malicious manifests
- **Path traversal**: Reading/writing files outside the repository
- **Injection**: Command injection, YAML/JSON injection
- **Registry poisoning**: Man-in-the-middle or registry compromise scenarios
- **Credential exposure**: Leaking tokens, API keys, or credentials
- **Denial of service**: Resource exhaustion attacks

### Out of Scope

Please **do not** report:

- Vulnerabilities in third-party dependencies (report to those projects)
- Issues requiring physical access to the machine
- Social engineering attacks
- Theoretical vulnerabilities without proof of concept

## Response Timeline

- **Initial response**: Within 48 hours
- **Confirmation/triage**: Within 7 days
- **Fix development**: Depends on severity (critical: days, others: weeks)
- **Public disclosure**: After patch is released and users have time to update

## Security Best Practices

When using uptool:

### 1. Pin Action Versions

**Do**:

```yaml
# Pin to specific version (most secure)
- uses: santosr2/uptool@v0.1.0

# Or pin to major version (recommended for convenience)
- uses: santosr2/uptool@v0
```

**Don't**:

```yaml
# Unpinned (can break unexpectedly)
- uses: santosr2/uptool@main
```

**Versioning Scheme**:
uptool uses [semantic versioning](https://semver.org/) with automated releases based on conventional commits. Version tags follow the format `vMAJOR.MINOR.PATCH` (e.g., `v0.1.0`).

When pinning versions:

- **Major version** (`@v0`) - Gets latest minor/patch updates automatically (e.g., 0.1.0 → 0.2.0)
- **Minor version** (`@v0.1`) - Gets latest patch updates only (e.g., 0.1.0 → 0.1.1)
- **Exact version** (`@v0.1.0`) - No automatic updates (most secure)

### 2. Limit GitHub Token Permissions

Use minimal permissions:

```yaml
permissions:
  contents: write        # Only if creating commits
  pull-requests: write   # Only if creating PRs
```

### 3. Review Generated PRs

Always review dependency update PRs before merging:

- Check changelogs for breaking changes
- Verify version bump is expected
- Run CI/CD tests

### 4. Use uptool to Monitor Itself

uptool practices what it preaches by using itself to monitor its own dependencies.

**Automated monitoring**: The [`dependency-hygiene.yml`](../.github/workflows/dependency-hygiene.yml) workflow runs weekly to:

- Scan all dependency manifests using uptool
- Check for available updates
- Create issues for manual review
- Optionally create PRs automatically

**Manual review workflow**:

```bash
# View the latest dependency check
# Go to: Actions → Dependency Hygiene → Latest run

# To apply updates automatically:
# Actions → Dependency Hygiene → Run workflow → Set auto_update: true
```

**Local development**:

```bash
# Scan for updates
uptool scan

# Preview available updates
uptool plan

# Apply updates
uptool update --diff

# Commit and create PR
git add .
git commit -m "chore(deps): update dependencies"
git push
```

**What gets monitored**:

- Go modules (`go.mod`)
- Pre-commit hooks (`.pre-commit-config.yaml`)
- mise tools (`mise.toml`)
- GitHub Actions (workflow files)
- Any other manifests uptool supports

### 5. Validate Manifest Changes

Before applying updates:

```bash
# Always dry-run first
uptool update --dry-run --diff

# Review changes
git diff

# Then apply
uptool update --diff
```

## Known Security Considerations

### Registry Integrity

uptool queries public registries (npm, Terraform Registry, Helm repos, GitHub Releases). We:

- Use HTTPS for all registry communications
- Validate response structures
- Do not execute arbitrary code from registries

**User responsibility**: Ensure your network is secure and not compromised.

### File System Access

uptool:

- Reads manifest files in the repository
- Writes updated manifests
- Runs native commands (e.g., `pre-commit autoupdate`)

**Mitigation**: uptool does not traverse outside the repository root.

### Command Injection

For integrations using native commands (e.g., pre-commit):

- Commands are executed with fixed arguments
- No user input is passed unsanitized
- Working directory is controlled

**User responsibility**: Ensure your pre-commit hooks and other tools are trusted.

### GitHub Token Scope

The GitHub Action requires a token with write permissions. We:

- Use the token only for creating commits and PRs
- Do not log or expose the token
- Recommend using `GITHUB_TOKEN` (auto-scoped) over PATs

## Security Updates

When security patches are released:

1. GitHub Security Advisory is published
2. Fixed version is tagged and released
3. Users are notified via GitHub notifications
4. CHANGELOG.md documents the fix

## Vulnerability Disclosure Policy

After a vulnerability is fixed:

1. Coordinated disclosure with reporter
2. Public advisory published
3. CVE requested (if applicable)
4. Credit given to reporter (if desired)

## Contact

For security concerns:

- **GitHub Security Advisories**: <https://github.com/santosr2/uptool/security/advisories>
- **Email**: <security@santosr2.dev> (if applicable)

For general questions:

- **Discussions**: <https://github.com/santosr2/uptool/discussions>
- **Issues**: <https://github.com/santosr2/uptool/issues> (non-security bugs)

---

Thank you for helping keep uptool secure!

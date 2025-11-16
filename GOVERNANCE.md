# Governance

This document outlines the governance model for the uptool project.

## Project Leadership

### Maintainers

Maintainers have write access to the repository and are responsible for:
- Reviewing and merging pull requests
- Triaging issues
- Releasing new versions
- Setting project direction
- Maintaining code quality

**Current Maintainers**:
- [@santosr2](https://github.com/santosr2) (Creator, Lead Maintainer)

### Becoming a Maintainer

Contributors who demonstrate:
- Consistent, high-quality contributions
- Understanding of the codebase and architecture
- Collaborative spirit and good judgment
- Active participation in discussions

...may be invited to become maintainers.

## Development Workflow

### Trunk-Based Development

We use **trunk-based development**, not Git Flow:

- **Main branch**: `main` (protected)
- **Feature branches**: Short-lived, merge directly to `main`
- **No develop/staging branches**: Features merge when ready
- **Releases**: Tagged from `main`

```
main (protected)
├── feature/add-python-support ──┐
├── fix/helm-parsing-bug ────────┤
└── docs/update-readme ──────────┴─> Merge to main
```

### Why Trunk-Based?

- **Simpler**: One source of truth (`main`)
- **Faster**: No long-lived branches to sync
- **Better CI/CD**: Continuous integration to `main`
- **Easier**: No complex merging strategies

## Decision Making

### Types of Decisions

**Minor decisions** (bug fixes, small features, docs):
- Any maintainer can approve and merge
- Require 1 approval

**Major decisions** (breaking changes, new integrations, architecture):
- Discuss in an issue or discussion first
- Require consensus from maintainers
- May require 2+ approvals

**Project direction** (roadmap, major initiatives):
- Lead maintainer makes final call
- After community input via discussions

### Consensus Model

We aim for consensus, not voting. If disagreement:
1. Discuss in the PR/issue
2. Seek compromise
3. Lead maintainer makes final decision if needed

## Pull Request Process

### For Contributors

1. **Create PR** with clear title and description
2. **Link related issues** using "Fixes #123" or "Closes #456"
3. **Respond to feedback** in a timely manner
4. **Update as needed** based on review

### For Maintainers

1. **Review within 3 days** (best effort)
2. **Be constructive** in feedback
3. **Approve when ready** (tests pass, code quality good)
4. **Merge to main** after approval

### PR Requirements

- [x] Tests pass (when implemented)
- [x] Code formatted (`go fmt`)
- [x] GoDoc comments for exported APIs
- [x] Documentation updated (if needed)
- [x] Conventional commit messages

### Merge Strategy

**Squash and merge** (default):
- Keeps `main` history clean
- Preserves full history in PR

**Rebase and merge** (for multi-commit features):
- When commit history is meaningful
- Ask in PR if you want this

## Release Process

### Versioning

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR** (1.0.0): Breaking changes
- **MINOR** (0.x.0): New features, backward compatible
- **PATCH** (0.0.x): Bug fixes

Current stage: **v0.x.x** (pre-1.0, API may change)

### Release Cadence

- **Minor releases**: Every 2-4 weeks (when features are ready)
- **Patch releases**: As needed for bugs
- **Major releases**: When breaking changes are necessary

### Release Process

Releases are **fully automated** using GitHub Actions and conventional commits. See [docs/versioning.md](docs/versioning.md) for complete details.

**Pre-Release**:
1. Trigger Pre-Release workflow via GitHub Actions
2. Select pre-release type (rc/beta/alpha)
3. Workflow calculates version from commits
4. Tests run automatically
5. **Approval gate**: Reviewers must approve
6. Artifacts built and pre-release published

**Stable Release**:
1. Test pre-release thoroughly
2. Trigger Promote workflow via GitHub Actions
3. Provide pre-release tag to promote
4. **Approval gate**: Multiple reviewers must approve
5. Stable release published
6. CHANGELOG updated automatically

**Maintainers don't need to**:
- Manually update version numbers
- Create git tags
- Build artifacts
- Write changelog entries

**Everything is automated via conventional commits!**

## Code Review Standards

### What Reviewers Check

- **Correctness**: Does it work as intended?
- **Tests**: Are edge cases covered?
- **Style**: Follows Go conventions?
- **Documentation**: Is it clear and accurate?
- **Simplicity**: Is it the simplest solution?

### Review Tone

- Be kind and constructive
- Ask questions, don't demand changes
- Explain the "why" behind suggestions
- Celebrate good work

### Author Responsibilities

- Respond to feedback promptly
- Don't take criticism personally
- Ask for clarification if needed
- Update the PR based on feedback

## Community Standards

### Code of Conduct

**Be respectful**:
- Welcoming to newcomers
- Patient with questions
- Constructive in criticism
- Inclusive in language

**Be collaborative**:
- Assume good intentions
- Seek to understand before being understood
- Find common ground
- Help others succeed

### Communication Channels

- **GitHub Issues**: Bug reports, feature requests
- **GitHub Discussions**: Questions, ideas, help
- **Pull Requests**: Code contributions, review
- **README/docs**: Project information

## Conflict Resolution

If conflicts arise:

1. **Direct discussion**: Talk it out in the PR/issue
2. **Maintainer mediation**: Maintainers facilitate resolution
3. **Lead decision**: Lead maintainer makes final call if needed

## Contributions

All contributors retain copyright to their contributions under the project's MIT License.

By submitting a PR, you agree to license your contribution under MIT.

## Changing Governance

This governance model can evolve. To propose changes:

1. Open a discussion with your proposal
2. Gather feedback from community
3. Maintainers decide on adoption
4. Update this document

---

Questions about governance? Open a [Discussion](https://github.com/santosr2/uptool/discussions).

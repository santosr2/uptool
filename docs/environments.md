# GitHub Environments Setup

uptool uses **GitHub Environments** with approval gates to control releases. This provides an additional layer of security and creates an audit trail for all releases.

## Overview

Two environments are configured in the release workflows:

1. **pre-release** - Used for creating pre-release versions (rc/beta/alpha)
2. **production** - Used for promoting pre-releases to stable versions

Both environments require manual approval from designated reviewers before the workflow can proceed with creating/promoting releases.

## Environment Configuration

### Prerequisites

- Repository administrator access
- Appropriate repository permissions (Settings → Environments)

### Creating Environments

#### 1. Navigate to Environments

1. Go to your repository on GitHub
2. Click **Settings** (top menu)
3. In the left sidebar, click **Environments**

#### 2. Create Pre-Release Environment

1. Click **New environment**
2. Name: `pre-release`
3. Click **Configure environment**

**Configure the following protection rules**:

- ✅ **Required reviewers**
  - Add maintainers who should approve pre-releases
  - Recommended: At least 1 reviewer
  - Prevents accidental pre-release creation

- ✅ **Wait timer** (optional)
  - Set to 0 minutes (no automatic delay)
  - Or add a delay if desired

- ✅ **Deployment branches and tags**
  - Select: **Selected branches and tags**
  - Add rule: `main` branch only
  - This ensures only main branch can create releases

**Environment secrets** (if needed):

- No additional secrets required for basic setup
- GitHub token is automatically provided

**Save** the environment

#### 3. Create Production Environment

1. Click **New environment**
2. Name: `production`
3. Click **Configure environment**

**Configure the following protection rules**:

- ✅ **Required reviewers**
  - Add maintainers who should approve stable releases
  - Recommended: At least 2 reviewers for production
  - Requires multiple approvals for critical releases

- ✅ **Wait timer** (optional)
  - Set to 0 minutes for immediate review
  - Or set to 10-30 minutes for time to review changes

- ✅ **Deployment branches and tags**
  - Select: **Selected branches and tags**
  - Add rule: `main` branch only
  - Ensures only main branch can promote to stable

- ⚠️ **Prevent self-review** (recommended)
  - Enable this to require approval from someone other than the workflow trigger

**Environment secrets** (if needed):

- No additional secrets required for basic setup

**Save** the environment

## Approval Workflow

### Pre-Release Creation

When someone triggers the **Pre-Release** workflow:

1. Workflow calculates version and updates files
2. Tests run automatically
3. **Workflow pauses at the `build` job**
4. GitHub sends notification to required reviewers
5. Reviewer(s) must **approve** or **reject** the deployment
6. If approved: artifacts are built and pre-release is created
7. If rejected: workflow is cancelled

**Approval screen shows**:

- Pre-release version (e.g., `v0.2.0-rc.1`)
- Commit SHA being released
- Link to the release page (once approved)
- Comment field for approval notes

### Stable Release Promotion

When someone triggers the **Promote to Stable Release** workflow:

1. Workflow validates pre-release tag exists
2. Updates version files to stable
3. Tests run automatically
4. **Workflow pauses at the `promote` job**
5. GitHub sends notification to required reviewers
6. Reviewer(s) must **approve** or **reject** the deployment
7. If approved: artifacts are promoted and stable release is created
8. If rejected: workflow is cancelled

**Approval screen shows**:

- Pre-release being promoted (e.g., `v0.2.0-rc.1`)
- Stable version (e.g., `v0.2.0`)
- Link to the stable release page (once approved)

## Approving a Deployment

### As a Reviewer

When you receive a deployment approval request:

1. **Check your email** or GitHub notifications
2. **Go to Actions** tab in the repository
3. Find the workflow run waiting for approval
4. **Review the details**:
   - Check the version being released
   - Review the commits included
   - Verify tests passed
   - Check CHANGELOG updates

5. Click **Review deployments**
6. Select the environment (`pre-release` or `production`)
7. Add a comment (optional but recommended):

   ```text
   Approved: Version v0.2.0 includes security fixes
   ```

8. Click **Approve and deploy** or **Reject**

### Review Checklist

Before approving a pre-release:

- [ ] Version number is correct
- [ ] All tests passed
- [ ] No breaking changes without documentation
- [ ] CHANGELOG is updated
- [ ] Commit messages follow conventional commits

Before approving a stable release:

- [ ] Pre-release was tested successfully
- [ ] No critical issues reported
- [ ] Documentation is accurate
- [ ] Version matches what was tested
- [ ] All artifacts are present

## Viewing Deployment History

GitHub tracks all deployments in the environment:

1. Go to **Settings** → **Environments**
2. Click on `pre-release` or `production`
3. View **Deployment history**:
   - Who triggered the workflow
   - Who approved/rejected
   - When it was deployed
   - Links to workflow runs
   - Comments from reviewers

This creates a complete audit trail for compliance.

## Bypassing Approval (Not Recommended)

**Warning**: Only repository administrators can bypass environment protection rules.

If you need to bypass approval (emergency only):

1. Go to **Settings** → **Environments**
2. Select the environment
3. Temporarily remove **Required reviewers**
4. Run the workflow
5. **Immediately re-enable** required reviewers

**Better approach**: Add yourself as an approved reviewer if needed.

## Troubleshooting

**Workflow stuck waiting**: Check Actions tab for "Review deployments" button, verify you're a required reviewer

**Cannot approve**: Ensure you're listed as reviewer and didn't trigger the workflow yourself (if self-review prevented)

**Environment not found**: Verify environment names in workflow files match **Settings** → **Environments** exactly

## Security Best Practices

### Required Reviewers

**Pre-release environment**:

- Minimum: 1 reviewer
- Recommended: 1-2 reviewers
- Should include: Project maintainers

**Production environment**:

- Minimum: 2 reviewers
- Recommended: 2-3 reviewers
- Should include: Senior maintainers, security lead

### Branch Protection

Combine environment protection with branch protection:

1. Go to **Settings** → **Branches**
2. Add rule for `main` branch:
   - ✅ Require pull request reviews (at least 1)
   - ✅ Require status checks to pass
   - ✅ Require branches to be up to date
   - ✅ Include administrators

This ensures:

- Code is reviewed before merging
- Tests pass before merging
- Releases require additional approval

### Audit Trail

GitHub automatically logs:

- Who triggered the workflow
- Who approved/rejected
- When deployment occurred
- Environment variables used
- Workflow run details

Export deployment logs regularly for compliance:

1. Go to **Settings** → **Environments** → [Environment]
2. View deployment history
3. Document approvals in release notes

## Example Release Flow

1. **Pre-Release**: Trigger workflow → Reviewer approves → `v0.2.0-rc.1` created
2. **Testing**: Test artifacts, fix issues, create new RC if needed
3. **Promotion**: Trigger promote workflow → Multiple reviewers approve → `v0.2.0` created
4. **Audit**: All approvals logged in environment history

## See Also

- [GitHub Environments Documentation](https://docs.github.com/en/actions/deployment/targeting-different-environments/using-environments-for-deployment)
- [Deployment Protection Rules](https://docs.github.com/en/actions/deployment/targeting-different-environments/using-environments-for-deployment#deployment-protection-rules)
- [Version Management Guide](versioning.md)
- [Contributing Guide](https://github.com/santosr2/uptool/blob/{{ extra.uptool_version }}/CONTRIBUTING.md)

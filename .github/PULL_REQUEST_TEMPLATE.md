# Pull Request

## Description

<!-- Provide a brief description of your changes -->

## Type of Change

<!-- Mark the relevant option with an "x" -->

- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update
- [ ] Code refactoring (no functional changes)
- [ ] Performance improvement
- [ ] Test coverage improvement
- [ ] CI/CD or build changes
- [ ] Dependency update

## Checklist

<!-- Mark completed items with an "x" -->

- [ ] My code follows the project's coding standards
- [ ] I have run `go fmt ./...` to format my code
- [ ] I have run `go vet ./...` and addressed all issues
- [ ] I have run `mise run lint` and addressed all issues
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] I have run `mise run test` and all tests pass
- [ ] New and existing unit tests pass locally with my changes
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] I have updated CHANGELOG.md with my changes
- [ ] My changes generate no new warnings or errors
- [ ] I have checked my code for security vulnerabilities
- [ ] Any dependent changes have been merged and published

## Testing

<!-- Describe the tests you ran to verify your changes -->

### Test Configuration

- **Go version**:
- **OS**:
- **Integration tested** (if applicable):

### Test Steps

1.
2.
3.

### Test Output

<!-- Paste relevant test output or attach screenshots -->

```text
# Paste test output here
```

## Integration-Specific Changes

<!-- If your PR affects a specific integration, fill this section -->

**Integration**: <!-- npm / helm / terraform / precommit / tflint / other -->

**Manifest files affected**:

- [ ] package.json
- [ ] Chart.yaml
- [ ] .pre-commit-config.yaml
- [ ] *.tf files
- [ ] .tflint.hcl
- [ ] Other: ______

**Registry client changes**:

- [ ] npm Registry
- [ ] Terraform Registry
- [ ] GitHub Releases
- [ ] Helm Repository
- [ ] Other: ______

## Breaking Changes

<!-- If this is a breaking change, describe: -->
<!-- 1. What breaks -->
<!-- 2. Why this change was necessary -->
<!-- 3. Migration path for users -->

## Additional Context

<!-- Add any other context, screenshots, or information about the PR here -->

## Related Issues

<!-- Link related issues using keywords like "Fixes #123" or "Closes #456" -->

Fixes #
Related to #

---

## For Maintainers

<!-- Maintainers: Fill this section during review -->

### Review Checklist

- [ ] Code quality and style reviewed
- [ ] Test coverage is adequate
- [ ] Documentation is complete and accurate
- [ ] CHANGELOG.md is updated
- [ ] No security concerns identified
- [ ] Performance impact is acceptable
- [ ] Breaking changes are clearly documented
- [ ] Commit message follows [Conventional Commits](https://www.conventionalcommits.org/)

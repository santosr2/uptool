# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Features

- **asdf**: Integration for `.tool-versions` runtime version manager (89.1% test coverage)
  - Detects and parses `.tool-versions` files
  - Version resolution and update capabilities
  - Edge case handling (comments, empty lines, invalid formats)
  - End-to-end workflow testing
- **mise**: Integration for `mise.toml` runtime version manager (90.5% test coverage)
  - Detects `mise.toml` and `.mise.toml` files
  - TOML format support (string and map formats)
  - Version resolution and update capabilities
  - Edge case handling (invalid TOML, empty files)
  - End-to-end workflow testing
- **npm**: Updates `package.json` with version constraint preservation
- **helm**: Updates `Chart.yaml` dependencies
- **pre-commit**: Uses native `pre-commit autoupdate` command
- **terraform**: Updates module versions in `*.tf` files
- **tflint**: Updates plugin versions in `.tflint.hcl`
- **config**: Configuration file support via `uptool.yaml`
  - Control which integrations are enabled/disabled
  - Per-integration update policies
  - Automatic loading from repository root
- **cli**: `--quiet` / `-q` flag to suppress informational output (errors only)
- **cli**: `--verbose` / `-v` flag for verbose debug output
- **cli**: Version display in help output
- **cli**: Filtering support via `--only` and `--exclude` flags
- **cli**: Multiple output formats (table, JSON)
- **cli**: Dry-run mode for safe previewing
- **engine**: Parallel execution with worker pools
- **engine**: Diff generation for all file changes
- **policy**: Semantic version resolution with policy enforcement
- **action**: GitHub Action support for automated dependency updates in CI/CD

### Bug Fixes

- **terraform**: Improved version comparison logic
  - Correctly strips version constraint prefixes (`~>`, `>=`, `=`)
  - Eliminates false positive update reports
  - Proper version normalization before comparison
- **core**: Fixed `go.mod` version from invalid `1.25.4` to `1.25`
- **config**: Configuration system now properly integrated with CLI (was previously unused)

### Security

- **registry**: Fixed 8 HTTP response body leaks across registry clients (npm, Terraform, Helm, GitHub)
- **engine**: Fixed file permissions for plan output files (0644 → 0600 for sensitive data)
- **pre-commit**: Explicitly acknowledged intentional error ignoring in integration
- **action**: GitHub Actions in `action.yml` now pinned to commit SHAs

### Performance

- **engine**: Modernized concurrency patterns to Go 1.25 standards
  - Uses `sync.WaitGroup.Go()` method (introduced in Go 1.25)
  - Replaced `wg.Add(1)` + `go func()` + `defer wg.Done()` pattern
  - Reduces boilerplate and prevents goroutine management bugs

### Refactor

- **core**: Removed architectural fragmentation (consolidated from dual engine/core to single engine-based architecture)
- **core**: Deleted unused `internal/core/` directory (architectural consolidation)
- **asdf/mise**: Separated into distinct integrations (asdf for `.tool-versions`, mise for `mise.toml`)

### Documentation

- **readme**: Comprehensive README with usage examples
- **readme**: Updated to reflect actual configuration capabilities
- **architecture**: ARCHITECTURE.md updated to reflect current codebase structure
- **claude**: CLAUDE.md updated with accurate integration status and agent system
- **contributing**: CONTRIBUTING.md with trunk-based workflow
- **security**: SECURITY.md with support policy and vulnerability reporting
- **governance**: GOVERNANCE.md for project management
- **docs**: Go 1.25 feature adoption analysis (`docs/go-1.25-analysis.md`)
- **docs**: GitHub Action usage guide
- **docs**: Versioning guide with release process details
- **docs**: Patch release workflow guide for maintainers
- **github**: Pull request template for contributors
- **github**: CODE_OF_CONDUCT.md (Contributor Covenant v2.1)
- **github**: .github/CODEOWNERS for code ownership

### Testing

- **rewrite**: Comprehensive test suite (87.7% coverage)
- **asdf**: Comprehensive test suite (89.1% coverage)
- **mise**: Comprehensive test suite (90.5% coverage)
- **terraform**: Version constraint fix verification tests (23.8% coverage)

### Miscellaneous Tasks

- **ci**: Comprehensive CI/CD infrastructure
  - Multi-platform testing (Ubuntu, macOS, Windows)
  - Security scanning (OSV, Dependency Review, CodeQL, govulncheck)
  - Automated release workflow with SBOM generation
  - Action validation and workflow linting
- **ci**: Pre-release workflow for rc/beta/alpha versions
- **ci**: Promote release workflow with automated release branch creation
- **ci**: Patch release workflow for backporting security fixes
- **ci**: Security patch coordination workflow for multi-version updates
- **ci**: Dependency hygiene workflow (uptool monitoring itself)
- **examples**: Example configuration files in `examples/` directory
  - `.tool-versions` - asdf example with multiple tools
  - `mise.toml` - mise example with simple string format
  - `.mise.toml` - mise hidden file variant with map format
  - `uptool.yaml` - Complete uptool configuration
  - Plugin examples (Python integration)
- **deps**: .github/dependabot.yml for comparison testing (temporary)
- **logging**: Default log level changed from INFO to WARN for cleaner CLI output

---

<!-- Releases will be automatically generated by git-cliff from this point forward -->

<!-- generated by git-cliff -->

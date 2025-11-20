# Contributing to uptool

Thank you for your interest in contributing to uptool! This document provides guidelines and instructions for contributing.

## Code of Conduct

Be respectful, inclusive, and constructive. We're all here to build great software together.

## Development Setup

### Prerequisites

- **Go 1.25+**: [Download Go](https://go.dev/dl/)
- **Git**: For version control
- **mise**: For consistent tool versions and task runner - [Install mise](https://mise.jdx.dev/)
- **A text editor**: VS Code, GoLand, vim, etc.

### Getting Started

1. **Fork the repository** on GitHub

2. **Clone your fork**:

   ```bash
   git clone https://github.com/YOUR_USERNAME/uptool.git
   cd uptool
   ```

3. **Add upstream remote**:

   ```bash
   git remote add upstream https://github.com/santosr2/uptool.git
   ```

4. **Install development tools** (optional but recommended):

   ```bash
   # Option 1: Use mise for consistent tool versions
   mise install

   # Option 2: Install tools manually
   # See mise.toml for required versions
   ```

5. **Build the project**:

   ```bash
   # Using mise (recommended)
   mise run build

   # Or use go directly
   go build -o uptool ./cmd/uptool
   ```

6. **Run tests**:

   ```bash
   # Using mise
   mise run test

   # Or use go directly
   go test ./...
   ```

7. **Verify it works**:

   ```bash
   ./dist/uptool scan
   ./dist/uptool plan
   ```

### Optional: Pre-commit Hooks

```bash
# Install (if not using mise)
pip install pre-commit

# Enable hooks
pre-commit install

# Run manually
pre-commit run --all-files
```

Hooks include Go formatting, linting, security checks, and tests. All checks also run in CI.

## Development Workflow (Trunk-Based)

We use **trunk-based development**—not Git Flow. All changes go directly to `main` after review.

### Making Changes

1. **Create a feature branch** from `main`:

   ```bash
   git checkout main
   git pull upstream main
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**:
   - Write clean, idiomatic Go code
   - Follow existing code patterns
   - Add tests for new functionality
   - Update documentation

3. **Test your changes**:

   ```bash
   # Run tests
   go test ./...

   # Format code
   go fmt ./...

   # Test locally
   ./uptool scan
   ./uptool plan --only=your-integration
   ./uptool update --only=your-integration --dry-run --diff
   ```

4. **Commit with conventional commits**:

   ```bash
   git add .
   git commit -m "feat: add support for Python pip"
   # or
   git commit -m "fix: handle empty package.json files"
   # or
   git commit -m "docs: update integration guide"
   ```

   **Commit message format**:
   - `feat:` - New features
   - `fix:` - Bug fixes
   - `docs:` - Documentation changes
   - `refactor:` - Code refactoring
   - `test:` - Test additions/changes
   - `chore:` - Maintenance tasks

5. **Push to your fork**:

   ```bash
   git push origin feature/your-feature-name
   ```

6. **Open a Pull Request** on GitHub

### Pull Request Guidelines

- **Title**: Use conventional commit format (e.g., "feat: add Python integration")
- **Description**: Explain what and why, not just how
- **Tests**: Include tests for new functionality
- **Documentation**: Update README and relevant docs
- **Small PRs**: Keep changes focused and reviewable

### PR Review Process

1. Maintainer reviews your PR
2. Feedback and discussion
3. Make requested changes (push to same branch)
4. Approval → merge to `main`

**No separate develop/staging branches**. Features merge directly to `main` when ready.

## Code Style

### Go Style Guidelines

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Run `go fmt ./...` before committing
- Use meaningful variable names
- Keep functions small and focused
- Write GoDoc comments for all exported types/functions

### GoDoc Comments

All exported items must have GoDoc comments:

```go
// Package npm implements npm package.json dependency updates.
package npm

// Integration implements the engine.Integration interface for npm.
type Integration struct {
    client *registry.NPMClient
}

// New creates a new npm integration with default settings.
func New() *Integration {
    return &Integration{
        client: registry.NewNPMClient(),
    }
}

// Detect scans for package.json files in the repository.
// It returns a list of manifests with extracted dependencies.
func (i *Integration) Detect(ctx context.Context, repoRoot string) ([]*engine.Manifest, error) {
    // implementation
}
```

### Error Handling

- Always check errors
- Wrap errors with context:

  ```go
  if err := doSomething(); err != nil {
      return fmt.Errorf("failed to do something: %w", err)
  }
  ```

- Don't panic unless absolutely necessary

### Testing

- Write table-driven tests
- Use testdata for fixtures
- Test edge cases and error paths

Example:

```go
func TestNPMDetect(t *testing.T) {
    tests := []struct {
        name    string
        fixture string
        want    int
    }{
        {"valid package.json", "testdata/npm/package.json", 4},
        {"empty deps", "testdata/npm/empty.json", 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

## Adding a New Integration

Want to add support for a new ecosystem? See the [Plugin Development Guide](docs/plugin-development.md) for complete details.

### Quick Overview

**Two options**:
1. **Built-in Integration** - Compiled into uptool (for widely-used ecosystems)
2. **Plugin** - External shared library (for custom/experimental integrations)

### Steps for Built-in Integration

1. **Create integration package**: `internal/integrations/<name>/<name>.go`
   - Implement `engine.Integration` interface (Detect, Plan, Apply, Validate)
   - See existing integrations (npm, helm) as reference

2. **Register integration**: Import in `internal/integrations/all/all.go`

3. **Add configuration**: Update `integrations.yaml` with metadata

4. **Write tests**: Create `<name>_test.go` with fixtures in `testdata/`
   - Target: >70% test coverage

5. **Document it**: Create `docs/integrations/<name>.md`
   - Use [Integration Template](docs/INTEGRATION_TEMPLATE.md) (target: 80-120 lines)

### Interface Requirements

```go
type Integration interface {
    Name() string
    Detect(ctx context.Context, repoRoot string) ([]*Manifest, error)
    Plan(ctx context.Context, manifest *Manifest) (*UpdatePlan, error)
    Apply(ctx context.Context, plan *UpdatePlan) (*ApplyResult, error)
    Validate(ctx context.Context, manifest *Manifest) error
}
```

### Reference Implementations

- **Simple**: `internal/integrations/asdf/` - Line-based parsing
- **YAML**: `internal/integrations/helm/` - YAML rewriting
- **JSON**: `internal/integrations/npm/` - Constraint preservation
- **HCL**: `internal/integrations/terraform/` - HCL parsing
- **Native**: `internal/integrations/precommit/` - Calls `pre-commit autoupdate`

For complete code examples and detailed steps, see [Plugin Development Guide](docs/plugin-development.md).

### Testing Your Integration

```bash
# Build
mise run build

# Test
./dist/uptool scan --only=yourintegration
./dist/uptool plan --only=yourintegration
./dist/uptool update --only=yourintegration --dry-run --diff

# Verify tests
mise run test
go test -cover ./internal/integrations/yourintegration  # Target >70%
```

### Submitting

1. Run `mise run check` (fmt + vet + lint + test)
2. Update CHANGELOG.md
3. Commit with conventional commit:
   ```bash
   git commit -m "feat(integration): add yourname support

   - Detects your-manifest.yaml files
   - Queries your-registry for versions
   - Updates dependencies with version constraints
   - Includes comprehensive tests and documentation"
   ```

5. Push and create PR

## Documentation Standards

### README Updates

When adding features, update:

- Features section
- Supported Integrations table
- Integration Details section
- Examples

### docs/ Files

Create detailed documentation in `docs/integrations/<name>.md`:

```markdown
# <Integration Name> Integration

## Overview

What this integration does and why.

## Manifest Files

Which files it processes.

## Update Strategy

How it updates files (native command, custom rewriting, etc.).

## Registry

Where it queries for versions.

## Examples

### Before
\`\`\`yaml
# example before
\`\`\`

### After
\`\`\`yaml
# example after
\`\`\`

## Configuration

Future: integration-specific config options.

## Limitations

Known limitations or edge cases.
```

### GoDoc

- First sentence should be a complete statement
- Don't repeat the function name unnecessarily
- Explain parameters and return values
- Include examples for complex functions

## Registry Clients

If your integration needs a new registry client, create it in `internal/registry/<name>.go`:

```go
package registry

// YourRegistryClient queries the XYZ registry for package information.
type YourRegistryClient struct {
    client  *http.Client
    baseURL string
}

// NewYourRegistryClient creates a new registry client.
func NewYourRegistryClient() *YourRegistryClient {
    return &YourRegistryClient{
        client: &http.Client{Timeout: 30 * time.Second},
        baseURL: "https://registry.example.com",
    }
}

// GetLatestVersion fetches the latest version for a package.
func (c *YourRegistryClient) GetLatestVersion(ctx context.Context, packageName string) (string, error) {
    // Implementation
}
```

## Testing Guidelines

### Unit Tests

- Test each integration method independently
- Use testdata fixtures
- Mock external API calls when appropriate

### Integration Tests

- Test real API calls (with rate limiting/caching)
- Document any API keys needed for testing
- Skip slow tests with `testing.Short()`

### Test Coverage

Aim for:

- Core engine: 80%+ coverage
- Integrations: 70%+ coverage
- Registry clients: 60%+ coverage

## Version Management

All commits **must** use [Conventional Commits](https://www.conventionalcommits.org/):

| Type | Impact | Example |
|------|--------|---------|
| `feat:` | Minor (0.1.0 → 0.2.0) | `feat: add Python integration` |
| `fix:` | Patch (0.1.0 → 0.1.1) | `fix: handle empty manifests` |
| `feat!:` | Major (0.1.0 → 1.0.0) | `feat!: redesign API` |
| `docs:`, `chore:` | No bump | `docs: update README` |

**Good**: `feat(npm): add peer dependencies support`
**Bad**: `added feature` (no type), `Fixed bug` (wrong case)

See [docs/versioning.md](docs/versioning.md) for complete details.

## Release Process

Releases are **fully automated** via GitHub Actions with approval gates. Contributors only need to:

1. Write good conventional commits
2. Submit PRs

Maintainers handle releases. See [docs/versioning.md](docs/versioning.md) and [SECURITY.md](SECURITY.md) for details.

## Mise Tasks

Common development tasks:

```bash
# Build
mise run build          # Build binary to dist/uptool
mise run build-all      # Build for all platforms

# Testing
mise run test           # Run all tests
mise run test-coverage  # Run tests with coverage report

# Code Quality
mise run fmt            # Format code
mise run lint           # Run golangci-lint
mise run vet            # Run go vet
mise run check          # Run all checks (fmt + vet + lint + test)

# Version Management
mise run version-show   # Display current version

# Cleanup
mise run clean          # Remove build artifacts

# Development
mise run run-scan       # Build and run scan
mise run run-plan       # Build and run plan
mise run run-update     # Build and run update --dry-run
```

## Getting Help

- **Questions**: Open a [Discussion](https://github.com/santosr2/uptool/discussions)
- **Bugs**: Open an [Issue](https://github.com/santosr2/uptool/issues)
- **Security**: See [SECURITY.md](SECURITY.md)

## Recognition

Contributors will be:

- Listed in release notes
- Credited in CHANGELOG.md
- Thanked in the community

Thank you for contributing to uptool! 🎉

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

We provide pre-commit hooks to automatically check code quality before commits:

1. **Install pre-commit**:

   ```bash
   # If you ran `mise install` it is already installed

   # Using pip
   pip install pre-commit

   # Or using homebrew (macOS)
   brew install pre-commit
   ```

2. **Install the git hooks**:

   ```bash
   pre-commit install
   ```

3. **Run manually** (optional):

   ```bash
   pre-commit run --all-files
   ```

Our pre-commit hooks include:

- **Go formatting** (`gofmt`, `goimports`)
- **Go linting** (`golangci-lint`)
- **Go security checks** (`gosec`)
- **Go tests** (`go test`)
- **General checks** (trailing whitespace, YAML/JSON validation, etc.)
- **YAML linting** (`yamllint`)
- **Markdown linting** (`markdownlint`)

**Note**: Pre-commit hooks are optional but highly recommended. All checks will also run in CI.

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

Want to add support for a new ecosystem? Great! You have two options:

1. **Built-in Integration** - Compiled into uptool binary (for widely-used ecosystems)
2. **Plugin** - External shared library loaded at runtime (for custom/experimental integrations)

See [Plugin Development Guide](docs/plugin-development.md) for external plugins. This section covers built-in integrations.

### Built-in Integration Development

#### Step 1: Create Integration Package

Create `internal/integrations/<name>/` directory with the following files:

**`internal/integrations/<name>/<name>.go`**:

```go
package yourintegration

import (
    "context"
    "fmt"
    "os"
    "path/filepath"

    "github.com/santosr2/uptool/internal/datasource"
    "github.com/santosr2/uptool/internal/engine"
    "github.com/santosr2/uptool/internal/integrations"
)

// Register the integration in init() so it's automatically available
func init() {
    integrations.Register("yourname", func() engine.Integration {
        return New()
    })
}

const integrationName = "yourname"

// Integration implements the engine.Integration interface.
type Integration struct {
    ds datasource.Datasource  // Optional: use datasource abstraction
}

// New creates a new integration instance.
func New() *Integration {
    return &Integration{
        ds: datasource.Get("your-datasource"), // Or create custom client
    }
}

// Name returns the integration identifier.
func (i *Integration) Name() string {
    return integrationName
}

// Detect finds manifest files for this integration.
// This should scan the repository and return a list of manifests with dependencies.
func (i *Integration) Detect(ctx context.Context, repoRoot string) ([]*engine.Manifest, error) {
    var manifests []*engine.Manifest

    // Walk repository to find manifest files
    err := filepath.Walk(repoRoot, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        // Skip directories
        if info.IsDir() {
            return nil
        }

        // Check if this is your manifest file (e.g., myapp.yaml, package.json, etc.)
        if filepath.Base(path) != "your-manifest.yaml" {
            return nil
        }

        // Parse the manifest
        deps, err := i.parseManifest(path)
        if err != nil {
            return fmt.Errorf("parsing %s: %w", path, err)
        }

        manifests = append(manifests, &engine.Manifest{
            Path:         path,
            Integration:  integrationName,
            Dependencies: deps,
        })

        return nil
    })

    if err != nil {
        return nil, fmt.Errorf("scanning for manifests: %w", err)
    }

    return manifests, nil
}

// Plan determines available updates by querying the registry.
func (i *Integration) Plan(ctx context.Context, manifest *engine.Manifest) (*engine.UpdatePlan, error) {
    var updates []*engine.DependencyUpdate

    for _, dep := range manifest.Dependencies {
        // Query datasource or registry for latest version
        latestVersion, err := i.ds.GetLatestVersion(ctx, dep.Name)
        if err != nil {
            // Log but continue
            fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
            continue
        }

        // Check if update is needed
        if dep.CurrentVersion != latestVersion {
            updates = append(updates, &engine.DependencyUpdate{
                Name:           dep.Name,
                CurrentVersion: dep.CurrentVersion,
                TargetVersion:  latestVersion,
                Impact:         i.calculateImpact(dep.CurrentVersion, latestVersion),
            })
        }
    }

    return &engine.UpdatePlan{
        Manifest: manifest,
        Updates:  updates,
    }, nil
}

// Apply executes the update plan by rewriting manifest files.
func (i *Integration) Apply(ctx context.Context, plan *engine.UpdatePlan) (*engine.ApplyResult, error) {
    // Read manifest
    content, err := os.ReadFile(plan.Manifest.Path)
    if err != nil {
        return nil, fmt.Errorf("reading manifest: %w", err)
    }

    // Apply updates (use rewrite package for YAML/TOML/JSON/HCL)
    updated := i.applyUpdates(string(content), plan.Updates)

    // Write back
    if err := os.WriteFile(plan.Manifest.Path, []byte(updated), 0644); err != nil {
        return nil, fmt.Errorf("writing manifest: %w", err)
    }

    return &engine.ApplyResult{
        Success: true,
        Applied: len(plan.Updates),
    }, nil
}

// Validate checks manifest validity (optional but recommended).
func (i *Integration) Validate(ctx context.Context, manifest *engine.Manifest) error {
    // Validate manifest structure
    // This can prevent corrupted updates
    return nil
}

// Helper methods

func (i *Integration) parseManifest(path string) ([]*engine.Dependency, error) {
    // Parse your manifest format
    // Return list of dependencies
    return nil, nil
}

func (i *Integration) calculateImpact(current, target string) engine.Impact {
    // Use internal/resolve package for semver comparison
    return engine.ImpactMinor
}

func (i *Integration) applyUpdates(content string, updates []*engine.DependencyUpdate) string {
    // Use internal/rewrite package for YAML/TOML rewriting
    // Or implement custom logic for your format
    return content
}
```

#### Step 2: Register Integration

The integration auto-registers via `init()` function. However, you need to import it in `internal/integrations/all/all.go`:

```go
package all

import (
    // Existing imports
    _ "github.com/santosr2/uptool/internal/integrations/helm"
    _ "github.com/santosr2/uptool/internal/integrations/npm"

    // Add your integration
    _ "github.com/santosr2/uptool/internal/integrations/yourintegration"
)
```

#### Step 3: Update integrations.yaml

Add your integration to `integrations.yaml` at the repository root:

```yaml
integrations:
  yourname:
    displayName: "Your Integration Name"
    description: "Brief description of what this integration does"
    filePatterns:
      - "your-manifest.yaml"
      - "config/*.yaml"  # Glob patterns supported
    datasources:
      - your-registry  # Reference to datasource below
    experimental: false  # Set to true for new integrations
    disabled: false
    url: "https://example.com"  # Tool homepage
    category: "package-manager"  # Reference to category below

# If you need a new datasource
datasources:
  your-registry:
    name: "Your Registry"
    url: "https://registry.example.com"
    type: "http-json"  # http-json, github-releases, etc.
    description: "Description of the registry"

# If you need a new category
categories:
  your-category:
    name: "Your Category Name"
    description: "Description of this category"
```

**Integration Fields**:

- `displayName`: Human-readable name shown in CLI
- `description`: One-line description of what it does
- `filePatterns`: Glob patterns for manifest files
- `datasources`: List of datasources this integration uses
- `experimental`: Mark new integrations as experimental initially
- `disabled`: Set to true to disable without removing code
- `url`: Homepage or documentation URL
- `category`: Category for grouping (see categories below)

**Categories** (existing):

- `runtime-manager`: asdf, mise, etc.
- `package-manager`: npm, pip, cargo, etc.
- `kubernetes`: Helm, kubectl, etc.
- `infrastructure`: Terraform, Pulumi, etc.
- `git-hooks`: pre-commit, etc.
- `linting`: golangci-lint, tflint, etc.

#### Step 4: Add Datasource (if needed)

If your integration needs a new registry client, create it in `internal/datasource/<name>.go`:

```go
package datasource

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

func init() {
    Register("your-registry", func() Datasource {
        return NewYourRegistry()
    })
}

// YourRegistry queries your custom registry.
type YourRegistry struct {
    client  *http.Client
    baseURL string
}

// NewYourRegistry creates a new registry client.
func NewYourRegistry() Datasource {
    return &YourRegistry{
        client: &http.Client{Timeout: 30 * time.Second},
        baseURL: "https://registry.example.com",
    }
}

// Name returns datasource identifier.
func (r *YourRegistry) Name() string {
    return "your-registry"
}

// GetLatestVersion fetches latest version for a package.
func (r *YourRegistry) GetLatestVersion(ctx context.Context, pkg string) (string, error) {
    url := fmt.Sprintf("%s/packages/%s", r.baseURL, pkg)

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return "", err
    }

    resp, err := r.client.Do(req)
    if err != nil {
        return "", fmt.Errorf("querying registry: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("registry returned %d", resp.StatusCode)
    }

    var data struct {
        LatestVersion string `json:"latest_version"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
        return "", fmt.Errorf("parsing response: %w", err)
    }

    return data.LatestVersion, nil
}

// GetVersions returns all available versions.
func (r *YourRegistry) GetVersions(ctx context.Context, pkg string) ([]string, error) {
    // Implementation
    return nil, nil
}

// GetPackageInfo returns package metadata.
func (r *YourRegistry) GetPackageInfo(ctx context.Context, pkg string) (*PackageInfo, error) {
    // Implementation
    return nil, nil
}
```

#### Step 5: Add Test Fixtures

Create test data in `internal/integrations/<name>/testdata/`:

```tree
internal/integrations/yourintegration/
├── yourintegration.go
├── yourintegration_test.go
└── testdata/
    ├── valid-manifest.yaml
    ├── empty-deps.yaml
    └── README.md  # Explain the fixtures
```

#### Step 6: Write Tests

Create `internal/integrations/<name>/<name>_test.go`:

```go
package yourintegration

import (
    "context"
    "testing"

    "github.com/santosr2/uptool/internal/engine"
)

func TestDetect(t *testing.T) {
    integration := New()

    manifests, err := integration.Detect(context.Background(), "testdata")
    if err != nil {
        t.Fatalf("Detect() error: %v", err)
    }

    if len(manifests) == 0 {
        t.Error("Expected to find manifests")
    }
}

func TestPlan(t *testing.T) {
    tests := []struct {
        name     string
        manifest *engine.Manifest
        wantErr  bool
    }{
        {
            name: "valid manifest",
            manifest: &engine.Manifest{
                Path: "testdata/valid-manifest.yaml",
                Dependencies: []*engine.Dependency{
                    {Name: "package1", CurrentVersion: "1.0.0"},
                },
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            integration := New()
            plan, err := integration.Plan(context.Background(), tt.manifest)

            if (err != nil) != tt.wantErr {
                t.Errorf("Plan() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if !tt.wantErr && plan == nil {
                t.Error("Expected non-nil plan")
            }
        })
    }
}
```

#### Step 7: Update Documentation

Create complete documentation in `docs/integrations/<name>.md`:

```markdown
# Your Integration

The yourname integration updates dependencies in your-manifest.yaml files.

## Overview

**Integration ID**: `yourname`

**Manifest Files**: `your-manifest.yaml`

**Update Strategy**: YAML parsing and rewriting

**Registry**: Your Registry API

**Status**: ✅ Stable (or 🧪 Experimental)

## What Gets Updated

[Explain what parts of the manifest get updated]

## Example

### Before
\`\`\`yaml
dependencies:
  package1: 1.0.0
  package2: 2.0.0
\`\`\`

### After
\`\`\`yaml
dependencies:
  package1: 1.5.0
  package2: 2.1.0
\`\`\`

## CLI Usage

[Include scan, plan, update examples]

## Configuration

[Show uptool.yaml configuration]

## Troubleshooting

[Common issues and solutions]

## Best Practices

[Recommendations]

## Limitations

[Known limitations]

## See Also

[Related links]
```

Also update:

- `README.md` - Add to Supported Integrations table
- `docs/integrations/README.md` - Add to integration list
- `examples/<name>.yaml` - Add example manifest

#### Step 8: Test End-to-End

```bash
# Build
mise run build

# Test scan
./dist/uptool scan --only=yourintegration

# Test plan
./dist/uptool plan --only=yourintegration

# Test update (dry-run)
./dist/uptool update --only=yourintegration --dry-run --diff

# Run all tests
mise run test

# Check coverage
go test -cover ./internal/integrations/yourintegration
```

#### Step 9: Create Pull Request

1. Ensure all tests pass
2. Run `mise run check` (fmt + vet + complexity + lint + test)
3. Update CHANGELOG.md with entry
4. Commit with conventional commit message:

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

uptool uses **automated semantic versioning** based on conventional commits. Understanding this is important for contributors.

### Conventional Commits (Required)

All commits **must** follow the [Conventional Commits](https://www.conventionalcommits.org/) format:

```text
<type>(<scope>): <subject>

[optional body]

[optional footer]
```

**Commit Types and Version Impact**:

| Type | Impact | Example |
|------|--------|---------|
| `feat:` | Minor bump (0.1.0 → 0.2.0) | `feat: add Python integration` |
| `fix:` | Patch bump (0.1.0 → 0.1.1) | `fix: handle empty manifests` |
| `feat!:` or `BREAKING CHANGE:` | Major bump (0.1.0 → 1.0.0) | `feat!: redesign API` |
| `docs:`, `chore:`, `test:`, etc. | No bump | `docs: update README` |

**Why This Matters**:

- Commits determine the next version automatically
- GitHub Actions uses commit history to calculate version bumps
- Incorrect commit types can cause wrong version numbers

**Good Examples**:

```bash
git commit -m "feat(npm): add peer dependencies support"
git commit -m "fix(helm): handle missing Chart.lock files"
git commit -m "docs: add configuration examples"
git commit -m "test: add integration tests for terraform"
git commit -m "chore: update dependencies"
```

**Bad Examples**:

```bash
git commit -m "added feature"           # ❌ No type prefix
git commit -m "Fixed bug"               # ❌ Wrong capitalization
git commit -m "WIP"                     # ❌ Not descriptive
git commit -m "updated stuff"           # ❌ Vague
```

### Commit Message Guidelines

**Subject Line**:

- Use imperative mood ("add" not "added" or "adds")
- Don't capitalize first letter
- No period at the end
- Keep under 72 characters

**Body** (optional but recommended for complex changes):

- Explain what and why, not how
- Wrap at 72 characters
- Separate from subject with blank line

**Footer** (for breaking changes):

```bash
git commit -m "feat!: redesign configuration format

BREAKING CHANGE: Configuration file now requires version: 2.
See docs/migration.md for upgrade instructions."
```

### Local Version Management

For testing version changes locally:

```bash
# Show current version
mise run version-show

# Manually bump for testing (don't commit these)
mise run version-bump-patch   # 0.1.0 → 0.1.1
mise run version-bump-minor   # 0.1.0 → 0.2.0
mise run version-bump-major   # 0.1.0 → 1.0.0
```

**Important**: Don't manually bump versions in PRs. The automated release process handles versioning.

## Release Process

Releases are **fully automated** based on conventional commits. Maintainers trigger releases via GitHub Actions.

### Pre-Release Process

1. **Development**: Contributors submit PRs with conventional commits
2. **Merge to main**: Maintainers merge approved PRs
3. **Pre-Release**: Maintainers trigger Pre-Release workflow
   - GitHub Action calculates version from commits
   - Updates VERSION file across codebase
   - Runs full test suite
4. **Approval Gate** ⚠️
   - Workflow pauses for manual approval
   - Designated reviewers approve/reject
   - See [docs/environments.md](docs/environments.md) for setup
5. **Build** (after approval):
   - Creates pre-release (e.g., `v0.2.0-rc.1`)
   - Builds and publishes artifacts
   - Creates mutable tags (v0-rc, v0.1-rc)
6. **Testing**: Community tests pre-release

### Stable Release Process

1. **Validation**: Pre-release is tested and verified
2. **Promotion**: Maintainers trigger Promote workflow
   - Extracts stable version (v0.2.0-rc.1 → v0.2.0)
   - Updates VERSION file
   - Runs full test suite
3. **Approval Gate** ⚠️
   - Workflow pauses for manual approval
   - Multiple reviewers approve/reject
   - See [docs/environments.md](docs/environments.md) for setup
4. **Release** (after approval):
   - Promotes artifacts
   - Updates CHANGELOG
   - Creates stable release
   - Creates mutable tags (v0, v0.2)
5. **Publication**: Stable release is published

**Contributors don't need to:**

- Manually update VERSION files
- Create git tags
- Build release artifacts
- Update CHANGELOG

**Everything is automated!** Just write good conventional commits.

### Version Support Policy

See [SECURITY.md](SECURITY.md) for the official support policy.

- Latest minor version: Full support
- Previous minor version: Security patches (6 months)
- Older versions: No support

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

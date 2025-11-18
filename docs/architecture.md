# Architecture Overview

This document provides a comprehensive overview of uptool's architecture, design decisions, and internal components.

## Table of Contents

- [Design Principles](#design-principles)
- [System Architecture](#system-architecture)
- [Core Components](#core-components)
- [Data Flow](#data-flow)
- [Integration Architecture](#integration-architecture)
- [Datasource Abstraction](#datasource-abstraction)
- [Manifest Rewriting](#manifest-rewriting)
- [Error Handling](#error-handling)
- [Performance Considerations](#performance-considerations)

---

## Design Principles

### 1. Manifest-First Philosophy

**Core Tenet**: Always update configuration files (manifests) directly, never rely solely on lockfile updates.

**Rationale**:

- Lockfiles can drift from manifests
- Manifests are the source of truth
- Better auditability and version control
- Consistent with developer workflows

**Implementation**:

- Integrations parse and rewrite manifest files
- Native commands used only when they update manifests
- Manual rewriting when native tools don't support manifest updates

### 2. Format Preservation

**Goal**: Maintain YAML comments, indentation, and structure when updating manifests.

**Challenges**:

- Standard YAML parsers lose comments
- Indentation styles vary
- Custom formatting preferences

**Solution**:

- Custom YAML rewriter in `internal/rewrite/yaml.go`
- Line-by-line processing
- Preserves everything except the version string

### 3. Extensibility

**Design**: Plugin-based architecture for integrations.

**Benefits**:

- Easy to add new package ecosystems
- Integrations are isolated
- Clear interfaces for testing

**Structure**:

```tree
internal/integrations/
├── registry.go          # Integration registration system
├── npm/
│   └── npm.go          # npm implementation
├── helm/
│   └── helm.go         # Helm implementation
└── ...

internal/engine/
└── types.go            # Integration interface definition
```

---

## System Architecture

```text
┌─────────────────────────────────────────────────────────────┐
│                         CLI Layer                           │
│                     (cmd/uptool)                            │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌───────────────────────────────────────────────────────────┐
│                      Engine Layer                         │
│                   (internal/engine)                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   Scan       │  │    Plan      │  │   Update     │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└────────────────────┬──────────────────────────────────────┘
                     │
        ┌────────────┼─────────────────┐
        │            │                 │
        ▼            ▼                 ▼
┌───────────┐  ┌────────────┐  ┌──────────┐
│Integration│  │ Datasource │  │ Rewrite  │
│  Layer    │  │   Layer    │  │  Layer   │
└───────────┘  └────────────┘  └──────────┘
                     │
                     ▼
             ┌──────────────┐
             │  Registry    │
             │   Clients    │
             └──────────────┘
                     │
                     ▼
┌──────────────────────────────────────┐
│        External Systems              │
│  ┌────┐ ┌────┐ ┌────┐ ┌────┐         │
│  │npm │ │PyPI│ │Helm│ │etc.│         │
│  └────┘ └────┘ └────┘ └────┘         │
└──────────────────────────────────────┘
```

---

## Core Components

### 1. CLI Layer (`cmd/uptool`)

**Responsibility**: Command-line interface and user interaction.

**Components**:

- Command parsing (Cobra)
- Flag validation
- Output formatting (table, JSON)
- Error display

**Files**:

- `cmd/uptool/main.go` - Entry point
- `cmd/uptool/cmd/root.go` - Root command
- `cmd/uptool/cmd/scan.go` - Scan command
- `cmd/uptool/cmd/plan.go` - Plan command
- `cmd/uptool/cmd/update.go` - Update command

### 2. Engine Layer (`internal/engine`)

**Responsibility**: Orchestrate the update workflow.

**Core Functions**:

```go
// Scan discovers available updates
func (e *Engine) Scan(ctx context.Context, opts ScanOptions) ([]Update, error)

// Plan shows what would be updated
func (e *Engine) Plan(ctx context.Context, opts PlanOptions) ([]Change, error)

// Apply executes the updates
func (e *Engine) Apply(ctx context.Context, opts ApplyOptions) error
```

**Workflow**:

1. Load configuration from `uptool.yaml`
2. Discover integrations (auto-detect manifest files)
3. Query registries for available versions
4. Apply policy rules (allow major/minor/patch)
5. Generate update plan
6. Execute updates (if not dry-run)

### 3. Integration Layer (`internal/integrations`)

**Responsibility**: Implement package ecosystem logic.

**Interface**:

```go
type Integration interface {
    // Name returns the integration identifier
    Name() string

    // Detect finds manifest files for this integration
    Detect(ctx context.Context, repoRoot string) ([]*Manifest, error)

    // Plan determines available updates for a manifest
    Plan(ctx context.Context, manifest *Manifest) (*UpdatePlan, error)

    // Apply executes the update plan
    Apply(ctx context.Context, plan *UpdatePlan) (*ApplyResult, error)

    // Validate checks if changes are valid (optional)
    Validate(ctx context.Context, manifest *Manifest) error
}
```

**Implementations**:

- `npm` - package.json
- `helm` - Chart.yaml
- `terraform` - versions.tf, *.tf
- `tflint` - .tflint.hcl
- `precommit` - .pre-commit-config.yaml (native command)
- `asdf` - .tool-versions
- `mise` - mise.toml, .mise.toml

### 4. Datasource Layer (`internal/datasource`)

**Responsibility**: Query package registries for available versions.

**Interface**:

```go
type Datasource interface {
    // Name returns the datasource identifier
    Name() string

    // GetLatestVersion returns the latest stable version for a package
    GetLatestVersion(ctx context.Context, pkg string) (string, error)

    // GetVersions returns all available versions for a package
    GetVersions(ctx context.Context, pkg string) ([]string, error)

    // GetPackageInfo returns detailed information about a package
    GetPackageInfo(ctx context.Context, pkg string) (*PackageInfo, error)
}
```

**Implementations**:

- `npm` - npm registry API
- `helm` - Artifact Hub / Helm repository
- `terraform` - Terraform Registry API
- `github` - GitHub Releases API (for tools like pre-commit, tflint)

**Architecture**: The datasource layer provides high-level abstractions that wrap low-level HTTP clients in `internal/registry/`. Each datasource (e.g., `internal/datasource/npm.go`) implements the `Datasource` interface by delegating to a registry client (e.g., `internal/registry/npm.go` with `NPMClient`).

**Caching**: Future enhancement (not yet implemented).

### 5. Rewrite Layer (`internal/rewrite`)

**Responsibility**: Preserve formatting while updating files.

**Key Features**:

- YAML comment preservation
- Line-by-line processing
- Minimal changes to file structure
- Diff generation

**Files**:

- `internal/rewrite/yaml.go` - YAML rewriter
- `internal/rewrite/diff.go` - Diff generator

### 6. Resolve Layer (`internal/resolve`)

**Responsibility**: Semantic version resolution and comparison.

**Functions**:

- Parse semantic versions
- Compare versions
- Filter by constraints (^1.0.0, ~2.3.0)
- Classify impact (major, minor, patch)

**Files**:

- `internal/resolve/semver.go`

### 7. Policy Layer (`internal/policy`)

**Responsibility**: Load and apply update policies.

**Configuration Example**:

```yaml
# uptool.yaml
version: 1

org_policy:
  signing:
    cosign_verify: false
  auto_merge:
    enabled: false
    guards: []
  require_signoff_from: []

integrations:
  - id: npm
    enabled: true
    policy:
      enabled: true
      update: minor  # none, patch, minor, major
      allow_prerelease: false
      pin: true
      cadence: weekly  # daily, weekly, monthly

  - id: terraform
    enabled: true
    policy:
      enabled: true
      update: patch
      allow_prerelease: false
      pin: true
```

**Files**:

- `internal/policy/config.go`

---

## Data Flow

### Scan Command Flow

```text
User runs: uptool scan
    │
    ▼
┌────────────────────┐
│ 1. Load Config     │  Read uptool.yaml
└────────┬───────────┘
         │
         ▼
┌────────────────────┐
│ 2. Detect Manifests│  Find package.json, Chart.yaml, etc.
└────────┬───────────┘
         │
         ▼
┌────────────────────┐
│ 3. Parse Manifests │  Extract dependencies
└────────┬───────────┘
         │
         ▼
┌────────────────────┐
│ 4. Query Registries│  Get latest versions (parallel)
└────────┬───────────┘
         │
         ▼
┌────────────────────┐
│ 5. Apply Policy    │  Filter by update rules
└────────┬───────────┘
         │
         ▼
┌────────────────────┐
│ 6. Format Output   │  Table or JSON
└────────────────────┘
```

### Update Command Flow

```text
User runs: uptool update
    │
    ▼
┌────────────────────┐
│ 1. Scan (above)    │  Get available updates
└────────┬───────────┘
         │
         ▼
┌────────────────────┐
│ 2. Generate Plan   │  What will change
└────────┬───────────┘
         │
         ▼
┌────────────────────┐
│ 3. Backup Manifests│  Create .bak files
└────────┬───────────┘
         │
         ▼
┌────────────────────┐
│ 4. Rewrite Files   │  Update versions
└────────┬───────────┘
         │
         ▼
┌────────────────────┐
│ 5. Validate        │  Check syntax
└────────┬───────────┘
         │
         ▼
┌────────────────────┐
│ 6. Run Native Cmds │  npm install, etc. (if configured)
└────────────────────┘
```

---

## Integration Architecture

### Integration Lifecycle

1. **Detection**: Integration discovers manifest files and extracts dependencies (via `Detect()`)
2. **Datasource Query**: Fetch available versions from registries (via datasource layer)
3. **Planning**: Determine what to update based on policy (via `Plan()`)
4. **Application**: Rewrite manifests with new versions (via `Apply()`)
5. **Validation**: Ensure updated manifests are valid (via `Validate()`)
6. **Post-Update**: Run native commands (optional, integration-specific)

### Integration Types

#### Type 1: Manifest Rewriting (npm, Helm, asdf, mise)

**Strategy**: Parse manifest, rewrite in-place, preserve formatting.

**Example (npm)**:

```javascript
// Before
{
  "dependencies": {
    "react": "^18.2.0"  // Keep this comment
  }
}

// After
{
  "dependencies": {
    "react": "^18.3.1"  // Keep this comment
  }
}
```

#### Type 2: Native Command (pre-commit)

**Strategy**: Delegate to native tool when it updates manifests.

**Example**:

```bash
# pre-commit updates .pre-commit-config.yaml
pre-commit autoupdate
```

**When to Use**:

- Tool already updates manifests correctly
- Format preservation is guaranteed
- No need for custom parsing

#### Type 3: Hybrid (Terraform)

**Strategy**: Parse HCL, rewrite with custom logic, validate with `terraform fmt`.

**Challenges**:

- HCL has complex syntax
- Multiple file locations (versions.tf, main.tf)
- Version constraints vary

---

## Datasource Abstraction

### Design Pattern

Each datasource implements a common interface, allowing integrations to be registry-agnostic. Datasources wrap low-level HTTP registry clients to provide a unified API.

**Benefits**:

- Support for private registries
- Caching at datasource level
- Consistent error handling
- Separation of concerns (high-level datasource logic vs low-level HTTP details)

### Datasource Implementations

#### npm Registry

**Endpoint**: `https://registry.npmjs.org/<package>`

**Response**:

```json
{
  "versions": {
    "1.0.0": {},
    "2.0.0": {}
  },
  "dist-tags": {
    "latest": "2.0.0"
  }
}
```

#### Helm (Artifact Hub)

**Endpoint**: `https://artifacthub.io/api/v1/packages/helm/<repo>/<chart>`

**Challenge**: Helm repositories are decentralized, need to query specific repos.

#### Terraform Registry

**Endpoint**: `https://registry.terraform.io/v1/providers/<namespace>/<name>/versions`

**Response**:

```json
{
  "versions": [
    {"version": "1.0.0"},
    {"version": "2.0.0"}
  ]
}
```

---

## Manifest Rewriting

### YAML Rewriter Algorithm

**Goal**: Update version strings while preserving everything else.

**Algorithm**:

```text
1. Read file line-by-line
2. For each line:
   a. Check if it contains a dependency key
   b. If yes, extract current version
   c. Replace with new version (exact string match)
   d. Preserve indentation, comments, quotes
3. Write modified lines to file
```

**Example**:

```yaml
# Input
dependencies:
  - name: nginx
    version: 1.24.0  # Production version
    repository: https://...

# Update nginx to 1.25.0

# Output (only version changed)
dependencies:
  - name: nginx
    version: 1.25.0  # Production version
    repository: https://...
```

### Diff Generation

**Purpose**: Show users exactly what will change.

**Implementation**:

- Unified diff format
- Color coding (green = additions, red = deletions)
- Context lines (3 before/after)

**Example**:

```diff
--- Chart.yaml
+++ Chart.yaml
@@ -3,7 +3,7 @@
 name: myapp
 description: My application
-version: 1.0.0
+version: 1.1.0
 dependencies:
   - name: postgresql
-    version: 12.0.0
+    version: 13.0.0
```

---

## Error Handling

### Error Categories

1. **User Errors** (exit code 1)
   - Invalid configuration
   - Missing manifest files
   - Invalid version constraints

2. **Registry Errors** (exit code 2)
   - Network failures
   - API rate limits
   - Package not found

3. **Update Errors** (exit code 3)
   - Manifest syntax errors after update
   - Native command failures
   - Permission issues

### Error Recovery

**Strategy**: Fail-safe operations.

**Mechanisms**:

- Automatic backups before updates
- Atomic file writes (write to temp, then rename)
- Rollback on validation failure
- Detailed error messages with recovery suggestions

---

## Performance Considerations

### Current State

- **Sequential processing**: One integration at a time
- **No caching**: Every run queries registries
- **Blocking I/O**: File operations are synchronous

### Future Optimizations

1. **Parallel Registry Queries**
   - Use goroutines for concurrent queries
   - Target: 5-10x speedup for large projects

2. **Caching Layer**
   - In-memory cache for registry responses
   - Disk cache with TTL
   - Target: 90% cache hit rate on repeated scans

3. **Incremental Scanning**
   - Only check changed manifests
   - Use file modification times
   - Target: Skip 80% of unchanged files

4. **Connection Pooling**
   - Reuse HTTP connections to registries
   - Target: Reduce connection overhead by 50%

---

## Testing Architecture

### Test Strategy

1. **Unit Tests**: Individual functions (target: 70% coverage)
2. **Integration Tests**: End-to-end workflows with real files
3. **Golden File Tests**: Compare output against expected results
4. **Benchmark Tests**: Performance regression detection

### Test Fixtures

Located in `testdata/`:

- Sample manifest files for each integration
- Expected outputs (golden files)
- Mock registry responses

---

## Security Considerations

### Threat Model

1. **Supply Chain Attacks**
   - Compromised registries
   - Malicious package versions
   - MITM attacks

2. **File System Access**
   - Permission issues
   - Symlink attacks
   - Path traversal

### Mitigations

- HTTPS for all registry communications
- Certificate pinning (future)
- File permission checks
- Input validation
- Dependency signing verification (future)

---

## Extension Points

Want to add a new integration? See:

- [Plugin Development Guide](plugin-development.md)
- [Integration Interface](https://github.com/santosr2/uptool/blob/{{ extra.uptool_version }}/internal/engine/types.go) - Core types and Integration interface
- [Integration Registry](https://github.com/santosr2/uptool/blob/{{ extra.uptool_version }}/internal/integrations/registry.go) - Integration registration

Want to add a new datasource? See:

- [NPM Datasource](https://github.com/santosr2/uptool/blob/{{ extra.uptool_version }}/internal/datasource/npm.go) - Example datasource implementation
- [NPM Registry Client](https://github.com/santosr2/uptool/blob/{{ extra.uptool_version }}/internal/registry/npm.go) - Example low-level HTTP client
- [Datasource Package](https://github.com/santosr2/uptool/tree/{{ extra.uptool_version }}/internal/datasource/) - All datasources
- [Registry Package](https://github.com/santosr2/uptool/tree/{{ extra.uptool_version }}/internal/registry/) - All registry HTTP clients

---

## Glossary

- **Manifest**: Configuration file (package.json, Chart.yaml, etc.)
- **Integration**: Package ecosystem support (npm, Helm, etc.)
- **Datasource**: High-level abstraction for querying package versions
- **Registry**: Package repository (npm registry, Artifact Hub, etc.) or low-level HTTP client
- **Dependency**: External package required by the project
- **Update**: Changing a dependency version
- **Policy**: Rules controlling which updates are allowed

---

Last updated: 2025-01-19

# Architecture

uptool's design, components, and data flow.

## Design Principles

1. **Manifest-First**: Update config files directly, not just lockfiles
2. **Format Preservation**: Maintain YAML comments, indentation, structure
3. **Extensibility**: Plugin-based integration architecture

## System Overview

```text
CLI (cmd/uptool)
  ↓
Engine (Scan/Plan/Update)
  ↓
Integrations (npm, helm, terraform, etc.)
  ↓
Datasources (registries) + Rewrite (manifest updates)
```

## Core Components

### CLI Layer (`cmd/uptool`)

Command handlers using Cobra:

- `scan` - Find and parse manifests
- `plan` - Query registries for updates
- `update` - Apply changes to manifests
- `list` - Show available integrations

### Engine Layer (`internal/engine`)

Orchestration logic:

- **Scan**: Parallel integration detection, manifest parsing
- **Plan**: Registry queries, version resolution, update policy
- **Update**: Manifest rewriting, validation, diff generation

### Integration Layer (`internal/integrations`)

Each integration implements:

```go
type Integration interface {
    Name() string
    Detect(ctx context.Context, repoRoot string) ([]*Manifest, error)
    Plan(ctx context.Context, manifest *Manifest, planCtx *PlanContext) (*UpdatePlan, error)
    Apply(ctx context.Context, plan *UpdatePlan) (*ApplyResult, error)
    Validate(ctx context.Context, manifest *Manifest) error
}
```

The `planCtx` parameter provides policy configuration following the precedence order:
`CLI flags > uptool.yaml policy > manifest constraints`.

**Types**:

1. **Manifest Rewriting** (npm, Helm, asdf, mise) - Parse and rewrite files
2. **Native Command** (pre-commit) - Execute tool's update command
3. **Hybrid** (Terraform) - Parse HCL, query registry, rewrite

### Datasource Layer (`internal/datasource`)

Registry abstraction:

- npm Registry API
- Helm/Artifact Hub
- Terraform Registry
- GitHub Releases (for tflint, asdf, mise)

### Rewrite Layer (`internal/rewrite`)

Format-preserving updates:

- **YAML**: Line-by-line rewriting, preserves comments
- **JSON**: Structured rewriting with indentation
- **TOML**: gopkg.in/toml-based updates
- **HCL**: hashicorp/hcl parser

### Resolve Layer (`internal/resolve`)

Semantic version resolution:

- Parse version constraints (`^4.0.0`, `~1.2.3`, `>=2.0.0`)
- Find compatible versions
- Handle pre-release tags

### Policy Layer (`internal/policy`)

Update policies from `uptool.yaml`:

- Maximum update level (none/patch/minor/major)
- Pre-release inclusion
- Pin vs range versions

## Data Flow

### Scan

1. CLI receives `scan` command
2. Engine loads integrations
3. Each integration detects manifests (parallel)
4. Parse manifests, extract dependencies
5. Return list of manifests with dependency counts

### Plan

1. For each manifest, query datasources
2. Resolve latest compatible versions
3. Apply update policy filters
4. Generate UpdatePlan with current → target versions
5. Return plans (shows what would change)

### Update

1. Execute Plan for each manifest
2. Integration rewrites manifest file
3. Generate diff (before/after)
4. Validate updated manifest
5. Return ApplyResult with changes

## Integration Patterns

### Pattern 1: Manifest Rewriting (npm)

```go
func (i *Integration) Apply(ctx context.Context, plan *UpdatePlan) (*ApplyResult, error) {
    data, _ := os.ReadFile(plan.Manifest.Path)
    var pkg PackageJSON
    json.Unmarshal(data, &pkg)

    // Update versions
    for dep, newVer := range plan.Updates {
        pkg.Dependencies[dep] = newVer
    }

    // Write back with formatting
    output, _ := json.MarshalIndent(pkg, "", "  ")
    os.WriteFile(plan.Manifest.Path, output, 0644)
    return &ApplyResult{Updated: len(plan.Updates)}, nil
}
```

### Pattern 2: Native Command (pre-commit)

```go
func (i *Integration) Apply(ctx context.Context, plan *UpdatePlan) (*ApplyResult, error) {
    cmd := exec.CommandContext(ctx, "pre-commit", "autoupdate")
    output, err := cmd.CombinedOutput()
    // Native command updates .pre-commit-config.yaml
    return &ApplyResult{Updated: countUpdates(output)}, err
}
```

## Error Handling

Errors categorized by severity:

- **Fatal**: Stop execution (invalid config, missing binary)
- **Retryable**: Temporary failures (network timeout, rate limit)
- **Skippable**: Non-critical (single integration failure)

## Performance

**Current**:

- Parallel integration detection
- Sequential registry queries (network bound)
- Single-threaded manifest rewriting

**Future**:

- Registry response caching
- Parallel registry queries with semaphore
- Batch updates for monorepos

## Testing Strategy

- **Unit tests**: Per-integration testing with testdata fixtures
- **Integration tests**: End-to-end with real registries
- **Golden file tests**: Manifest rewriting verification
- **Target coverage**: >80% overall, >80% for core engine

## File Structure

```tree
uptool/
├── cmd/uptool/              # CLI entry point
│   ├── main.go
│   └── cmd/                 # Command handlers
├── internal/
│   ├── engine/              # Core orchestration
│   ├── integrations/        # npm, helm, terraform, etc.
│   ├── datasource/          # Registry clients
│   ├── rewrite/             # Format-preserving updates
│   ├── resolve/             # Version resolution
│   └── policy/              # Update policy engine
├── testdata/                # Test fixtures
└── examples/                # Sample configs
```

## Extension Points

1. **New Integration**: Implement `Integration` interface in `internal/integrations/`
2. **New Datasource**: Add client in `internal/datasource/`
3. **New Format**: Add rewriter in `internal/rewrite/`
4. **Plugin**: External `.so` implementing integration interface

## See Also

- [Plugin Development](plugin-development.md) - Create custom integrations
- [CONTRIBUTING.md](CONTRIBUTING.md) - Development guidelines
- [Integration Examples](https://github.com/santosr2/uptool/blob/{{ extra.uptool_version }}/internal/integrations/) - Source code examples

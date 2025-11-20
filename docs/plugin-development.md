# Plugin Development

Create external plugins to extend uptool with custom integrations.

## Overview

**Built-in vs Plugin**:
- **Built-in**: Compiled into uptool (npm, Helm, Terraform) - for widely-used ecosystems
- **Plugin**: External `.so` library - for custom/experimental/proprietary integrations

Plugins allow custom integrations without forking uptool.

## Plugin Interface

Implement the `engine.Integration` interface:

```go
type Integration interface {
    Name() string
    Detect(ctx context.Context, repoRoot string) ([]*Manifest, error)
    Plan(ctx context.Context, manifest *Manifest) (*UpdatePlan, error)
    Apply(ctx context.Context, plan *UpdatePlan) (*ApplyResult, error)
    Validate(ctx context.Context, manifest *Manifest) error
}
```

Export a `RegisterWith` function:

```go
func RegisterWith(register func(name string, constructor func() engine.Integration)) {
    register("yourintegration", New)
}
```

## Creating a Plugin

### 1. Project Structure

```
my-plugin/
├── go.mod
├── plugin.go       # Integration implementation
└── main.go         # Plugin entry point
```

### 2. Implement Integration

```go
// plugin.go
package main

import (
    "context"
    "github.com/santosr2/uptool/internal/engine"
)

type MyIntegration struct{}

func New() engine.Integration {
    return &MyIntegration{}
}

func (i *MyIntegration) Name() string {
    return "myintegration"
}

func (i *MyIntegration) Detect(ctx context.Context, repoRoot string) ([]*engine.Manifest, error) {
    // Find manifest files
    return manifests, nil
}

func (i *MyIntegration) Plan(ctx context.Context, manifest *engine.Manifest) (*engine.UpdatePlan, error) {
    // Query registry for updates
    return plan, nil
}

func (i *MyIntegration) Apply(ctx context.Context, plan *engine.UpdatePlan) (*engine.ApplyResult, error) {
    // Update manifest file
    return result, nil
}

func (i *MyIntegration) Validate(ctx context.Context, manifest *engine.Manifest) error {
    // Validate manifest syntax
    return nil
}
```

### 3. Plugin Entry Point

```go
// main.go
package main

import "github.com/santosr2/uptool/internal/engine"

func RegisterWith(register func(name string, constructor func() engine.Integration)) {
    register("myintegration", New)
}

func main() {}
```

### 4. Build Plugin

```bash
# Build as shared library
go build -buildmode=plugin -o myintegration.so .
```

## Plugin Discovery

uptool searches for plugins in these locations (in order):

1. `./plugins/` - Current directory
2. `~/.config/uptool/plugins/` - User config
3. `/etc/uptool/plugins/` - System-wide
4. `$UPTOOL_PLUGIN_DIR` - Custom location

**Install**:
```bash
mkdir -p ~/.config/uptool/plugins
cp myintegration.so ~/.config/uptool/plugins/
```

**Verify**:
```bash
uptool list --experimental
# Should show "myintegration"
```

## Testing

### Unit Tests

```go
// plugin_test.go
package main

import (
    "context"
    "testing"
)

func TestDetect(t *testing.T) {
    integration := New()
    manifests, err := integration.Detect(context.Background(), "./testdata")
    if err != nil {
        t.Fatal(err)
    }
    if len(manifests) != 1 {
        t.Errorf("expected 1 manifest, got %d", len(manifests))
    }
}
```

### Integration Testing

```bash
# Build and test
go build -buildmode=plugin -o myintegration.so .
cp myintegration.so ~/.config/uptool/plugins/
uptool scan --only=myintegration
```

## Best Practices

1. **Version compatibility**: Match uptool's Go version and dependencies
2. **Error handling**: Return descriptive errors with context
3. **Logging**: Use structured logging, avoid print statements
4. **Context**: Respect context cancellation
5. **Resource cleanup**: Close files, network connections
6. **Testing**: >70% coverage target
7. **Documentation**: Add README with usage examples

## Example Plugin

See [`examples/plugins/python/`](../../examples/plugins/) for a complete example:
- Detects `pyproject.toml`, `requirements.txt`, `Pipfile`
- Queries PyPI registry
- Updates dependency versions

## Distribution

### GitHub Release

```yaml
# .goreleaser.yml
builds:
  - id: plugin
    main: .
    flags:
      - -buildmode=plugin
    goos: [linux, darwin]
    goarch: [amd64, arm64]
```

### Installation Script

```bash
#!/bin/bash
PLUGIN_DIR="${HOME}/.config/uptool/plugins"
mkdir -p "$PLUGIN_DIR"
curl -LO "https://github.com/you/plugin/releases/latest/download/plugin-$(uname -s)-$(uname -m).so"
mv plugin-*.so "$PLUGIN_DIR/myplugin.so"
```

## Limitations

- Plugin must be compiled with same Go version as uptool
- Shared libraries are OS/arch specific
- Cannot modify core engine behavior
- Plugin crashes may crash uptool

## See Also

- [Integration Examples](../internal/integrations/) - Built-in integration code
- [API Reference](api/README.md) - Engine API documentation
- [CONTRIBUTING.md](../CONTRIBUTING.md) - Development guidelines

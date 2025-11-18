# uptool Plugin Examples

This directory contains complete, working examples of external plugins for uptool.

## Overview

Plugins allow you to extend uptool with custom integrations without modifying the core codebase. Each plugin is a standalone Go module compiled as a shared library (.so file) and loaded at runtime.

## Available Examples

### Python (`python/`)

**Status**: Complete working example

Complete plugin for managing Python `requirements.txt` dependencies:

- Detects `requirements.txt` files
- Queries PyPI for latest versions
- Updates version constraints
- Preserves comments and formatting

**Features demonstrated**:

- PyPI JSON API integration
- Requirements.txt parsing
- Version constraint handling
- Comprehensive test suite
- Build automation (Makefile)

See [python/README.md](python/README.md) for details.

## Quick Start

### Build a Plugin

```bash
cd python
make build
```

### Install Plugin

```bash
# User installation (recommended)
make install

# System-wide installation
sudo make install-system
```

### Test Plugin

```bash
# Verify plugin is loaded
uptool list

# Test with example
cd python/testdata
uptool scan --only=python
uptool plan --only=python
```

## Creating Your Own Plugin

### 1. Study the Example

Start by examining the Python example:

```bash
cd python
tree  # See structure
cat README.md  # Read documentation
cat main.go  # See entry point
cat integration.go  # See implementation
```

### 2. Use as Template

Copy the example and modify:

```bash
# Copy example
cp -r python ../my-plugin
cd ../my-plugin

# Update go.mod
sed -i 's/python/my-plugin/g' go.mod

# Modify code
# - main.go: Update plugin name
# - integration.go: Implement your logic
# - Add your registry client
```

### 3. Follow the Guide

See the comprehensive [Plugin Development Guide](../../docs/plugin-development.md) for:

- Plugin architecture
- Interface requirements
- Best practices
- Distribution strategies

## Plugin Structure

Each plugin follows this standard structure:

```
plugin-name/
├── README.md              # Documentation
├── go.mod                 # Go module definition
├── main.go                # Plugin entry point (RegisterWith)
├── integration.go         # Integration implementation
├── parser.go              # Format-specific parser (optional)
├── registry.go            # Registry client (optional)
├── integration_test.go    # Tests
├── testdata/             # Test fixtures
│   └── example-manifest
├── Makefile              # Build automation
└── build.sh              # Build script
```

## Key Files Explained

### `main.go` - Entry Point

```go
package main

import "github.com/santosr2/uptool/internal/engine"

// RegisterWith is called by uptool to register this plugin's integrations.
// This function MUST be exported and have this exact signature.
func RegisterWith(register func(name string, constructor func() engine.Integration)) {
    register("myplugin", New)
}

func main() {
    // Not used in plugin mode
}
```

### `integration.go` - Implementation

Must implement `engine.Integration` interface:

```go
type Integration interface {
    Name() string
    Detect(ctx context.Context, repoRoot string) ([]*Manifest, error)
    Plan(ctx context.Context, manifest *Manifest) (*UpdatePlan, error)
    Apply(ctx context.Context, plan *UpdatePlan) (*ApplyResult, error)
    Validate(ctx context.Context, manifest *Manifest) error
}
```

### `Makefile` - Build Automation

```makefile
build:
    go build -buildmode=plugin -o plugin.so .

install: build
    mkdir -p ~/.uptool/plugins
    cp plugin.so ~/.uptool/plugins/
```

## Plugin Discovery

uptool searches for plugins in these locations (in order):

1. **`./plugins/`** - Current directory (for development)
2. **`~/.uptool/plugins/`** - User plugins (recommended)
3. **`/usr/local/lib/uptool/plugins/`** - System-wide (Unix-like)
4. **`$UPTOOL_PLUGIN_DIR`** - Custom via environment variable

## Testing Plugins

### Unit Tests

```bash
cd python
go test -v ./...
```

### Integration Tests

```bash
# Build plugin
make build
mkdir -p plugins
cp python.so plugins/

# Create test manifest
cat > requirements.txt <<EOF
requests==2.28.0
EOF

# Test with uptool
uptool scan --only=python -v
uptool plan --only=python
uptool update --only=python --dry-run --diff
```

## Distribution

### GitHub Releases

Build for multiple architectures:

```bash
#!/bin/bash
PLUGIN_NAME="python"

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -buildmode=plugin -o ${PLUGIN_NAME}-linux-amd64.so .
GOOS=linux GOARCH=arm64 go build -buildmode=plugin -o ${PLUGIN_NAME}-linux-arm64.so .
GOOS=darwin GOARCH=amd64 go build -buildmode=plugin -o ${PLUGIN_NAME}-darwin-amd64.so .
GOOS=darwin GOARCH=arm64 go build -buildmode=plugin -o ${PLUGIN_NAME}-darwin-arm64.so .

# Create release
gh release create v1.0.0 *.so
```

### Installation Script

Provide easy installation:

```bash
#!/bin/bash
# install.sh
PLUGIN="python"
VERSION="latest"
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

[ "$ARCH" = "x86_64" ] && ARCH="amd64"
[ "$ARCH" = "aarch64" ] && ARCH="arm64"

URL="https://github.com/you/plugin/releases/download/${VERSION}/${PLUGIN}-${OS}-${ARCH}.so"

mkdir -p ~/.uptool/plugins
curl -L "$URL" -o ~/.uptool/plugins/${PLUGIN}.so
echo "✓ Installed ${PLUGIN} plugin"
```

## Development Tips

### Hot Reload

During development, plugins can be reloaded:

```bash
# Terminal 1: Watch and rebuild
while true; do
  make build
  cp plugin.so ~/.uptool/plugins/
  sleep 2
done

# Terminal 2: Test
uptool scan --only=myplugin -v
```

### Debugging

Build with debug symbols:

```bash
go build -buildmode=plugin -gcflags="all=-N -l" -o plugin.so .
```

Enable verbose logging:

```bash
uptool scan --only=myplugin -v
```

### Version Compatibility

Ensure plugin is built with compatible Go and uptool versions:

```go
// version.go
package main

const (
    PluginVersion = "1.0.0"
    RequiredUptoolVersion = "0.1.0"
)
```

## Best Practices

1. **Error Handling**: Provide helpful error messages
2. **Logging**: Use structured logging
3. **Context**: Respect context cancellation
4. **Resources**: Clean up resources properly
5. **Testing**: Include comprehensive tests
6. **Documentation**: Document manifest format and usage

## Common Issues

### Plugin Not Loading

**Symptoms**: Plugin not visible in `uptool list`

**Causes**:

- Plugin not in search path
- Incorrect file extension (.so required)
- Build mode not set to plugin
- RegisterWith function not exported

**Solutions**:

```bash
# Verify plugin location
ls -la ~/.uptool/plugins/

# Verify it's a shared library
file ~/.uptool/plugins/python.so
# Should output: "Mach-O 64-bit dynamically linked shared library"

# Rebuild with correct flags
go build -buildmode=plugin -o python.so .

# Check uptool can find it
uptool list -v
```

### Symbol Not Found

**Symptoms**: Error loading plugin: "symbol not found"

**Cause**: Plugin built with different Go version than uptool

**Solution**: Rebuild plugin with same Go version:

```bash
go version  # Check uptool's Go version
go build -buildmode=plugin -o plugin.so .
```

### Interface Mismatch

**Symptoms**: "RegisterWith has wrong signature"

**Cause**: Incorrect function signature

**Solution**: Ensure exact signature:

```go
func RegisterWith(register func(name string, constructor func() engine.Integration))
```

## Resources

### Documentation

- [Plugin Development Guide](../../docs/plugin-development.md) - Comprehensive guide
- [uptool Documentation](../../docs/README.md) - Main documentation
- [CONTRIBUTING.md](../../CONTRIBUTING.md) - Built-in integration development

### Code Examples

- [Python Plugin](python/) - Complete working example
- [Integration Tests](python/integration_test.go) - Test examples
- [Built-in Integrations](../../internal/integrations/) - Reference implementations

### API Reference

- [engine.Integration](https://pkg.go.dev/github.com/santosr2/uptool/internal/engine#Integration) - Interface definition
- [Go Packages](https://pkg.go.dev/github.com/santosr2/uptool) - Full API documentation

## Contributing

Found a bug or want to improve the examples?

1. Open an issue describing the problem
2. Submit a PR with your improvements
3. Update documentation if needed

## License

Same as uptool - Apache License 2.0

---

**Happy Plugin Development!** 🎉

For questions, open a [Discussion](https://github.com/santosr2/uptool/discussions) or [Issue](https://github.com/santosr2/uptool/issues).

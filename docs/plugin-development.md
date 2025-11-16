# Plugin Development Guide

This guide explains how to create external plugins for uptool, extending it with custom integrations without modifying the core codebase.

## Table of Contents

- [Overview](#overview)
- [Built-in vs Plugin Integrations](#built-in-vs-plugin-integrations)
- [Plugin Architecture](#plugin-architecture)
- [Creating a Plugin](#creating-a-plugin)
- [Plugin Discovery](#plugin-discovery)
- [Building and Installing](#building-and-installing)
- [Testing Plugins](#testing-plugins)
- [Best Practices](#best-practices)
- [Example Plugins](#example-plugins)

## Overview

uptool supports two types of integrations:

1. **Built-in Integrations** - Compiled into the binary (npm, Helm, Terraform, etc.)
2. **Plugin Integrations** - External shared libraries loaded at runtime

Plugins allow you to:
- Add custom integrations without forking uptool
- Develop proprietary integrations
- Experiment with new ecosystems
- Share integrations with the community

## Built-in vs Plugin Integrations

| Aspect | Built-in | Plugin |
|--------|----------|--------|
| **Compile Time** | Compiled with uptool | Compiled separately |
| **Distribution** | Bundled in binary | Separate `.so` file |
| **Loading** | Always available | Loaded at runtime |
| **Performance** | Slightly faster | Minimal overhead |
| **Use Case** | Core integrations | Custom/experimental |
| **Maintenance** | Part of uptool releases | Independent versioning |

**When to use Built-in**:
- Widely-used ecosystems (npm, Terraform, etc.)
- Part of uptool's core value proposition
- Ready for mainline support

**When to use Plugin**:
- Company-specific tools
- Experimental integrations
- Niche ecosystems
- Proprietary manifest formats

## Plugin Architecture

### Plugin Interface

Plugins must implement the standard `engine.Integration` interface:

```go
type Integration interface {
    Name() string
    Detect(ctx context.Context, repoRoot string) ([]*Manifest, error)
    Plan(ctx context.Context, manifest *Manifest) (*UpdatePlan, error)
    Apply(ctx context.Context, plan *UpdatePlan) (*ApplyResult, error)
    Validate(ctx context.Context, manifest *Manifest) error
}
```

### Registration Pattern

Plugins must export a `RegisterWith` function that uptool calls to register integrations:

```go
// RegisterWith is called by uptool to register this plugin's integrations.
// The provided register function should be called with the integration name and constructor.
func RegisterWith(register func(name string, constructor func() engine.Integration)) {
    register("yourintegration", New)
}
```

### Plugin Discovery

uptool searches for plugins in these locations (in order):

1. `./plugins/` - Current directory
2. `~/.uptool/plugins/` - User's home directory
3. `/usr/local/lib/uptool/plugins/` - System-wide (Unix-like systems)
4. `$UPTOOL_PLUGIN_DIR` - Custom directory via environment variable

## Creating a Plugin

### Step 1: Project Structure

Create a new Go module for your plugin:

```bash
mkdir uptool-plugin-example
cd uptool-plugin-example
go mod init github.com/yourname/uptool-plugin-example
```

Project structure:

```
uptool-plugin-example/
├── go.mod
├── go.sum
├── main.go              # Plugin entry point with RegisterWith
├── integration.go       # Integration implementation
├── registry.go          # Optional: Custom registry client
├── testdata/           # Test fixtures
│   └── example.yaml
└── README.md
```

### Step 2: Import Dependencies

```go
// go.mod
module github.com/yourname/uptool-plugin-example

go 1.25

require (
    github.com/santosr2/uptool v0.1.0
    // Add other dependencies as needed
)
```

### Step 3: Implement Integration

```go
// integration.go
package main

import (
    "context"
    "fmt"
    "os"
    "path/filepath"

    "github.com/santosr2/uptool/internal/engine"
)

const integrationName = "example"

// Integration implements the engine.Integration interface for example manifests.
type Integration struct {
    // Add any clients, configuration, etc.
}

// New creates a new integration instance.
func New() engine.Integration {
    return &Integration{}
}

// Name returns the integration identifier.
func (i *Integration) Name() string {
    return integrationName
}

// Detect finds manifest files in the repository.
func (i *Integration) Detect(ctx context.Context, repoRoot string) ([]*engine.Manifest, error) {
    var manifests []*engine.Manifest

    // Walk the repository looking for manifest files
    err := filepath.Walk(repoRoot, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        // Skip directories and non-manifest files
        if info.IsDir() || filepath.Base(path) != "example.yaml" {
            return nil
        }

        // Read and parse manifest
        content, err := os.ReadFile(path)
        if err != nil {
            return fmt.Errorf("reading %s: %w", path, err)
        }

        // Parse dependencies from manifest
        deps, err := i.parseDependencies(content)
        if err != nil {
            return fmt.Errorf("parsing %s: %w", path, err)
        }

        // Create manifest
        manifests = append(manifests, &engine.Manifest{
            Path:         path,
            Integration:  integrationName,
            Dependencies: deps,
        })

        return nil
    })

    if err != nil {
        return nil, err
    }

    return manifests, nil
}

// Plan generates an update plan for a manifest.
func (i *Integration) Plan(ctx context.Context, manifest *engine.Manifest) (*engine.UpdatePlan, error) {
    var updates []*engine.DependencyUpdate

    for _, dep := range manifest.Dependencies {
        // Query registry for latest version
        latestVersion, err := i.getLatestVersion(ctx, dep.Name)
        if err != nil {
            return nil, fmt.Errorf("querying %s: %w", dep.Name, err)
        }

        // Check if update is needed
        if dep.CurrentVersion != latestVersion {
            updates = append(updates, &engine.DependencyUpdate{
                Name:           dep.Name,
                CurrentVersion: dep.CurrentVersion,
                TargetVersion:  latestVersion,
                Impact:         engine.ImpactMinor, // Calculate impact
            })
        }
    }

    return &engine.UpdatePlan{
        Manifest: manifest,
        Updates:  updates,
    }, nil
}

// Apply executes the update plan.
func (i *Integration) Apply(ctx context.Context, plan *engine.UpdatePlan) (*engine.ApplyResult, error) {
    // Read manifest file
    content, err := os.ReadFile(plan.Manifest.Path)
    if err != nil {
        return nil, fmt.Errorf("reading manifest: %w", err)
    }

    // Update versions in content
    updated := string(content)
    for _, update := range plan.Updates {
        // Replace old version with new version
        // Implementation depends on manifest format
        updated = i.replaceVersion(updated, update)
    }

    // Write updated manifest
    if err := os.WriteFile(plan.Manifest.Path, []byte(updated), 0644); err != nil {
        return nil, fmt.Errorf("writing manifest: %w", err)
    }

    return &engine.ApplyResult{
        Success: true,
        Applied: len(plan.Updates),
    }, nil
}

// Validate checks if a manifest is valid.
func (i *Integration) Validate(ctx context.Context, manifest *engine.Manifest) error {
    // Optional: Validate manifest structure
    return nil
}

// Helper methods

func (i *Integration) parseDependencies(content []byte) ([]*engine.Dependency, error) {
    // Parse manifest format and extract dependencies
    // Implementation depends on your manifest format
    return nil, nil
}

func (i *Integration) getLatestVersion(ctx context.Context, packageName string) (string, error) {
    // Query your registry for latest version
    // Implementation depends on your registry
    return "", nil
}

func (i *Integration) replaceVersion(content string, update *engine.DependencyUpdate) string {
    // Replace version in manifest
    // Implementation depends on your manifest format
    return content
}
```

### Step 4: Add Plugin Entry Point

```go
// main.go
package main

import "github.com/santosr2/uptool/internal/engine"

// RegisterWith is called by uptool to register this plugin's integrations.
// This function MUST be exported and have this exact signature.
func RegisterWith(register func(name string, constructor func() engine.Integration)) {
    register("example", New)
}

// main is not used when building as a plugin, but helps during development/testing
func main() {
    // Optional: Add standalone testing code here
}
```

### Step 5: Build Plugin

Build as a Go plugin (shared library):

```bash
# Build plugin
go build -buildmode=plugin -o example.so .

# Verify it's a valid shared library
file example.so
# Output: example.so: Mach-O 64-bit dynamically linked shared library arm64
```

## Plugin Discovery

### Installation Locations

#### Local Development
```bash
# Project-specific plugins
mkdir -p plugins
mv example.so plugins/
```

#### User-Level
```bash
# User plugins (recommended for personal use)
mkdir -p ~/.uptool/plugins
mv example.so ~/.uptool/plugins/
```

#### System-Wide
```bash
# System-wide plugins (requires sudo)
sudo mkdir -p /usr/local/lib/uptool/plugins
sudo mv example.so /usr/local/lib/uptool/plugins/
```

#### Custom Location
```bash
# Use environment variable
export UPTOOL_PLUGIN_DIR=/opt/uptool/plugins
mkdir -p $UPTOOL_PLUGIN_DIR
mv example.so $UPTOOL_PLUGIN_DIR/
```

### Verifying Plugin Loading

```bash
# List all integrations (includes plugins)
uptool list

# Run with verbose logging to see plugin loading
uptool scan -v
```

## Building and Installing

### Development Workflow

```bash
# 1. Build plugin
go build -buildmode=plugin -o example.so .

# 2. Install to local plugins directory
mkdir -p plugins
cp example.so plugins/

# 3. Test with uptool
uptool scan --only=example -v

# 4. Iterate on changes
# Edit code...
go build -buildmode=plugin -o example.so .
cp example.so plugins/
uptool scan --only=example
```

### Build Script

Create `build.sh`:

```bash
#!/bin/bash
set -e

PLUGIN_NAME="example"
VERSION="${VERSION:-dev}"

echo "Building ${PLUGIN_NAME} plugin..."

# Build plugin
go build -buildmode=plugin -o "${PLUGIN_NAME}.so" .

echo "✓ Built ${PLUGIN_NAME}.so"

# Optionally install to user directory
if [ "$1" == "install" ]; then
    INSTALL_DIR="$HOME/.uptool/plugins"
    mkdir -p "$INSTALL_DIR"
    cp "${PLUGIN_NAME}.so" "$INSTALL_DIR/"
    echo "✓ Installed to $INSTALL_DIR"
fi
```

Usage:
```bash
chmod +x build.sh
./build.sh          # Build only
./build.sh install  # Build and install
```

### Makefile

```makefile
PLUGIN_NAME := example
BUILD_DIR := .
INSTALL_DIR := $(HOME)/.uptool/plugins

.PHONY: build install clean test

build:
	@echo "Building $(PLUGIN_NAME) plugin..."
	go build -buildmode=plugin -o $(PLUGIN_NAME).so .

install: build
	@echo "Installing to $(INSTALL_DIR)..."
	mkdir -p $(INSTALL_DIR)
	cp $(PLUGIN_NAME).so $(INSTALL_DIR)/
	@echo "✓ Plugin installed"

clean:
	rm -f $(PLUGIN_NAME).so

test:
	go test -v ./...
```

## Testing Plugins

### Unit Tests

```go
// integration_test.go
package main

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
    integration := New()

    manifest := &engine.Manifest{
        Path:        "testdata/example.yaml",
        Integration: "example",
        Dependencies: []*engine.Dependency{
            {Name: "package1", CurrentVersion: "1.0.0"},
        },
    }

    plan, err := integration.Plan(context.Background(), manifest)
    if err != nil {
        t.Fatalf("Plan() error: %v", err)
    }

    if plan == nil {
        t.Error("Expected non-nil plan")
    }
}
```

### Integration Testing

Test with real uptool binary:

```bash
# Build plugin
go build -buildmode=plugin -o example.so .
mkdir -p plugins
cp example.so plugins/

# Create test repository
mkdir -p testdata
cat > testdata/example.yaml <<EOF
dependencies:
  package1: 1.0.0
  package2: 2.0.0
EOF

# Test scan
uptool scan --only=example

# Test plan
uptool plan --only=example

# Test update (dry-run)
uptool update --only=example --dry-run --diff
```

## Best Practices

### 1. Version Compatibility

Ensure plugin is built with compatible uptool version:

```go
// version.go
package main

const (
    // PluginVersion is this plugin's version
    PluginVersion = "1.0.0"

    // RequiredUptoolVersion is the minimum compatible uptool version
    RequiredUptoolVersion = "0.1.0"
)

// Optionally check version compatibility
func init() {
    // Check uptool version and warn if incompatible
}
```

### 2. Error Handling

Provide helpful error messages:

```go
func (i *Integration) Detect(ctx context.Context, repoRoot string) ([]*engine.Manifest, error) {
    manifests, err := i.findManifests(repoRoot)
    if err != nil {
        return nil, fmt.Errorf("example: detecting manifests in %s: %w", repoRoot, err)
    }
    return manifests, nil
}
```

### 3. Logging

Use structured logging:

```go
import "log/slog"

type Integration struct {
    logger *slog.Logger
}

func New() engine.Integration {
    return &Integration{
        logger: slog.Default().With("integration", "example"),
    }
}

func (i *Integration) Detect(ctx context.Context, repoRoot string) ([]*engine.Manifest, error) {
    i.logger.Info("scanning for manifests", "root", repoRoot)
    // ...
}
```

### 4. Context Handling

Respect context cancellation:

```go
func (i *Integration) Plan(ctx context.Context, manifest *engine.Manifest) (*engine.UpdatePlan, error) {
    for _, dep := range manifest.Dependencies {
        // Check context before expensive operations
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        default:
        }

        version, err := i.queryRegistry(ctx, dep.Name)
        // ...
    }
}
```

### 5. Resource Cleanup

Clean up resources:

```go
type Integration struct {
    client *http.Client
}

func New() engine.Integration {
    return &Integration{
        client: &http.Client{Timeout: 30 * time.Second},
    }
}

// Optional: Implement cleanup if needed
func (i *Integration) Close() error {
    i.client.CloseIdleConnections()
    return nil
}
```

### 6. Testing

Include comprehensive tests:

```bash
# Test plugin before distribution
go test -v ./...
go test -race ./...
go test -cover ./...
```

### 7. Documentation

Document your plugin:

```markdown
# Example Plugin

Custom uptool integration for example.yaml manifests.

## Installation

```bash
mkdir -p ~/.uptool/plugins
curl -LO https://github.com/you/plugin/releases/latest/example.so
mv example.so ~/.uptool/plugins/
```

## Usage

```bash
uptool scan --only=example
uptool plan --only=example
uptool update --only=example
```

## Manifest Format

```yaml
# example.yaml
dependencies:
  package1: 1.0.0
  package2: 2.0.0
```
```

## Example Plugins

Complete example plugins are available in the `examples/plugins/` directory:

- **Python (pip)** - `examples/plugins/python/` - Manages requirements.txt
- **Ruby (bundler)** - `examples/plugins/ruby/` - Manages Gemfile
- **Custom YAML** - `examples/plugins/custom-yaml/` - Generic YAML dependency updater

Each example includes:
- Full source code
- Tests
- Build scripts
- Documentation

## Distribution

### GitHub Releases

Create releases with compiled plugins:

```bash
# Build for multiple architectures
GOOS=linux GOARCH=amd64 go build -buildmode=plugin -o example-linux-amd64.so .
GOOS=linux GOARCH=arm64 go build -buildmode=plugin -o example-linux-arm64.so .
GOOS=darwin GOARCH=amd64 go build -buildmode=plugin -o example-darwin-amd64.so .
GOOS=darwin GOARCH=arm64 go build -buildmode=plugin -o example-darwin-arm64.so .

# Create GitHub release and attach .so files
gh release create v1.0.0 *.so
```

### Installation Script

Provide an easy installation script:

```bash
#!/bin/bash
# install.sh
set -e

PLUGIN_NAME="example"
VERSION="${VERSION:-latest}"
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
esac

DOWNLOAD_URL="https://github.com/you/plugin/releases/download/${VERSION}/${PLUGIN_NAME}-${OS}-${ARCH}.so"

echo "Installing ${PLUGIN_NAME} for ${OS}/${ARCH}..."
mkdir -p ~/.uptool/plugins
curl -L "$DOWNLOAD_URL" -o ~/.uptool/plugins/${PLUGIN_NAME}.so
echo "✓ Installed ${PLUGIN_NAME} plugin"
```

## See Also

- [Integration Development Guide](../CONTRIBUTING.md#adding-a-new-integration) - Built-in integrations
- [Engine Interface](../internal/engine/types.go) - Integration interface definition
- [Example Plugins](../examples/plugins/) - Complete working examples
- [API Documentation](https://pkg.go.dev/github.com/santosr2/uptool) - Go package docs

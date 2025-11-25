# Python uptool Plugin

Example external plugin for uptool that manages Python `requirements.txt` dependencies.

## Overview

This plugin demonstrates how to create an external integration for uptool. It:

- Detects `requirements.txt` files
- Queries PyPI for latest package versions
- Updates version constraints in `requirements.txt`
- Preserves comments and formatting

## Structure

```tree
python/
├── README.md              # This file
├── go.mod                 # Go module definition
├── main.go                # Plugin entry point with RegisterWith
├── integration.go         # Integration implementation
├── pypi.go                # PyPI registry client
├── parser.go              # requirements.txt parser
├── integration_test.go    # Tests
├── testdata/             # Test fixtures
│   └── requirements.txt
├── Makefile              # Build automation
└── build.sh              # Build script
```

## Building

### Prerequisites

- Go 1.25+
- uptool installed

### Build Plugin

```bash
# Build the plugin
make build

# Or use the build script
./build.sh

# Or build manually
go build -buildmode=plugin -o python.so .
```

## Installation

### Option 1: User Installation (Recommended)

```bash
# Build and install
make install

# Or manually
mkdir -p ~/.uptool/plugins
cp python.so ~/.uptool/plugins/
```

### Option 2: Project-Local

```bash
# For testing in a specific project
mkdir -p plugins
cp python.so plugins/
```

### Option 3: System-Wide

```bash
# Requires sudo
sudo make install-system

# Or manually
sudo mkdir -p /usr/local/lib/uptool/plugins
sudo cp python.so /usr/local/lib/uptool/plugins/
```

## Usage

### Scan for requirements.txt

```bash
uptool scan --only=python

# Expected output:
# Type                 Path                    Dependencies
# ----------------------------------------------------------------
# python               requirements.txt        5
```

### Plan Updates

```bash
uptool plan --only=python

# Expected output:
# requirements.txt (python):
# Package          Current         Target          Impact
# --------------------------------------------------------
# requests         2.28.0          2.31.0          minor
# flask            2.2.0           3.0.0           major
# pytest           7.0.0           8.0.0           major
```

### Apply Updates

```bash
# Dry run first
uptool update --only=python --dry-run --diff

# Apply updates
uptool update --only=python
```

## Supported Formats

### Simple Version

```python
requests==2.28.0
flask==2.2.0
pytest>=7.0.0
```

### Version Constraints

```python
requests>=2.28.0,<3.0.0
flask~=2.2.0
pytest>=7.0.0
```

### Comments

```python
# Web framework
flask==2.2.0

# Testing
pytest>=7.0.0  # Latest stable
```

### Extras

```python
requests[security]==2.28.0
flask[async]>=2.2.0
```

## Configuration

Add to `uptool.yaml`:

```yaml
version: 1

integrations:
  - id: python
    enabled: true
    policy:
      update: minor
      allow_prerelease: false
      pin: true
```

## Limitations

1. **PyPI only**: Only queries PyPI (no support for private indexes yet)
2. **No lock file**: Doesn't update `poetry.lock`, `Pipfile.lock`, etc.
3. **Simple parsing**: May not handle very complex version specifications
4. **No dependency resolution**: Doesn't check for dependency conflicts

## Development

### Run Tests

```bash
# Run all tests
make test

# Run with coverage
go test -cover ./...

# Run with race detection
go test -race ./...
```

### Debugging

```bash
# Build with debug symbols
go build -buildmode=plugin -gcflags="all=-N -l" -o python.so .

# Run with verbose logging
uptool scan --only=python -v
```

### Hot Reload

During development, rebuild and reload:

```bash
# Terminal 1: Watch and rebuild
while true; do
  make build
  sleep 2
done

# Terminal 2: Test
uptool scan --only=python
```

## Architecture

### Components

1. **Integration** (`integration.go`):
   - Implements `engine.Integration` interface
   - Orchestrates detection, planning, and updates

2. **Parser** (`parser.go`):
   - Parses `requirements.txt` format
   - Handles comments, constraints, extras

3. **PyPI Client** (`pypi.go`):
   - Queries PyPI JSON API
   - Fetches package metadata and versions

4. **Entry Point** (`main.go`):
   - Exports `RegisterWith` function
   - Registers integration with uptool

### Data Flow

```text
uptool
  ↓
RegisterWith() - Plugin registration
  ↓
Detect() - Find requirements.txt files
  ↓
Parse() - Extract dependencies
  ↓
Plan() - Query PyPI for updates
  ↓
Apply() - Rewrite requirements.txt
```

## Testing

### Unit Tests

Test individual components:

```go
func TestParseRequirements(t *testing.T) {
    input := `requests==2.28.0
flask>=2.2.0  # Web framework`

    deps, err := ParseRequirements(input)
    // Assert expectations
}
```

### Integration Tests

Test with real uptool:

```bash
# Create test environment
mkdir -p testdata
cat > testdata/requirements.txt <<EOF
requests==2.28.0
flask==2.2.0
EOF

# Build plugin
make build
mkdir -p plugins
cp python.so plugins/

# Test
cd testdata
uptool scan --only=python
uptool plan --only=python
```

## Contributing

To improve this plugin:

1. Fork the repository
2. Create a feature branch
3. Make changes
4. Add tests
5. Submit a pull request

## License

Same as uptool - MIT License

## See Also

- [Plugin Development Guide](../../../docs/plugin-development.md)
- [uptool Documentation](../../../docs/index.md)
- [PyPI JSON API](https://warehouse.pypa.io/api-reference/json.html)

# Installation

This page covers various methods to install uptool on your system.

---

## Prerequisites

- **Go 1.25+** (if installing from source)
- **Git** (for version control)
- **Internet connection** (for downloading binaries or building from source)

---

## Installation Methods

### Method 1: Go Install (Recommended for Go users)

If you have Go 1.25+ installed:

```bash
go install github.com/santosr2/uptool/cmd/uptool@latest
```

This will install the latest version of uptool to your `$GOPATH/bin` directory.

!!! tip "Add to PATH"
    Ensure `$GOPATH/bin` is in your `$PATH`:
    ```bash
    export PATH="$PATH:$(go env GOPATH)/bin"
    ```

---

### Method 2: Pre-built Binaries

Download pre-compiled binaries from the [GitHub Releases](https://github.com/santosr2/uptool/releases) page.

=== "Linux (AMD64)"

  ```bash
  curl -LO https://github.com/santosr2/uptool/releases/latest/download/uptool-linux-amd64
  chmod +x uptool-linux-amd64
  sudo mv uptool-linux-amd64 /usr/local/bin/uptool
  ```

=== "Linux (ARM64)"

  ```bash
  curl -LO https://github.com/santosr2/uptool/releases/latest/download/uptool-linux-arm64
  chmod +x uptool-linux-arm64
  sudo mv uptool-linux-arm64 /usr/local/bin/uptool
  ```

=== "macOS (Apple Silicon)"

  ```bash
  curl -LO https://github.com/santosr2/uptool/releases/latest/download/uptool-darwin-arm64
  chmod +x uptool-darwin-arm64
  sudo mv uptool-darwin-arm64 /usr/local/bin/uptool
  ```

=== "macOS (Intel)"

  ```bash
  curl -LO https://github.com/santosr2/uptool/releases/latest/download/uptool-darwin-amd64
  chmod +x uptool-darwin-amd64
  sudo mv uptool-darwin-amd64 /usr/local/bin/uptool
  ```

=== "Windows (AMD64)"

  ```powershell
  # Download from GitHub Releases
  Invoke-WebRequest -Uri https://github.com/santosr2/uptool/releases/latest/download/uptool-windows-amd64.exe -OutFile uptool.exe

  # Move to a directory in your PATH
  Move-Item uptool.exe C:\Windows\System32\uptool.exe
  ```

---

### Method 3: Build from Source

Clone the repository and build from source:

```bash
# Clone the repository
git clone https://github.com/santosr2/uptool.git
cd uptool

# Build the binary
mise run build

# Install to $GOPATH/bin
mise run install

# Or manually copy the binary
sudo cp dist/uptool /usr/local/bin/
```

---

### Method 4: Using mise (Development Environment)

If you use [mise](https://mise.jdx.dev/) for managing development tools:

```bash
# Add to your mise.toml
echo 'uptool = "latest"' >> mise.toml

# Install
mise install
```

---

## Verification

Verify the installation by checking the version:

```bash
uptool version
```

Expected output:

```text
uptool version 0.1.0
```

---

## Configuration

After installation, you may want to configure uptool for your project. See the [Configuration Guide](configuration.md) for details.

---

## Updating uptool

### Go Install

```bash
go install github.com/santosr2/uptool/cmd/uptool@latest
```

### Pre-built Binaries

Download the latest release and replace your existing binary.

### From Source

```bash
cd uptool
git pull origin main
mise run build
sudo cp dist/uptool /usr/local/bin/
```

---

## Uninstallation

### Go Install

```bash
rm $(which uptool)
```

### Manual Installation

```bash
sudo rm /usr/local/bin/uptool
```

---

## Next Steps

- [Quick Start Guide](quickstart.md) - Get started with your first project
- [Configuration](configuration.md) - Learn about uptool configuration
- [GitHub Action Usage](action-usage.md) - Use uptool in CI/CD

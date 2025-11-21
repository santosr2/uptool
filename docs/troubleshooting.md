# Troubleshooting

Common issues and solutions for uptool.

## Quick Diagnostics

```bash
# Run with verbose logging
uptool scan --verbose

# Check version
uptool --version

# Verify manifest files exist
ls package.json Chart.yaml mise.toml .tool-versions
```

## Common Issues

### No manifests detected

**Check**: Run from correct directory, verify manifest files exist, check `.gitignore`

```bash
pwd && ls -la package.json Chart.yaml
uptool scan --verbose
```

### Registry query failed

**Causes**: Network issues, rate limiting, authentication

**Solutions**:

- Test connectivity: `curl -I https://registry.npmjs.org`
- For private packages: Configure `.npmrc` (npm) or `helm repo add` (helm)
- Check rate limits: Use `GITHUB_TOKEN` env var

### Manifest parsing failed

**Check**: Validate syntax with `yamllint`, `jq`, or online validators

```bash
yamllint Chart.yaml
jq . package.json
```

### Permission denied

**Solutions**:

```bash
# Fix binary permissions
chmod +x /usr/local/bin/uptool

# Or install to user directory
go install github.com/santosr2/uptool/cmd/uptool@latest
```

## Installation Issues

### Command not found

**Check PATH**:

```bash
echo $PATH
which uptool

# Add to PATH
export PATH="$PATH:/usr/local/bin"
```

### Installation script fails

**Alternatives**:

```bash
# Download binary directly
curl -LO https://github.com/santosr2/uptool/releases/latest/download/uptool-$(uname -s)-$(uname -m)

# Or build from source
git clone https://github.com/santosr2/uptool
cd uptool && mise run build
```

## Integration-Specific

### npm

**Lockfile out of sync**: Run `npm install` after uptool updates
**Peer dependency conflict**: Check `npm install` output for warnings

### Helm

**Repository not found**: Add repository: `helm repo add <name> <url>`
**API version mismatch**: Update Helm client to compatible version

### Terraform

**Provider constraint invalid**: Check `.tf` syntax with `terraform validate`
**Provider not available**: Run `terraform init` after updates

### mise/asdf

**Tool not installed**: Run `mise install` or `asdf install` after updates

## Performance

### Slow scans

- Use `--only` flag to limit integrations
- Check network latency to registries
- Increase timeout: `--timeout=60s`

### High memory usage

- Scan one integration at a time: `--only=npm`
- Reduce concurrency in large monorepos

## Debug Mode

### Environment Variables

```bash
# Enable debug logging
export UPTOOL_LOG_LEVEL=debug

# Increase timeout
export UPTOOL_TIMEOUT=60

# Use GitHub token (higher rate limits)
export GITHUB_TOKEN=ghp_xxxx
```

### Verbose Output

```bash
# All commands support -v flag
uptool scan -v
uptool plan --verbose
uptool update -v --dry-run
```

## Configuration Issues

### Config not loaded

**Check**: File must be named `uptool.yaml` (not `.yml`) in repository root

### Integration not running

**Verify**:

- `enabled: true` in config
- No CLI overrides (`--exclude`)
- File patterns match: `uptool scan -v`

## Getting Help

1. **Check logs**: Run with `--verbose`
2. **Search issues**: [GitHub Issues](https://github.com/santosr2/uptool/issues)
3. **Ask questions**: [GitHub Discussions](https://github.com/santosr2/uptool/discussions)
4. **Report bugs**: Include `uptool --version`, logs, and manifest example

### Useful Information for Bug Reports

```bash
# System info
uptool --version
go version
uname -a

# Verbose output
uptool scan --verbose > debug.log 2>&1

# Configuration
cat uptool.yaml
```

## See Also

- [Configuration Guide](configuration.md) - Config file options
- [Integration Guides](integrations/README.md) - Integration-specific docs
- [CLI Reference](cli/commands.md) - Command usage

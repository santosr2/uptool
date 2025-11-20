# Overview

Universal, manifest-first dependency updater for multiple ecosystems.

## What is uptool?

Updates **manifest files** directly (package.json, Chart.yaml, *.tf) instead of just lockfiles, preserving your declared dependencies while keeping them current.

## Manifest-First Philosophy

1. Update manifests first (package.json, Chart.yaml, *.tf)
2. Use native commands only when they update manifests (`pre-commit autoupdate` ✅, `npm update` ❌)
3. Then run lockfile updates (`npm install`, `terraform init`)

## Supported Integrations

npm, Helm, Terraform, tflint, pre-commit, asdf (experimental), mise (experimental)

See [Integrations](integrations/README.md) for complete list.

## Key Features

Multi-ecosystem support • Manifest-first updates • CLI & GitHub Action • Safe by default • Concurrent execution • Flexible filtering

## Documentation

- [Installation](installation.md) • [Quick Start](quickstart.md) • [Configuration](configuration.md)
- [CLI Reference](cli/commands.md) • [Integrations](integrations/README.md) • [GitHub Action](action-usage.md)
- [Architecture](architecture.md) • [Main README](../README.md)

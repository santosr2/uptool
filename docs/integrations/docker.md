# Docker Integration

Update Docker image versions in Dockerfiles and docker-compose files.

## Overview

**Integration ID**: `docker`

**Manifest Files**: `Dockerfile`, `Dockerfile.*`, `docker-compose.yml`, `docker-compose.yaml`, `compose.yml`, `compose.yaml`

**Update Strategy**: Text rewriting (preserves formatting and comments)

**Registry**: Docker Hub API

**Status**: ✅ Stable

## What Gets Updated

**Dockerfiles**:

- `FROM` instructions with tagged images (e.g., `FROM node:20` → `FROM node:22`)
- Multi-stage builds (all `FROM` stages are scanned)
- Platform-specific images (e.g., `FROM --platform=linux/amd64 node:20`)

**Docker Compose**:

- `image:` fields in service definitions (e.g., `image: postgres:15` → `image: postgres:16`)

**Not Updated**:

- `FROM scratch` - no versioning needed
- Build args (e.g., `FROM ${BASE_IMAGE}`) - cannot resolve dynamically
- Digest-pinned images (e.g., `node@sha256:...`) - kept for reproducibility
- `latest` tag - no specific version to update from

## Example

**Before** (`Dockerfile`):

```dockerfile
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci

FROM node:20-alpine
WORKDIR /app
COPY --from=builder /app/node_modules ./node_modules
COPY . .
CMD ["node", "server.js"]
```

**After**:

```dockerfile
FROM node:22-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci

FROM node:22-alpine
WORKDIR /app
COPY --from=builder /app/node_modules ./node_modules
COPY . .
CMD ["node", "server.js"]
```

**Before** (`docker-compose.yml`):

```yaml
services:
  db:
    image: postgres:15
    environment:
      POSTGRES_PASSWORD: secret

  redis:
    image: redis:7.2
```

**After**:

```yaml
services:
  db:
    image: postgres:16
    environment:
      POSTGRES_PASSWORD: secret

  redis:
    image: redis:7.4
```

## Integration-Specific Behavior

- **Semantic version filtering**: Only considers semver-like tags (e.g., `16`, `7.2`, `1.0.0`), skips non-version tags like `alpine`, `slim`, `bullseye`
- **Official images**: Handles official Docker Hub images (e.g., `node`, `postgres`) by querying `library/<image>`
- **Custom registries**: Supports images with namespaces (e.g., `myorg/myimage:1.0`)
- **Comment preservation**: All comments and formatting in Dockerfiles are preserved
- **Multi-file support**: Detects all Dockerfiles including `Dockerfile.prod`, `Dockerfile.dev`, etc.

## Configuration

Example `uptool.yaml` configuration:

```yaml
version: 1

integrations:
  - id: docker
    enabled: true
    policy:
      update: minor        # Recommended for production
      allow_prerelease: false
```

## Limitations

1. **Docker Hub only**: Currently only queries Docker Hub; private registries and other registries (ghcr.io, gcr.io) are not supported
2. **No variant handling**: Doesn't track variants like `-alpine`, `-slim` separately
3. **No digest updates**: SHA256 digest-pinned images are not updated

## See Also

- [CLI Reference](../cli/commands.md)
- [Configuration Guide](../configuration.md)
- [GitHub Actions Integration](actions.md) - For updating workflow files

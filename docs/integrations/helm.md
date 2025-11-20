# Helm Integration

Updates Kubernetes Helm chart dependencies in `Chart.yaml` files.

## Overview

**Integration ID**: `helm`

**Manifest Files**: `Chart.yaml`

**Update Strategy**: YAML parsing and rewriting

**Registry**: Helm chart repositories (`index.yaml`)

**Status**: ✅ Stable

## What Gets Updated

Chart dependency versions in the `dependencies` list:

- `dependencies[].version` - Version of each chart dependency

**Not updated**: Chart `version` (your app version) or `appVersion` (packaged app version)

## Example

**Before**:

```yaml
apiVersion: v2
name: my-application
version: 1.0.0
dependencies:
  - name: postgresql
    version: 12.0.0
    repository: https://charts.bitnami.com/bitnami
  - name: redis
    version: 17.0.0
    repository: https://charts.bitnami.com/bitnami
```

**After**:

```yaml
apiVersion: v2
name: my-application
version: 1.0.0     # Unchanged - your chart version
dependencies:
  - name: postgresql
    version: 18.1.8  # Updated
    repository: https://charts.bitnami.com/bitnami
  - name: redis
    version: 23.2.12 # Updated
    repository: https://charts.bitnami.com/bitnami
```

## Integration-Specific Behavior

### Repository Types

| Type | Example | Support |
|------|---------|---------|
| Public | `https://charts.bitnami.com/bitnami` | ✅ Full |
| Private | `https://charts.company.internal` | ✅ With auth |
| OCI Registry | `oci://registry.example.com/charts` | ✅ With config |

### Repository Authentication

For private repositories, configure Helm authentication:

```bash
helm repo add myrepo https://charts.company.internal \
  --username=user \
  --password=pass
```

uptool respects Helm's repository configuration in `~/.config/helm/repositories.yaml`.

### Chart.lock Handling

uptool updates **only** `Chart.yaml`. Run `helm dependency update` after to regenerate lockfile:

```bash
uptool update --only helm
helm dependency update charts/myapp
```

**Monorepo**: Each `Chart.yaml` updated independently.

## Configuration

```yaml
version: 1

integrations:
  - id: helm
    enabled: true
    match:
      files:
        - "Chart.yaml"
        - "charts/*/Chart.yaml"    # Monorepo pattern
    policy:
      update: minor
      allow_prerelease: false
```

## Limitations

1. **No Chart.lock updates**: Only `Chart.yaml` modified. Run `helm dependency update` after.
2. **No version constraint validation**: Test with `helm lint` after updating.
3. **Repository must be configured**: Ensure repositories added via `helm repo add`.

## See Also

- [CLI Reference](../cli/commands.md) - `uptool scan --only helm`
- [Configuration Guide](../configuration.md) - Policy settings
- [Helm Chart Dependencies](https://helm.sh/docs/helm/helm_dependency/)
- [Helm Repositories](https://helm.sh/docs/topics/chart_repository/)

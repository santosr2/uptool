# Helm Integration

The Helm integration updates Kubernetes Helm chart dependencies in `Chart.yaml` files.

## Overview

**Integration ID**: `helm`

**Manifest Files**: `Chart.yaml`

**Update Strategy**: YAML parsing and rewriting

**Registry**: Helm chart repositories (index.yaml)

**Status**: ✅ Stable

## What Gets Updated

The Helm integration updates chart dependency versions in `Chart.yaml`:

- `dependencies[].version` - Version of each chart dependency

## Example

**Before** (`Chart.yaml`):

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
  - name: nginx
    version: 13.0.0
    repository: https://charts.bitnami.com/bitnami
```

**After** (uptool update):

```yaml
apiVersion: v2
name: my-application
version: 1.0.0
dependencies:
  - name: postgresql
    version: 18.1.8
    repository: https://charts.bitnami.com/bitnami
  - name: redis
    version: 23.2.12
    repository: https://charts.bitnami.com/bitnami
  - name: nginx
    version: 18.2.5
    repository: https://charts.bitnami.com/bitnami
```

## Chart Metadata

uptool **does NOT** update:

- Chart `version` (your application version)
- Chart `appVersion` (application version packaged)
- Chart metadata fields

Only dependency versions are updated.

## CLI Usage

### Scan for Helm Charts

```bash
uptool scan --only=helm
```

Output:

```text
Type                 Path                            Dependencies
------------------------------------------------------------------------
helm                 charts/myapp/Chart.yaml         3
helm                 charts/api/Chart.yaml           2
helm                 infra/charts/monitoring/Chart.yaml  5

Total: 3 manifests
```

### Plan Helm Updates

```bash
uptool plan --only=helm
```

Output:

```text
charts/myapp/Chart.yaml (helm):
Chart            Current         Target          Impact
--------------------------------------------------------
postgresql       12.0.0          18.1.8          major
redis            17.0.0          23.2.12         major
nginx            13.0.0          18.2.5          major

charts/api/Chart.yaml (helm):
Chart            Current         Target          Impact
--------------------------------------------------------
mysql            9.0.0           11.1.16         major

Total: 4 updates across 2 manifests
```

### Apply Helm Updates

```bash
# Dry run first
uptool update --only=helm --dry-run --diff

# Apply updates
uptool update --only=helm

# Then update Chart.lock
helm dependency update charts/myapp
helm dependency update charts/api
```

## Helm Repository Types

### Public Repositories

Most common: Bitnami, official charts

```yaml
dependencies:
  - name: postgresql
    version: 18.1.8
    repository: https://charts.bitnami.com/bitnami
```

uptool queries the repository's `index.yaml`:

```bash
curl https://charts.bitnami.com/bitnami/index.yaml
```

### OCI Registries

Helm charts in OCI registries (Docker-like):

```yaml
dependencies:
  - name: my-chart
    version: 1.0.0
    repository: oci://registry.example.com/charts
```

**Note**: OCI registry support depends on Helm client configuration.

### Chart Museum

Private chart repositories:

```yaml
dependencies:
  - name: internal-chart
    version: 2.0.0
    repository: https://charts.company.internal
```

uptool attempts to fetch `index.yaml` from the repository URL.

## Configuration

### Update Policy

```yaml
# uptool.yaml
version: 1

integrations:
  - id: helm
    enabled: true
    policy:
      update: minor              # none, patch, minor, major
      allow_prerelease: false    # Don't update to pre-release versions
```

**Update Levels**:

- `none` - No updates
- `patch` - Only patch updates (12.0.0 → 12.0.1)
- `minor` - Patch + minor updates (12.0.0 → 12.1.0)
- `major` - All updates including major (12.0.0 → 18.0.0)

### Exclude Specific Charts

To pin a chart version, use exact version without range:

```yaml
dependencies:
  - name: critical-chart
    version: 1.2.3    # Won't be updated (exact version)
```

## Repository Authentication

### Private Repositories

Configure Helm authentication separately:

```bash
# Add private repository with credentials
helm repo add myrepo https://charts.company.internal \
  --username=user \
  --password=pass

# Or use token
helm repo add myrepo https://charts.company.internal \
  --username=token \
  --password=$HELM_TOKEN
```

uptool respects Helm's repository configuration in `~/.config/helm/repositories.yaml`.

### Helm Registries

For OCI registries, authenticate with Helm:

```bash
helm registry login registry.example.com -u username
```

## Chart.lock File

uptool **does NOT** update `Chart.lock` (lockfile).

After updating `Chart.yaml`, regenerate the lockfile:

```bash
# For single chart
helm dependency update charts/myapp

# For multiple charts
for dir in charts/*/; do
  helm dependency update "$dir"
done
```

## Monorepo Pattern

Common structure:

```tree
my-monorepo/
├── charts/
│   ├── frontend/
│   │   ├── Chart.yaml       # Updated independently
│   │   └── Chart.lock
│   ├── backend/
│   │   ├── Chart.yaml       # Updated independently
│   │   └── Chart.lock
│   └── database/
│       ├── Chart.yaml       # Updated independently
│       └── Chart.lock
```

Each `Chart.yaml` is scanned and updated independently.

## Limitations

1. **No Chart.lock updates**: uptool only updates `Chart.yaml`
   - Solution: Run `helm dependency update` after

2. **No version constraint validation**: uptool doesn't validate Helm version constraints
   - Solution: Test with `helm lint` after updating

3. **No chart availability check**: Assumes charts exist in repositories
   - Solution: Run `helm dependency build` to verify

4. **Repository access**: Requires repository to be accessible and configured
   - Solution: Ensure `helm repo add` is configured

## Troubleshooting

### Repository Not Found

**Problem**: "Failed to fetch chart from repository"

**Causes**:

1. Repository URL incorrect in `Chart.yaml`
2. Repository not added to Helm
3. Network connectivity issues

**Solutions**:

```bash
# List configured repositories
helm repo list

# Add missing repository
helm repo add bitnami https://charts.bitnami.com/bitnami

# Update repository indexes
helm repo update

# Test chart availability
helm search repo postgresql
```

### Authentication Errors

**Problem**: "403 Forbidden" or "401 Unauthorized"

**Causes**:

1. Private repository requires authentication
2. Credentials expired or incorrect

**Solutions**:

```bash
# Re-add repository with credentials
helm repo remove myrepo
helm repo add myrepo https://charts.company.internal \
  --username=$USER \
  --password=$PASS

# Verify access
helm search repo myrepo/
```

### Chart.lock Out of Sync

**Problem**: After updating, `helm install` fails with dependency errors

**Solution**:

```bash
# Delete Chart.lock and rebuild
rm Chart.lock
helm dependency update .

# Or rebuild dependencies
helm dependency build .
```

### Version Not Available

**Problem**: uptool wants to update to version that doesn't exist

**Causes**:

1. Repository index is stale
2. Chart was removed from repository

**Solutions**:

```bash
# Update repository indexes
helm repo update

# Check available versions
helm search repo postgresql --versions

# Pin to known good version in Chart.yaml
```

## Best Practices

1. **Always regenerate Chart.lock**:

   ```bash
   uptool update --only=helm
   helm dependency update charts/myapp
   git add charts/myapp/Chart.yaml charts/myapp/Chart.lock
   git commit -m "chore(helm): update chart dependencies"
   ```

2. **Test charts after updating**:

   ```bash
   helm lint charts/myapp
   helm template charts/myapp
   helm install --dry-run test charts/myapp
   ```

3. **Review major version updates**:

   ```bash
   # Check release notes for breaking changes
   helm show readme bitnami/postgresql --version 18.1.8
   ```

4. **Use separate PRs for major updates**:

   ```bash
   # Minor/patch updates together
   uptool update --only=helm  # (with policy: minor)

   # Major updates separately per chart
   # Review each carefully
   ```

5. **Keep repository indexes fresh**:

   ```bash
   helm repo update
   ```

6. **Pin critical dependencies**:

   ```yaml
   dependencies:
     - name: critical-database
       version: 12.0.0  # Exact version
   ```

## Helm Version Compatibility

uptool works with Helm v3.x chart repositories.

**Helm v2** (deprecated):

- Not officially supported
- May work if repository format is compatible

**Helm v3**:

- ✅ Fully supported
- Works with `Chart.yaml` v2 (apiVersion: v2)

## See Also

- [Helm Chart Dependencies](https://helm.sh/docs/helm/helm_dependency/)
- [Helm Repositories](https://helm.sh/docs/topics/chart_repository/)
- [Chart.yaml Specification](https://helm.sh/docs/topics/charts/#the-chartyaml-file)
- [Manifest Files Reference](../manifests.md)

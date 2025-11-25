# Quick Start

Get up and running with uptool in 5 minutes!

---

## Prerequisites

Make sure you have uptool installed. If not, see the [Installation Guide](installation.md).

```bash
uptool version
```

---

## Step 1: Initialize Your Project

Navigate to your project directory:

```bash
cd your-project
```

---

## Step 2: Scan for Dependencies

Scan your project to detect supported manifest files and check for outdated dependencies:

```bash
uptool scan
```

Example output:

```text
Found 5 manifest files:
  ✅ package.json (npm)
  ✅ Chart.yaml (helm)
  ✅ main.tf (terraform)
  ✅ .pre-commit-config.yaml (precommit)
  ✅ mise.toml (mise)

Scanning for updates...
```

---

## Step 3: Plan Updates

Preview what would be updated without making changes:

```bash
uptool plan
```

Example output:

```text
Updates available:

npm (package.json):
  - react: 18.2.0 → 18.3.1
  - typescript: 5.0.0 → 5.4.5

terraform (main.tf):
  - aws: 5.0.0 → 5.70.0

precommit (.pre-commit-config.yaml):
  - golangci-lint: v1.63.4 → v2.6.2
```

!!! info "Dry Run"
    The `plan` command never modifies files. It only shows what would change.

---

## Step 4: Apply Updates

Apply the updates with a diff preview:

```bash
uptool update --diff
```

This will:

1. Update manifest files
2. Show a diff of changes
3. Preserve formatting and comments
4. Validate the changes

Example diff output:

```diff
--- package.json
+++ package.json
@@ -5,7 +5,7 @@
   "dependencies": {
-    "react": "^18.2.0",
+    "react": "^18.3.1",
-    "typescript": "^5.0.0"
+    "typescript": "^5.4.5"
   }
 }
```

---

## Step 5: Review and Commit

After uptool applies the updates, review the changes:

```bash
git diff
```

Commit the changes:

```bash
git add .
git commit --signoff -m "chore(deps): update dependencies via uptool"
git push
```

---

## Advanced Usage

### Filter by Integration

Update only specific integrations:

```bash
# Update only npm packages
uptool update --only npm

# Update everything except terraform
uptool update --exclude terraform
```

### Dry Run Mode

Preview changes without applying:

```bash
uptool update --dry-run
```

### Quiet Mode

Suppress informational output (errors only):

```bash
uptool update --quiet
```

### Verbose Mode

Get detailed debug output:

```bash
uptool scan --verbose
```

---

## Configuration File

Create a `uptool.yaml` configuration file to customize behavior:

```yaml
version: 1

integrations:
  - id: npm
    enabled: true
    policy:
      update: minor           # none, patch, minor, major
      allow_prerelease: false

  - id: terraform
    enabled: true
    policy:
      update: major
      allow_prerelease: false

  - id: helm
    enabled: false  # Skip Helm charts
```

See the [Configuration Guide](configuration.md) for more details.

---

## Using as a GitHub Action

Add uptool to your CI/CD pipeline:

```yaml
name: Dependency Updates

on:
  schedule:
    - cron: '0 0 * * 1'  # Weekly on Monday
  workflow_dispatch:

jobs:
  update-dependencies:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Run uptool
        uses: santosr2/uptool@{{ extra.uptool_version }}
        with:
          command: update
          create-pr: true
```

See the [GitHub Action Usage Guide](action-usage.md) for more examples.

---

## Example Configurations

See the [examples/](https://github.com/santosr2/uptool/blob/{{ extra.uptool_version }}/examples/) directory for sample configurations: [uptool.yaml](https://github.com/santosr2/uptool/blob/{{ extra.uptool_version }}/examples/uptool.yaml), [uptool-minimal.yaml](https://github.com/santosr2/uptool/blob/{{ extra.uptool_version }}/examples/uptool-minimal.yaml), [uptool-monorepo.yaml](https://github.com/santosr2/uptool/blob/{{ extra.uptool_version }}/examples/uptool-monorepo.yaml)

---

## Next Steps

- [Configuration Guide](configuration.md) - Customize uptool behavior
- [Integrations](integrations/README.md) - Learn about supported ecosystems
- [GitHub Action Usage](action-usage.md) - Automate dependency updates
- [Plugin Development](plugin-development.md) - Add custom integrations

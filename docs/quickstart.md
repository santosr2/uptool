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
git commit -m "chore(deps): update dependencies via uptool"
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
# Enable/disable specific integrations
integrations:
  npm:
    enabled: true
  terraform:
    enabled: true
  helm:
    enabled: false  # Skip Helm charts

# Update policies
policies:
  npm:
    allow_major: false  # Only minor and patch updates
    allow_prerelease: false

  terraform:
    allow_major: true
    version_constraints: "~>"  # Use pessimistic constraints
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
        uses: santosr2/uptool@v0.1
        with:
          command: update
          create-pr: true
```

See the [GitHub Action Usage Guide](action-usage.md) for more examples.

---

## Example Projects

Check out example configurations:

- **JavaScript Project**: [examples/javascript](https://github.com/santosr2/uptool/tree/main/examples/javascript)
- **Terraform Project**: [examples/terraform](https://github.com/santosr2/uptool/tree/main/examples/terraform)
- **Multi-Language**: [examples/monorepo](https://github.com/santosr2/uptool/tree/main/examples/monorepo)

---

## Common Workflows

### Daily Dependency Scan

```bash
# Add to crontab
0 9 * * * cd /path/to/project && uptool scan
```

### Pre-commit Hook

```bash
# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: uptool-scan
        name: Scan for outdated dependencies
        entry: uptool scan
        language: system
        pass_filenames: false
```

### Integration with CI

```bash
# In your CI script
uptool scan || exit 1
uptool plan --format json > updates.json
```

---

## Troubleshooting

### No Updates Found

If `uptool scan` doesn't find updates:

1. Check that manifest files exist and are valid
2. Verify internet connectivity (uptool queries registries)
3. Run with `--verbose` to see debug output

### Permission Denied

If you get permission errors:

```bash
# On Linux/macOS
sudo chmod +x /usr/local/bin/uptool

# Or install to user directory
go install github.com/santosr2/uptool/cmd/uptool@latest
```

### Rate Limiting

If you hit API rate limits:

```bash
# Use authentication for higher limits
export NPM_TOKEN=your-npm-token
export GITHUB_TOKEN=your-github-token
uptool scan
```

---

## Next Steps

- [Configuration Guide](configuration.md) - Customize uptool behavior
- [Integrations](integrations/README.md) - Learn about supported ecosystems
- [GitHub Action Usage](action-usage.md) - Automate dependency updates
- [Plugin Development](plugin-development.md) - Add custom integrations

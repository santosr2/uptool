# Documentation Templates

This directory contains static HTML templates used by the GitHub Pages documentation site.

## Files

### `index.html`

The main landing page template for the uptool documentation site at `https://santosr2.github.io/uptool/`.

**Purpose**: Provides a clean, GitHub-style landing page with:
- Quick start guide
- Links to API, CLI, and user documentation
- Supported ecosystems table
- GitHub Action usage example

**Usage**: This template is copied to `_site/index.html` during the documentation build process (`.github/workflows/docs-deploy.yml`).

**Important**: Do NOT generate this file at build time. It must exist in the repository as a static file.

## Modifying Templates

To update the documentation site appearance:

1. Edit the HTML template in this directory
2. Test locally if possible
3. Commit changes to the repository
4. The next documentation build will use the updated template

## Template Philosophy

Templates in this directory follow uptool's **manifest-first** philosophy:

- Templates are **source files** (manifests), not generated artifacts
- They exist in the repository, not generated during CI/CD builds
- Changes are versioned and reviewable via Git

This ensures:
- Reproducible builds
- Version-controlled presentation
- No build-time failures from template generation
- Easy local testing

## See Also

- GitHub Pages workflow: `.github/workflows/docs-deploy.yml`
- Documentation structure: `docs/README.md` (if exists)
- Main documentation: `../README.md`

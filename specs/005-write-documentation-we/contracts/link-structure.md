# Contract: Documentation Link Structure

**Date**: 2025-10-27
**Feature**: VitePress Documentation Site
**Type**: Data Structure Contract

## Overview

This contract defines the structure and validation rules for links in the documentation.

## Link Format Specification

### Internal Links (Between Pages)

**Format**:
```markdown
[Link Text](./relative/path/to/file.md)
```

**Rules**:
- MUST use relative paths (not absolute URLs)
- MUST include `.md` extension in source
- MUST point to existing markdown files
- SHOULD use descriptive link text (not "click here")
- Resolved to `/category/file` (without .md) at build time

**Examples**:
```markdown
# From getting-started/quick-start.md to getting-started/configuration.md
[Configure Aether](./configuration.md)

# From guides/torch-integration.md to getting-started/installation.md
[Installation instructions](../getting-started/installation.md)

# From development/testing.md to guides/pipeline-steps.md
[Pipeline step details](../../guides/pipeline-steps.md)
```

### Anchor Links (Within Page)

**Format**:
```markdown
[Section Link](#section-title)
```

**Rules**:
- Anchor MUST match H2-H6 heading in SAME file
- VitePress auto-generates anchors from headings
- Anchor = heading text converted: lowercase, spaces to hyphens, special chars removed
- MUST be lowercase with hyphens

**Examples**:
```markdown
# Same page anchor
[Jump to Installation](#installation-steps)

# Link to specific section in another page
[See Architecture overview](./architecture.md#system-components)
```

### External Links

**Format**:
```markdown
[Link Text](https://external-site.com/path)
```

**Rules**:
- MUST use full https:// URL
- SHOULD have descriptive text
- Timeout validation: 10 second timeout for link checks
- Can be to GitHub, official docs, external resources

**Examples**:
```markdown
[Vue 3 Documentation](https://vuejs.org/)
[Vite Guide](https://vitejs.dev/)
[FHIR Standards](https://www.hl7.org/fhir/)
```

## File Structure Contract

### Required Files

```
docs/
├── index.md                    # Home page (REQUIRED)
├── getting-started/
│   ├── installation.md         # How to install
│   ├── quick-start.md          # First steps
│   └── configuration.md        # Configuration options
├── guides/
│   ├── torch-integration.md    # TORCH feature guide
│   ├── dimp-pseudonymization.md # DIMP feature guide
│   └── pipeline-steps.md       # Pipeline overview
├── api-reference/
│   ├── cli-commands.md         # CLI command reference
│   └── config-reference.md     # Configuration reference
├── development/
│   ├── architecture.md         # System architecture
│   ├── testing.md              # Testing guide
│   ├── contributing.md         # Contribution guide
│   └── coding-guidelines.md    # Code standards
└── .vitepress/
    └── config.ts               # VitePress configuration
```

Each `.md` file MUST:
- Have exactly one H1 heading (becomes page title)
- Use kebab-case filename (e.g., `quick-start.md`)
- Be referenced in `.vitepress/config.ts` sidebar

### Navigation Sidebar Contract

```typescript
// .vitepress/config.ts must define sidebar with structure:
sidebar: {
  '/': [
    {
      text: 'Getting Started',
      items: [
        { text: 'Installation', link: '/getting-started/installation' },
        { text: 'Quick Start', link: '/getting-started/quick-start' },
        { text: 'Configuration', link: '/getting-started/configuration' }
      ]
    },
    {
      text: 'Guides',
      items: [
        { text: 'TORCH Integration', link: '/guides/torch-integration' },
        { text: 'DIMP Pseudonymization', link: '/guides/dimp-pseudonymization' },
        { text: 'Pipeline Steps', link: '/guides/pipeline-steps' }
      ]
    },
    {
      text: 'API Reference',
      items: [
        { text: 'CLI Commands', link: '/api-reference/cli-commands' },
        { text: 'Configuration', link: '/api-reference/config-reference' }
      ]
    },
    {
      text: 'Development',
      items: [
        { text: 'Architecture', link: '/development/architecture' },
        { text: 'Testing', link: '/development/testing' },
        { text: 'Contributing', link: '/development/contributing' },
        { text: 'Coding Guidelines', link: '/development/coding-guidelines' }
      ]
    }
  ]
}
```

**Validation**:
- Each link in sidebar MUST point to existing file
- Link text SHOULD match H1 in target file
- Order in sidebar determines display order
- No links should have `.md` extension in config

## Validation Rules

### Pre-Build Validation

**Internal Links**:
```bash
# Pseudo-code for validation
for each markdown file:
  for each [text](path.md) link:
    if not file_exists(path.md):
      ERROR: broken internal link
    if not path.endswith('.md'):
      ERROR: internal link missing .md extension

for each #anchor link:
  if not anchor_exists_in_target_file:
    ERROR: broken anchor link
```

**Sidebar Config**:
```bash
for each link in sidebar config:
  if not file_exists(linked_file.md):
    ERROR: sidebar link points to non-existent file
```

### Build-Time Validation

- VitePress build MUST complete without errors
- Search index MUST be generated
- All pages MUST render to valid HTML
- No 404 errors for internal links

### Post-Deploy Validation

- All pages MUST be accessible via sidebar navigation
- Search MUST find pages by title and content keywords
- Mobile view MUST render correctly (320px+ width)

## Content Migration Links

When migrating from existing docs:

| Source | Target | Link Type | Update Rule |
|--------|--------|-----------|------------|
| README.md sections | getting-started/ | Internal | `(docs/getting-started/...)` |
| specs/*/spec.md | api-reference/ | Internal | `(/api-reference/...)` |
| specs/*/plan.md | development/architecture.md | Internal | `(../../development/architecture.md)` |
| docs/*.md | appropriate section | Internal | Match destination structure |
| External links | Keep as-is | External | No changes needed |

## Build Output Contract

After successful build:

```
docs/.vitepress/dist/
├── index.html           # Home page
├── getting-started/
│   ├── installation/index.html
│   ├── quick-start/index.html
│   └── configuration/index.html
├── guides/
│   ├── torch-integration/index.html
│   ├── dimp-pseudonymization/index.html
│   └── pipeline-steps/index.html
├── api-reference/
│   ├── cli-commands/index.html
│   └── config-reference/index.html
├── development/
│   ├── architecture/index.html
│   ├── testing/index.html
│   ├── contributing/index.html
│   └── coding-guidelines/index.html
├── assets/
│   ├── style-*.css
│   ├── main-*.js
│   └── search-*.js
└── search/
    └── index.js
```

**Properties**:
- Each `.md` file becomes a directory with `index.html`
- No `.md` extension in final URLs
- Assets cached with hash-based filenames
- Static HTML, CSS, JS (no server required)

## Testing Strategy

**Unit Tests** (pre-migration):
- Link format validation (regex patterns)
- Anchor generation rules

**Integration Tests** (post-build):
- All sidebar links resolve correctly
- No 404s when following links
- Search index includes all pages
- Page titles match sidebar text

**Manual Tests** (pre-launch):
- Navigate via sidebar (all sections accessible)
- Search for key terms (find relevant pages)
- Test internal links (click and verify destination)
- Test on mobile browser (responsiveness)

## Assumptions

- VitePress default link handling behavior
- GitHub Pages serves from `gh-pages` branch
- No custom link rewriting plugins needed
- Standard markdown syntax for links

# Data Model: VitePress Documentation Site

**Date**: 2025-10-27
**Feature**: VitePress Documentation Site
**Status**: Phase 1 Design

## Overview

This document defines the data structures and content organization model for the Aether documentation site.

## Core Entities

### 1. Documentation Page

**Definition**: A single markdown file representing one documented topic

**Fields**:
- `path`: File path relative to `docs/` (e.g., `getting-started/installation.md`)
- `title`: Human-readable page title (from markdown H1)
- `section`: Top-level category (getting-started | guides | api-reference | development)
- `subsection`: Optional second-level category (e.g., "CLI Commands" under api-reference)
- `content`: Raw markdown body
- `metadata`: YAML front-matter (optional, e.g., description, last-updated)

**Validation Rules**:
- Each file MUST have exactly one H1 heading (becomes page title)
- File names MUST be kebab-case (e.g., `quick-start.md`)
- MUST NOT contain broken internal links
- External links SHOULD have descriptive text (not bare URLs)
- Code blocks SHOULD specify language for syntax highlighting

**State**: Static (no modifications needed during documentation lifecycle)

### 2. Navigation Structure

**Definition**: Hierarchical organization of pages for sidebar navigation

**Structure**:
```
docs/
├── index.md (root)
├── getting-started/ (section)
│   ├── installation.md
│   ├── quick-start.md
│   └── configuration.md
├── guides/ (section)
│   ├── torch-integration.md
│   ├── dimp-pseudonymization.md
│   └── pipeline-steps.md
├── api-reference/ (section)
│   ├── cli-commands.md
│   └── config-reference.md
└── development/ (section)
    ├── architecture.md
    ├── testing.md
    ├── contributing.md
    └── coding-guidelines.md
```

**VitePress Config Representation**:
```typescript
// .vitepress/config.ts
sidebar: {
  '/getting-started/': [
    { text: 'Installation', link: '/getting-started/installation' },
    { text: 'Quick Start', link: '/getting-started/quick-start' },
    { text: 'Configuration', link: '/getting-started/configuration' }
  ],
  // ... more sections
}
```

**Validation Rules**:
- Each section folder MUST have at least one .md file
- Sidebar config MUST match actual file structure
- Links in sidebar config MUST not have `.md` extension
- Section order in config determines display order

### 3. Content Link

**Definition**: Reference from one documentation page to another

**Types**:
- **Internal**: Links between pages (markdown links: `[text](./other-file.md)`)
- **Anchor**: Links to sections within pages (markdown anchors: `#section-title`)
- **External**: Links to external websites

**Validation Rules**:
- Internal links MUST use relative paths (not absolute)
- Internal links MUST include `.md` extension in source
- VitePress converts to extension-less URLs at build time
- Anchor links MUST match actual headings in target file
- All links MUST be validated before deployment

**Data Structure**:
```
Link {
  source: DocumentationPage  // Page containing the link
  target: string              // URL or file path
  text: string                // Link text
  type: "internal" | "anchor" | "external"
  valid: boolean              // Result of validation
}
```

### 4. Build Artifact

**Definition**: The compiled static site output by VitePress

**Structure**:
```
docs/.vitepress/dist/
├── index.html
├── getting-started/
│   ├── installation/index.html
│   ├── quick-start/index.html
│   └── configuration/index.html
├── assets/
│   ├── style-*.css
│   └── main-*.js
└── search/
    └── index.js (search index)
```

**Properties**:
- Files are static HTML/CSS/JS (no server required)
- Ready to deploy to GitHub Pages
- Search index pre-generated at build time
- All internal links resolved and validated

**Validation**:
- Build completes without warnings
- All pages render correctly (HTML valid)
- Search index includes all pages
- No 404s when following internal links

### 5. Deployment Target

**Definition**: GitHub Pages repository serving the documentation

**Configuration**:
- **Repository**: `trobanga/aether`
- **Branch**: `gh-pages` (contains compiled site)
- **URL**: `https://trobanga.github.io/aether/`
- **Source**: Automatically from GitHub Pages settings

**Requirements**:
- Branch MUST be created and set as Pages source
- Branch SHOULD contain only build artifacts (not source markdown)
- VitePress generates static HTML that GitHub Pages serves directly

## Content Classification

### By Audience

**End-Users**: Getting-started guides, feature guides, configuration reference
- Assumes: Familiar with command-line basics
- Goal: Set up and run Aether successfully

**Developers**: Architecture docs, testing guide, contributing guide, coding guidelines
- Assumes: Familiar with Go, testing practices
- Goal: Understand codebase, contribute changes

**Operators**: Configuration reference, deployment guide, troubleshooting
- Assumes: Familiar with system administration
- Goal: Deploy and maintain Aether in production

### By Content Type

**Tutorial**: Step-by-step instructions (e.g., quick-start.md)
- Has numbered steps
- Includes example commands/output
- Guides to a working state

**Reference**: Structured information lookup (e.g., cli-commands.md)
- Alphabetically or logically organized
- Defines options, formats, examples
- No narrative flow required

**Conceptual**: Explanatory content (e.g., architecture.md)
- Explains why/how system works
- May include diagrams/examples
- Context-heavy

**Procedural**: Task-focused guides (e.g., contributing.md)
- How to accomplish specific goal
- Prerequisites stated clearly
- Troubleshooting included

## Metadata

### Front-matter Template

Optional YAML front-matter at top of each markdown file:

```markdown
---
title: Page Title                    # Override H1 if needed
description: Short summary
lastUpdated: 2025-10-27
editUrl: https://github.com/...     # Link to edit on GitHub
---

# Page Title (if not in frontmatter)
...
```

### Search Index

VitePress automatically generates search index from:
- Page titles (H1)
- Section headings (H2-H6)
- Page content
- Custom metadata (description)

No manual index maintenance needed.

## Constraints & Validation

| Constraint | Validation Method | When Checked |
|-----------|-------------------|--------------|
| 100% content preservation | Content diff after migration | Before merge |
| Zero broken links | Link validation script | Pre-build check |
| Proper formatting | Markdown linting | CI/CD pipeline |
| Build succeeds | `npm run docs:build` | On every commit |
| Mobile responsive | Browser testing (320px+) | Pre-launch |
| Search works | Manual test of key terms | Pre-launch |

## Implementation Notes

- All entities are immutable (markdown files don't change during runtime)
- VitePress handles HTML generation (no custom logic needed)
- Configuration is declarative (config.ts)
- Links validated via npm package or custom script
- No database or complex state management

## Next Steps

→ Phase 1: Generate contracts/ and quickstart.md

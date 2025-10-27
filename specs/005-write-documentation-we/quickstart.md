# Quick Start: VitePress Documentation Setup

**Date**: 2025-10-27
**Feature**: VitePress Documentation Site
**Audience**: Developers implementing documentation tasks

## Overview

This guide walks through setting up the VitePress documentation site from scratch.

## Prerequisites

- Node.js 18 or later (`node --version`)
- npm or pnpm (`npm --version`)
- Git and GitHub repository access
- Basic familiarity with markdown and terminal

## Phase 1: Initial Setup

### Step 1: Initialize VitePress

```bash
# From repository root
npm install -D vitepress vue

# Create docs directory with example
npx vitepress init docs
# Respond to prompts:
# - Project name: aether
# - Language: English
# - Theme: Default
```

### Step 2: Verify Installation

```bash
cd docs

# Start development server
npm run docs:dev

# Browser opens to http://localhost:5173
# You should see the default VitePress home page
```

### Step 3: Stop Server

```bash
# Press Ctrl+C in terminal
```

## Phase 2: Directory Structure Setup

### Create Content Directories

```bash
# From docs/ directory
mkdir -p getting-started guides api-reference development public/assets

# Create placeholder files to establish structure
touch getting-started/installation.md
touch getting-started/quick-start.md
touch getting-started/configuration.md
touch guides/torch-integration.md
touch guides/dimp-pseudonymization.md
touch guides/pipeline-steps.md
touch api-reference/cli-commands.md
touch api-reference/config-reference.md
touch development/architecture.md
touch development/testing.md
touch development/contributing.md
touch development/coding-guidelines.md
```

### Result: Directory Structure

```
docs/
├── .vitepress/
│   ├── config.ts          # Configuration (auto-generated)
│   ├── cache/
│   └── dist/              # Build output
├── getting-started/
│   ├── installation.md
│   ├── quick-start.md
│   └── configuration.md
├── guides/
│   ├── torch-integration.md
│   ├── dimp-pseudonymization.md
│   └── pipeline-steps.md
├── api-reference/
│   ├── cli-commands.md
│   └── config-reference.md
├── development/
│   ├── architecture.md
│   ├── testing.md
│   ├── contributing.md
│   └── coding-guidelines.md
├── index.md               # Home page
├── package.json
└── public/
    └── assets/
```

## Phase 3: Configure VitePress

### Edit `.vitepress/config.ts`

Key sections to configure:

**1. Site metadata**:
```typescript
export default defineConfig({
  title: 'Aether',
  description: 'CLI for orchestrating DUP pipelines for medical FHIR data',
  base: '/aether/',  // For GitHub Pages subdirectory
})
```

**2. Navigation sidebar**:
```typescript
themeConfig: {
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
}
```

**3. Search configuration**:
```typescript
themeConfig: {
  search: {
    provider: 'local'  // VitePress built-in search
  }
}
```

## Phase 4: Create Home Page

Edit `docs/index.md`:

```markdown
---
layout: home

hero:
  name: Aether
  text: CLI for DUP Pipelines
  tagline: Orchestrate FHIR data processing workflows
  actions:
    - theme: brand
      text: Get Started
      link: /getting-started/installation
    - theme: alt
      text: View Guides
      link: /guides/torch-integration

features:
  - title: Easy Setup
    details: Install and configure in minutes
  - title: FHIR Data Processing
    details: Handle TORCH extractions and DIMP pseudonymization
  - title: Well Documented
    details: Comprehensive guides for users and developers
---
```

## Phase 5: Build and Test

### Build for Production

```bash
npm run docs:build

# Creates docs/.vitepress/dist/ with static site
# Check build output - should show no errors
```

### Verify Build Output

```bash
# List build files
ls -la .vitepress/dist/

# Should contain index.html, assets/, and CSS/JS files
```

### Preview Built Site

```bash
npm run docs:preview

# Opens preview server at http://localhost:4173
# Test navigation and search functionality
```

## Phase 6: Set Up GitHub Deployment

### Create GitHub Actions Workflow

Create `.github/workflows/docs-deploy.yml`:

```yaml
name: Build and Deploy Docs

on:
  push:
    branches: [main]
    paths: ['docs/**', '.github/workflows/docs-deploy.yml']

jobs:
  build-and-deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: 18

      - name: Install dependencies
        run: npm install
        working-directory: docs

      - name: Build docs
        run: npm run docs:build
        working-directory: docs

      - name: Deploy to GitHub Pages
        uses: peaceiris/actions-gh-pages@v3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./docs/.vitepress/dist
```

### Configure GitHub Pages

1. Go to repository Settings → Pages
2. Set "Source" to `Deploy from a branch`
3. Select branch: `gh-pages`
4. Click Save

### First Deployment

```bash
# Push changes to trigger workflow
git add .
git commit -m "docs: initialize VitePress site"
git push origin 005-write-documentation-we
```

Monitor: GitHub Actions tab → docs-deploy workflow → should complete in <2 minutes

## Verification Checklist

Before declaring Phase 1 complete:

- [ ] VitePress initializes without errors
- [ ] Development server starts (`npm run docs:dev`)
- [ ] All page files created in correct directories
- [ ] `config.ts` updated with sidebar configuration
- [ ] Home page (index.md) displays correctly
- [ ] Navigation sidebar shows all sections
- [ ] Build succeeds (`npm run docs:build`)
- [ ] Build output in `.vitepress/dist/` is created
- [ ] Preview shows site with navigation working
- [ ] GitHub Actions workflow configured
- [ ] Pages deployed to GitHub (check Actions status)
- [ ] Site accessible at GitHub Pages URL

## Common Issues

| Problem | Solution |
|---------|----------|
| Node version too old | Update to 18+: `nvm install 18` or download from nodejs.org |
| `npm: command not found` | Install Node.js (includes npm) |
| Build fails with TypeScript errors | Update vitepress: `npm update vitepress` |
| GitHub Actions fails | Check workflow file indentation (YAML is whitespace-sensitive) |
| Site not appearing online | Wait 2-3 minutes, then hard-refresh browser (Ctrl+Shift+R) |

## Next Steps

→ Phase 2: Content migration (existing docs → VitePress structure)
→ Phase 3: Link validation and final deployment

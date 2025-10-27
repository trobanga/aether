# Research: VitePress Documentation Site Setup

**Date**: 2025-10-27
**Feature**: VitePress Documentation Site
**Status**: Research Complete

## Overview

This document consolidates research findings for setting up a VitePress documentation site following patterns from similar projects (Torch, FTSnext).

## Key Research Findings

### 1. VitePress Best Practices

**Decision**: Use VitePress 1.x with default theme + minimal customization

**Rationale**:
- Industry standard for technical documentation (used by Vue, Vite, and many OSS projects)
- Built-in dark/light theme support
- Automatic sidebar generation from file structure
- Built-in full-text search (no external dependencies)
- Excellent mobile responsiveness out-of-the-box

**Alternatives Considered**:
- Docusaurus: More feature-rich but heavier, better for multi-versioned docs
- MkDocs: Python-based, excellent for API docs but less suited to mixed content
- Jekyll: GitHub native but less modern, fewer features
- **Selected**: VitePress (sweet spot for Aether's mixed content needs)

### 2. Documentation Structure

**Decision**: Organize by user role (getting-started, guides, api-reference, development)

**Rationale**:
- Matches typical user journeys: new users → feature guides → API → development
- Clear mental model for navigation
- Supports sidebar hierarchy in VitePress config
- Matches Torch/FTSnext patterns

**Alternative** (rejected): Single flat structure - harder to navigate with 20+ documents

### 3. Content Migration Strategy

**Decision**: Incremental migration with validation at each step

**Steps**:
1. Initialize VitePress with default theme
2. Copy README content to `getting-started/`
3. Organize specs/ docs into appropriate sections
4. Update internal links (README links → markdown links)
5. Validate all links work in built site
6. Deploy to GitHub Pages

**Why this approach**:
- Reduces risk of content loss
- Catches broken links before deployment
- Can be done incrementally in separate PRs

### 4. GitHub Pages Deployment

**Decision**: Automated via GitHub Actions workflow

**Workflow Steps**:
1. Trigger on push to main
2. Install Node.js + dependencies
3. Run `npm run docs:build`
4. Deploy `docs/.vitepress/dist/` to `gh-pages` branch
5. GitHub Pages serves from `gh-pages` automatically

**Configuration**:
- Repository settings: Set Pages source to `gh-pages` branch
- Site URL: `https://trobanga.github.io/aether/` (or similar)

### 5. Search Functionality

**Decision**: Use VitePress built-in search (MiniSearch)

**Why**:
- Zero external dependencies
- Works offline
- Fast (~100-200ms for typical searches)
- Configured via `config.ts`
- Requires no backend

**Alternatives Considered**:
- Algolia: Powerful but adds complexity, requires external service
- meilisearch: Self-hosted option but overkill for small docs
- **Selected**: Built-in search (aligns with KISS principle)

### 6. Code Example Highlighting

**Decision**: Use Shiki (VitePress default) for syntax highlighting

**Supported Languages**: Go, YAML, JSON, Shell, TypeScript, Markdown, etc.
- Automatic detection or explicit language specification
- Theme matches VitePress default (handles dark/light mode)
- No additional configuration needed

### 7. Mobile Responsiveness

**Decision**: Leverage VitePress default responsive design

**Support**:
- Mobile-first CSS (320px and up)
- Touch-friendly navigation
- Readable typography on all sizes
- Default theme includes breakpoints for tablets/mobile

### 8. Link Validation

**Decision**: Implement link validation script (npm package: `check-links` or custom)

**When to run**:
- Pre-build check (fail build if broken links found)
- CI/CD integration
- Can be extended to check external links

**What to validate**:
- Internal markdown links (e.g., `[text](./other-file.md)`)
- Anchor links (e.g., `#section-title`)
- External links (with timeout handling)

### 9. Build Performance

**Decision**: Target <1 minute build time

**Optimization strategies**:
- Use VitePress production build (optimized by default)
- Minimize custom plugins
- Cache dependencies in GitHub Actions

**Typical timeline for 20 docs**: 15-30 seconds with modern hardware

### 10. Versioning Strategy

**Decision**: Version documentation alongside code (git tags only, no separate version branches)

**Approach**:
- Each release gets tag (e.g., `v1.0.0`)
- Documentation updated in same PR as code changes
- Historical docs available via git history
- No multi-version site (too complex for current scale)

**Future**: If needed, can add versioning plugin when docs grow

## Research Conclusions

✅ **Ready to proceed**: All technical decisions documented, no blockers identified.

**Key takeaways**:
1. VitePress provides optimal balance of features vs. simplicity
2. GitHub Actions deployment is straightforward and native to GitHub
3. Built-in features (search, highlighting, responsive design) eliminate need for complex plugins
4. Content migration can be phased safely with validation
5. Aligns with project's KISS and functional programming principles

## Next Steps

→ Proceed to Phase 1: Generate data-model.md, contracts/, and quickstart.md

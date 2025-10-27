# Implementation Plan: VitePress Documentation Site

**Branch**: `005-write-documentation-we` | **Date**: 2025-10-27 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/005-write-documentation-we/spec.md`

## Summary

Build a VitePress-based documentation site that consolidates user guides, API references, and development documentation. The site will provide search functionality, responsive design, and automated deployment to GitHub Pages. Three prioritized user stories: P1 (VitePress setup), P2 (content migration/organization), P3 (CI/CD deployment).

## Technical Context

**Language/Version**: Node.js 18+ (VitePress requirement)
**Primary Dependencies**: VitePress 1.x, Vue 3, Vite (static site generator stack)
**Storage**: File-based (markdown files in `docs/` directory)
**Testing**: Markdown linting, link validation, build verification
**Target Platform**: Static site hosted on GitHub Pages (browser-based)
**Project Type**: Static site generator (documentation site)
**Performance Goals**: Load time <2s on 4G, search results <500ms, build time <1min
**Constraints**: 100% content preservation during migration, zero broken links in final output, mobile-responsive (320px+)
**Scale/Scope**: ~20+ markdown documents, 3-4 top-level navigation sections, multi-role audience (end-user, developer, operator)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Functional Programming
**Status**: ✅ PASS (N/A for documentation site)
- Documentation content is immutable (markdown files)
- Configuration is declarative (VitePress config file)
- No complex stateful logic required

### Principle II: Test-Driven Development
**Status**: ⚠️ MODIFIED APPROACH
- TDD applies to build/deployment automation (GitHub Actions workflows tested before deployment)
- Content migration testing: validation scripts verify all links work, no content lost
- Build verification: automated checks ensure all markdown renders correctly
- *Justification*: Documentation content isn't "code" in traditional sense; focus on process automation and content validation

### Principle III: Keep It Simple, Stupid
**Status**: ✅ PASS
- Use VitePress built-in features (no custom plugins unless necessary)
- Leverage GitHub Pages native support (no complex deployment infrastructure)
- Markdown as-is (minimal processing, leverage VitePress defaults)
- *Simplicity choice*: Static site generator + markdown > custom frontend + database

**GATE RESULT**: ✅ PASS - All principles satisfied or appropriately adapted for documentation domain

## Project Structure

### Documentation (this feature)

```
specs/005-write-documentation-we/
├── plan.md              # This file (implementation plan)
├── research.md          # Phase 0 output (research findings)
├── data-model.md        # Phase 1 output (content structure)
├── quickstart.md        # Phase 1 output (setup guide)
├── contracts/           # Phase 1 output (link structure contract)
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```
docs/                           # VitePress documentation source
├── .vitepress/
│   ├── config.ts               # VitePress configuration
│   ├── theme/                  # Custom theme settings
│   └── dist/                   # Build output (gitignored)
├── index.md                    # Home page
├── getting-started/            # User guides
│   ├── installation.md
│   ├── quick-start.md
│   └── configuration.md
├── guides/                      # Feature guides
│   ├── torch-integration.md
│   ├── dimp-pseudonymization.md
│   └── pipeline-steps.md
├── api-reference/              # API documentation
│   ├── cli-commands.md
│   └── config-reference.md
├── development/                # Developer guides
│   ├── architecture.md
│   ├── testing.md
│   ├── contributing.md
│   └── coding-guidelines.md
└── public/                     # Static assets (images, etc.)
    └── assets/

.github/workflows/
└── docs-deploy.yml            # GitHub Actions deployment workflow

package.json                    # VitePress dependencies
```

**Structure Decision**: Single VitePress project with modular content organization by audience (getting-started for end-users, development for contributors, api-reference for operators). GitHub Pages deployment via Actions workflow.

## Complexity Tracking

No constitution violations. KISS principle fully satisfied with static site generator approach.

## Phase 0: Research - COMPLETE ✅

**Deliverable**: `research.md`

**Findings**:
- VitePress 1.x selected as primary framework (industry standard, minimal configuration)
- Built-in search sufficient (MiniSearch), no external services needed
- Incremental content migration strategy with validation at each step
- GitHub Actions automation for deployment (native to GitHub, no external CI/CD)
- Link validation via npm package (integrated into build process)
- Documentation versioning via git tags (no separate version branches needed for current scale)

**Status**: All unknowns resolved, no blockers identified

## Phase 1: Design - COMPLETE ✅

**Deliverables**: `data-model.md`, `quickstart.md`, `contracts/link-structure.md`

### Data Model (`data-model.md`)
Defines core entities:
- **Documentation Page**: Markdown file with metadata and content
- **Navigation Structure**: Hierarchical sidebar configuration
- **Content Link**: Internal, anchor, and external links with validation rules
- **Build Artifact**: Static site output
- **Deployment Target**: GitHub Pages configuration

### Quick Start Guide (`quickstart.md`)
Step-by-step setup guide:
1. Initialize VitePress with npm
2. Create directory structure
3. Configure VitePress (site metadata, sidebar, search)
4. Create home page
5. Build and test locally
6. Set up GitHub Actions deployment
7. Verification checklist

### Link Structure Contract (`contracts/link-structure.md`)
Specifies:
- Internal link format (relative paths with .md extension)
- Anchor link format (lowercase with hyphens)
- External link rules (https:// URLs with descriptive text)
- File structure requirements
- Sidebar configuration structure
- Validation rules (pre-build, build-time, post-deploy)
- Content migration mapping rules
- Build output directory structure
- Testing strategy

**Status**: Phase 1 design complete, ready for Phase 2 task generation

## Phase 2: Task Generation

Next step: Run `/speckit.tasks` to generate `tasks.md` with:
- User story breakdown into actionable tasks
- Task dependencies and ordering
- Estimation (T-shirt sizes: S/M/L/XL)
- Definition of Done criteria per task
- Risk mitigation strategies

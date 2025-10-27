# Tasks: VitePress Documentation Site

**Input**: Design documents from `/specs/005-write-documentation-we/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/link-structure.md ✅

**Organization**: Tasks organized by user story to enable independent implementation and testing. Tests are NOT included (not explicitly requested in feature spec - focus is on setup and content organization).

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files/directories, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Paths follow project structure: `docs/` for source, `.github/workflows/` for automation

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Initialize VitePress project and basic structure

- [x] T001 [P] Initialize VitePress project structure per plan.md in `docs/` directory
- [x] T002 [P] Create package.json with VitePress, Vue 3, and Vite dependencies in `docs/package.json`
- [x] T003 [P] Initialize Node.js project with npm in `docs/` directory
- [x] T004 [P] Create empty directory structure in `docs/` (getting-started/, guides/, api-reference/, development/, public/assets/)
- [x] T005 Create `.vitepress/config.ts` with basic site metadata in `docs/.vitepress/config.ts`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST complete before any user story can proceed

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T006 [P] Configure VitePress sidebar structure in `docs/.vitepress/config.ts` per link-structure contract
- [x] T007 [P] Configure VitePress search settings (enable built-in MiniSearch) in `docs/.vitepress/config.ts`
- [x] T008 [P] Configure syntax highlighting for code blocks (Go, YAML, JSON, Shell) in `docs/.vitepress/config.ts`
- [x] T009 Create home page template in `docs/index.md` with hero section and feature list
- [x] T010 [P] Create stub files for all navigation pages (12 files total across 4 sections) - use placeholder content
- [x] T011 [P] Configure GitHub Pages base URL in `docs/.vitepress/config.ts` (base: '/aether/')
- [x] T012 Skip (VitePress generates static HTML directly, no Jekyll needed)

**Checkpoint**: ✅ Foundation complete - user story implementation can now begin

---

## Phase 3: User Story 1 - Set up VitePress Documentation Site (Priority: P1) 🎯 MVP

**Goal**: Build and verify a functional VitePress documentation site that renders markdown content with navigation and search

**Independent Test**: Documentation site builds successfully, renders all pages with proper navigation and search functionality, and can be served locally

### Implementation for User Story 1

**Local Build & Development Setup**:

- [x] T013 [US1] Run `npm install` in `docs/` directory to install VitePress and dependencies
- [x] T014 [US1] Verify VitePress development server starts with `npm run docs:dev` in `docs/`
- [x] T015 [US1] Update home page (`docs/index.md`) with proper hero section, tagline, and quick-start links per spec
- [x] T016 [P] [US1] Create getting-started section pages with actual content in `docs/getting-started/`:
  - `installation.md` (prerequisites, installation steps, verification)
  - `quick-start.md` (first steps after installation)
  - `configuration.md` (how to configure Aether)
- [x] T017 [P] [US1] Create guides section pages with actual content in `docs/guides/`:
  - `torch-integration.md` (TORCH feature overview and usage)
  - `dimp-pseudonymization.md` (DIMP feature overview and usage)
  - `pipeline-steps.md` (pipeline architecture and step details)
- [x] T018 [P] [US1] Create api-reference section pages with actual content in `docs/api-reference/`:
  - `cli-commands.md` (all CLI commands and options)
  - `config-reference.md` (configuration file options and examples)
- [x] T019 [P] [US1] Create development section pages with actual content in `docs/development/`:
  - `architecture.md` (system architecture and design)
  - `testing.md` (testing guidelines and strategies)
  - `contributing.md` (contribution workflow and guidelines)
  - `coding-guidelines.md` (code style and standards)
- [x] T020 [US1] Update all internal links in markdown files to use correct relative paths with .md extension per link-structure contract
- [x] T021 [US1] Test VitePress build locally with `npm run docs:build` in `docs/` - verify zero errors
- [ ] T022 [US1] Verify search index is generated and search functionality works for key terms in built site
- [ ] T023 [US1] Test mobile responsiveness (320px+ width) of rendered site in browser developer tools
- [ ] T024 [US1] Verify breadcrumb navigation and sidebar hierarchy display correctly in local preview
- [ ] T025 [US1] Run link validation to ensure all internal links resolve correctly pre-build

**Checkpoint**: ✅ User Story 1 complete - VitePress site builds, renders, navigates, and searches correctly

---

## Phase 4: User Story 2 - Migrate and Organize Documentation Content (Priority: P2)

**Goal**: Consolidate all existing documentation from README.md, specs/, and docs/ into the VitePress site with proper organization and working links

**Independent Test**: All existing documentation accessible from VitePress site with no broken links; content preserved; navigation by role (end-user, developer, operator) works

### Implementation for User Story 2

**Content Migration from Existing Sources**:

- [ ] T026 [US2] Extract and migrate installation/quick-start content from current README.md to `docs/getting-started/` pages
- [ ] T027 [US2] Extract and migrate configuration examples from README.md to `docs/getting-started/configuration.md`
- [ ] T028 [US2] Extract and migrate TORCH integration content from `specs/002-import-via-torch/quickstart.md` to `docs/guides/torch-integration.md`
- [ ] T029 [US2] Extract and migrate DIMP content from README.md and existing docs to `docs/guides/dimp-pseudonymization.md`
- [ ] T030 [US2] Extract and migrate pipeline overview content to `docs/guides/pipeline-steps.md`
- [ ] T031 [US2] Extract and migrate CLI commands from current documentation to `docs/api-reference/cli-commands.md`
- [ ] T032 [US2] Extract and migrate configuration reference to `docs/api-reference/config-reference.md`
- [ ] T033 [US2] Extract and migrate architecture documentation to `docs/development/architecture.md`
- [ ] T034 [US2] Extract and migrate testing guidance to `docs/development/testing.md`
- [ ] T035 [US2] Extract and migrate contributing workflow to `docs/development/contributing.md`
- [ ] T036 [US2] Extract and migrate coding guidelines to `docs/development/coding-guidelines.md`

**Link Update & Validation**:

- [ ] T037 [US2] Audit all old README.md links and update references to point to new VitePress pages
- [ ] T038 [US2] Update all cross-references between documentation pages to use correct markdown link format per contract
- [ ] T039 [US2] Validate all anchor links (#section-titles) match actual headings in target files
- [ ] T040 [US2] Run comprehensive link validation script (all internal, anchor, and external links)
- [ ] T041 [US2] Verify zero broken links in built site with `npm run docs:build` in `docs/`

**Content Organization & Navigation**:

- [ ] T042 [US2] Create clear user role taxonomy (end-user, developer, operator) and organize sidebar accordingly
- [ ] T043 [US2] Update VitePress sidebar config in `docs/.vitepress/config.ts` to reflect final content organization
- [ ] T044 [US2] Create quick navigation links on home page (`docs/index.md`) for common user journeys
- [ ] T045 [US2] Test sidebar navigation - verify all sections and pages accessible in 2 clicks or fewer
- [ ] T046 [US2] Test search functionality - verify search finds pages by keywords (e.g., "TORCH", "installation", "configuration")
- [ ] T047 [US2] Verify all code examples render correctly with proper syntax highlighting

**Checkpoint**: ✅ User Story 2 complete - All documentation migrated, organized by role, with zero broken links

---

## Phase 5: User Story 3 - Configure GitHub Pages Deployment (Priority: P3)

**Goal**: Automate documentation deployment to GitHub Pages via GitHub Actions; ensure updates are live within 3 minutes of push

**Independent Test**: GitHub Actions workflow builds and deploys documentation automatically; site accessible via GitHub Pages; updates appear within 3 minutes of merge to main

### Implementation for User Story 3

**GitHub Actions Workflow Setup**:

- [x] T048 [US3] Create GitHub Actions workflow file at `.github/workflows/docs-deploy.yml`
- [x] T049 [US3] Configure workflow trigger for pushes to main branch with docs/ path filter
- [x] T050 [US3] Configure workflow step: Setup Node.js 18.x environment
- [x] T051 [US3] Configure workflow step: Install dependencies with `npm install` in docs/ directory
- [x] T052 [US3] Configure workflow step: Run build with `npm run docs:build` in docs/ directory
- [x] T053 [US3] Configure workflow step: Deploy `docs/.vitepress/dist/` to `gh-pages` branch using actions-deploy-pages
- [x] T054 [US3] Test workflow file YAML syntax and structure before committing

**GitHub Pages Configuration**:

- [x] T055 [US3] Verify GitHub repository Settings → Pages configured to deploy from `gh-pages` branch
- [x] T056 [US3] Ensure site URL in Pages settings matches configuration in `docs/.vitepress/config.ts` (base: '/aether/')
- [x] T057 [US3] Skip (VitePress static output doesn't require Jekyll prevention)

**Deployment Testing**:

- [ ] T058 [US3] Push sample change to main branch to trigger workflow
- [ ] T059 [US3] Monitor GitHub Actions → docs-deploy workflow → verify build completes within 2 minutes
- [ ] T060 [US3] Verify site appears at GitHub Pages URL and is accessible via browser
- [ ] T061 [US3] Verify documentation loads correctly with proper styling and navigation on GitHub Pages
- [ ] T062 [US3] Test that subsequent pushes to main trigger workflow and update site within 3 minutes
- [ ] T063 [US3] Verify GitHub Pages deployment doesn't show any 404 errors or broken assets

**Checkpoint**: ✅ User Story 3 complete - Documentation automatically deployed to GitHub Pages; updates live within 3 minutes

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final validation, optimization, and documentation quality improvements

- [ ] T064 [P] Verify site loads in under 2 seconds on 4G network using browser dev tools
- [ ] T065 [P] Verify search returns results in under 500ms for typical queries
- [ ] T066 [P] Run final comprehensive link validation (internal, anchor, external) in built site
- [ ] T067 [P] Verify code examples in all pages render with correct syntax highlighting
- [ ] T068 [P] Test site navigation on multiple devices (desktop, tablet, mobile 320px)
- [ ] T069 [P] Verify all external links resolve correctly (90%+ success rate per SC-010)
- [ ] T070 Verify at least 90% of external links in documentation are valid
- [ ] T071 Add CONTRIBUTING.md to repository root with link to `docs/development/contributing.md`
- [ ] T072 Update README.md with prominent link to VitePress documentation site
- [ ] T073 Create GitHub Pages GitHub discussion/announcement about new documentation
- [ ] T074 Run final build verification - zero errors, valid output
- [ ] T075 Verify quickstart.md workflow completes successfully end-to-end

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately ✅
- **Foundational (Phase 2)**: Depends on Setup completion → BLOCKS all user stories
  - Must complete before any story implementation starts
- **User Stories (Phase 3+)**: All depend on Foundational completion
  - **US1 (P1)**: Can start after Foundational → MVP goal
  - **US2 (P2)**: Can start after Foundational → Content migration
  - **US3 (P3)**: Can start after Foundational → Deployment automation
  - Stories can proceed in parallel or sequentially
- **Polish (Phase 6)**: Depends on all desired user stories being complete

### User Story Dependencies

- **US1 (VitePress Setup)**: No dependencies on other stories - fully independent
- **US2 (Content Migration)**: No dependencies on other stories - uses US1 as foundation, fully independent
- **US3 (GitHub Deployment)**: No dependencies on other stories - uses US1/US2 as foundation, fully independent

### Within Each Phase

- Phase 1: All [P] tasks can run in parallel
- Phase 2: Tasks ordered by dependencies (config before stub files)
- Phase 3-5: [P] tasks can run in parallel; sequential tasks depend on earlier stages
- Phase 6: [P] tasks can run in parallel; final verification sequential

---

## Parallel Opportunities

### Phase 1 (Setup)
```
T001, T002, T003, T004 can run in parallel (different files/directories)
T005 depends on T001-T004 (uses config.ts location)
```

### Phase 2 (Foundational)
```
T006, T007, T008, T011 can run in parallel (all edit config.ts but independent sections)
T010 can run in parallel (creates stub files)
T009 depends on T005 (home page needs config)
T012 skipped (no Jekyll dependency needed)
```

### Phase 3 (User Story 1)
```
T016, T017, T018, T019 can run in parallel (different content directories)
T020, T021, T025 can run in parallel (link update + build validation)
T022, T023, T024 sequential (require T021 build to be complete)
```

### Phase 3 Example: Parallel Content Creation
```
Developer A: T016 (getting-started section)
Developer B: T017 (guides section)
Developer C: T018 (api-reference section)
Developer D: T019 (development section)
All complete independently, then T020-T025 validate together
```

### Phase 4 (User Story 2)
```
T026-T036 can run in parallel (different source→destination migrations)
T037-T040 sequential (must audit links, then validate)
T041-T047 sequential (require built site)
```

### Phase 5 (User Story 3)
```
T048-T054 sequential (workflow steps depend on each other)
T055-T057 sequential (setup required before testing)
T058-T063 sequential (deployment steps)
```

### Phase 6 (Polish)
```
T064-T069 can run in parallel (validation tasks)
T070-T075 sequential (final checks)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. **Complete Phase 1**: Setup (T001-T005)
2. **Complete Phase 2**: Foundational (T006-T012) ← CRITICAL, BLOCKS stories
3. **Complete Phase 3**: User Story 1 (T013-T025)
4. **STOP and VALIDATE**:
   - Site builds without errors
   - All pages render correctly
   - Navigation and search work
   - Site accessible via GitHub Pages (T055-T063 first)
5. **Demo/Deploy**: MVP ready - functional documentation site

**Timeline**: ~2-3 days (depends on content writing effort for T016-T019)

### Incremental Delivery (Full Feature)

1. **Foundation Ready** (Phase 1 + 2): VitePress setup complete
2. **Add US1**: VitePress site builds and navigates (Phase 3)
3. **Add US2**: Content migrated from old docs with links working (Phase 4) → Deploy
4. **Add US3**: GitHub Actions automation live (Phase 5) → Continuous deployment enabled
5. **Polish**: Final optimizations and validation (Phase 6)

**Timeline**: ~1 week total

### Parallel Team Strategy

With multiple developers:

1. **Day 1**: All developers complete Phase 1 + Phase 2 together (shared infrastructure)
2. **Day 2-3**:
   - Developer A: T016-T019 (content creation for US1)
   - Developer B: T026-T036 (content migration for US2)
   - Developer C: T048-T054 (workflow setup for US3)
3. **Day 4**: All converge on T020-T025, T037-T047, T055-T063 (validation)
4. **Day 5**: Polish and final verification

---

## Success Criteria (Story Checkpoints)

### ✅ User Story 1 Complete When:
- `npm run docs:build` produces zero errors
- Site renders at http://localhost:4173 with all pages accessible
- Navigation sidebar shows all 4 sections (getting-started, guides, api-reference, development)
- Search finds at least 3 sample pages when searching for key terms
- All pages display with proper syntax highlighting for code blocks
- Mobile view (320px) renders correctly

### ✅ User Story 2 Complete When:
- All 12 content pages migrated from old sources to docs/
- `npm run docs:build` produces zero warnings about broken links
- Link validation script runs cleanly with no errors
- Sidebar navigation shows all content under correct sections
- Search finds pages by keywords from migrated content
- All internal links work (verified by clicking through in browser)
- Content preserved (word count ±5% from original)

### ✅ User Story 3 Complete When:
- GitHub Actions workflow file present at `.github/workflows/docs-deploy.yml`
- Workflow shows successful build+deploy in Actions tab
- Site accessible at GitHub Pages URL
- Documentation loads with correct styling and navigation
- Pushing a test change triggers workflow and updates site within 3 minutes
- No 404 errors on deployed site

---

## Task Completion Tracking

**Total Tasks**: 75
**By Phase**:
- Phase 1 (Setup): 5 tasks
- Phase 2 (Foundational): 7 tasks
- Phase 3 (US1): 13 tasks
- Phase 4 (US2): 22 tasks
- Phase 5 (US3): 16 tasks
- Phase 6 (Polish): 12 tasks

**By Parallelization**:
- [P] Parallelizable: 28 tasks
- Sequential: 47 tasks

**MVP Scope** (Phase 1 + 2 + 3): 25 tasks = ~2-3 days effort
**Full Scope** (All phases): 75 tasks = ~1 week effort

---

## Notes

- [P] tasks = can run in parallel (different directories/files, no dependencies)
- [Story] label (US1, US2, US3) maps task to user story for traceability
- Each user story can be completed and tested independently
- Commits recommended after each phase completion or logical task group
- Stop at any checkpoint to validate that story works independently
- Avoid: combining sequential tasks in parallel, cross-story dependencies that break independence

# Documentation Changelog

Record of all documentation updates performed by the doc-sync skill. Newest entries first.

---

<!-- Format:
### YYYY-MM-DD — [TRIGGER_DESCRIPTION]
- **Trigger**: What code change prompted this update
- **Files Updated**: List of docs modified
- **Changes**: Brief description of what was updated in each file
-->

### 2026-03-22 — Project File Tree & File Registry Created
- **Trigger**: Need for a running inventory of all project files with descriptions for orientation and staleness detection
- **Files Updated**: .github/skills/doc-sync/references/file-tree.md (NEW), .github/skills/doc-sync/SKILL.md, .github/skills/doc-sync/references/changelog.md
- **Changes**:
  - Created `references/file-tree.md` — complete file tree of every file in the project with icons, descriptions, implementation status table, and update protocol
  - Updated `SKILL.md` — added file-tree.md to Tier 5 registry and business-logic-reference.md to Tier 6 registry

### 2026-03-22 — Business Logic Reference Extracted from guimba-debt-tracker
- **Trigger**: Full extraction of business logic from original Python/FastAPI project (guimba-debt-tracker v3.1.0) for behavioral parity during Go rewrite
- **Files Updated**: documentation/prompts/business-logic-reference.md (NEW), documentation/README.md, MASTERPLAN.md, .github/skills/doc-sync/references/changelog.md
- **Changes**:
  - Created `documentation/prompts/business-logic-reference.md` — comprehensive reference covering all 8 domain entities, 3 value objects, 10+ business rule sets, 10 service workflows, auth flows, background jobs, dashboard stats, OCR, audit system, error taxonomy, and 62+ API endpoints
  - Updated `documentation/README.md` — added business-logic-reference.md to directory map and quick links table
  - Updated `MASTERPLAN.md` — added "Business Logic Reference" section before Phase 1 with summary and usage rule

### 2026-03-19 — MongoDB Added (Polyglot Persistence)
- **Trigger**: Decision to add MongoDB alongside PostgreSQL for flexible per-module database choice
- **Files Updated**: MASTERPLAN.md, copilot-instructions.md, go-backend.instructions.md, database.instructions.md, api-builder.agent.md, docker-compose-services/SKILL.md, compose-patterns.md, documentation/project/tech-stack.md, documentation/project/setup-guide.md, documentation/architecture/clean-architecture.md
- **Changes**:
  - MASTERPLAN: Added MongoDB to environment snapshot, Phase 1, Phase 2 (polyglot persistence strategy table, persistence/pg/ + persistence/mongo/ split), Phase 3 tasks, Key Libraries
  - copilot-instructions.md: Updated tech stack with MongoDB
  - go-backend.instructions.md: Infrastructure layer now documents pg/ and mongo/ subdirectories
  - database.instructions.md: Added polyglot persistence section, MongoDB collection/query standards, DB selection guide
  - api-builder.agent.md: Step 4 now offers PostgreSQL or MongoDB per resource
  - docker-compose-services: Added MongoDB service, connection command, health check
  - documentation/: Updated tech-stack.md, setup-guide.md, clean-architecture.md with polyglot persistence example

### 2026-03-19 — Documentation Hub Created
- **Trigger**: Decision to consolidate all system docs, prompts, and Copilot config mirrors into `documentation/`
- **Files Updated**: MASTERPLAN.md, doc-sync/SKILL.md, new `documentation/` directory (11 files)
- **Changes**:
  - MASTERPLAN: Added `documentation/` to folder structure with 6 subdirectories
  - doc-sync/SKILL.md: Added Tier 6 (documentation hub) with 9 files to the registry; added `documentation/README.md` to Tier 1
  - Created: documentation/README.md (index), project/setup-guide.md, project/tech-stack.md, project/conventions.md, architecture/clean-architecture.md, architecture/testing-strategy.md, architecture/design-system.md, api/README.md, prompts/connected-trio-original.md, prompts/design-decisions.md, copilot-config/README.md

### 2026-03-19 — Consolidated Design System Added
- **Trigger**: Decision to enforce UI consistency with consolidated CSS and design tokens
- **Files Updated**: MASTERPLAN.md, nextjs-frontend.instructions.md, doc-sync/SKILL.md, new skill design-system/
- **Changes**:
  - MASTERPLAN: Expanded frontend folder structure with `src/styles/` (tokens, typography, layouts, components), `src/components/ui/` primitives, `tailwind.config.ts`; Phase 5 tasks expanded with design system foundation + UI primitives
  - nextjs-frontend.instructions.md: Replaced simple styling section with full design system architecture (token flow, rules, component hierarchy, violation examples)
  - doc-sync/SKILL.md: Added design-system to Tier 4 and Tier 5 registries
  - New skill: design-system/ with SKILL.md (enforcement rules, violation detection, audit flow, component patterns, tokens↔tailwind connection) and references/token-registry.md (full token inventory, component CSS classes, primitive component list)

### 2026-03-19 — Playwright E2E + Visual Regression Added
- **Trigger**: Decision to add Playwright for full-stack browser E2E and visual regression testing
- **Files Updated**: MASTERPLAN.md, doc-sync/SKILL.md, new skill playwright-testing/
- **Changes**:
  - MASTERPLAN: Added `tests/playwright/` to folder structure with specs/, pages/, fixtures/, helpers/, snapshots/; updated testing strategy table; added Playwright commands; added to Phase 2 tasks and Phase 5 tasks; added to Key Libraries
  - doc-sync/SKILL.md: Added playwright-testing to Tier 4 and Tier 5 registries
  - New skill: playwright-testing/ with SKILL.md (POM, fixtures, full-stack validation, visual regression, troubleshooting) and references/playwright-patterns.md (config, CI, locator strategy, tagging)

### 2026-03-19 — Clean Architecture + Centralized Tests Adoption
- **Trigger**: Architecture decision to adopt Clean Architecture and separate tests folder
- **Files Updated**: MASTERPLAN.md, copilot-instructions.md, go-backend.instructions.md, api-builder.agent.md, go-testing/SKILL.md
- **Changes**: 
  - MASTERPLAN Phase 2: replaced flat `internal/` layout with Clean Architecture layers (domain → usecase → infrastructure → delivery); added `tests/` folder structure (unit/integration/e2e/fixtures/mocks/helpers)
  - MASTERPLAN Phase 3 & 4: updated tasks to reference Clean Architecture paths
  - copilot-instructions.md: updated Go backend section with Clean Architecture rules
  - go-backend.instructions.md: full rewrite to document all 4 layers with dependency rules
  - api-builder.agent.md: updated scaffolding workflow to generate files layer-by-layer (domain first, delivery last)
  - go-testing/SKILL.md: updated test paths to `tests/`, mock paths to `tests/mocks/`, run commands to `./tests/...`

### 2026-03-19 — Initial Setup
- **Trigger**: Phase 0 completion — Copilot customization layer created
- **Files Updated**: All (initial creation)
- **Changes**: Created full documentation layer: copilot-instructions.md, AGENTS.md, 3 path instructions, 3 agents, 5 skills (docker-compose-services, swagger-gen, go-testing, bug-tracker, doc-sync)

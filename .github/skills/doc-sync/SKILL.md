---
name: doc-sync
description: "Keeps all project documentation, instructions, and knowledge files in sync with code changes. Use when user says 'update docs', 'sync documentation', 'update readme', 'refresh instructions', 'update memory', 'what changed', after any feature is completed, after any refactor, or after any bug fix. Also triggers when files in .github/, AGENTS.md, or references/ may be stale."
---

# Documentation Sync — Long-Term Memory Manager

This is the project's **long-term memory**. It ensures all documentation, instructions, skills, and knowledge files stay accurate as the codebase evolves.

## Core Principle
> Code without updated docs is a future hallucination. Every meaningful code change must ripple through the documentation layer.

## The Documentation Registry

These are all the files this skill is responsible for keeping in sync:

### Tier 1: Always Check (high-impact, loaded every session)
| File | Purpose | Update When |
|:---|:---|:---|
| `.github/copilot-instructions.md` | Global project context, tech stack, guardrails | Tech stack changes, new conventions adopted, new guardrails needed |
| `AGENTS.md` | Project overview, naming conventions, error format | Architecture changes, new naming rules, API contract changes |
| `documentation/README.md` | Documentation hub index | New docs added, structure changes |

### Tier 2: Check When Relevant Layer Changes
| File | Purpose | Update When |
|:---|:---|:---|
| `.github/instructions/go-backend.instructions.md` | Go coding rules | New Go patterns adopted, architecture layer changes |
| `.github/instructions/nextjs-frontend.instructions.md` | Frontend coding rules | New frontend patterns, component conventions change |
| `.github/instructions/database.instructions.md` | SQL/migration rules | New DB conventions, migration strategy changes |

### Tier 3: Check When Related Feature/Workflow Changes
| File | Purpose | Update When |
|:---|:---|:---|
| `.github/agents/api-builder.agent.md` | API scaffolding workflow | New libraries added, endpoint patterns change |
| `.github/agents/frontend-builder.agent.md` | Frontend scaffolding workflow | Component patterns change, new UI libraries |
| `.github/agents/db-migrator.agent.md` | Migration workflow | Migration tooling changes, new DB patterns |
| `.github/agents/feature-orchestrator.agent.md` | Full vertical feature slice orchestration | New layers, patterns, or tools change |

### Tier 4: Skills — Check When Domain Expertise Evolves
| File | Purpose | Update When |
|:---|:---|:---|
| `.github/skills/docker-compose-services/SKILL.md` | Docker workflow | New services added, port changes, config changes |
| `.github/skills/swagger-gen/SKILL.md` | Swagger generation | New annotation patterns, tooling changes |
| `.github/skills/go-testing/SKILL.md` | Testing patterns | New test patterns discovered, testing tools change |
| `.github/skills/bug-tracker/SKILL.md` | Bug tracking process | Process improvements, new escalation rules |
| `.github/skills/playwright-testing/SKILL.md` | Playwright E2E & visual regression | New test patterns, config changes, POM conventions change |
| `.github/skills/design-system/SKILL.md` | Consolidated CSS, design tokens, UI primitives | New tokens, components, or styling patterns adopted |
| `.github/skills/auth-rbac/SKILL.md` | JWT auth, RBAC, login/register flows | Auth patterns, role hierarchy, or middleware changes |
| `.github/skills/api-client/SKILL.md` | Frontend API client, typing, error handling | API client patterns, error format, or auth flow changes |
| `.github/skills/error-handling/SKILL.md` | Standardized error codes and propagation | New error codes, handler patterns, or error format changes |
| `.github/skills/ci-cd/SKILL.md` | GitHub Actions CI/CD pipelines | Pipeline stages, caching, or deployment strategy changes |
| `.github/skills/security-hardening/SKILL.md` | CORS, CSP, rate limiting, OWASP | Security middleware, headers, or policy changes |
| `.github/skills/redis-caching/SKILL.md` | Redis caching patterns and invalidation | Cache strategy, key naming, or TTL changes |
| `.github/skills/seed-data/SKILL.md` | Database seed data and test fixtures | New entities, fixture format, or seed runner changes |
| `.github/skills/env-config/SKILL.md` | Environment config, .env, secret handling | New config fields, environment hierarchy changes |
| `.github/skills/bug-tracker/references/bug-log.md` | Resolved bugs history | Every bug fix (handled by bug-tracker skill) |

### Tier 5: Reference Files — Deep Knowledge
| File | Purpose | Update When |
|:---|:---|:---|
| `.github/skills/docker-compose-services/references/compose-patterns.md` | Docker Compose patterns | New patterns discovered, config best practices change |
| `.github/skills/swagger-gen/references/annotation-guide.md` | Swagger annotation reference | New annotation types used |
| `.github/skills/go-testing/references/test-patterns.md` | Test code patterns | New testing idioms adopted |
| `.github/skills/playwright-testing/references/playwright-patterns.md` | Playwright config, POM, visual regression patterns | New browser test patterns, CI config changes |
| `.github/skills/design-system/references/token-registry.md` | Color/spacing/typography tokens, UI component inventory | New tokens, new primitives, variant changes |
| `.github/skills/auth-rbac/references/auth-patterns.md` | JWT claims, middleware chain, role hierarchy | Auth implementation patterns change |
| `.github/skills/api-client/references/client-patterns.md` | API client template, error types, pagination | Client implementation changes |
| `.github/skills/error-handling/references/error-codes.md` | Error code table, response examples, handler helpers | New error codes or response format changes |
| `.github/skills/ci-cd/references/workflow-templates.md` | GitHub Actions workflow templates | Pipeline configuration changes |
| `.github/skills/security-hardening/references/security-checklist.md` | OWASP checklist, middleware stack, CSP template | Security requirements or middleware changes |
| `.github/skills/redis-caching/references/cache-patterns.md` | Cache wrapper template, key naming, TTL cheat sheet | Caching patterns or strategy changes |
| `.github/skills/seed-data/references/fixture-templates.md` | SQL/JSON fixture templates, seed runner code | New entities or fixture format changes |
| `.github/skills/env-config/references/config-template.md` | Config struct, loader, .env.example template | Config fields or loading strategy changes |
| `.github/skills/doc-sync/references/changelog.md` | Record of all doc updates | Every time this skill runs |

### Tier 6: Documentation Hub (`documentation/`)
| File | Purpose | Update When |
|:---|:---|:---|
| `documentation/project/setup-guide.md` | Dev environment setup | New tools, changed commands, new services |
| `documentation/project/tech-stack.md` | Tech stack with rationale | New libraries adopted, tools changed |
| `documentation/project/conventions.md` | Naming, commit, branch rules | Conventions change |
| `documentation/architecture/clean-architecture.md` | Architecture layers explained | Layer structure changes, new patterns |
| `documentation/architecture/testing-strategy.md` | Test pyramid and locations | New test types, tooling changes |
| `documentation/architecture/design-system.md` | CSS & token architecture | Design system evolves |
| `documentation/api/README.md` | API overview | API versioning or auth changes |
| `documentation/prompts/design-decisions.md` | Prompt architecture rationale | New skills added, architecture decisions |
| `documentation/copilot-config/README.md` | Copilot config index | New skills, agents, or instructions added |

## Workflow

### After Any Code Change (Feature, Fix, or Refactor)

#### Step 1: Identify Impact Scope
Determine what kind of change was made:
- **New feature** → check Tier 1-4 (does the feature add new conventions, patterns, or capabilities?)
- **Bug fix** → check Tier 1 (guardrails), Tier 4 (bug-tracker), Tier 5 (test patterns if test-related)
- **Refactor** → check Tier 1-3 (did the architecture or conventions shift?)
- **New dependency** → check Tier 1 (tech stack list), relevant Tier 3 agent, relevant Tier 4 skill
- **Config change** → check Tier 1 (if global), relevant Tier 4 skill

#### Step 2: Scan Each Affected File
For each file in the impact scope:
1. Read the current content
2. Compare against the actual state of the codebase
3. Identify stale, missing, or incorrect information

#### Step 3: Apply Updates
- Use **diff-style edits** — never rewrite entire files
- Preserve existing structure and voice
- Add new information in the correct section
- Mark removed features/patterns as deprecated before deleting

#### Step 4: Log the Update
Append an entry to `references/changelog.md`:
```markdown
### YYYY-MM-DD — [TRIGGER_DESCRIPTION]
- **Trigger**: What code change prompted this update
- **Files Updated**: List of docs modified
- **Changes**: Brief description of what was updated in each file
```

## When Invoked Manually ("update docs" / "sync documentation")

Perform a **full audit**:
1. Walk through ALL tiers (1-5)
2. For each file, verify accuracy against current codebase state
3. Report findings as a checklist:
   ```
   ✅ copilot-instructions.md — up to date
   ⚠️ go-backend.instructions.md — missing new validator pattern (updated)
   ✅ api-builder.agent.md — up to date
   ❌ docker-compose-services/SKILL.md — ports changed (updated)
   ```
4. Apply all fixes
5. Log everything in `references/changelog.md`

## Rules
- **Never delete documented knowledge** unless the feature/pattern is confirmed removed from codebase
- **Never rewrite from scratch** — always use targeted edits
- **Always log changes** in `references/changelog.md` for auditability
- **Cross-reference skills** — if a doc update reveals a gap in a skill, flag it for update
- **Keep files under 5000 words** — move overflow content to `references/` subfolder

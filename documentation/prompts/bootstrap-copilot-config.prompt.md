# Bootstrap `.github/` and `.vscode/` Copilot Configuration for a New Project

> **Usage**: Copy everything below the `---` line and paste it as a prompt to GitHub Copilot (CLI, VS Code chat, or any LLM). Fill in the `[PLACEHOLDERS]` with your project's details first.

---

## The Prompt

```
You are a senior developer setting up GitHub Copilot's project-aware configuration for a new codebase. Your goal is to generate the full `.github/` and `.vscode/` directory structures that give Copilot deep project context — custom instructions, reusable skills, specialized agents, MCP server connections, and guardrails.

## My Project Details

- **Project name**: [PROJECT_NAME]
- **Description**: [ONE_SENTENCE_DESCRIPTION]
- **Tech stack**:
  - Backend: [LANGUAGE/FRAMEWORK — e.g., Go/Chi, Python/FastAPI, Node/Express, Rust/Axum]
  - Frontend: [FRAMEWORK — e.g., Next.js 15, React + Vite, SvelteKit, Vue/Nuxt]
  - Database(s): [e.g., PostgreSQL, MongoDB, SQLite, Redis]
  - Auth: [e.g., JWT, OAuth2, session-based, Clerk, Supabase Auth]
  - Other tools: [e.g., Docker, Swagger, Terraform, Prisma, SQLC]
- **Architecture pattern**: [e.g., Clean Architecture, MVC, Hexagonal, Monolith, Microservices]
- **Monorepo or multi-repo?**: [e.g., monorepo with backend/ + frontend/, single Go module, Nx workspace]
- **Primary language naming conventions**: [e.g., Go: PascalCase exported, camelCase unexported; TS: camelCase vars, PascalCase types]
- **API style**: [e.g., REST with /api/v1/ prefix, GraphQL, gRPC, tRPC]
- **Testing strategy**: [e.g., Go table-driven tests, Jest, Vitest, Playwright E2E]
- **CI/CD**: [e.g., GitHub Actions, GitLab CI, none yet]
- **Docker services needed**: [e.g., PostgreSQL, Redis, MongoDB — list with ports and credentials if known]

---

## What To Generate

### 1. `.github/copilot-instructions.md` — Global Project Context

Create a comprehensive project instructions file. This is loaded into every Copilot session. Structure it as:

```markdown
# [PROJECT_NAME] Project Instructions

## Project Overview
[One paragraph describing what this project does]

## Tech Stack
[Bullet list of all technologies with versions]

## Coding Standards
### [Backend Language]
- Architecture rules (layer dependencies, where things go)
- Error handling conventions
- Naming conventions
- Import ordering rules

### [Frontend Framework]
- Component conventions (server vs client, file organization)
- State management approach
- Styling approach

### Database
- Migration format and tooling
- Naming conventions (tables, columns, indexes, foreign keys)
- Query safety rules

## Guardrails

### Anti-Redundancy Check
Before generating new code, verify:
1. Does a similar function/handler/component already exist?
2. Can an existing utility be reused or extended?
3. Check skill descriptions — has a skill already been created for this task?
If duplication is found, extend or refactor instead of creating new files.

### Structured Bug Logging Convention
When a bug is encountered and resolved, log it in `.github/skills/bug-tracker/references/bug-log.md`:
[Include the 5-field format: Issue, Root Cause, Resolution, Files Changed, Prevention]

### Before Debugging
1. Search bug-log.md for related keywords
2. Check if the same root cause has appeared before
3. If a match is found, apply the documented resolution first

### Documentation Sync (Long-Term Memory)
After completing any meaningful code change:
1. Consider which documentation files may be affected
2. Use the doc-sync skill to update all relevant docs
3. Every doc update is logged in `.github/skills/doc-sync/references/changelog.md`

### Iterative Refinement
- Do not rewrite from scratch — use targeted, diff-style changes
- Preserve existing working logic
- Only modify what's necessary

## MCP Servers (Tool Providers)
[Table of configured MCP servers with columns: Server | What It Provides | When to Use]

## Available Skills & Agents
[Tables listing all skills and agents with their purpose]
```

### 2. `.github/instructions/` — Path-Specific Rules

Create one instruction file per major code domain. Each file uses frontmatter `applyTo` globs so Copilot automatically applies the rules when editing matching files.

**Pattern**:
```markdown
---
applyTo: "[GLOB_PATTERN]"
---
# [Domain] Instructions

## Architecture / Layer Rules
[Layer-by-layer breakdown with dependency rules]

## MCP-Assisted Development
[Which MCP servers to use before writing code in this domain]

## Conventions
[Domain-specific naming, patterns, anti-patterns]
```

**Generate these files** (adapt to your stack):
- `[backend-language].instructions.md` → applies to `backend/**/*.[ext]`
- `database.instructions.md` → applies to `**/*.sql,[migration_path]/**,[persistence_path]/**`
- `[frontend-framework].instructions.md` → applies to `frontend/**/*.{ts,tsx,js,jsx,vue,svelte}`

### 3. `.github/agents/` — Specialized Workflow Agents

Create agents for repetitive multi-step workflows. Each agent file follows this format:

```markdown
---
name: [agent-name]
description: "[Trigger phrases and use cases]"
---
# [Agent Title]

[Narrative explaining what this agent does]

## When to Use
[Trigger phrases that invoke this agent]

## Workflow
### Step 1: [Title]
[Detailed instructions]

### Step 2: [Title]
[Detailed instructions]

[Continue for all steps...]

## Rules
- [Hard rule 1]
- [Hard rule 2]

## File Locations
[Map of where each artifact goes in the project tree]
```

**Generate these agents** (adapt to your stack):
- **api-builder** — Scaffolds complete API endpoints (entity → handler)
- **db-migrator** — Handles database schema changes and migration files
- **frontend-builder** — Creates pages, components, hooks
- **feature-orchestrator** — Orchestrates full vertical features (DB + backend + frontend)

### 4. `.github/skills/` — Reusable Domain Knowledge

Each skill is a directory with a `SKILL.md` and a `references/` folder:

```
.github/skills/[skill-name]/
├── SKILL.md                    # Skill definition (triggers, workflow, rules)
└── references/
    └── [topic].md              # Deep knowledge, templates, patterns
```

**SKILL.md format**:
```markdown
---
name: [skill-name]
description: "[Trigger phrases that activate this skill]"
---
# [Skill Title]

## Purpose
[What this skill manages]

## When to Use
[Trigger phrases]

## Workflow
[Step-by-step process]

## Rules
[Hard constraints]
```

**Generate these skills** (adapt to your stack):
- **bug-tracker** — Persistent bug memory with structured logging format and pattern recognition. Include empty `references/bug-log.md` template.
- **doc-sync** — Documentation sync orchestrator with 5-tier registry (global → layer → workflow → domain → reference). Include `references/changelog.md` and `references/file-tree.md`.
- **error-handling** — Standardized error types, codes, and propagation patterns.
- **auth-rbac** — Authentication (JWT/session), authorization (RBAC), login/register flows.
- **api-client** — Frontend API client patterns, typed requests, error handling, token refresh.
- **docker-compose-services** — Local dev environment setup (databases, cache, etc.).
- **go-testing** / **jest-testing** / **vitest-testing** — Test patterns for your backend/frontend.
- **env-config** — Environment variable management, .env files, secret handling.
- **ci-cd** — CI/CD pipeline configuration and conventions.
- **security-hardening** — CORS, CSP, rate limiting, input sanitization.
- **design-system** — Frontend styling tokens, component library patterns (if applicable).
- **seed-data** — Database seed data and test fixtures.

Only generate skills that are relevant to the tech stack described above. Skip any that don't apply.

### 5. `.github/copilot/mcp.json` — MCP Server Configuration

```json
{
  "mcpServers": {
    // Only include servers relevant to the project's tech stack
  }
}
```

**Available MCP servers to include** (only add what's relevant):

| Server | Package | When to Include |
|:---|:---|:---|
| `postgres` | `@modelcontextprotocol/server-postgres` | Project uses PostgreSQL |
| `mongodb` | `mongodb-mcp-server` | Project uses MongoDB |
| `redis` | `@modelcontextprotocol/server-redis` | Project uses Redis |
| `memory` | `@modelcontextprotocol/server-memory` | Always (session context) |
| `playwright` | `@playwright/mcp@latest` | Project has E2E tests |
| `chrome-devtools` | `chrome-devtools-mcp@latest` | Project has a web frontend |
| `context7` | `@upstash/context7-mcp@latest` | Always (library doc lookup) |
| `markitdown` | `markitdown-mcp-npx` | Need to convert docs/PDFs |
| `filesystem` | `@modelcontextprotocol/server-filesystem` | Always (file access) |

Use placeholder credentials (`[DB_USER]`, `[DB_PASS]`, `[DB_NAME]`) — I'll fill them in.

### 6. `.vscode/mcp.json` — VS Code MCP Configuration

Same servers as `.github/copilot/mcp.json` but in VS Code's format:

```json
{
  "servers": {
    "[server-name]": {
      "type": "stdio",
      "command": "npx",
      "args": ["-y", "[package]", "[connection-string]"]
    }
  }
}
```

---

## Quality Criteria

Before outputting, verify:

1. **No hardcoded project-specific values** — all connection strings use placeholders
2. **Consistent cross-references** — skills mentioned in `copilot-instructions.md` all exist as directories
3. **Agents reference correct file paths** — adapted to the actual project structure
4. **Instructions use correct `applyTo` globs** — matching the real file extensions and paths
5. **Every skill has both `SKILL.md` and `references/` directory** with at least one reference file
6. **Bug-tracker and doc-sync skills are always included** — they are universal guardrails
7. **MCP servers match the declared tech stack** — don't add MongoDB server if no MongoDB
8. **All trigger phrases in skill descriptions are realistic** — things a developer would actually say

---

## Output Format

Output each file with its full path as a header, then the complete file content. Group by directory:

1. `.github/copilot-instructions.md`
2. `.github/copilot/mcp.json`
3. `.github/instructions/*.instructions.md`
4. `.github/agents/*.agent.md`
5. `.github/skills/*/SKILL.md` + `references/*.md`
6. `.vscode/mcp.json`
```

---

## How to Use This Prompt

1. **Copy everything** inside the code fence above
2. **Fill in the `[PLACEHOLDERS]`** in the "My Project Details" section
3. **Paste into Copilot** (CLI, VS Code chat, or any LLM with file creation ability)
4. **Review the output** — adjust skill list, agent workflows, and MCP servers to your needs
5. **Create the files** in your repo and commit them

## Customization Tips

- **Remove skills you don't need** — if no Redis, drop `redis-caching`; if no frontend, drop `design-system`, `api-client`, `frontend-builder`
- **Add project-specific skills** — e.g., `stripe-billing`, `email-templates`, `i18n`
- **Adjust agent workflows** — the 4 default agents cover most CRUD apps; add agents for your domain-specific workflows (e.g., `deployment-agent`, `migration-rollback-agent`)
- **Update MCP credentials** — replace placeholders with real connection strings from your `.env`
- **Keep doc-sync and bug-tracker** — these two are universal and provide long-term memory across sessions

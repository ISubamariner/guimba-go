# Copilot Configuration Reference

This is a **readable mirror** of the live Copilot configuration in `.github/`. The source of truth is always `.github/` — update there first, then sync here.

## 4 Layers of Copilot Customization

### Layer 1: Global Instructions (always loaded)
- **Source**: `.github/copilot-instructions.md`
- **Mirror**: See source directly
- **Contents**: Tech stack, coding standards, anti-redundancy guardrails, bug logging convention, doc-sync rule, MCP server usage rules

### MCP Server Configuration
- **Reference**: `documentation/copilot-config/mcp-servers.md`
- **Config files**: `~/.copilot/mcp-config.json` (CLI), `.vscode/mcp.json` (VS Code), `.github/copilot/mcp.json` (remote agent)
- **Servers**: postgres, mongodb, redis, memory, filesystem, playwright, chrome-devtools, context7, markitdown
- **Key rule**: Query live databases before writing persistence code; use context7 for up-to-date library docs

### Layer 2: Path-Specific Instructions (loaded when matching files are open)

| File | Applies To | Key Rules |
|:---|:---|:---|
| `go-backend.instructions.md` | `backend/**/*.go` | Clean Architecture layers, dependency rules, handler/service/repo patterns |
| `nextjs-frontend.instructions.md` | `frontend/**/*.{ts,tsx}` | App Router, Server Components, design system rules, token flow |
| `database.instructions.md` | `**/*.sql, backend/migrations/**` | Migration format, SQL standards, naming, query safety |

### Layer 3: Custom Agents (invoked on demand)

| Agent | Trigger Phrases | What It Does |
|:---|:---|:---|
| `api-builder` | "create endpoint", "scaffold handler", "new CRUD" | Generates full Clean Architecture stack: entity → interface → use case → persistence → DTO → handler → routes → tests |
| `frontend-builder` | "create page", "add component", "build UI" | Creates Next.js pages, components, hooks using design system |
| `db-migrator` | "create migration", "add column", "alter table" | Generates migration pairs (up/down SQL) + updates Go models |

### Layer 4: Skills (auto-detected based on context)

| Skill | Trigger Phrases | Category |
|:---|:---|:---|
| `docker-compose-services` | "start services", "docker compose", "reset database" | Workflow |
| `swagger-gen` | "generate swagger", "API docs", "OpenAPI" | Documentation |
| `go-testing` | "write tests", "test coverage", "create mock" | Quality |
| `bug-tracker` | "log bug", "debug", "troubleshoot", "why is this failing" | Long-term memory |
| `doc-sync` | "update docs", "sync documentation", "refresh instructions" | Long-term memory |
| `playwright-testing` | "write e2e test", "playwright", "visual regression" | Quality |
| `design-system` | "add component style", "update tokens", "fix styling" | UI consistency |

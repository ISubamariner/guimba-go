# MCP Server Configuration Reference

This documents all Model Context Protocol (MCP) servers configured for the Guimba-GO project. MCP servers extend Copilot with live tool access — direct database queries, browser automation, file operations, and more.

## Config File Locations

| File | Used By | Key Format |
|:---|:---|:---|
| `~/.copilot/mcp-config.json` | **Copilot CLI** (terminal `copilot` command) | `mcpServers` |
| `.vscode/mcp.json` | **VS Code Chat** (Copilot Chat in editor) | `servers` |
| `.github/copilot/mcp.json` | **Remote Copilot coding agent** (GitHub.com) | `mcpServers` |

> **Important**: When adding or removing a server, update all three files to keep them in sync.

## Server Inventory

### postgres — PostgreSQL Direct Access
- **Package**: `@modelcontextprotocol/server-postgres`
- **Connection**: `postgres://guimba:guimba_secret@localhost:5432/guimba_db`
- **Capabilities**: Execute SQL queries, list tables, describe schemas
- **Requires**: Docker Compose running (`docker compose up -d postgres`)
- **Best for**:
  - Verifying table schemas before writing Go repository code
  - Checking migration results
  - Testing complex queries before embedding in Go
  - Inspecting seed data
- **Example prompts**:
  - "Query the social_programs table schema"
  - "Show me all foreign keys on the beneficiaries table"
  - "Run `SELECT * FROM users WHERE role = 'admin'`"

### mongodb — MongoDB Read-Only Access
- **Package**: `mongodb-mcp-server`
- **Connection**: `mongodb://guimba:guimba_secret@localhost:27017/guimba_db?authSource=admin`
- **Mode**: Read-only (`MDB_MCP_READ_ONLY=true`)
- **Capabilities**: Query collections, list databases, describe schemas
- **Requires**: Docker Compose running (`docker compose up -d mongodb`)
- **Best for**:
  - Inspecting audit log entries
  - Checking document schemas in CQRS read models
  - Verifying MongoDB indexes
- **Example prompts**:
  - "List all collections in guimba_db"
  - "Query the audit_logs collection for recent entries"

### redis — Redis Key/Value Access
- **Package**: `@modelcontextprotocol/server-redis`
- **Connection**: `redis://:guimba_secret@localhost:6380`
- **Capabilities**: GET/SET keys, list keys, check TTLs
- **Requires**: Docker Compose running (`docker compose up -d redis`)
- **Port note**: External port is 6380 (mapped to container's 6379)
- **Best for**:
  - Checking cached data (cache-aside pattern)
  - Inspecting token blocklist entries (`token_blocklist:{jti}`)
  - Verifying TTL values
  - Debugging cache invalidation issues
- **Example prompts**:
  - "List all Redis keys matching `token_blocklist:*`"
  - "Check the TTL on key `program:123`"

### memory — Persistent Key-Value Memory
- **Package**: `@modelcontextprotocol/server-memory`
- **Connection**: Local (no external service)
- **Capabilities**: Store/retrieve entities and relations as a knowledge graph
- **Best for**:
  - Tracking multi-step task progress across conversation turns
  - Remembering design decisions made during a session
  - Storing temporary lookup data

### filesystem — Project File Operations
- **Package**: `@modelcontextprotocol/server-filesystem`
- **Scope**: `c:/Users/Ian/Documents/dcode/Guimba-GO` (project root only)
- **Capabilities**: Read/write files, list directories, search files, get file metadata
- **Best for**:
  - Bulk file operations
  - Directory traversal and file discovery
  - Reading files outside the normal code context

### playwright — Browser Automation & E2E Testing
- **Package**: `@playwright/mcp`
- **Connection**: Local browser
- **Capabilities**: Navigate pages, click elements, fill forms, take screenshots, run accessibility audits
- **Best for**:
  - Running E2E test scenarios against the running frontend
  - Capturing screenshots for visual regression
  - Validating UI flows end-to-end
  - Debugging frontend rendering issues
- **Requires**: Frontend running (`npm run dev` in `frontend/`)

### chrome-devtools — Chrome DevTools Protocol
- **Package**: `chrome-devtools-mcp`
- **Connection**: Chrome instance with remote debugging
- **Capabilities**: Network inspection, console logs, DOM queries, performance profiling, CSS inspection
- **Best for**:
  - Debugging API request/response cycles from the browser
  - Inspecting network waterfall and timing
  - Auditing CSS applied to elements
  - Performance profiling (Core Web Vitals)
- **Requires**: Chrome launched with `--remote-debugging-port=9222`

### context7 — Library Documentation Lookup
- **Package**: `@upstash/context7-mcp`
- **Connection**: Cloud API (requires internet)
- **Capabilities**: Fetch current documentation for any library/package
- **Best for**:
  - Looking up current API signatures for Go/npm packages
  - Checking breaking changes in library updates
  - Getting accurate usage examples instead of guessing from training data
- **Rule**: Always use this instead of guessing library APIs
- **Example prompts**:
  - "Look up the pgx v5 QueryRow API"
  - "Get the current chi router middleware docs"
  - "Check the go-playground/validator struct tag options"

### markitdown — File-to-Markdown Converter
- **Package**: `markitdown-mcp-npx`
- **Connection**: Local
- **Capabilities**: Convert PDF, DOCX, XLSX, PPTX, images to Markdown
- **Best for**:
  - Extracting text from uploaded PDF requirements documents
  - Converting spreadsheet data for analysis
  - Making non-text files searchable and analyzable

## Usage Patterns

### Database-First Development
1. Use `postgres` MCP to check current schema
2. Write migration SQL
3. Use `postgres` MCP to verify migration applied correctly
4. Write Go repository code that matches verified schema

### Debug Workflow
1. Use `chrome-devtools` to inspect failing network request
2. Use `postgres`/`mongodb` to verify data state
3. Use `redis` to check cache state
4. Fix the code with full context

### Library Integration
1. Use `context7` to get current API docs for the library
2. Write integration code with accurate signatures
3. Use `playwright` to verify end-to-end behavior

## Troubleshooting

| Problem | Solution |
|:---|:---|
| MCP server not loading | Restart Copilot CLI (`exit` → `copilot`) |
| Database connection refused | Run `docker compose up -d` |
| Redis connection refused | Check port is 6380 (not 6379) |
| MongoDB write denied | Expected — configured as read-only |
| Chrome DevTools not connecting | Launch Chrome with `--remote-debugging-port=9222` |
| context7 timeout | Check internet connection |

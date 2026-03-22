---
name: bug-tracker
description: "Tracks and retrieves resolved bugs for pattern recognition and prevention. Use when user says 'log bug', 'track issue', 'what bugs have we seen', 'check bug history', 'debug', 'troubleshoot', 'why is this failing', or when investigating any error."
---

# Bug Tracker Skill

Persistent memory of resolved bugs. Prevents re-investigation of known issues and identifies recurring patterns.

## When to Use This Skill

### Before Debugging (MANDATORY)
Before investigating any bug or error:
1. Search `references/bug-log.md` for **keywords related to the current error**
2. Check if the same root cause has appeared before
3. If a match exists, apply the documented resolution first
4. If no match, proceed with normal debugging

### After Resolving a Bug
After any bug is fixed, add an entry to `references/bug-log.md` using the format below.

## Bug Log Entry Format

```markdown
### [SHORT_TITLE] — YYYY-MM-DD
- **Issue**: What went wrong (symptoms, error messages)
- **Root Cause**: Why it happened (the actual underlying problem)
- **Resolution**: What was changed to fix it
- **Files Changed**: List of modified files
- **Prevention**: What rule or check would prevent recurrence
```

## Example Entry

```markdown
### PostgreSQL Connection Timeout — 2026-03-19
- **Issue**: `dial tcp: connect: connection refused` when starting the Go server
- **Root Cause**: Docker Compose `depends_on` didn't wait for PostgreSQL to be ready, only for the container to start
- **Resolution**: Added `healthcheck` to PostgreSQL service and `condition: service_healthy` to the backend service
- **Files Changed**: `docker-compose.yml`
- **Prevention**: Always use health checks with `depends_on` conditions, never bare `depends_on`
```

## Pattern Analysis
When the bug log grows, look for:
- **Same root cause appearing multiple times** → a systemic fix is needed (update a skill or instruction)
- **Same file appearing in multiple bug entries** → the file may need refactoring
- **Same error type recurring** → add a guardrail to the relevant instruction file

## Escalation
If a root cause appears 3+ times in the bug log:
1. Identify which skill or instruction should prevent it
2. Update that skill/instruction with a specific rule
3. Note the update in the bug log entry under **Prevention**

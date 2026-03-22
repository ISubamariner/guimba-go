# Prompt Architecture Design Decisions

## Context
We started with the "Connected Trio" prompt (see `connected-trio-original.md`) and needed to decide how to implement its ideas within the Guimba-GO project.

## Decision: Use Copilot-Native Architecture, Not a Custom Prompt System

### What We Kept (Good Ideas)
| Trio Concept | Where It Lives Now |
|:---|:---|
| **Learner/Debugger** (bug logging) | `bug-tracker` skill — `.github/skills/bug-tracker/` |
| **Skill Monitor** (doc/skill auditing) | `doc-sync` skill — `.github/skills/doc-sync/` |
| Anti-redundancy checks | Global instructions — `.github/copilot-instructions.md` |
| "Search before debugging" | Enforced in `go-testing` skill + `bug-tracker` skill |
| Structured bug format (Issue→Cause→Fix) | Convention in global instructions + `bug-tracker` template |
| Iterative refinement (diff-only updates) | Convention in global instructions |

### What We Changed
| Trio Concept | Why We Changed It |
|:---|:---|
| `.skills/` custom directory | Conflicts with Copilot's native `.github/skills/` — using native system instead |
| `skill-registry.md` | Copilot handles discovery natively via `description` field matching — no manual registry needed |
| Version bumping after every bug | Excessive overhead; we log bugs but version bumps are for meaningful capability changes |
| "Read last 5 entries in bug-log" | Arbitrary window; we search by **related keywords** instead |
| YAML `version`, `last-updated`, `related-skills` | Replaced with Copilot-native `name`, `description`, optional `metadata` |
| Three "concurrent" agents | LLMs don't run concurrent processes; we use behavioral modes within skills instead |

### Why Not a Parallel System?
Running both `.skills/` (Trio) and `.github/skills/` (Copilot) would create:
- Two competing skill discovery systems
- Duplicated configuration
- Confusion about which is the source of truth
- Wasted context window on meta-management

### Result: 7 Skills in Copilot-Native Format
1. `docker-compose-services` — local dev environment
2. `swagger-gen` — API documentation generation
3. `go-testing` — Go test patterns with "check bug-log" rule
4. `bug-tracker` — persistent bug memory (Trio's Learner/Debugger)
5. `doc-sync` — documentation sync & audit (Trio's Skill Monitor)
6. `playwright-testing` — browser E2E + visual regression
7. `design-system` — consolidated CSS & UI consistency

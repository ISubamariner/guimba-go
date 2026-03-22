# Guimba-GO Documentation Hub

Central reference for all project documentation, architecture decisions, system prompts, and Copilot configuration.

> **Note**: The `.github/` directory contains the *live* Copilot configuration files that are actively loaded. This `documentation/` folder is a **readable mirror** and extended reference. When updating Copilot config, update the source in `.github/` first, then sync here via the `doc-sync` skill.

## Directory Map

```
documentation/
├── README.md                       ← You are here
│
├── project/                        ← Project-level docs
│   ├── setup-guide.md              ← How to set up the dev environment
│   ├── tech-stack.md               ← Full tech stack with rationale
│   └── conventions.md              ← Naming, commit, branch conventions
│
├── architecture/                   ← Architecture & design decisions
│   ├── clean-architecture.md       ← Clean Architecture layers explained
│   ├── testing-strategy.md         ← Unit/integration/e2e/Playwright strategy
│   └── design-system.md            ← Consolidated CSS & token architecture
│
├── api/                            ← API documentation
│   └── README.md                   ← API overview (Swagger lives in backend/docs/)
│
├── prompts/                        ← System prompts & AI instructions
│   ├── business-logic-reference.md ← Complete business logic extracted from guimba-debt-tracker
│   ├── connected-trio-original.md  ← The original "Connected Trio" prompt (reference)
│   ├── design-decisions.md         ← Why we chose this prompt architecture
│   └── rebrand-and-extend.prompt.md ← Rebrand & extension prompt for Go rewrite
│
└── copilot-config/                 ← Mirror of .github/ Copilot configuration
    ├── README.md                   ← Index of all Copilot layers
    ├── instructions/               ← Mirrors of .github/instructions/
    ├── agents/                     ← Mirrors of .github/agents/
    └── skills/                     ← Mirrors of .github/skills/
```

## Quick Links

| What | Where |
|:---|:---|
| Master plan | [`/MASTERPLAN.md`](../MASTERPLAN.md) |
| **Business logic reference** | [`documentation/prompts/business-logic-reference.md`](prompts/business-logic-reference.md) |
| Global Copilot instructions | [`.github/copilot-instructions.md`](../.github/copilot-instructions.md) |
| Project conventions | [`/AGENTS.md`](../AGENTS.md) |
| Bug log | [`.github/skills/bug-tracker/references/bug-log.md`](../.github/skills/bug-tracker/references/bug-log.md) |
| Doc change log | [`.github/skills/doc-sync/references/changelog.md`](../.github/skills/doc-sync/references/changelog.md) |
| Design tokens | [`.github/skills/design-system/references/token-registry.md`](../.github/skills/design-system/references/token-registry.md) |

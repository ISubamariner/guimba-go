# Design System Architecture

## Core Principle
> Every visual decision lives in `src/styles/` and `tailwind.config.ts`. Components consume tokens — they never invent their own styles.

## Token Flow

```
tokens.css (CSS custom properties)
    ↓
tailwind.config.ts (maps tokens into Tailwind theme)
    ↓
components.css (@apply-based component classes)
    ↓
src/components/ui/*.tsx (React components with variant props)
    ↓
Feature components & pages (import from ui/)
```

## File Responsibilities

| File | Role |
|:---|:---|
| `src/styles/tokens.css` | CSS custom properties: colors, spacing, radii, shadows, z-index |
| `src/styles/typography.css` | Font families, size scale, weights, line-heights |
| `src/styles/layouts.css` | Reusable layout patterns (page shell, sidebar, grid) |
| `src/styles/components.css` | `@apply`-based classes: `.btn`, `.card`, `.input`, `.badge`, `.table` |
| `tailwind.config.ts` | Extends Tailwind with design tokens from `tokens.css` |
| `src/components/ui/*.tsx` | React wrappers with typed variant props |
| `src/components/ui/index.ts` | Barrel export for all primitives |
| `src/lib/cn.ts` | `cn()` = clsx + tailwind-merge |

## Dark Mode
- Implemented via CSS custom property overrides in `.dark` class
- Tokens in `tokens.css` define both `:root` (light) and `.dark` values
- Toggle with `darkMode: 'class'` in `tailwind.config.ts`

## UI Primitives

| Component | Variants | Sizes |
|:---|:---|:---|
| Button | primary, secondary, outline, ghost, destructive | sm, md, lg |
| Input | default, error | sm, md, lg |
| Card | default, outlined | — |
| Modal | default, wide | — |
| Table | — | — |
| Badge | default, success, warning, error, info | — |
| Alert | info, success, warning, error | — |

## Rules
1. Never hardcode colors, spacing, or font sizes in components
2. Always use semantic tokens (`bg-primary`, not `bg-blue-500`)
3. Build pages from UI primitives — never raw `<button>`, `<input>`, `<table>`
4. New visual patterns go to `components.css` first, then React wrapper
5. One-off positioning uses Tailwind utilities; visual identity uses component classes

## Full token registry
See [`.github/skills/design-system/references/token-registry.md`](../../.github/skills/design-system/references/token-registry.md)

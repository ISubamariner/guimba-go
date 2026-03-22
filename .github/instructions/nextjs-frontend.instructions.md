---
applyTo: "frontend/**/*.{ts,tsx}"
---

# Next.js Frontend Instructions

## Framework
- Next.js 15+ with App Router
- TypeScript in strict mode
- Tailwind CSS for styling

## File Organization
- Pages: `src/app/<route>/page.tsx`
- Layouts: `src/app/<route>/layout.tsx`
- Components: `src/components/<ComponentName>.tsx`
- API client: `src/lib/api.ts`
- Custom hooks: `src/hooks/use<Name>.ts`
- Types: `src/types/<domain>.ts`

## Component Rules
- Use Server Components by default; add `'use client'` only when needed (state, effects, browser APIs)
- Props interfaces defined above the component in the same file
- Prefer composition over prop drilling — use React Context for deep state

## API Communication
- All backend calls go through `src/lib/api.ts`
- Use `fetch` with proper error handling
- Type all API responses with interfaces from `src/types/`

## MCP-Assisted Development
- Use `playwright` MCP for browser automation, E2E testing, and screenshot capture
- Use `chrome-devtools` MCP to inspect network requests, debug CSS, and audit performance
- Use `context7` MCP to look up current Next.js/React/Tailwind APIs instead of guessing

## Styling — Consolidated Design System

### Architecture: Single Source of Truth
All visual decisions live in `src/styles/` and `tailwind.config.ts`. Components consume tokens — they never hardcode colors, spacing, or typography.

```
src/styles/tokens.css      → CSS custom properties (--color-primary, --spacing-4, --radius-md)
src/styles/typography.css   → Font families, scale, weights
src/styles/layouts.css      → Reusable layout patterns (@apply-based)
src/styles/components.css   → Component classes (.btn, .card, .input, .badge)
tailwind.config.ts          → Extends Tailwind theme with CSS custom properties
```

### Rules
- **Never hardcode raw values** in components (no `bg-blue-500`, `text-[14px]`, `p-[12px]`)
- **Always use semantic tokens**: `bg-primary`, `text-heading`, `p-4` (mapped to design tokens)
- **Use `cn()` utility** for conditional classes: `cn('btn', variant === 'primary' && 'btn-primary')`
- **Build pages from UI primitives** in `src/components/ui/` — never write raw HTML for buttons, inputs, cards, tables
- **New visual patterns** → add to `src/styles/components.css` first, then consume in components
- **One-off overrides** → use Tailwind utilities on top of component classes, never create new CSS files per component
- Tailwind utility classes are fine for layout/positioning; semantic component classes for visual identity

### Component Hierarchy
```
Page (src/app/**/page.tsx)
  └── Layout Components (sidebar, page-shell — from layouts.css)
       └── Feature Components (ProgramList, UserTable — in src/components/)
            └── UI Primitives (Button, Card, Input — in src/components/ui/)
                 └── Design Tokens (tokens.css → tailwind.config.ts)
```

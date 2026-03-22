---
name: design-system
description: "Manages the consolidated CSS design system, design tokens, UI primitives, and visual consistency. Use when user says 'add component style', 'create UI component', 'update design tokens', 'change theme', 'fix styling', 'make consistent', 'add color', 'update button', or when working with frontend/src/styles/, frontend/src/components/ui/, or frontend/tailwind.config.ts."
---

# Design System Skill — UI Consistency Guardian

Ensures all frontend styling flows through a single consolidated design system. Prevents CSS sprawl, enforces token usage, and maintains visual consistency.

## Core Principle
> Every visual decision lives in `src/styles/` and `tailwind.config.ts`. Components consume tokens — they never invent their own colors, spacing, or typography.

## File Ownership

| File | What It Controls | When to Modify |
|:---|:---|:---|
| `src/styles/tokens.css` | Colors, spacing scale, border radii, shadows, z-index, transitions | New brand colors, spacing needs, shadow variants |
| `src/styles/typography.css` | Font families, size scale, weights, line-heights, letter-spacing | New font, new text size needed |
| `src/styles/layouts.css` | Page shells, sidebar, grid systems, container widths | New layout pattern used in 2+ pages |
| `src/styles/components.css` | Component classes (`.btn`, `.card`, `.input`, `.badge`, `.table`) | New UI primitive, new variant of existing primitive |
| `tailwind.config.ts` | Maps CSS custom properties into Tailwind theme | Whenever `tokens.css` changes |
| `src/components/ui/*.tsx` | React wrappers around component classes with props/variants | New primitive component, new variant |
| `src/components/ui/index.ts` | Barrel export for all UI primitives | Every time a new primitive is added |
| `src/lib/cn.ts` | `cn()` utility (clsx + tailwind-merge) | Never (set once) |

## The Token Flow

```
tokens.css (CSS custom properties)
    ↓
tailwind.config.ts (extends Tailwind theme)
    ↓
components.css (@apply-based component classes)
    ↓
src/components/ui/*.tsx (React components with variant props)
    ↓
Page components (import from src/components/ui/)
```

## Enforcement Rules

### Before Adding Any Style
1. **Check tokens.css** — Does the color/spacing/shadow already exist as a token?
2. **Check components.css** — Does a component class already cover this pattern?
3. **Check src/components/ui/** — Does a primitive component already exist?

If yes → **use it**. If no → **add it to the right layer** (token → class → component), don't inline it.

### Violations to Flag
| Violation | Example | Fix |
|:---|:---|:---|
| Hardcoded color | `bg-blue-500`, `text-[#333]` | Use `bg-primary`, `text-foreground` (semantic token) |
| Hardcoded spacing | `p-[12px]`, `mt-[30px]` | Use `p-3`, `mt-8` (from spacing scale) |
| Hardcoded font size | `text-[14px]` | Use `text-sm` (from typography scale) |
| Raw HTML button | `<button className="bg-blue-500 ...">` | Use `<Button variant="primary">` from `ui/` |
| Raw HTML input | `<input className="border ...">` | Use `<Input>` from `ui/` |
| One-off CSS file | `ProgramList.module.css` | Move pattern to `components.css` or use Tailwind utilities |
| Duplicated visual pattern | Same card styling in 3 different components | Extract to `<Card>` primitive |

### When Creating a New UI Component
1. Define the visual pattern in `src/styles/components.css` first using `@apply`
2. Create the React component in `src/components/ui/<name>.tsx`
3. Support variants via props (e.g., `variant="primary"`, `size="md"`)
4. Use `cn()` to compose base class + variant class + custom className
5. Export from `src/components/ui/index.ts`
6. Add to the token registry in `references/token-registry.md`

### Pattern: UI Primitive Component
```tsx
// src/components/ui/button.tsx
import { cn } from '@/lib/cn';

const variants = {
  primary: 'btn-primary',
  secondary: 'btn-secondary',
  outline: 'btn-outline',
  ghost: 'btn-ghost',
  destructive: 'btn-destructive',
} as const;

const sizes = {
  sm: 'btn-sm',
  md: 'btn-md',
  lg: 'btn-lg',
} as const;

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: keyof typeof variants;
  size?: keyof typeof sizes;
}

export function Button({
  variant = 'primary',
  size = 'md',
  className,
  ...props
}: ButtonProps) {
  return (
    <button
      className={cn('btn', variants[variant], sizes[size], className)}
      {...props}
    />
  );
}
```

### Pattern: tokens.css → tailwind.config.ts Connection
```css
/* src/styles/tokens.css */
:root {
  /* Colors */
  --color-primary: 220 90% 56%;
  --color-primary-hover: 220 90% 48%;
  --color-secondary: 210 40% 96%;
  --color-destructive: 0 84% 60%;
  --color-foreground: 222 47% 11%;
  --color-muted: 210 40% 96%;
  --color-border: 214 32% 91%;
  --color-background: 0 0% 100%;

  /* Spacing (base: 4px) — use Tailwind's default scale */

  /* Radii */
  --radius-sm: 0.25rem;
  --radius-md: 0.375rem;
  --radius-lg: 0.5rem;
  --radius-full: 9999px;

  /* Shadows */
  --shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.05);
  --shadow-md: 0 4px 6px -1px rgb(0 0 0 / 0.1);
  --shadow-lg: 0 10px 15px -3px rgb(0 0 0 / 0.1);
}

/* Dark mode override */
.dark {
  --color-primary: 220 90% 66%;
  --color-foreground: 210 40% 98%;
  --color-background: 222 47% 11%;
  --color-border: 217 33% 17%;
}
```

```typescript
// tailwind.config.ts
import type { Config } from 'tailwindcss';

const config: Config = {
  darkMode: 'class',
  content: ['./src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        primary: 'hsl(var(--color-primary) / <alpha-value>)',
        'primary-hover': 'hsl(var(--color-primary-hover) / <alpha-value>)',
        secondary: 'hsl(var(--color-secondary) / <alpha-value>)',
        destructive: 'hsl(var(--color-destructive) / <alpha-value>)',
        foreground: 'hsl(var(--color-foreground) / <alpha-value>)',
        muted: 'hsl(var(--color-muted) / <alpha-value>)',
        border: 'hsl(var(--color-border) / <alpha-value>)',
        background: 'hsl(var(--color-background) / <alpha-value>)',
      },
      borderRadius: {
        sm: 'var(--radius-sm)',
        md: 'var(--radius-md)',
        lg: 'var(--radius-lg)',
      },
      boxShadow: {
        sm: 'var(--shadow-sm)',
        md: 'var(--shadow-md)',
        lg: 'var(--shadow-lg)',
      },
    },
  },
};

export default config;
```

## Auditing: Manual "Check Consistency" Flow

When invoked with "check consistency" or "audit styles":

1. **Scan all `.tsx` files** in `src/components/` and `src/app/`
2. Flag violations:
   - Hardcoded hex/rgb colors → should use semantic token
   - Hardcoded pixel values → should use spacing scale
   - Raw `<button>`, `<input>`, `<table>` → should use UI primitives
   - Duplicate styling patterns → should extract to component
3. Report findings
4. Fix violations by migrating to design system tokens/components

## Troubleshooting

### Tailwind Classes Not Working After Token Change
**Cause**: Tailwind config not synced with `tokens.css`
**Fix**: Update `tailwind.config.ts` `extend` section to match new CSS custom properties

### Dark Mode Looks Wrong
**Cause**: Component hardcodes light-mode colors instead of using tokens
**Fix**: Replace hardcoded colors with semantic tokens that have `.dark` overrides in `tokens.css`

### Component Looks Different Across Pages
**Cause**: Overriding component styles inline instead of using variant props
**Fix**: Add a new variant to the component if a legitimate visual difference is needed

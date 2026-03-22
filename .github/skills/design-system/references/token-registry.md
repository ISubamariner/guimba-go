# Design Token & Component Registry

Track all design tokens and UI primitives here. Keep in sync with actual CSS files.

---

## Color Tokens (`tokens.css`)

| Token | Light | Dark | Usage |
|:---|:---|:---|:---|
| `--color-primary` | `220 90% 56%` | `220 90% 66%` | Primary actions, links, active states |
| `--color-primary-hover` | `220 90% 48%` | — | Primary button/link hover |
| `--color-secondary` | `210 40% 96%` | — | Secondary backgrounds, subtle fills |
| `--color-destructive` | `0 84% 60%` | — | Delete, errors, danger actions |
| `--color-foreground` | `222 47% 11%` | `210 40% 98%` | Body text, headings |
| `--color-muted` | `210 40% 96%` | — | Disabled text, placeholder |
| `--color-border` | `214 32% 91%` | `217 33% 17%` | Borders, dividers |
| `--color-background` | `0 0% 100%` | `222 47% 11%` | Page/card backgrounds |

## Radius Tokens

| Token | Value | Usage |
|:---|:---|:---|
| `--radius-sm` | `0.25rem` | Small elements (badges, chips) |
| `--radius-md` | `0.375rem` | Inputs, buttons |
| `--radius-lg` | `0.5rem` | Cards, modals |
| `--radius-full` | `9999px` | Avatars, pills |

## Shadow Tokens

| Token | Usage |
|:---|:---|
| `--shadow-sm` | Subtle elevation (cards at rest) |
| `--shadow-md` | Medium elevation (dropdowns, popovers) |
| `--shadow-lg` | High elevation (modals, toasts) |

## Typography Scale (`typography.css`)

| Class | Size | Weight | Usage |
|:---|:---|:---|:---|
| `text-display` | 2.25rem / 36px | 700 | Page titles, hero text |
| `text-heading` | 1.5rem / 24px | 600 | Section headings |
| `text-subheading` | 1.125rem / 18px | 600 | Subsections, card titles |
| `text-body` | 1rem / 16px | 400 | Default body text |
| `text-small` | 0.875rem / 14px | 400 | Labels, helper text |
| `text-caption` | 0.75rem / 12px | 400 | Timestamps, metadata |

## UI Primitive Components (`src/components/ui/`)

| Component | File | Variants | Sizes |
|:---|:---|:---|:---|
| Button | `button.tsx` | primary, secondary, outline, ghost, destructive | sm, md, lg |
| Input | `input.tsx` | default, error | sm, md, lg |
| Textarea | `textarea.tsx` | default, error | — |
| Select | `select.tsx` | default, error | sm, md, lg |
| Card | `card.tsx` | default, outlined | — |
| Modal | `modal.tsx` | default, wide | — |
| Table | `table.tsx` | — | — |
| Badge | `badge.tsx` | default, success, warning, error, info | — |
| Alert | `alert.tsx` | info, success, warning, error | — |
| Toast | `toast.tsx` | info, success, warning, error | — |

## Component Classes (`components.css`)

```css
/* Buttons */
.btn { @apply inline-flex items-center justify-center font-medium rounded-md transition-colors focus-visible:outline-none focus-visible:ring-2; }
.btn-sm { @apply h-8 px-3 text-sm; }
.btn-md { @apply h-10 px-4 text-sm; }
.btn-lg { @apply h-12 px-6 text-base; }
.btn-primary { @apply bg-primary text-white hover:bg-primary-hover; }
.btn-secondary { @apply bg-secondary text-foreground hover:bg-muted; }
.btn-outline { @apply border border-border bg-transparent hover:bg-secondary; }
.btn-ghost { @apply bg-transparent hover:bg-secondary; }
.btn-destructive { @apply bg-destructive text-white hover:opacity-90; }

/* Cards */
.card { @apply rounded-lg border border-border bg-background shadow-sm; }
.card-header { @apply px-6 py-4 border-b border-border; }
.card-body { @apply px-6 py-4; }
.card-footer { @apply px-6 py-4 border-t border-border; }

/* Inputs */
.input { @apply w-full rounded-md border border-border bg-background px-3 py-2 text-sm placeholder:text-muted focus:outline-none focus:ring-2 focus:ring-primary; }
.input-error { @apply border-destructive focus:ring-destructive; }

/* Badges */
.badge { @apply inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium; }
.badge-default { @apply bg-secondary text-foreground; }
.badge-success { @apply bg-green-100 text-green-800; }
.badge-warning { @apply bg-yellow-100 text-yellow-800; }
.badge-error { @apply bg-red-100 text-red-800; }

/* Tables */
.table { @apply w-full text-sm; }
.table-header { @apply bg-secondary text-left font-medium; }
.table-row { @apply border-b border-border hover:bg-muted/50 transition-colors; }
.table-cell { @apply px-4 py-3; }
```

---

<!-- Update this registry whenever tokens, typography, or components change -->

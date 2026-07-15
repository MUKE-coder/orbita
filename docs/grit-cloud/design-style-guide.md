# Grit Cloud / Orbita — Design & Style Guide

The feeling we're after: **techy yet effortless.** It should look like a serious infrastructure tool a developer trusts with production — dense, precise, dark-first, monospace where it counts — but never intimidating. A freelancer deploying their first client app should feel guided, not lost. Think *Vercel's calm × Railway's playfulness × Linear's precision*, self-hosted.

This guide governs the Orbita React SPA and any CLI output styling. It respects the existing stack: **React 18 + Vite + Tailwind CSS v4 + shadcn/ui + Zustand + TanStack Query + xterm.js.** Do not introduce a different UI framework.

---

## 1. Design principles

1. **Clarity over cleverness.** Every screen answers "what is running, is it healthy, what do I do next." Status is never ambiguous.
2. **Dark-first, light-supported.** Infrastructure tools live in the dark. Ship a polished dark theme as default; light theme must work but is secondary.
3. **Dense but breathable.** Show real data (containers, logs, metrics) without clutter. Generous line-height and spacing keep density from becoming noise.
4. **Terminal-honest.** Logs, IDs, commands, env vars, and connection strings render in monospace and are one-click copyable. Never prettify a value the user needs to paste verbatim.
5. **Progressive disclosure.** Beginners see the happy path (Create App → Deploy). Power features (cgroup quotas, node scaling, exec) are one click deeper, not on the front page.
6. **Guided, not gated.** Empty states teach. Errors explain the cause *and* the fix (mirroring the VPS-harden blog's tone).

---

## 2. Color system

Define as CSS variables / Tailwind v4 `@theme` tokens. Dark is the source of truth.

### Brand
- **Orbita primary** — an orbital indigo/violet: `#6D5CE7` (hover `#5B4BD1`, active `#4C3EB8`). Used for primary actions, active nav, focus rings.
- **Accent / signal** — electric cyan `#22D3EE` for links, live indicators, "streaming" states. Use sparingly; it's the "something is happening" color.

### Dark theme (default)
- `--bg`         `#0B0D12`  (app background — near-black, faint blue)
- `--bg-elev-1`  `#12151C`  (cards, panels)
- `--bg-elev-2`  `#1A1E27`  (popovers, inputs, hover rows)
- `--border`     `#252A34`  (hairlines; 1px)
- `--text`       `#E6E8EC`  (primary text)
- `--text-muted` `#9AA0AA`  (secondary/labels)
- `--text-faint` `#5C626D`  (timestamps, placeholders)

### Light theme
- `--bg` `#FAFBFC` · `--bg-elev-1` `#FFFFFF` · `--bg-elev-2` `#F1F3F6` · `--border` `#E2E5EA` · `--text` `#14171C` · `--text-muted` `#5A616B` · `--text-faint` `#9AA0AA`

### Status (semantic — same hues both themes, tuned for contrast)
- **Success / Running / Healthy** — green `#22C55E`
- **Warning / Deploying / Pending** — amber `#F59E0B`
- **Error / Crashed / Failed** — red `#EF4444`
- **Info / Neutral / Stopped** — slate `#64748B`
- **Live / Streaming** — cyan `#22D3EE` (with a subtle pulse animation)

Status is **always** communicated by color **and** a label/icon (never color alone — accessibility + colorblind users).

---

## 3. Typography

- **UI / sans:** Inter (or `Geist`) — the whole interface. Weights: 400 body, 500 labels, 600 headings, 700 rare emphasis.
- **Mono:** `JetBrains Mono` or `Geist Mono` — logs, IDs, env keys, commands, connection strings, code, metrics numbers. Monospace is a signal that "this is a real technical value."
- **Scale (rem):** `12` micro/timestamps · `13` table cells, secondary · `14` body (base) · `16` section titles · `20` page titles · `28` hero/empty-state headline.
- **Line-height:** 1.5 body, 1.3 headings, 1.6 in log/terminal views.
- Tabular numbers (`font-variant-numeric: tabular-nums`) for all metrics, counts, and tables so digits don't jitter.

---

## 4. Spacing, radius, elevation

- **Spacing scale (px):** 2, 4, 8, 12, 16, 24, 32, 48. Default gutter 16; card padding 24; tight lists 8–12.
- **Radius:** `6px` inputs/buttons · `10px` cards/panels · `14px` modals · `999px` pills/badges. Never sharp corners; never overly round — precise, not bubbly.
- **Elevation:** flat by default. Depth via `--bg-elev-*` layering + 1px `--border`, not heavy shadows. Modals/popovers get one soft shadow: `0 8px 24px rgba(0,0,0,.4)` (dark).
- **Borders:** 1px hairline `--border` everywhere; this is the "techy" texture — visible structure, gridlines, panel seams.

---

## 5. Core components (shadcn/ui, restyled to tokens)

- **Buttons:** Primary (Orbita indigo, white text). Secondary (transparent, `--border`, `--text`). Ghost (text only, hover `--bg-elev-2`). Destructive (red). Always a loading spinner state; deploy/destroy buttons show progress inline.
- **Status badge:** pill with a 6px dot + label (`● Running`, `● Deploying`, `● Crashed`). Color from the status system.
- **Cards / resource tiles:** elevated panel, 1px border, title row + status badge + key metric. App/DB/cron tiles use this shape consistently.
- **Tables:** dense, hairline row separators, hover highlights the row (`--bg-elev-2`), monospace for IDs/values, tabular-nums for numbers. Sticky header. Right-aligned actions column.
- **Tabs:** underline-style (active tab = indigo underline). Used inside a resource for Overview / Env / Domains / Logs / Metrics / Terminal.
- **Inputs / forms:** `--bg-elev-2` fill, 1px border, indigo focus ring. React Hook Form + Zod validation; inline error text in red beneath the field.
- **Copyable value:** any ID/URL/connection string renders mono in a subtle pill with a copy icon that confirms "Copied". This is everywhere.
- **Toasts:** bottom-right, status-colored left border, auto-dismiss except errors (which persist until dismissed and include the fix).

---

## 6. Signature surfaces

### Dashboard (org overview)
Top: a row of stat tiles (running apps, databases, cron jobs, CPU/RAM used vs quota) with tabular-nums and tiny sparklines. Below: a resource list grouped by type, each row a status badge + name + primary metric + quick actions. The quota bar (cgroup usage) is a slim horizontal meter that turns amber near limit. **First thing the eye lands on: overall health.**

### Deploy view (the money screen)
A vertical **step timeline** of the deploy: `Pull/Build → Migrate → Health-check → Cutover → Live`. Each step shows a live spinner → check/cross, with elapsed time. Build and migration logs stream below in the terminal component. On success, a green banner with the **live URL** + **Pulse** + **Sentinel** dashboard links, all copyable. This screen must feel *alive* — it's the moment the product delivers.

### Logs / Terminal
Full-height xterm.js on `--bg` (true near-black), mono, with a "● Live" cyan pulsing indicator when streaming, a pause/resume toggle, search, and wrap toggle. Never truncate; let it scroll. The in-browser container terminal uses the same surface so logs and shell feel continuous.

### Empty states
Every empty resource list is a teaching moment: a short headline, one sentence explaining the concept, and a single primary CTA. E.g. Apps empty: *"No apps yet. Deploy your first Grit app in one command, or create one here."* with a copyable `grit deploy --host prod`.

---

## 7. Motion

- Fast and functional: 120–160ms ease-out for hovers, tab switches, popovers. Nothing decorative over 250ms.
- **Live pulse** for streaming/deploying indicators (gentle opacity 0.5↔1, ~1.5s).
- Deploy step transitions animate the spinner→check swap.
- Respect `prefers-reduced-motion`: drop pulses and transitions to instant.

---

## 8. Iconography & imagery

- **Lucide** icons (line style) throughout — matches the shadcn ecosystem, stays crisp and technical.
- Icons always paired with text labels in nav and actions (icon-only allowed only in dense action columns with tooltips).
- Keep the Orbita logo/orbital motif; use the orbit as a subtle loading/empty-state graphic, not decoration everywhere.

---

## 9. CLI aesthetic (grit cloud / grit deploy)

The terminal output is part of the brand. Keep it consistent with the dashboard:
- Indigo for the `grit` prompt/headers, cyan for live/streaming lines, green/amber/red for status matching the web status colors.
- Step output mirrors the deploy timeline: `✔ Build`, `⠹ Migrate`, `✔ Health-check`, `✔ Live → https://...`.
- Print copyable URLs and dashboard links at the end, plainly.
- Errors: one red line for the cause, one dim line for the fix, matching the web toast pattern and the VPS-harden blog voice.

---

## 10. Accessibility

- Contrast ≥ 4.5:1 for text (verify status colors on both themes).
- Never status-by-color-alone — always a label or icon.
- Full keyboard nav; visible indigo focus rings; ESC closes modals/popovers.
- `prefers-reduced-motion` honored.
- xterm and log views screen-reader-labeled; copy actions announce success.

---

## 11. Do / Don't

**Do:** dark-first, hairline borders, mono for real values, status by color+label, teach in empty states, make every ID copyable, keep the deploy screen alive.

**Don't:** heavy drop-shadows, rounded-bubbly cards, color-only status, hiding errors, prettifying pasteable values, decorative animation, introducing a non-shadcn component kit, or swapping the established Tailwind v4 + shadcn foundation.

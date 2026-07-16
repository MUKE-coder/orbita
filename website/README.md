# Orbita website

Landing page + documentation for [Orbita](https://github.com/MUKE-coder/orbita), built with
Next.js 14 (App Router), MDX, Shiki, Tailwind CSS, and framer-motion.

## Develop

```bash
cd website
npm install
npm run dev        # http://localhost:3000
```

## Build

```bash
npm run build      # standalone output in .next/standalone
npm start
```

## Structure

```
app/
  page.tsx              landing page (animated hero + features + CTA)
  docs/
    layout.tsx          docs shell (sidebar + prev/next)
    <slug>/page.mdx     one MDX file per docs page
components/             Navbar, Footer, Sidebar, Callout, Tabs, CodeCard, …
lib/site.ts             nav + sidebar config (edit here to change navigation)
mdx-components.tsx      maps MDX elements to styled React components
```

To add a docs page: create `app/docs/<slug>/page.mdx` and add it to `sidebar` in `lib/site.ts`.

## Deploy (Dokploy / Docker)

The included `Dockerfile` produces a standalone Next.js server image. Point Dokploy (or Orbita
itself) at this repo with build context `website/` and it runs on port 3000.

## Attribution

Design and layout adapted from [animateicons](https://github.com/avijit07x/animateicons) by
Avijit Dey (MIT). See [`NOTICE`](./NOTICE).

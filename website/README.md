# Orbita website + docs

The Orbita / Grit Cloud landing page and documentation, built with
[VitePress](https://vitepress.dev). Dark-first, brand-themed (indigo/cyan, mono for real values)
per the Grit Cloud design guide.

## Develop

```bash
cd website
npm install
npm run dev        # http://localhost:5173
```

## Build

```bash
npm run build      # static site → .vitepress/dist/
npm run preview    # preview the production build
```

## Structure

```
website/
├── index.md                    # landing page (VitePress home layout)
├── docs/                       # documentation pages
│   ├── what-is-orbita.md
│   ├── why.md
│   ├── architecture.md
│   ├── quickstart.md
│   ├── install.md
│   ├── deploy.md
│   ├── grit-yaml.md
│   ├── cli.md
│   └── troubleshooting.md
├── public/logo.svg
└── .vitepress/
    ├── config.mjs              # site config, nav, sidebar
    └── theme/
        ├── index.js
        └── brand.css           # brand theme (colors, mono, hint boxes)
```

## Deploy

The static output in `.vitepress/dist/` can be served by anything (Netlify, Cloudflare Pages,
GitHub Pages, or an Orbita app). For GitHub Pages, set `base: '/orbita/'` in `config.mjs`.

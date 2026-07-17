// Central site config: nav + docs sidebar. Edit here to change navigation.

export const site = {
  name: 'Orbita',
  github: 'https://github.com/MUKE-coder/orbita',
  version: 'v1.0.0',
}

export const nav = [
  { title: 'Docs', href: '/docs/what-is-orbita' },
  { title: 'Quickstart', href: '/docs/quickstart' },
  { title: 'GitHub', href: site.github, external: true },
]

export type DocLink = { title: string; href: string }
export type DocSection = { title: string; items: DocLink[] }

// Order matters: it drives the sidebar and prev/next. The general case leads —
// Grit is a showcase guide, never a prerequisite.
export const sidebar: DocSection[] = [
  {
    title: 'Introduction',
    items: [
      { title: 'What is Orbita?', href: '/docs/what-is-orbita' },
      { title: 'Why it exists', href: '/docs/why' },
      { title: 'Architecture', href: '/docs/architecture' },
    ],
  },
  {
    title: 'Get started',
    items: [{ title: 'Getting started', href: '/docs/quickstart' }],
  },
  {
    title: 'Guides',
    items: [{ title: 'Deploying Grit apps', href: '/docs/grit-apps' }],
  },
  {
    title: 'Reference',
    items: [
      { title: 'orbita.yaml spec', href: '/docs/orbita-yaml' },
      { title: 'CLI reference', href: '/docs/cli' },
      { title: 'Troubleshooting', href: '/docs/troubleshooting' },
    ],
  },
]

// Flat ordered list for prev/next.
export const docOrder: DocLink[] = sidebar.flatMap((s) => s.items)

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
    items: [
      { title: 'Quickstart', href: '/docs/quickstart' },
      { title: 'Install on a fresh server', href: '/docs/install' },
      { title: 'Deploy a Grit app', href: '/docs/deploy' },
    ],
  },
  {
    title: 'Reference',
    items: [
      { title: 'grit.yaml spec', href: '/docs/grit-yaml' },
      { title: 'CLI reference', href: '/docs/cli' },
      { title: 'Troubleshooting', href: '/docs/troubleshooting' },
    ],
  },
]

// Flat ordered list for prev/next.
export const docOrder: DocLink[] = sidebar.flatMap((s) => s.items)

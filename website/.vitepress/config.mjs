import { defineConfig } from 'vitepress'

// Orbita / Grit Cloud docs site. Dark-first, techy, brand-themed (indigo/cyan,
// mono for real values) per the Grit Cloud design guide.
export default defineConfig({
  title: 'Orbita',
  description: 'Self-hosted multi-tenant PaaS — and the control plane for Grit Cloud. Deploy a full Grit app in two commands.',
  lang: 'en-US',
  cleanUrls: true,
  appearance: 'dark',
  // The repo README for this dir shouldn't become a published page.
  srcExclude: ['README.md'],

  head: [
    ['meta', { name: 'theme-color', content: '#6D5CE7' }],
    ['meta', { property: 'og:title', content: 'Orbita — self-hosted PaaS + Grit Cloud' }],
    ['meta', { property: 'og:description', content: 'One VPS. Many clients. Full isolation. Deploy a full Grit app in two commands.' }],
  ],

  markdown: {
    theme: { light: 'github-light', dark: 'github-dark' },
    lineNumbers: false,
  },

  themeConfig: {
    logo: '/logo.svg',
    siteTitle: 'Orbita',

    nav: [
      { text: 'Home', link: '/' },
      { text: 'Docs', link: '/docs/what-is-orbita' },
      { text: 'Quickstart', link: '/docs/quickstart' },
      {
        text: 'v1.0.0',
        items: [
          { text: 'grit.yaml spec', link: '/docs/grit-yaml' },
          { text: 'CLI reference', link: '/docs/cli' },
          { text: 'Troubleshooting', link: '/docs/troubleshooting' },
          { text: 'GitHub', link: 'https://github.com/MUKE-coder/orbita' },
        ],
      },
    ],

    sidebar: {
      '/docs/': [
        {
          text: 'Introduction',
          collapsed: false,
          items: [
            { text: 'What is Orbita?', link: '/docs/what-is-orbita' },
            { text: 'Why it exists', link: '/docs/why' },
            { text: 'Architecture', link: '/docs/architecture' },
          ],
        },
        {
          text: 'Get started',
          collapsed: false,
          items: [
            { text: 'Quickstart (2 commands)', link: '/docs/quickstart' },
            { text: 'Install on a fresh server', link: '/docs/install' },
            { text: 'Deploy a Grit app', link: '/docs/deploy' },
          ],
        },
        {
          text: 'Reference',
          collapsed: false,
          items: [
            { text: 'grit.yaml spec', link: '/docs/grit-yaml' },
            { text: 'CLI reference', link: '/docs/cli' },
            { text: 'Troubleshooting', link: '/docs/troubleshooting' },
          ],
        },
      ],
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/MUKE-coder/orbita' },
    ],

    search: { provider: 'local' },

    footer: {
      message: 'MIT Licensed. Built with Go, React, and Grit.',
      copyright: 'Orbita — self-hosted PaaS + Grit Cloud',
    },

    outline: { level: [2, 3], label: 'On this page' },
    docFooter: { prev: 'Previous', next: 'Next' },
  },
})

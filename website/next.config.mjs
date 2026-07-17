import createMDX from '@next/mdx'
import remarkGfm from 'remark-gfm'
import rehypeSlug from 'rehype-slug'
import rehypePrettyCode from 'rehype-pretty-code'

// Shiki syntax highlighting via rehype-pretty-code (build-time, no client JS).
// Single dark theme on purpose: the dual {dark,light} form only emits
// --shiki-dark/--shiki-light CSS vars and needs extra CSS to apply a color,
// which silently rendered every block unhighlighted. The site is dark-only.
const prettyCodeOptions = {
  theme: 'github-dark-default',
  keepBackground: false, // the figure card supplies the surface
  defaultLang: 'text',
}

const withMDX = createMDX({
  extension: /\.mdx?$/,
  options: {
    remarkPlugins: [remarkGfm],
    // rehype-slug gives every heading an id — used by anchors + the page TOC.
    rehypePlugins: [rehypeSlug, [rehypePrettyCode, prettyCodeOptions]],
  },
})

/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone', // small production image for Dokploy
  pageExtensions: ['ts', 'tsx', 'md', 'mdx'],
  reactStrictMode: true,
}

export default withMDX(nextConfig)

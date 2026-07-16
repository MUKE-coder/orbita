import createMDX from '@next/mdx'
import remarkGfm from 'remark-gfm'
import rehypePrettyCode from 'rehype-pretty-code'

// Shiki syntax highlighting via rehype-pretty-code (build-time, no client JS).
const prettyCodeOptions = {
  theme: { dark: 'github-dark', light: 'github-light' },
  keepBackground: false,
}

const withMDX = createMDX({
  extension: /\.mdx?$/,
  options: {
    remarkPlugins: [remarkGfm],
    rehypePlugins: [[rehypePrettyCode, prettyCodeOptions]],
  },
})

/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone', // small production image for Dokploy
  pageExtensions: ['ts', 'tsx', 'md', 'mdx'],
  reactStrictMode: true,
}

export default withMDX(nextConfig)

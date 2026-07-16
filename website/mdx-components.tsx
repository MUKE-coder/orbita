import type { MDXComponents } from 'mdx/types'
import Link from 'next/link'
import { Callout } from '@/components/Callout'
import { Tabs, TabItem } from '@/components/Tabs'

// Global MDX component map. Custom components (Callout, Tabs, TabItem) are also
// available inside every .mdx file without importing them.
export function useMDXComponents(components: MDXComponents): MDXComponents {
  return {
    a: ({ href = '', children, ...props }) =>
      href.startsWith('/') ? (
        <Link href={href} {...props}>
          {children}
        </Link>
      ) : (
        <a href={href} target="_blank" rel="noreferrer" {...props}>
          {children}
        </a>
      ),
    Callout,
    Tabs,
    TabItem,
    ...components,
  }
}

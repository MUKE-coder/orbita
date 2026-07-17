import type { MDXComponents } from 'mdx/types'
import Link from 'next/link'
import { Callout } from '@/components/Callout'
import { Tabs, TabItem } from '@/components/Tabs'
import { Cards, Card } from '@/components/Cards'
import { CodeFigure } from '@/components/CodeFigure'

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
    // rehype-pretty-code wraps every code block in a <figure>; CodeFigure adds
    // the header bar (environment badge + language) and the copy button.
    figure: CodeFigure,
    Callout,
    Tabs,
    TabItem,
    Cards,
    Card,
    ...components,
  }
}

'use client'

// "On this page" panel. Reads headings from the rendered article on mount
// (rehype-slug has already given them ids) and highlights the one in view.

import { useEffect, useState } from 'react'
import { usePathname } from 'next/navigation'

type Heading = { id: string; text: string; level: number }

export function Toc() {
  const path = usePathname()
  const [headings, setHeadings] = useState<Heading[]>([])
  const [active, setActive] = useState<string>('')

  useEffect(() => {
    const nodes = Array.from(
      document.querySelectorAll<HTMLHeadingElement>('article h2[id], article h3[id]')
    )
    setHeadings(
      nodes.map((n) => ({
        id: n.id,
        text: n.textContent ?? '',
        level: Number(n.tagName[1]),
      }))
    )

    if (nodes.length === 0) return

    const observer = new IntersectionObserver(
      (entries) => {
        const visible = entries
          .filter((e) => e.isIntersecting)
          .sort((a, b) => a.boundingClientRect.top - b.boundingClientRect.top)
        if (visible[0]) setActive(visible[0].target.id)
      },
      // Trigger when a heading sits in the upper part of the viewport.
      { rootMargin: '-80px 0px -70% 0px', threshold: 0 }
    )
    nodes.forEach((n) => observer.observe(n))
    return () => observer.disconnect()
  }, [path])

  if (headings.length === 0) return null

  return (
    <nav aria-label="On this page" className="text-sm">
      <div className="mb-3 text-[11px] font-medium uppercase tracking-wide text-textMuted">
        On this page
      </div>
      <ul className="space-y-1 border-l border-hair/60">
        {headings.map((h) => (
          <li key={h.id}>
            <a
              href={`#${h.id}`}
              className={`-ml-px block border-l py-1 pr-2 transition-colors ${
                h.level === 3 ? 'pl-6' : 'pl-3'
              } ${
                active === h.id
                  ? 'border-primary font-medium text-primary'
                  : 'border-transparent text-textSecondary hover:border-hair hover:text-textPrimary'
              }`}
            >
              {h.text}
            </a>
          </li>
        ))}
      </ul>
    </nav>
  )
}

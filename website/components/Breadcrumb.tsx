'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { ChevronRight } from 'lucide-react'
import { sidebar } from '@/lib/site'

export function Breadcrumb() {
  const path = usePathname()
  const section = sidebar.find((s) => s.items.some((i) => i.href === path))
  const page = section?.items.find((i) => i.href === path)
  if (!section || !page) return null

  return (
    <nav aria-label="Breadcrumb" className="mb-3 flex items-center gap-1.5 text-xs text-textMuted">
      <Link href="/docs/what-is-orbita" className="hover:text-textSecondary">
        Docs
      </Link>
      <ChevronRight size={13} className="text-textDisabled" />
      <span>{section.title}</span>
      <ChevronRight size={13} className="text-textDisabled" />
      <span className="text-textSecondary">{page.title}</span>
    </nav>
  )
}

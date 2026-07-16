'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { ArrowLeft, ArrowRight } from 'lucide-react'
import { docOrder } from '@/lib/site'

export function PrevNext() {
  const path = usePathname()
  const i = docOrder.findIndex((d) => d.href === path)
  if (i === -1) return null
  const prev = i > 0 ? docOrder[i - 1] : null
  const next = i < docOrder.length - 1 ? docOrder[i + 1] : null
  return (
    <div className="mt-12 flex items-center justify-between gap-4 border-t border-hair pt-6 text-sm">
      {prev ? (
        <Link href={prev.href} className="flex items-center gap-2 text-muted hover:text-ink">
          <ArrowLeft size={15} /> {prev.title}
        </Link>
      ) : (
        <span />
      )}
      {next ? (
        <Link href={next.href} className="flex items-center gap-2 text-muted hover:text-ink">
          {next.title} <ArrowRight size={15} />
        </Link>
      ) : (
        <span />
      )}
    </div>
  )
}

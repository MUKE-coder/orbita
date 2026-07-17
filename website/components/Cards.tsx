'use client'

// Choice cards for docs. Used where a guide forks on the reader's situation
// (CLI vs no CLI, has an SSH key vs not) so nobody has to read past steps that
// don't apply to them.
//
// Usage in MDX:
//   <Cards>
//     <Card title="…" badge="Recommended" href="#anchor" icon="terminal">body</Card>
//   </Cards>

import type { ReactNode } from 'react'
import Link from 'next/link'
import { Terminal, MousePointerClick, KeyRound, Lock, ArrowRight } from 'lucide-react'

const ICONS = {
  terminal: Terminal,
  browser: MousePointerClick,
  key: KeyRound,
  lock: Lock,
} as const

export type CardIcon = keyof typeof ICONS

export function Cards({ children }: { children: ReactNode }) {
  return <div className="my-6 grid gap-4 sm:grid-cols-2">{children}</div>
}

export function Card({
  title,
  badge,
  href,
  icon,
  children,
}: {
  title: string
  badge?: string
  href?: string
  icon?: CardIcon
  children: ReactNode
}) {
  const Icon = icon ? ICONS[icon] : null

  const inner = (
    <>
      <div className="mb-3 flex items-start justify-between gap-3">
        {Icon ? (
          <span className="inline-flex size-9 shrink-0 items-center justify-center rounded-lg border border-primary/20 bg-gradient-to-b from-primary/15 to-primary/5 text-primary">
            <Icon size={17} />
          </span>
        ) : null}
        {badge ? (
          <span className="shrink-0 rounded-full border border-primary/30 bg-primary/10 px-2 py-0.5 text-[10px] font-medium uppercase tracking-wide text-primary">
            {badge}
          </span>
        ) : null}
      </div>

      <h3 className="mb-1.5 text-sm font-semibold text-textPrimary">{title}</h3>
      <div className="text-xs leading-relaxed text-textSecondary [&>p]:my-0">{children}</div>

      {href ? (
        <span className="mt-3 inline-flex items-center gap-1 text-xs font-medium text-primary">
          Go there <ArrowRight size={13} className="transition-transform group-hover:translate-x-0.5" />
        </span>
      ) : null}
    </>
  )

  const cls =
    'group glass glass-bevel flex flex-col p-5 transition-all duration-300 hover:border-primary/40 hover:shadow-[0_8px_24px_-12px_rgba(244,91,72,0.3)]'

  return href ? (
    <Link href={href} className={cls}>
      {inner}
    </Link>
  ) : (
    <div className={cls}>{inner}</div>
  )
}

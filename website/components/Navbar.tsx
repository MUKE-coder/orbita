'use client'

import Link from 'next/link'
import { Github } from 'lucide-react'
import { Logo } from './Logo'
import { nav, site } from '@/lib/site'

export function Navbar() {
  return (
    <header className="sticky top-0 z-40 border-b border-hair/70 bg-bg/80 backdrop-blur">
      <div className="mx-auto flex h-14 max-w-6xl items-center justify-between px-5">
        <Link href="/" className="flex items-center gap-2.5">
          <Logo size={26} />
          <span className="font-semibold tracking-tight">Orbita</span>
          <span className="hidden rounded-full border border-hair px-2 py-0.5 font-mono text-[11px] text-faint sm:inline">
            {site.version}
          </span>
        </Link>

        <nav className="flex items-center gap-1 text-sm">
          {nav.map((item) =>
            item.external ? (
              <a
                key={item.href}
                href={item.href}
                target="_blank"
                rel="noreferrer"
                className="flex items-center gap-1.5 rounded-md px-3 py-1.5 text-muted hover:bg-elev2 hover:text-ink"
              >
                <Github size={15} />
                <span className="hidden sm:inline">GitHub</span>
              </a>
            ) : (
              <Link
                key={item.href}
                href={item.href}
                className="rounded-md px-3 py-1.5 text-muted hover:bg-elev2 hover:text-ink"
              >
                {item.title}
              </Link>
            )
          )}
        </nav>
      </div>
    </header>
  )
}

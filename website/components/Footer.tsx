import Link from 'next/link'
import { Logo } from './Logo'
import { site } from '@/lib/site'

export function Footer() {
  return (
    <footer className="mt-24 border-t border-hair/70">
      <div className="mx-auto flex max-w-6xl flex-col gap-4 px-5 py-10 text-sm text-faint sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-2.5">
          <Logo size={22} />
          <span className="text-muted">Orbita</span>
          <span>— self-hosted PaaS + Grit Cloud</span>
        </div>
        <div className="flex items-center gap-4">
          <Link href="/docs/what-is-orbita" className="hover:text-ink">
            Docs
          </Link>
          <a href={site.github} target="_blank" rel="noreferrer" className="hover:text-ink">
            GitHub
          </a>
          <span>MIT Licensed</span>
        </div>
      </div>
    </footer>
  )
}

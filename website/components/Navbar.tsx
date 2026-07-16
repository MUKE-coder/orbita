import Link from 'next/link'
import { Star } from 'lucide-react'
import { Logo } from './Logo'
import { GithubMark } from './icons/GithubMark'
import { site } from '@/lib/site'

export function Navbar() {
  return (
    <header className="sticky top-0 z-50">
      <nav className="bg-bgDark/80 backdrop-blur-xl">
        <div className="mx-auto max-w-6xl px-4 lg:px-6">
          <div className="flex h-14 items-center justify-between">
            <Link href="/" className="flex items-center gap-2.5">
              <Logo size={26} />
              <span className="text-lg font-semibold text-white">Orbita</span>
              <span className="hidden rounded-full border border-hair px-2 py-0.5 font-mono text-[11px] text-textMuted sm:inline">
                {site.version}
              </span>
            </Link>

            <div className="flex items-center gap-1 text-sm">
              <Link
                href="/docs/what-is-orbita"
                className="hidden items-center gap-2 rounded-md px-2.5 py-1.5 font-medium text-textPrimary hover:bg-surface hover:text-primary-hover md:flex"
              >
                Docs
              </Link>
              <Link
                href="/docs/quickstart"
                className="hidden items-center gap-2 rounded-md px-2.5 py-1.5 font-medium text-textPrimary hover:bg-surface hover:text-primary-hover md:flex"
              >
                Quickstart
              </Link>
              <span className="mx-1 hidden h-4 w-px bg-hair md:block" />
              <a
                href={site.github}
                target="_blank"
                rel="noreferrer"
                className="inline-flex items-center gap-2 rounded-full border border-hair bg-surface px-3 py-1.5 text-textPrimary transition-colors hover:bg-surfaceHover"
              >
                <GithubMark className="size-4" />
                <span className="hidden items-center gap-1 sm:inline-flex">
                  <Star size={13} className="fill-textMuted text-textMuted" />
                  Star
                </span>
              </a>
            </div>
          </div>
        </div>
      </nav>
    </header>
  )
}

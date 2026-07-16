import Link from 'next/link'
import { site } from '@/lib/site'

export function Footer() {
  return (
    <footer className="border-t border-divider/50">
      <div className="mx-auto max-w-6xl px-4 py-10">
        <div className="flex flex-col gap-8 md:flex-row md:items-center md:justify-between">
          <div className="space-y-2">
            <h3 className="text-sm font-semibold text-textPrimary">Orbita</h3>
            <p className="max-w-xs text-xs text-textMuted">
              A self-hosted, multi-tenant PaaS in a single Go binary — and the control plane for
              Grit Cloud. Secure a server and deploy in two commands.
            </p>
          </div>

          <div className="flex flex-wrap gap-6 text-sm">
            <Link href="/docs/what-is-orbita" className="text-textSecondary hover:text-textPrimary">
              Docs
            </Link>
            <Link href="/docs/quickstart" className="text-textSecondary hover:text-textPrimary">
              Quickstart
            </Link>
            <a href={site.github} target="_blank" rel="noreferrer" className="text-textSecondary hover:text-textPrimary">
              GitHub
            </a>
          </div>
        </div>

        <div className="mt-10 flex flex-col items-center justify-between gap-4 border-t border-divider/50 pt-6 text-xs text-textMuted md:flex-row">
          <span>Open source · MIT licensed</span>
          <span>
            Design adapted from{' '}
            <a href="https://github.com/avijit07x/animateicons" target="_blank" rel="noreferrer" className="hover:underline">
              animateicons
            </a>{' '}
            (MIT, © Avijit Dey)
          </span>
        </div>
      </div>
    </footer>
  )
}

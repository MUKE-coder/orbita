'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { sidebar } from '@/lib/site'

export function Sidebar() {
  const path = usePathname()
  return (
    <nav className="space-y-6 text-sm">
      {sidebar.map((section) => (
        <div key={section.title}>
          <div className="mb-2 font-mono text-[11px] uppercase tracking-wide text-faint">
            {section.title}
          </div>
          <ul className="space-y-0.5">
            {section.items.map((item) => {
              const active = path === item.href
              return (
                <li key={item.href}>
                  <Link
                    href={item.href}
                    className={`block rounded-md px-2.5 py-1.5 transition ${
                      active
                        ? 'bg-indigo-soft font-medium text-ink'
                        : 'text-muted hover:bg-elev2 hover:text-ink'
                    }`}
                  >
                    {item.title}
                  </Link>
                </li>
              )
            })}
          </ul>
        </div>
      ))}
    </nav>
  )
}

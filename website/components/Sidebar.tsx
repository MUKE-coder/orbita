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
          <div className="mb-2 text-[11px] font-medium uppercase tracking-wide text-textMuted">
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
                        ? 'bg-surfaceElevated font-medium text-primary'
                        : 'text-textSecondary hover:bg-surfaceHover hover:text-textPrimary'
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

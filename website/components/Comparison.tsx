'use client'

import { Check, Minus } from 'lucide-react'

// Orbita vs the two tools people actually compare it to. Kept honest: the rows
// where all three tick are still shown, so the rows where Orbita stands alone
// (true multi-tenancy, per-tenant encryption, resource quotas, server
// hardening, footprint) carry weight instead of looking cherry-picked.

type Cell = boolean | string
type Row = { label: string; orbita: Cell; dokploy: Cell; coolify: Cell; highlight?: boolean }

const rows: Row[] = [
  { label: 'True multi-tenancy (per-org isolation)', orbita: true, dokploy: false, coolify: false, highlight: true },
  { label: 'Per-tenant AES-256 encryption keys', orbita: true, dokploy: false, coolify: false, highlight: true },
  { label: 'Per-client CPU / RAM quotas (cgroup v2)', orbita: true, dokploy: false, coolify: false, highlight: true },
  { label: 'Hardens the server for you', orbita: true, dokploy: false, coolify: false, highlight: true },
  { label: 'Idle RAM footprint', orbita: '~50 MB', dokploy: '~200 MB', coolify: '~500 MB', highlight: true },
  { label: 'Written in', orbita: 'Go — 1 binary', dokploy: 'Node.js', coolify: 'PHP / Laravel' },
  { label: 'Zero-config framework fast path', orbita: 'Grit', dokploy: false, coolify: false, highlight: true },
  { label: 'DB migrations under an advisory lock', orbita: true, dokploy: false, coolify: false, highlight: true },
  { label: 'Deploy Docker images + Dockerfile repos', orbita: true, dokploy: true, coolify: true },
  { label: 'Nixpacks build (no Dockerfile)', orbita: true, dokploy: true, coolify: true },
  { label: 'Docker Compose stacks', orbita: true, dokploy: true, coolify: true },
  { label: 'Automatic HTTPS (Let’s Encrypt)', orbita: true, dokploy: true, coolify: true },
  { label: 'Push-to-deploy + CLI', orbita: true, dokploy: true, coolify: true },
  { label: '4-role RBAC + teams', orbita: true, dokploy: true, coolify: true },
  { label: 'Built-in observability', orbita: 'Pulse + Sentinel', dokploy: 'Basic', coolify: 'Basic' },
  { label: 'Open source', orbita: true, dokploy: true, coolify: true },
]

function CellView({ v, accent }: { v: Cell; accent?: boolean }) {
  if (v === true)
    return (
      <span className="inline-flex items-center justify-center">
        <Check size={17} className={accent ? 'text-primary' : 'text-success'} />
      </span>
    )
  if (v === false)
    return (
      <span className="inline-flex items-center justify-center">
        <Minus size={16} className="text-textDisabled" />
      </span>
    )
  return <span className={`text-xs ${accent ? 'font-medium text-primary' : 'text-textSecondary'}`}>{v}</span>
}

export function Comparison() {
  return (
    <div className="overflow-x-auto">
      <table className="w-full min-w-[640px] border-collapse text-sm">
        <thead>
          <tr className="border-b border-hair/60">
            <th className="w-[46%] px-4 py-3 text-left font-medium text-textSecondary">How it compares</th>
            <th className="px-4 py-3 text-center">
              <span className="inline-flex items-center gap-1.5 rounded-full border border-primary/30 bg-primary/10 px-3 py-1 text-sm font-semibold text-primary">
                Orbita
              </span>
            </th>
            <th className="px-4 py-3 text-center font-semibold text-textPrimary">Dokploy</th>
            <th className="px-4 py-3 text-center font-semibold text-textPrimary">Coolify</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((r) => (
            <tr
              key={r.label}
              className={`border-b border-hair/40 ${r.highlight ? 'bg-primary/[0.03]' : ''}`}
            >
              <td className="px-4 py-2.5 text-left text-textSecondary">{r.label}</td>
              <td className="px-4 py-2.5 text-center">
                <CellView v={r.orbita} accent />
              </td>
              <td className="px-4 py-2.5 text-center">
                <CellView v={r.dokploy} />
              </td>
              <td className="px-4 py-2.5 text-center">
                <CellView v={r.coolify} />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

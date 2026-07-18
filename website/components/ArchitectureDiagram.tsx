import type { ReactNode } from 'react'
import { ArrowDown, Globe } from 'lucide-react'
import {
  SiTraefikproxy,
  SiDocker,
  SiLetsencrypt,
  SiGo,
  SiReact,
  SiPostgresql,
  SiRedis,
  SiNextdotjs,
  SiLaravel,
  SiDjango,
  SiNodedotjs,
  SiGithub,
} from 'react-icons/si'
import { Logo } from '@/components/Logo'

// Real brand marks (Simple Icons via react-icons), each tinted to its official
// colour but nudged brighter where the true hue would vanish on black.
type Tile = { icon: ReactNode; label: string; sub?: string }

function IconTile({ icon, label, sub }: Tile) {
  return (
    <div className="flex items-center gap-2.5 rounded-lg border border-white/10 bg-white/[0.03] px-3 py-2">
      <span className="flex h-7 w-7 shrink-0 items-center justify-center">{icon}</span>
      <span className="leading-tight">
        <span className="block text-[13px] font-medium text-textPrimary">{label}</span>
        {sub && <span className="block text-[11px] text-textMuted">{sub}</span>}
      </span>
    </div>
  )
}

// A dashed group box with a floating title, echoing the classic "VPS" panel.
function Group({
  title,
  badge,
  accent,
  children,
}: {
  title: string
  badge?: ReactNode
  accent?: boolean
  children: ReactNode
}) {
  return (
    <div
      className={`relative rounded-xl border border-dashed p-4 pt-5 ${
        accent ? 'border-primary/40 bg-primary/[0.04]' : 'border-hair bg-white/[0.015]'
      }`}
    >
      <div className="absolute -top-2.5 left-4 flex items-center gap-2 bg-surface px-2">
        <span
          className={`text-[11px] font-semibold uppercase tracking-wider ${
            accent ? 'text-primary' : 'text-textSecondary'
          }`}
        >
          {title}
        </span>
      </div>
      {badge && <div className="absolute -top-2.5 right-4 bg-surface px-2">{badge}</div>}
      {children}
    </div>
  )
}

function Flow() {
  return (
    <div className="mx-auto flex w-full max-w-md justify-center py-1.5 text-textDisabled">
      <ArrowDown className="h-5 w-5" />
    </div>
  )
}

const sz = 20

export function ArchitectureDiagram() {
  return (
    <div className="not-prose overflow-x-auto rounded-2xl border border-hair bg-surface p-5 sm:p-7">
      <div className="mx-auto flex min-w-[300px] max-w-2xl flex-col">
        {/* Internet */}
        <Group title="Internet">
          <div className="flex flex-wrap justify-center gap-3">
            <IconTile
              icon={<Globe className="h-5 w-5 text-info" />}
              label="Your users"
              sub="Browser + CLI"
            />
            <IconTile
              icon={<SiGithub size={sz} className="text-white" />}
              label="GitHub"
              sub="git push → webhook"
            />
          </div>
        </Group>

        <Flow />

        {/* VPS */}
        <Group
          title="Your VPS"
          accent
          badge={
            <span className="flex items-center gap-1.5">
              <SiDocker size={18} color="#2496ED" />
              <span className="text-[11px] font-medium text-textSecondary">Docker Swarm</span>
            </span>
          }
        >
          <div className="flex flex-col gap-4">
            {/* Reverse proxy */}
            <Group title="Reverse proxy">
              <div className="flex flex-wrap items-center justify-center gap-3">
                <IconTile
                  icon={<SiTraefikproxy size={sz} color="#24A1C1" />}
                  label="Traefik v3"
                  sub="Routing + TLS"
                />
                <IconTile
                  icon={<SiLetsencrypt size={sz} color="#5BB0EB" />}
                  label="Let's Encrypt"
                  sub="Automatic HTTPS"
                />
              </div>
            </Group>

            <Flow />

            {/* Orbita control plane */}
            <Group title="Orbita — one Go binary">
              <div className="mb-3 flex items-center justify-center gap-2">
                <Logo size={22} />
                <span className="text-sm font-semibold text-textPrimary">Control plane</span>
              </div>
              <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
                <IconTile icon={<SiGo size={sz} color="#00ADD8" />} label="Go API" />
                <IconTile icon={<SiReact size={sz} color="#61DAFB" />} label="Dashboard" />
                <IconTile
                  icon={<SiPostgresql size={sz} color="#5A8DEE" />}
                  label="Postgres"
                  sub="metadata"
                />
                <IconTile
                  icon={<SiRedis size={sz} color="#FF4438" />}
                  label="Redis"
                  sub="cache / RL"
                />
              </div>
            </Group>

            <Flow />

            {/* Tenant apps */}
            <Group title="Your apps — isolated per org">
              <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
                <IconTile icon={<SiNextdotjs size={sz} className="text-white" />} label="Next.js" />
                <IconTile icon={<SiLaravel size={sz} color="#FF2D20" />} label="Laravel" />
                <IconTile icon={<SiDjango size={sz} color="#20AA76" />} label="Django" />
                <IconTile icon={<SiNodedotjs size={sz} color="#5FA04E" />} label="Node.js" />
                <IconTile
                  icon={<SiPostgresql size={sz} color="#5A8DEE" />}
                  label="Postgres"
                  sub="provisioned"
                />
                <IconTile icon={<SiRedis size={sz} color="#FF4438" />} label="Redis" sub="provisioned" />
              </div>
              <p className="mt-3 text-center text-[11px] text-textMuted">
                Each org runs on its own Docker network, cgroup slice + AES-256 key — tenants never
                see each other.
              </p>
            </Group>
          </div>
        </Group>
      </div>
    </div>
  )
}

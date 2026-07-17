'use client'

import { useRef, useState } from 'react'
import Link from 'next/link'
import { motion, Variants } from 'framer-motion'
import {
  ArrowRight,
  Boxes,
  Zap,
  Lock,
  Globe,
  Activity,
  Feather,
  ShieldCheck,
  Layers,
  GitBranch,
  Server,
  Terminal,
  Database,
  Check,
  Copy,
} from 'lucide-react'
import { CommandCard } from '@/components/CommandCard'
import { Section, SectionHeader } from '@/components/Section'
import { GithubMark } from '@/components/icons/GithubMark'
import { ArchitectureBoard } from '@/components/ArchitectureBoard'
import { Comparison } from '@/components/Comparison'

const container: Variants = {
  hidden: { opacity: 0 },
  show: { opacity: 1, transition: { staggerChildren: 0.12, delayChildren: 0.1 } },
}
const item: Variants = {
  hidden: { opacity: 0, y: 20 },
  show: { opacity: 1, y: 0, transition: { duration: 0.5, ease: 'easeOut' } },
}

// Glowing coral CTA — the animateicons "Browse icons" button treatment.
const ctaClass =
  'group relative inline-flex items-center justify-center gap-1.5 overflow-hidden rounded-full bg-gradient-to-b from-primary to-primary/85 px-6 py-2.5 text-sm font-semibold text-[var(--cta-text)] shadow-[0_1px_0_rgba(255,255,255,0.18)_inset,0_10px_28px_-8px_rgba(244,91,72,0.55)] ring-1 ring-inset ring-white/15 transition-all duration-200 hover:shadow-[0_1px_0_rgba(255,255,255,0.22)_inset,0_14px_36px_-8px_rgba(244,91,72,0.7)] hover:brightness-110 active:scale-[0.98]'

const highlights = [
  {
    icon: Boxes,
    title: 'True multi-tenancy',
    body: 'Per-org Docker networks, AES-256 keys, cgroup v2 CPU/RAM quotas, and 4-role RBAC. Run every client on one cheap box — they never see each other.',
  },
  {
    icon: ShieldCheck,
    title: 'Secure by the first command',
    body: 'orbita init hardens the server — deploy user, SSH keys, locked-down root, UFW, Fail2ban — then installs Orbita on your HTTPS subdomain.',
  },
  {
    icon: Layers,
    title: 'Any app, no Dockerfile',
    body: 'Dockerfile, Compose, a raw image, or Nixpacks for Laravel/Django/Rails/Node/static. Grit apps get a zero-config fast path on top.',
  },
  {
    icon: GitBranch,
    title: 'Migrations under a lock',
    body: 'Deploys build, then run migrations under a Postgres advisory lock before cutover. A failed migration aborts — never a schema-mismatched cutover.',
  },
]

const features = [
  {
    icon: Server,
    title: 'One ~30 MB binary',
    body: 'The Go control plane embeds the React dashboard and idles under 50 MB of RAM — leaving nearly all of your server for the apps you run.',
  },
  {
    icon: Globe,
    title: 'HTTPS by default',
    body: 'Traefik v3 with automatic Let’s Encrypt, HTTP→HTTPS redirect, and per-app routing generated from orbita.yaml. Only the proxy binds the public host.',
  },
  {
    icon: Activity,
    title: 'Observable from day one',
    body: 'Live logs, metrics and an in-browser terminal for every app. Grit apps additionally get Pulse and Sentinel mounted automatically.',
  },
  {
    icon: Zap,
    title: 'Push to deploy',
    body: 'After the first deploy, every git push to your branch redeploys via webhook. The CLI becomes optional — the platform keeps shipping.',
  },
  {
    icon: Lock,
    title: 'Isolated secrets',
    body: 'Each org’s secrets are encrypted with an AES-256 key HKDF-derived from a master key + org ID — never the master key directly.',
  },
  {
    icon: Feather,
    title: 'The CLI is optional',
    body: 'Deploy from the dashboard or by pushing to git — no install needed. Prefer a terminal? The optional orbita CLI does it all from your machine.',
  },
]

// Small glass code panel used in the "two commands" section.
function CodePanel({ label, file, children }: { label: string; file?: string; children: React.ReactNode }) {
  const ref = useRef<HTMLPreElement>(null)
  const [copied, setCopied] = useState(false)

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(ref.current?.textContent ?? '')
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    } catch {
      /* clipboard can be blocked — the snippet stays selectable */
    }
  }

  return (
    <div className="glass glass-bevel">
      <div className="flex items-center justify-between gap-2 border-b border-hair/40 px-4 py-2 text-[11px] uppercase tracking-wide text-textSecondary">
        <span>{label}</span>
        <div className="flex items-center gap-2">
          {file ? <span className="truncate text-textMuted">{file}</span> : null}
          <button
            type="button"
            onClick={copy}
            aria-label={copied ? 'Copied' : 'Copy code'}
            className="inline-flex size-7 shrink-0 items-center justify-center rounded-md text-textSecondary ring-1 ring-inset ring-transparent transition-colors hover:bg-white/[0.06] hover:text-textPrimary hover:ring-hair/60"
          >
            {copied ? <Check size={14} className="text-success" /> : <Copy size={14} />}
          </button>
        </div>
      </div>
      <pre ref={ref} className="overflow-x-auto px-4 py-3.5 font-mono text-[13px] leading-6">
        {children}
      </pre>
    </div>
  )
}

export default function Home() {
  return (
    <main>
      <div className="relative overflow-hidden">
        {/* Hero */}
        <div className="relative flex min-h-[calc(100dvh-8rem)] items-center justify-center overflow-hidden">
          <div className="bg-grid" />

          <motion.div
            variants={container}
            initial="hidden"
            animate="show"
            className="relative z-10 mx-auto flex w-full max-w-3xl flex-col items-center justify-center gap-8 px-4 py-16 text-center"
          >
            <motion.a
              variants={item}
              href="https://github.com/MUKE-coder/orbita"
              target="_blank"
              rel="noreferrer"
              className="-mb-2 inline-flex items-center gap-2 rounded-full border border-hair bg-surface px-4 py-2 text-xs text-textPrimary hover:bg-surfaceHover"
            >
              <GithubMark className="size-4" />
              <span className="font-medium">Open Source</span>
              <span className="rounded-full border border-hair px-2 py-0.5 text-[10px] text-textSecondary">
                MIT
              </span>
            </motion.a>

            <motion.h1
              variants={item}
              className="text-4xl font-bold tracking-tight sm:text-5xl lg:text-6xl"
            >
              <span className="text-primary">Harden a server,</span>
              <br />
              <span className="font-medium text-textPrimary">then deploy anything</span>
            </motion.h1>

            <motion.div variants={item} className="max-w-xl space-y-2 text-sm leading-relaxed text-zinc-300">
              <p>
                A self-hosted PaaS that <span className="text-textPrimary">secures your server</span>{' '}
                and gives <span className="text-textPrimary">every client true isolation</span> —
                in one ~30&nbsp;MB Go binary. Deploy Laravel, Django, Rails, Next.js, a static site,
                or any Docker image. Where Coolify and Dokploy are single-tenant panels, Orbita is
                built for running many clients on one box.
              </p>
            </motion.div>

            <motion.div variants={item} className="flex w-full items-center justify-center">
              <CommandCard />
            </motion.div>

            <motion.div variants={item} className="flex flex-wrap items-center justify-center gap-3">
              <Link href="/docs/quickstart" className={ctaClass}>
                <span
                  aria-hidden
                  className="pointer-events-none absolute inset-x-6 top-px h-px bg-gradient-to-r from-transparent via-white/40 to-transparent"
                />
                <span>Quickstart</span>
                <ArrowRight className="size-4 transition-transform duration-300 group-hover:translate-x-0.5" />
              </Link>
              <Link
                href="/docs/what-is-orbita"
                className="inline-flex items-center gap-1.5 rounded-full px-4 py-2.5 text-sm font-medium text-textPrimary transition-colors hover:text-primary"
              >
                What is Orbita?
              </Link>
            </motion.div>
          </motion.div>
        </div>

        {/* Two commands, zero config */}
        <Section id="two-commands" noBorder>
          <SectionHeader
            title="Two commands, zero config"
            subtitle="Harden the box, install Orbita, then ship any containerised app — from the dashboard or the terminal, whichever you prefer."
            spacing="tight"
          />

          <div className="grid items-center gap-6 sm:gap-8 lg:grid-cols-2 lg:gap-12">
            <ul className="grid h-full min-w-0 gap-4 sm:grid-cols-2 sm:gap-5">
              {highlights.map(({ icon: Icon, title, body }) => (
                <li key={title} className="glass glass-bevel group flex flex-col gap-3 p-5 transition-all duration-300 hover:border-primary/40 hover:shadow-[0_8px_24px_-12px_rgba(244,91,72,0.3)]">
                  <div className="inline-flex size-10 items-center justify-center rounded-xl border border-primary/20 bg-gradient-to-b from-primary/15 to-primary/5 text-primary">
                    <Icon size={20} />
                  </div>
                  <h3 className="text-sm font-semibold text-textPrimary">{title}</h3>
                  <p className="text-xs leading-relaxed text-textSecondary">{body}</p>
                </li>
              ))}
            </ul>

            <div className="flex min-w-0 flex-col gap-4">
              <CodePanel label="Provision" file="on your laptop">
                <code>
                  <span className="select-none text-primary">$ </span>
                  <span className="text-textPrimary">orbita init</span>
                  {'\n'}
                  <span className="text-success">  ✔ Server hardened (score 94/100)</span>
                  {'\n'}
                  <span className="text-success">  ✔ Orbita live at https://orbita.example.com</span>
                </code>
              </CodePanel>

              <CodePanel label="Deploy" file="in your Grit app">
                <code>
                  <span className="select-none text-primary">$ </span>
                  <span className="text-textPrimary">orbita deploy --host prod</span>
                  {'\n'}
                  <span className="text-success">  ✔ Migrations applied (advisory lock)</span>
                  {'\n'}
                  <span className="text-success">  ✔ Live</span>
                  {'\n'}
                  <span className="text-textSecondary">    App:  https://rental.example.com</span>
                  {'\n'}
                  <span className="text-textSecondary">    API:  https://api.rental.example.com</span>
                  {'\n'}
                  <span className="text-textMuted">  git push → auto-deploys via webhook</span>
                </code>
              </CodePanel>

              <div className="flex flex-wrap items-center gap-3 pt-1">
                <Link href="/docs/quickstart" className={ctaClass}>
                  <span aria-hidden className="pointer-events-none absolute inset-x-5 top-px h-px bg-gradient-to-r from-transparent via-white/40 to-transparent" />
                  Read the quickstart
                  <ArrowRight className="size-4 transition-transform duration-300 group-hover:translate-x-0.5" />
                </Link>
                <Link href="/docs/deploy" className="inline-flex items-center gap-1.5 rounded-full px-4 py-2 text-sm font-medium text-textPrimary transition-colors hover:text-primary">
                  How deploy works
                </Link>
              </div>
            </div>
          </div>
        </Section>

        {/* Architecture illustration */}
        <Section id="architecture">
          <SectionHeader
            title="How it fits together"
            subtitle="Your CLI and browser talk to one Go binary. It drives Traefik, Postgres, Redis, and Docker Swarm on the host — every org fully isolated."
            spacing="tight"
          />

          <div className="glass glass-bevel mx-auto max-w-4xl p-2 sm:p-4">
            <ArchitectureBoard />
          </div>

          <div className="mt-8 grid gap-6 sm:grid-cols-3">
            {[
              {
                icon: Terminal,
                title: 'Entry points',
                body: 'The grit CLI (over an orb_ token) and the embedded dashboard both hit the same REST API — nothing else is exposed.',
              },
              {
                icon: Boxes,
                title: 'One control plane',
                body: 'A single ~30 MB Go binary: Gin router → services → GORM, plus the orchestrator that speaks the Docker SDK and writes Traefik config.',
              },
              {
                icon: Database,
                title: 'Backing services',
                body: 'Postgres holds metadata + encrypted secrets, Redis handles cache and rate limits, and Swarm runs every workload with rolling updates.',
              },
            ].map(({ icon: Icon, title, body }) => (
              <div
                key={title}
                className="glass glass-bevel p-5"
              >
                <div className="mb-3 inline-flex size-9 items-center justify-center rounded-lg border border-primary/20 bg-gradient-to-b from-primary/15 to-primary/5 text-primary">
                  <Icon size={18} />
                </div>
                <h3 className="mb-1.5 text-sm font-semibold text-textPrimary">{title}</h3>
                <p className="text-xs leading-relaxed text-textSecondary">{body}</p>
              </div>
            ))}
          </div>
        </Section>

        {/* Feature grid */}
        <Section>
          <SectionHeader
            title="Built for one box, many clients"
            subtitle="True isolation, automatic HTTPS, and observability — without a heavy control plane."
          />

          <motion.div
            variants={container}
            initial="hidden"
            whileInView="show"
            viewport={{ once: true, margin: '-80px' }}
            className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3"
          >
            {features.map(({ icon: Icon, title, body }) => (
              <motion.div
                key={title}
                variants={item}
                className="glass glass-bevel group p-6 transition-all duration-300 hover:border-primary/40 hover:shadow-[0_8px_24px_-12px_rgba(244,91,72,0.3)]"
              >
                <div className="mb-4 inline-flex size-10 items-center justify-center rounded-xl border border-primary/20 bg-gradient-to-b from-primary/15 to-primary/5 text-primary">
                  <Icon size={22} />
                </div>
                <h3 className="mb-2 text-sm font-semibold text-textPrimary">{title}</h3>
                <p className="text-sm leading-relaxed text-textSecondary">{body}</p>
              </motion.div>
            ))}
          </motion.div>
        </Section>

        {/* Comparison */}
        <Section id="compare">
          <SectionHeader
            title="Orbita vs Coolify vs Dokploy"
            subtitle="Coolify and Dokploy are excellent single-tenant panels. Orbita is built for isolation and density — many clients on one box, each sealed off, on a fraction of the RAM."
          />
          <div className="glass glass-bevel mx-auto max-w-4xl p-2 sm:p-4">
            <Comparison />
          </div>
          <p className="mx-auto mt-4 max-w-2xl text-center text-xs text-textMuted">
            Rows highlighted in coral are where Orbita is designed differently. Everything else the
            three tools share — Orbita just does it in a 30&nbsp;MB binary.
          </p>
        </Section>

        {/* Closing CTA */}
        <Section>
          <div className="glass glass-bevel mx-auto max-w-3xl p-8 text-center sm:p-12">
            <h2 className="text-2xl font-semibold tracking-tight text-white sm:text-3xl">
              Secure a VPS and run many clients on it — from one binary.
            </h2>
            <p className="mx-auto mt-3 max-w-xl text-sm text-textSecondary">
              True per-tenant isolation, automatic HTTPS, migrations under a lock, and built-in
              observability — self-hosted on a box that costs a few dollars a month.
            </p>
            <div className="mt-6 flex justify-center">
              <Link href="/docs/quickstart" className={ctaClass}>
                <span aria-hidden className="pointer-events-none absolute inset-x-6 top-px h-px bg-gradient-to-r from-transparent via-white/40 to-transparent" />
                Get started
                <ArrowRight className="size-4 transition-transform duration-300 group-hover:translate-x-0.5" />
              </Link>
            </div>
          </div>
        </Section>
      </div>
    </main>
  )
}

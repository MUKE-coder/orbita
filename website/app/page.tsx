'use client'

import Link from 'next/link'
import { motion } from 'framer-motion'
import {
  ArrowRight,
  Boxes,
  Zap,
  Lock,
  Globe,
  Activity,
  Feather,
} from 'lucide-react'
import { CodeCard } from '@/components/CodeCard'

const features = [
  {
    icon: Boxes,
    title: 'True multi-tenancy',
    body: 'Isolated Docker networks, per-org AES-256 keys, cgroup v2 CPU/RAM quotas, and 4-role RBAC. Run every client on one cheap box — they never see each other.',
  },
  {
    icon: Zap,
    title: 'Grit-aware, zero-config',
    body: 'A Grit app has a known shape, so Orbita reads it — builds each service from the Dockerfiles Grit ships, provisions Postgres/Redis/MinIO, wires domains, and migrates.',
  },
  {
    icon: Lock,
    title: 'Migrations under a lock',
    body: 'Deploys build, then run migrations under a Postgres advisory lock before cutover. A failed migration aborts — never a schema-mismatched cutover.',
  },
  {
    icon: Globe,
    title: 'HTTPS by default',
    body: 'Traefik v3 with automatic Let’s Encrypt, HTTP→HTTPS redirect, and per-app routing generated from grit.yaml. Nothing binds the public host but the proxy.',
  },
  {
    icon: Activity,
    title: 'Observable from day one',
    body: 'Pulse (latency/SQL/errors) and Sentinel (WAF/rate-limit/anomaly) mount on every Grit app by default. Live logs, metrics, and an in-browser terminal.',
  },
  {
    icon: Feather,
    title: 'One ~30 MB binary',
    body: 'The Go control plane embeds the React dashboard and idles under 50 MB of RAM — leaving nearly all of your server for the apps you actually run.',
  },
]

const fade = {
  hidden: { opacity: 0, y: 14 },
  show: (i: number) => ({
    opacity: 1,
    y: 0,
    transition: { delay: i * 0.05, duration: 0.4, ease: 'easeOut' },
  }),
}

export default function Home() {
  return (
    <main className="mx-auto max-w-6xl px-5">
      {/* Hero */}
      <section className="grid items-center gap-10 py-16 lg:grid-cols-2 lg:py-24">
        <div>
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.4 }}
            className="inline-flex items-center gap-2 rounded-full border border-hair bg-elev1 px-3 py-1 font-mono text-[11px] text-muted"
          >
            <span className="h-1.5 w-1.5 animate-pulse-dot rounded-full bg-cyan" />
            SELF-HOSTED · MULTI-TENANT · GRIT-AWARE
          </motion.div>

          <motion.h1
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.45, delay: 0.05 }}
            className="mt-5 text-4xl font-semibold leading-[1.1] tracking-tight sm:text-5xl"
          >
            One control plane for{' '}
            <span className="bg-gradient-to-r from-indigo to-cyan bg-clip-text text-transparent">
              your whole stack.
            </span>
          </motion.h1>

          <motion.p
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.45, delay: 0.12 }}
            className="mt-5 max-w-xl text-[15px] leading-7 text-muted"
          >
            A self-hosted, multi-tenant PaaS in a single Go binary — and the control plane for Grit
            Cloud. Go from a bare, unsecured VPS to a hardened box running a live, migrated, HTTPS
            app in two commands.
          </motion.p>

          <motion.div
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.45, delay: 0.18 }}
            className="mt-8 flex flex-wrap items-center gap-3"
          >
            <Link
              href="/docs/quickstart"
              className="inline-flex items-center gap-2 rounded-lg bg-gradient-to-r from-indigo to-[#7d6ef0] px-4 py-2.5 text-sm font-medium text-white transition hover:brightness-110"
            >
              Quickstart <ArrowRight size={16} />
            </Link>
            <Link
              href="/docs/what-is-orbita"
              className="inline-flex items-center gap-2 rounded-lg border border-hair px-4 py-2.5 text-sm text-ink transition hover:bg-elev2"
            >
              What is Orbita?
            </Link>
          </motion.div>
        </div>

        <motion.div
          initial={{ opacity: 0, scale: 0.98 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.5, delay: 0.15 }}
        >
          <CodeCard
            title="two commands, from bare server to live app"
            lines={[
              { text: '$ grit cloud init', tone: 'cmd' },
              { text: '  ✔ Server hardened (score 94/100)', tone: 'ok' },
              { text: '  ✔ Orbita live at https://orbita.example.com', tone: 'ok' },
              { text: '', tone: 'muted' },
              { text: '$ grit deploy --host prod', tone: 'cmd' },
              { text: '  ✔ Migrations applied (advisory lock)', tone: 'ok' },
              { text: '  ✔ Live', tone: 'ok' },
              { text: '    App:  https://rental.example.com', tone: 'accent' },
              { text: '    API:  https://api.rental.example.com', tone: 'accent' },
              { text: '  git push → auto-deploys via webhook', tone: 'muted' },
            ]}
          />
        </motion.div>
      </section>

      {/* Features */}
      <section className="py-8">
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {features.map((f, i) => (
            <motion.div
              key={f.title}
              custom={i}
              variants={fade}
              initial="hidden"
              whileInView="show"
              viewport={{ once: true, margin: '-40px' }}
              className="group rounded-xl border border-hair bg-elev1 p-5 transition hover:-translate-y-0.5 hover:border-indigo"
            >
              <div className="mb-3 inline-flex rounded-lg bg-indigo-soft p-2.5">
                <f.icon size={18} className="text-indigo" />
              </div>
              <h3 className="text-[15px] font-semibold">{f.title}</h3>
              <p className="mt-1.5 text-sm leading-6 text-muted">{f.body}</p>
            </motion.div>
          ))}
        </div>
      </section>

      {/* Closing CTA */}
      <section className="py-16">
        <div className="rounded-2xl border border-hair bg-gradient-to-b from-elev1 to-bg p-8 text-center sm:p-12">
          <h2 className="text-2xl font-semibold tracking-tight sm:text-3xl">
            Secure a VPS and ship a full-stack app in two commands.
          </h2>
          <p className="mx-auto mt-3 max-w-xl text-muted">
            Self-hosted on a €4.5/mo box, with true isolation, migrations, and observability. That’s
            Grit Cloud.
          </p>
          <Link
            href="/docs/quickstart"
            className="mt-6 inline-flex items-center gap-2 rounded-lg bg-gradient-to-r from-indigo to-[#7d6ef0] px-5 py-2.5 text-sm font-medium text-white transition hover:brightness-110"
          >
            Get started <ArrowRight size={16} />
          </Link>
        </div>
      </section>
    </main>
  )
}

'use client'

// Hero install surface — a glass pill toggle over a code card with a copy
// button. Mirrors the animateicons CmdSection visual vocabulary (gradient
// pill, motion layout active state, coral active label), with Orbita's flow.

import { useState } from 'react'
import { motion } from 'framer-motion'
import { Check, Copy } from 'lucide-react'

type Method = { value: string; label: string; hint: string; code: string }

// Every command here must be real and copy-pasteable — these are the URLs the
// docs use, not placeholders.
const METHODS: Method[] = [
  {
    value: 'harden',
    label: '1. Harden',
    hint: 'Run on your server over SSH — deploy user, keys, UFW, Fail2ban',
    code:
      'curl -sSL https://raw.githubusercontent.com/MUKE-coder/vps-harden/main/vps-harden.sh -o h.sh \\\n' +
      '  && sudo bash h.sh --no-dokploy',
  },
  {
    value: 'install',
    label: '2. Install',
    hint: 'Brings its own Docker, Swarm, Postgres, Redis and Traefik',
    code:
      'curl -sSL https://raw.githubusercontent.com/MUKE-coder/orbita/main/install.sh \\\n' +
      '  | sudo bash -s -- --yes',
  },
  {
    value: 'cli',
    label: 'Or: one command',
    hint: 'Optional CLI — does both of the above from your own machine',
    code: 'orbita init',
  },
]

export function CommandCard() {
  const [active, setActive] = useState(0)
  const [copied, setCopied] = useState(false)
  const method = METHODS[active]

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(method.code.replace(/\\\n\s*/g, ' '))
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    } catch {
      /* clipboard can be blocked — the snippet stays selectable */
    }
  }

  return (
    <div className="flex w-full flex-col items-stretch gap-3 lg:max-w-[42rem]">
      <div
        role="tablist"
        aria-label="Install method"
        className="inline-flex self-center rounded-full border border-hair/80 bg-gradient-to-b from-surface to-surfaceElevated p-1 text-xs shadow-[0_1px_0_rgba(255,255,255,0.04)_inset,0_8px_24px_-12px_rgba(0,0,0,0.6)] backdrop-blur"
      >
        {METHODS.map((opt, i) => {
          const isActive = i === active
          return (
            <button
              key={opt.value}
              role="tab"
              type="button"
              aria-selected={isActive}
              title={opt.hint}
              onClick={() => setActive(i)}
              className={`relative inline-flex items-center rounded-full px-4 py-1.5 font-medium transition-colors ${
                isActive ? 'text-primary' : 'text-textSecondary hover:text-textPrimary'
              }`}
            >
              {isActive && (
                <motion.span
                  layoutId="cmd-method-pill"
                  className="absolute inset-0 -z-10 rounded-full bg-gradient-to-b from-white/[0.06] to-transparent ring-1 ring-inset ring-primary/30"
                  transition={{ type: 'spring', stiffness: 380, damping: 32 }}
                />
              )}
              {opt.label}
            </button>
          )
        })}
      </div>

      <div className="glass glass-bevel w-full text-start">
        <div className="flex items-center justify-between gap-2 border-b border-hair/40 px-4 py-2 text-[11px] uppercase tracking-wide text-textSecondary">
          <span>Terminal</span>
          <button
            type="button"
            onClick={copy}
            aria-label={copied ? 'Copied' : 'Copy command'}
            className="inline-flex size-7 items-center justify-center rounded-md text-textSecondary ring-1 ring-inset ring-transparent transition-colors hover:bg-white/[0.04] hover:text-textPrimary hover:ring-hair/60"
          >
            {copied ? <Check size={14} className="text-success" /> : <Copy size={14} />}
          </button>
        </div>
        <pre className="overflow-x-auto px-4 py-3.5 font-mono text-[13.5px] leading-6">
          <code>
            <span className="select-none text-primary">$ </span>
            <span className="text-textPrimary">{method.code}</span>
          </code>
        </pre>
      </div>
    </div>
  )
}

'use client'

// Wraps every rehype-pretty-code <figure> in a card with a header bar
// (environment badge + language) and a copy button.
//
// Terminal blocks declare WHERE they run via the code-fence title, e.g.
//   ```bash title="Local terminal (your PC)"
//   ```bash title="Server terminal (SSH)"
// The badge colour keys off that prefix so it's obvious at a glance whether a
// command belongs on your machine or on the VPS.

import {
  Children,
  isValidElement,
  useRef,
  useState,
  type ComponentPropsWithoutRef,
  type ReactNode,
} from 'react'
import { Check, Copy, Laptop, Server } from 'lucide-react'

type Env = 'local' | 'server' | null

function detectEnv(title?: string): Env {
  if (!title) return null
  const t = title.toLowerCase()
  if (t.startsWith('local')) return 'local'
  if (t.startsWith('server') || t.startsWith('remote')) return 'server'
  return null
}

function CopyButton({ getText }: { getText: () => string }) {
  const [copied, setCopied] = useState(false)
  const copy = async () => {
    try {
      await navigator.clipboard.writeText(getText())
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    } catch {
      /* clipboard can be blocked — the snippet stays selectable */
    }
  }
  return (
    <button
      type="button"
      onClick={copy}
      aria-label={copied ? 'Copied' : 'Copy code'}
      className="inline-flex size-7 shrink-0 items-center justify-center rounded-md text-textSecondary ring-1 ring-inset ring-transparent transition-colors hover:bg-white/[0.06] hover:text-textPrimary hover:ring-hair/60"
    >
      {copied ? <Check size={14} className="text-success" /> : <Copy size={14} />}
    </button>
  )
}

export function CodeFigure({ children, ...props }: ComponentPropsWithoutRef<'figure'>) {
  const ref = useRef<HTMLElement>(null)

  // Not a code block (plain <figure> in MDX) — render untouched.
  if (!('data-rehype-pretty-code-figure' in props)) {
    return <figure {...props}>{children}</figure>
  }

  const kids = Children.toArray(children)
  const caption = kids.find((c) => isValidElement(c) && c.type === 'figcaption')
  const rest = kids.filter((c) => c !== caption)

  const title =
    caption && isValidElement(caption)
      ? String((caption.props as { children?: ReactNode }).children ?? '')
      : undefined

  const pre = rest.find((c) => isValidElement(c) && c.type === 'pre')
  const lang =
    pre && isValidElement(pre)
      ? ((pre.props as Record<string, string>)['data-language'] ?? undefined)
      : undefined

  const env = detectEnv(title)
  const getText = () => ref.current?.querySelector('pre')?.textContent ?? ''

  // Label priority: explicit title > language name.
  const label = title ?? (lang && lang !== 'text' ? lang : 'Code')

  return (
    <figure
      ref={ref}
      className="group my-6 overflow-hidden rounded-xl border border-hair/60 bg-surfaceElevated"
      {...props}
    >
      <div className="flex items-center justify-between gap-3 border-b border-hair/60 bg-white/[0.02] px-3 py-2">
        <div className="flex min-w-0 items-center gap-2">
          {env === 'local' && <Laptop size={13} className="shrink-0 text-success" />}
          {env === 'server' && <Server size={13} className="shrink-0 text-warning" />}
          <span
            className={`truncate text-[11px] font-medium tracking-wide ${
              env === 'local'
                ? 'text-success'
                : env === 'server'
                  ? 'text-warning'
                  : 'text-textMuted uppercase'
            }`}
          >
            {label}
          </span>
        </div>
        <div className="flex shrink-0 items-center gap-2">
          {title && lang && lang !== 'text' && (
            <span className="text-[10px] uppercase tracking-wide text-textDisabled">{lang}</span>
          )}
          <CopyButton getText={getText} />
        </div>
      </div>
      {rest}
    </figure>
  )
}

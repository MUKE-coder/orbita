import { Info, Lightbulb, AlertTriangle, ShieldAlert } from 'lucide-react'

type Kind = 'tip' | 'note' | 'warning' | 'danger'

const styles: Record<Kind, { border: string; text: string; icon: typeof Info; label: string }> = {
  tip: { border: 'border-l-primary', text: 'text-primary', icon: Lightbulb, label: 'Tip' },
  note: { border: 'border-l-info', text: 'text-info', icon: Info, label: 'Note' },
  warning: { border: 'border-l-warning', text: 'text-warning', icon: AlertTriangle, label: 'Warning' },
  danger: { border: 'border-l-error', text: 'text-error', icon: ShieldAlert, label: 'Careful' },
}

export function Callout({
  type = 'note',
  title,
  children,
}: {
  type?: Kind
  title?: string
  children: React.ReactNode
}) {
  const s = styles[type]
  const Icon = s.icon
  return (
    <div className={`my-5 rounded-xl border border-hair/60 ${s.border} border-l-[3px] bg-gradient-to-b from-white/[0.03] to-white/[0.01] p-4`}>
      <div className={`mb-1 flex items-center gap-2 text-sm font-semibold ${s.text}`}>
        <Icon size={15} />
        {title ?? s.label}
      </div>
      <div className="text-sm leading-6 text-textSecondary [&>p]:my-1.5 [&>p:first-child]:mt-0 [&>p:last-child]:mb-0">
        {children}
      </div>
    </div>
  )
}

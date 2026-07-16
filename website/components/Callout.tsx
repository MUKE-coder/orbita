import { Info, Lightbulb, AlertTriangle, ShieldAlert } from 'lucide-react'

type Kind = 'tip' | 'note' | 'warning' | 'danger'

const styles: Record<Kind, { border: string; text: string; icon: typeof Info; label: string }> = {
  tip: { border: 'border-l-cyan', text: 'text-cyan', icon: Lightbulb, label: 'Tip' },
  note: { border: 'border-l-indigo', text: 'text-indigo', icon: Info, label: 'Note' },
  warning: { border: 'border-l-[#F59E0B]', text: 'text-[#F59E0B]', icon: AlertTriangle, label: 'Warning' },
  danger: { border: 'border-l-[#EF4444]', text: 'text-[#EF4444]', icon: ShieldAlert, label: 'Careful' },
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
    <div className={`my-5 rounded-lg border border-hair ${s.border} border-l-[3px] bg-elev1 p-4`}>
      <div className={`mb-1 flex items-center gap-2 text-sm font-semibold ${s.text}`}>
        <Icon size={15} />
        {title ?? s.label}
      </div>
      <div className="text-sm leading-6 text-muted [&>p]:my-1.5 [&>p:first-child]:mt-0 [&>p:last-child]:mb-0">
        {children}
      </div>
    </div>
  )
}

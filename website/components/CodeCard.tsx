// A static, styled terminal/code card for the landing page (no highlighting
// dependency — deliberately simple and dependency-light).
export function CodeCard({
  title,
  lines,
}: {
  title: string
  lines: { text: string; tone?: 'cmd' | 'ok' | 'muted' | 'accent' }[]
}) {
  const tone = (t?: string) =>
    t === 'ok'
      ? 'text-[#22C55E]'
      : t === 'accent'
        ? 'text-cyan'
        : t === 'muted'
          ? 'text-faint'
          : 'text-ink'
  return (
    <div className="overflow-hidden rounded-xl border border-hair bg-[#0d1017]">
      <div className="flex items-center gap-2 border-b border-hair/70 px-4 py-2.5">
        <span className="h-2.5 w-2.5 rounded-full bg-[#EF4444]/70" />
        <span className="h-2.5 w-2.5 rounded-full bg-[#F59E0B]/70" />
        <span className="h-2.5 w-2.5 rounded-full bg-[#22C55E]/70" />
        <span className="ml-2 font-mono text-[11px] text-faint">{title}</span>
      </div>
      <pre className="overflow-x-auto p-4 font-mono text-[13px] leading-6">
        {lines.map((l, i) => (
          <div key={i} className={tone(l.tone)}>
            {l.text || ' '}
          </div>
        ))}
      </pre>
    </div>
  )
}

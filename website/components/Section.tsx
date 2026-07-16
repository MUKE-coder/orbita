// Shared shell for landing-page sections — consistent vertical rhythm and a
// hairline top border. Mirrors the animateicons Section component.
export function Section({
  children,
  noBorder = false,
  id,
}: {
  children: React.ReactNode
  noBorder?: boolean
  id?: string
}) {
  return (
    <section
      id={id}
      className={`relative py-18 lg:py-24 ${noBorder ? '' : 'border-t border-divider/50'}`}
    >
      <div className="mx-auto max-w-6xl px-4">{children}</div>
    </section>
  )
}

// Centered title + subtitle pair for section headers.
export function SectionHeader({
  title,
  subtitle,
  spacing = 'default',
}: {
  title: string
  subtitle?: string
  spacing?: 'default' | 'tight'
}) {
  return (
    <div className={`${spacing === 'tight' ? 'mb-14' : 'mb-16'} text-center`}>
      <h2 className="text-2xl font-semibold text-white sm:text-3xl">{title}</h2>
      {subtitle ? <p className="mt-3 text-sm text-textSecondary">{subtitle}</p> : null}
    </div>
  )
}

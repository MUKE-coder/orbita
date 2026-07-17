'use client'

import { useState, Children, isValidElement } from 'react'

// Code-group tabs. Usage in MDX:
//   <Tabs labels={["One-line", "Manual"]}>
//     <TabItem> ...code block... </TabItem>
//     <TabItem> ...code block... </TabItem>
//   </Tabs>

export function Tabs({ labels, children }: { labels: string[]; children: React.ReactNode }) {
  const [active, setActive] = useState(0)
  const panels = Children.toArray(children).filter(isValidElement)
  return (
    <div className="my-5 overflow-hidden rounded-xl border border-hair/60 bg-gradient-to-b from-white/[0.03] to-white/[0.01]">
      <div className="flex gap-1 border-b border-hair/60 px-2 pt-2">
        {labels.map((label, i) => (
          <button
            key={label}
            onClick={() => setActive(i)}
            className={`rounded-t-md px-3 py-1.5 text-[13px] transition ${
              i === active ? 'bg-surfaceElevated text-primary' : 'text-textSecondary hover:text-textPrimary'
            }`}
          >
            {label}
          </button>
        ))}
      </div>
      {/* The code figure brings its own card chrome — flatten it inside a tab. */}
      <div className="[&_figure]:my-0 [&_figure]:rounded-none [&_figure]:border-0 [&_figure]:bg-transparent">
        {panels[active]}
      </div>
    </div>
  )
}

export function TabItem({ children }: { children: React.ReactNode }) {
  return <>{children}</>
}

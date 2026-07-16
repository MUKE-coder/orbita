'use client'

import { useState, Children, isValidElement } from 'react'

// Simple code-group tabs. Usage in MDX:
//   <Tabs labels={["One-line", "Manual"]}>
//     <TabItem> ...code block... </TabItem>
//     <TabItem> ...code block... </TabItem>
//   </Tabs>

export function Tabs({ labels, children }: { labels: string[]; children: React.ReactNode }) {
  const [active, setActive] = useState(0)
  const panels = Children.toArray(children).filter(isValidElement)
  return (
    <div className="my-5 overflow-hidden rounded-[10px] border border-hair">
      <div className="flex gap-1 border-b border-hair bg-elev1 px-2 pt-2">
        {labels.map((label, i) => (
          <button
            key={label}
            onClick={() => setActive(i)}
            className={`rounded-t-md px-3 py-1.5 text-[13px] transition ${
              i === active
                ? 'bg-[#0d1017] text-ink'
                : 'text-muted hover:text-ink'
            }`}
          >
            {label}
          </button>
        ))}
      </div>
      <div className="[&_pre]:my-0 [&_pre]:rounded-none [&_pre]:border-0">{panels[active]}</div>
    </div>
  )
}

export function TabItem({ children }: { children: React.ReactNode }) {
  return <>{children}</>
}

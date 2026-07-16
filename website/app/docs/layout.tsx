import { Sidebar } from '@/components/Sidebar'
import { PrevNext } from '@/components/PrevNext'

export default function DocsLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="mx-auto max-w-6xl px-5">
      <div className="grid gap-10 py-10 lg:grid-cols-[220px_minmax(0,1fr)]">
        <aside className="hidden lg:block">
          <div className="sticky top-20">
            <Sidebar />
          </div>
        </aside>

        <article className="min-w-0">
          <div className="prose-orbita max-w-content">{children}</div>
          <div className="max-w-content">
            <PrevNext />
          </div>
        </article>
      </div>
    </div>
  )
}

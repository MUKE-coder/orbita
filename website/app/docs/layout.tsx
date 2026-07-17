import { Sidebar } from '@/components/Sidebar'
import { PrevNext } from '@/components/PrevNext'
import { Toc } from '@/components/Toc'
import { Breadcrumb } from '@/components/Breadcrumb'

export default function DocsLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="mx-auto max-w-[88rem] px-5 lg:px-8">
      <div className="grid gap-x-10 py-10 lg:grid-cols-[15rem_minmax(0,1fr)] xl:grid-cols-[15rem_minmax(0,1fr)_14rem]">
        {/* Left: section nav */}
        <aside className="hidden lg:block">
          <div className="sticky top-20 max-h-[calc(100dvh-6rem)] overflow-y-auto pb-10 pr-2">
            <Sidebar />
          </div>
        </aside>

        {/* Middle: article */}
        <article className="min-w-0 pb-4">
          <Breadcrumb />
          <div className="prose-orbita">{children}</div>
          <PrevNext />
        </article>

        {/* Right: on-this-page */}
        <aside className="hidden xl:block">
          <div className="sticky top-20 max-h-[calc(100dvh-6rem)] overflow-y-auto pb-10">
            <Toc />
          </div>
        </aside>
      </div>
    </div>
  )
}

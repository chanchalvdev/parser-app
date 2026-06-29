import { TopBar } from './TopBar'
import { Sidebar } from './Sidebar'
import { useUiStore } from '@/stores/uiStore'
import type { ReactNode } from 'react'

export const AppShell = ({ children }: { children: ReactNode }) => {
  const collapsed = useUiStore((state) => state.sidebarCollapsed)

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100">
      <TopBar />
      <div className="mx-auto flex min-h-[calc(100vh-3.5rem)] w-full max-w-screen-2xl gap-4 px-3 py-4 sm:px-5 lg:px-6">
        <Sidebar />
        <main className={`w-full transition-all ${collapsed ? 'ml-0' : ''}`}>{children}</main>
      </div>
    </div>
  )
}

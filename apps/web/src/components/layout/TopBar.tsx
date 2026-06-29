import { useUiStore } from '@/stores/uiStore'

export const TopBar = () => {
  const tenantId = useUiStore((state) => state.tenantId)
  const setTenantId = useUiStore((state) => state.setTenantId)
  const toggleSidebar = useUiStore((state) => state.toggleSidebar)

  return (
    <header className="sticky top-0 z-20 border-b border-slate-700/70 bg-slate-900/80 backdrop-blur">
      <div className="mx-auto flex min-h-14 w-full max-w-screen-2xl items-center justify-between gap-4 px-4 sm:px-6">
        <button
          aria-label="toggle sidebar"
          type="button"
          className="rounded-lg border border-slate-600 px-2 py-2 text-slate-200 transition hover:bg-slate-800"
          onClick={() => toggleSidebar()}
        >
          ☰
        </button>
        <div className="text-sm text-slate-200">Enterprise File Ingestion Platform</div>
        <div className="flex items-center gap-3 text-sm">
          <span className="text-slate-300">Tenant</span>
          <input
            value={tenantId}
            className="w-80 rounded-md border border-slate-700 bg-slate-900/80 px-2 py-1 text-xs"
            onChange={(event) => setTenantId(event.target.value)}
          />
        </div>
      </div>
    </header>
  )
}

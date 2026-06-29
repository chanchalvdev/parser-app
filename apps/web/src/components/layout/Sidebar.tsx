import { Link, useLocation } from 'react-router-dom'
import { useUiStore } from '@/stores/uiStore'

type NavItem = {
  path: string
  label: string
}

const items: NavItem[] = [
  { path: '/dashboard', label: 'Dashboard' },
  { path: '/upload', label: 'Upload' },
  { path: '/files', label: 'Files' },
  { path: '/jobs', label: 'Jobs' },
  { path: '/search', label: 'Search' },
  { path: '/admin/settings', label: 'Admin Settings' },
  { path: '/audit-logs', label: 'Audit Logs' },
]

export const Sidebar = () => {
  const location = useLocation()
  const collapsed = useUiStore((state) => state.sidebarCollapsed)

  return (
    <aside
      className={`panel h-screen shrink-0 border-r border-slate-700/70 bg-slate-900/80 p-4 transition-all ${
        collapsed ? 'w-20' : 'w-64'
      }`}
    >
      <div className="mb-4 flex items-center justify-between">
        <div className={`text-base font-bold tracking-tight text-white ${collapsed ? 'hidden' : ''}`}>Workspace</div>
      </div>
      <nav className="space-y-2">
        {items.map((item) => {
          const active = location.pathname === item.path
          return (
            <Link
              key={item.path}
              to={item.path}
              className={`block rounded-lg px-3 py-2 text-sm ${
                active ? 'bg-blue-500/30 text-blue-100' : 'text-slate-200 hover:bg-slate-800/60'
              }`}
            >
              {collapsed ? item.label[0] : item.label}
            </Link>
          )
        })}
      </nav>
    </aside>
  )
}

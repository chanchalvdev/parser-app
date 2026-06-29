import { NavLink } from 'react-router-dom'
import {
  LayoutDashboard,
  UploadCloud,
  ListChecks,
  Search,
  FolderTree,
  Settings as SettingsIcon,
  HardDrive,
} from './Icon'

export interface SidebarProps {
  open: boolean
  onNavigate: () => void
}

interface NavConfig {
  to: string
  label: string
  Icon: typeof LayoutDashboard
  end?: boolean
}

const NAV_ITEMS: NavConfig[] = [
  { to: '/',         label: 'Dashboard', Icon: LayoutDashboard, end: true },
  { to: '/upload',   label: 'Upload',    Icon: UploadCloud },
  { to: '/jobs',     label: 'Jobs',      Icon: ListChecks },
  { to: '/search',   label: 'Search',    Icon: Search },
  { to: '/files',    label: 'Files',     Icon: FolderTree },
  { to: '/settings', label: 'Settings',  Icon: SettingsIcon },
]

export function Sidebar({ open, onNavigate }: SidebarProps) {
  return (
    <aside
      className={`sidebar ${open ? 'open' : ''}`.trim()}
      aria-label="Primary navigation"
    >
      <div className="sidebar-brand">
        <span className="brand-mark" aria-hidden="true">FI</span>
        <span className="brand-text">
          <span className="brand-wordmark">File Ingestion</span>
          <span className="brand-tagline">Local enterprise ingestion lab</span>
        </span>
      </div>

      <nav className="sidebar-nav">
        {NAV_ITEMS.map(({ to, label, Icon, end }) => (
          <NavLink
            key={to}
            to={to}
            end={end}
            className={({ isActive }) =>
              `sidebar-item ${isActive ? 'active' : ''}`.trim()
            }
            onClick={onNavigate}
          >
            <Icon size={18} aria-hidden="true" />
            <span>{label}</span>
          </NavLink>
        ))}
      </nav>

      <div className="sidebar-footer">
        <span className="sidebar-footer-chip">
          <span className="sidebar-footer-chip-dot" aria-hidden="true" />
          <HardDrive size={11} aria-hidden="true" />
          <span>v0.1.0 · local</span>
        </span>
      </div>
    </aside>
  )
}

import { useEffect, useMemo, useState } from 'react'
import { useLocation } from 'react-router-dom'
import {
  Menu,
  Sun,
  Moon,
  Info,
  Search as SearchIcon,
} from './Icon'
import { Breadcrumbs, type BreadcrumbItem } from './Breadcrumbs'

export interface TopBarProps {
  onOpenSidebar: () => void
}

const ROUTE_TITLES: Record<string, string> = {
  '/': 'Dashboard',
  '/upload': 'Upload',
  '/jobs': 'Jobs',
  '/search': 'Search',
  '/files': 'Files',
  '/settings': 'Settings',
  '/help': 'Help',
}

const ROUTE_LABELS: Record<string, string> = {
  '/': 'Dashboard',
  '/upload': 'Upload',
  '/jobs': 'Jobs',
  '/search': 'Search',
  '/files': 'Files',
  '/settings': 'Settings',
}

function humanize(segment: string): string {
  return segment
    .replace(/[-_]+/g, ' ')
    .replace(/\b\w/g, (c) => c.toUpperCase())
}

function resolveTitle(pathname: string): string {
  if (ROUTE_TITLES[pathname]) return ROUTE_TITLES[pathname]
  const top = '/' + (pathname.split('/').filter(Boolean)[0] ?? '')
  return ROUTE_TITLES[top] ?? humanize(pathname.split('/').filter(Boolean).pop() ?? 'Page')
}

function buildBreadcrumbs(pathname: string): BreadcrumbItem[] {
  const segments = pathname.split('/').filter(Boolean)
  if (segments.length === 0) {
    return [{ label: 'Dashboard' }]
  }
  const items: BreadcrumbItem[] = [{ label: 'Home', to: '/' }]
  let acc = ''
  segments.forEach((segment, index) => {
    acc += `/${segment}`
    const label = ROUTE_LABELS[acc] ?? humanize(segment)
    const isLast = index === segments.length - 1
    items.push(isLast ? { label } : { label, to: acc })
  })
  return items
}

type Theme = 'light' | 'dark'

function getInitialTheme(): Theme {
  if (typeof document === 'undefined') return 'dark'
  const attr = document.documentElement.getAttribute('data-theme')
  if (attr === 'light' || attr === 'dark') return attr
  return 'dark'
}

export function TopBar({ onOpenSidebar }: TopBarProps) {
  const location = useLocation()
  const title = useMemo(() => resolveTitle(location.pathname), [location.pathname])
  const crumbs = useMemo(() => buildBreadcrumbs(location.pathname), [location.pathname])
  const [theme, setTheme] = useState<Theme>(getInitialTheme)

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme)
    try {
      window.localStorage.setItem('theme', theme)
    } catch {
      // Storage unavailable; ignore.
    }
  }, [theme])

  const toggleTheme = () => setTheme((prev) => (prev === 'light' ? 'dark' : 'light'))

  return (
    <header className="topbar" role="banner">
      <div className="topbar-left">
        <button
          type="button"
          className="topbar-icon-btn topbar-hamburger"
          onClick={onOpenSidebar}
          aria-label="Open navigation menu"
        >
          <Menu size={20} />
        </button>

        <h1 className="topbar-title">{title}</h1>

        <span className="topbar-divider" aria-hidden="true" />

        <Breadcrumbs items={crumbs} />
      </div>

      <div className="topbar-right">
        <a className="topbar-search-hint" href="/search">
          <SearchIcon size={14} aria-hidden="true" />
          <span>Search</span>
          <kbd>⌘K</kbd>
        </a>

        <button
          type="button"
          className="topbar-icon-btn"
          onClick={toggleTheme}
          aria-label={theme === 'light' ? 'Switch to dark theme' : 'Switch to light theme'}
          title={theme === 'light' ? 'Dark theme' : 'Light theme'}
        >
          {theme === 'light' ? <Moon size={18} /> : <Sun size={18} />}
        </button>

        <a
          className="topbar-icon-btn"
          href="#help"
          aria-label="Help"
          title="Help"
        >
          <Info size={18} />
        </a>
      </div>
    </header>
  )
}

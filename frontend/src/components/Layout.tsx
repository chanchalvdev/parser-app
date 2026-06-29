import { useEffect, useState } from 'react'
import { Outlet, useLocation } from 'react-router-dom'
import { Sidebar } from './Sidebar'
import { TopBar } from './TopBar'

export default function Layout() {
  const location = useLocation()
  const [drawerOpen, setDrawerOpen] = useState(false)

  useEffect(() => {
    setDrawerOpen(false)
  }, [location.pathname])

  useEffect(() => {
    if (!drawerOpen) return
    const previousOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    const handleKey = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setDrawerOpen(false)
    }
    document.addEventListener('keydown', handleKey)
    return () => {
      document.body.style.overflow = previousOverflow
      document.removeEventListener('keydown', handleKey)
    }
  }, [drawerOpen])

  const closeDrawer = () => setDrawerOpen(false)
  const openDrawer = () => setDrawerOpen(true)

  return (
    <>
      <a className="skip-link" href="#main-content">
        Skip to content
      </a>

      <div className="app-shell">
        <Sidebar open={drawerOpen} onNavigate={closeDrawer} />

        {drawerOpen && (
          <div
            className="sidebar-backdrop"
            onClick={closeDrawer}
            role="presentation"
          />
        )}

        <div className="app-main">
          <TopBar onOpenSidebar={openDrawer} />
          <main className="content fade-in" id="main-content">
            <Outlet />
          </main>
        </div>
      </div>
    </>
  )
}

import type { ReactNode } from 'react'

export const Card = ({ title, subtitle, children }: { title?: string; subtitle?: string; children: ReactNode }) => {
  return (
    <section className="panel p-5">
      {(title || subtitle) && (
        <header className="mb-4 border-b border-slate-700/50 pb-3">
          {title ? <h3 className="panel-title text-white">{title}</h3> : null}
          {subtitle ? <p className="text-xs text-slate-300">{subtitle}</p> : null}
        </header>
      )}
      {children}
    </section>
  )
}

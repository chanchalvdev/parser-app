import type { HTMLAttributes, ReactNode } from 'react'

export type CardPadding = 'sm' | 'md' | 'lg' | 'none'

export interface CardProps extends HTMLAttributes<HTMLElement> {
  title?: ReactNode
  subtitle?: ReactNode
  actions?: ReactNode
  padding?: CardPadding
  children?: ReactNode
}

const padClass: Record<CardPadding, string> = {
  sm: 'card-pad-sm',
  md: 'card-pad-md',
  lg: 'card-pad-lg',
  none: 'card-pad-none',
}

export function Card({
  title,
  subtitle,
  actions,
  padding = 'md',
  children,
  className,
  ...rest
}: CardProps) {
  const hasHeader = title !== undefined || subtitle !== undefined || actions !== undefined

  const classes = ['card', className ?? ''].filter(Boolean).join(' ')

  return (
    <section className={classes} {...rest}>
      {hasHeader && (
        <header className="card-header">
          <div className="card-title-wrap">
            {title !== undefined && (
              typeof title === 'string' ? <h3>{title}</h3> : title
            )}
            {subtitle !== undefined && (
              <span className="card-subtitle">{subtitle}</span>
            )}
          </div>
          {actions !== undefined && (
            <div className="card-actions">{actions}</div>
          )}
        </header>
      )}
      {children !== undefined && (
        <div className={`card-body ${padClass[padding]}`}>{children}</div>
      )}
    </section>
  )
}

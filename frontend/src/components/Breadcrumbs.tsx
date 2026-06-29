import { Link } from 'react-router-dom'
import { ChevronRight } from './Icon'

export interface BreadcrumbItem {
  label: string
  to?: string
}

export interface BreadcrumbsProps {
  items: BreadcrumbItem[]
  className?: string
}

export function Breadcrumbs({ items, className }: BreadcrumbsProps) {
  if (!items.length) return null
  return (
    <nav aria-label="Breadcrumb" className={`breadcrumb ${className ?? ''}`.trim()}>
      {items.map((item, index) => {
        const isLast = index === items.length - 1
        const content = isLast || !item.to ? (
          <span className="breadcrumb-item current" aria-current={isLast ? 'page' : undefined}>
            {item.label}
          </span>
        ) : (
          <Link to={item.to} className="breadcrumb-item">
            {item.label}
          </Link>
        )
        return (
          <span key={`${item.label}-${index}`} className="row" style={{ gap: 6 }}>
            {content}
            {!isLast && (
              <span className="breadcrumb-separator" aria-hidden="true">
                <ChevronRight size={14} />
              </span>
            )}
          </span>
        )
      })}
    </nav>
  )
}

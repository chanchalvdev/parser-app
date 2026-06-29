import type { ReactNode } from 'react'
import type { LucideIcon } from 'lucide-react'

export interface EmptyStateProps {
  icon: LucideIcon
  title: ReactNode
  description?: ReactNode
  action?: ReactNode
  className?: string
}

export function EmptyState({ icon: Icon, title, description, action, className }: EmptyStateProps) {
  return (
    <div className={`empty-state ${className ?? ''}`.trim()} role="status">
      <div className="empty-state-icon" aria-hidden="true">
        <Icon size={28} />
      </div>
      <div className="empty-state-title">{title}</div>
      {description !== undefined && (
        <div className="empty-state-description">{description}</div>
      )}
      {action !== undefined && (
        <div className="empty-state-action">{action}</div>
      )}
    </div>
  )
}

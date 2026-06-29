import type { HTMLAttributes, ReactNode } from 'react'

export type BadgeVariant =
  | 'neutral'
  | 'info'
  | 'success'
  | 'warning'
  | 'danger'
  | 'accent'

export type BadgeSize = 'sm' | 'md'

export interface BadgeProps extends HTMLAttributes<HTMLSpanElement> {
  variant?: BadgeVariant
  size?: BadgeSize
  children: ReactNode
}

const variantClass: Record<BadgeVariant, string> = {
  neutral: 'badge-neutral',
  info: 'badge-info',
  success: 'badge-success',
  warning: 'badge-warning',
  danger: 'badge-danger',
  accent: 'badge-accent',
}

export function Badge({
  variant = 'neutral',
  size = 'sm',
  className,
  children,
  ...rest
}: BadgeProps) {
  const classes = [
    'badge',
    variantClass[variant],
    size === 'md' ? 'badge-md' : '',
    className ?? '',
  ]
    .filter(Boolean)
    .join(' ')

  return (
    <span className={classes} {...rest}>
      {children}
    </span>
  )
}

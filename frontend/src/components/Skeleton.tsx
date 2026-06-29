import { useMemo } from 'react'

export interface SkeletonProps {
  width?: number | string
  height?: number | string
  circle?: boolean
  count?: number
  className?: string
}

function toCssSize(value: number | string | undefined): string | undefined {
  if (value === undefined) return undefined
  return typeof value === 'number' ? `${value}px` : value
}

function SkeletonUnit({ width, height, circle, className }: Omit<SkeletonProps, 'count'>) {
  const style: React.CSSProperties = {
    width: toCssSize(width) ?? '100%',
    height: toCssSize(height) ?? '14px',
  }
  const classes = [
    'skeleton',
    circle ? 'skeleton-circle' : '',
    className ?? '',
  ]
    .filter(Boolean)
    .join(' ')

  return <span className={classes} style={style} aria-hidden="true" />
}

export function Skeleton({ count = 1, ...rest }: SkeletonProps) {
  const items = useMemo(() => Array.from({ length: count }, (_, i) => i), [count])
  if (count <= 1) return <SkeletonUnit {...rest} />
  return (
    <span className="stack-sm" aria-hidden="true">
      {items.map((i) => (
        <SkeletonUnit key={i} {...rest} />
      ))}
    </span>
  )
}

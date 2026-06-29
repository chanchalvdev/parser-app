export type SpinnerSize = 'sm' | 'md' | 'lg'

export interface SpinnerProps {
  size?: SpinnerSize
  label?: string
  className?: string
}

export function Spinner({ size = 'md', label, className }: SpinnerProps) {
  const sizeClass = size === 'sm' ? 'spinner-sm' : size === 'lg' ? 'spinner-lg' : 'spinner-md'
  const spinner = (
    <span
      className={`spinner ${sizeClass} ${className ?? ''}`.trim()}
      role="status"
      aria-hidden="true"
    />
  )

  if (!label) return spinner

  return (
    <span className="spinner-label">
      {spinner}
      <span>{label}</span>
    </span>
  )
}

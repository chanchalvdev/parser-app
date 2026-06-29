import type { ReactNode } from 'react'
import { AlertCircle, CheckCircle2, Info, X, AlertTriangle } from './Icon'

export type ToastVariant = 'success' | 'error' | 'info' | 'warning'

export interface ToastData {
  id: string
  variant: ToastVariant
  message: ReactNode
}

export interface ToastProps {
  toast: ToastData
  onDismiss: (id: string) => void
}

const variantClass: Record<ToastVariant, string> = {
  success: 'toast-success',
  error: 'toast-error',
  info: 'toast-info',
  warning: 'toast-warning',
}

function VariantIcon({ variant }: { variant: ToastVariant }) {
  const size = 18
  switch (variant) {
    case 'success':
      return <CheckCircle2 size={size} aria-hidden="true" />
    case 'error':
      return <AlertCircle size={size} aria-hidden="true" />
    case 'warning':
      return <AlertTriangle size={size} aria-hidden="true" />
    case 'info':
    default:
      return <Info size={size} aria-hidden="true" />
  }
}

export function Toast({ toast, onDismiss }: ToastProps) {
  const isError = toast.variant === 'error'
  return (
    <div
      className={`toast ${variantClass[toast.variant]}`}
      role={isError ? 'alert' : 'status'}
      aria-live={isError ? 'assertive' : 'polite'}
      aria-atomic="true"
    >
      <span className="toast-icon">
        <VariantIcon variant={toast.variant} />
      </span>
      <span className="toast-message">{toast.message}</span>
      <button
        type="button"
        className="toast-close"
        onClick={() => onDismiss(toast.id)}
        aria-label="Dismiss notification"
      >
        <X size={16} />
      </button>
    </div>
  )
}

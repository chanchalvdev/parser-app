import { useEffect, useRef, type MouseEvent, type ReactNode } from 'react'
import { createPortal } from 'react-dom'
import { X } from './Icon'

export type ModalSize = 'sm' | 'md' | 'lg' | 'xl'

export interface ModalProps {
  open: boolean
  onClose: () => void
  title?: ReactNode
  children?: ReactNode
  footer?: ReactNode
  size?: ModalSize
  closeOnBackdrop?: boolean
  closeOnEscape?: boolean
  ariaLabelledBy?: string
}

const sizeClass: Record<ModalSize, string> = {
  sm: 'modal-sm',
  md: 'modal-md',
  lg: 'modal-lg',
  xl: 'modal-xl',
}

const FOCUSABLE_SELECTOR = [
  'a[href]',
  'button:not([disabled])',
  'input:not([disabled])',
  'select:not([disabled])',
  'textarea:not([disabled])',
  '[tabindex]:not([tabindex="-1"])',
].join(',')

export function Modal({
  open,
  onClose,
  title,
  children,
  footer,
  size = 'md',
  closeOnBackdrop = true,
  closeOnEscape = true,
  ariaLabelledBy,
}: ModalProps) {
  const modalRef = useRef<HTMLDivElement | null>(null)
  const previouslyFocusedRef = useRef<HTMLElement | null>(null)

  useEffect(() => {
    if (!open) return

    const previouslyFocused = document.activeElement as HTMLElement | null
    previouslyFocusedRef.current = previouslyFocused

    const originalOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'

    const focusFirst = () => {
      const modal = modalRef.current
      if (!modal) return
      const focusables = modal.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR)
      const target = focusables[0] ?? modal
      target.focus()
    }

    const handleKey = (event: KeyboardEvent) => {
      if (event.key === 'Escape' && closeOnEscape) {
        event.stopPropagation()
        onClose()
        return
      }
      if (event.key !== 'Tab') return
      const modal = modalRef.current
      if (!modal) return
      const focusables = Array.from(
        modal.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR),
      ).filter((el) => !el.hasAttribute('disabled') && el.offsetParent !== null)
      if (focusables.length === 0) {
        event.preventDefault()
        modal.focus()
        return
      }
      const first = focusables[0]
      const last = focusables[focusables.length - 1]
      const active = document.activeElement as HTMLElement | null
      if (event.shiftKey) {
        if (active === first || !modal.contains(active)) {
          event.preventDefault()
          last.focus()
        }
      } else {
        if (active === last) {
          event.preventDefault()
          first.focus()
        }
      }
    }

    document.addEventListener('keydown', handleKey)
    const raf = window.requestAnimationFrame(focusFirst)

    return () => {
      window.cancelAnimationFrame(raf)
      document.removeEventListener('keydown', handleKey)
      document.body.style.overflow = originalOverflow
      const toFocus = previouslyFocusedRef.current
      if (toFocus && typeof toFocus.focus === 'function') {
        toFocus.focus()
      }
    }
  }, [open, onClose, closeOnEscape])

  if (!open) return null

  const handleBackdropMouseDown = (event: MouseEvent<HTMLDivElement>) => {
    if (!closeOnBackdrop) return
    if (event.target === event.currentTarget) onClose()
  }

  const titleId = ariaLabelledBy ?? 'modal-title'

  const modalNode = (
    <div
      className="modal-backdrop"
      onMouseDown={handleBackdropMouseDown}
      role="presentation"
    >
      <div
        ref={modalRef}
        className={`modal ${sizeClass[size]}`}
        role="dialog"
        aria-modal="true"
        aria-labelledby={title ? titleId : undefined}
        tabIndex={-1}
      >
        {title !== undefined && (
          <header className="modal-header">
            <h2 className="modal-title" id={titleId}>
              {title}
            </h2>
            <button
              type="button"
              className="topbar-icon-btn"
              onClick={onClose}
              aria-label="Close dialog"
            >
              <X size={18} />
            </button>
          </header>
        )}
        <div className="modal-body">{children}</div>
        {footer !== undefined && <footer className="modal-footer">{footer}</footer>}
      </div>
    </div>
  )

  return createPortal(modalNode, document.body)
}

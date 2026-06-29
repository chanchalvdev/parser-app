import { useEffect, useRef, useState } from 'react'
import { Check, Copy } from './Icon'

export interface CopyButtonProps {
  text: string
  label?: string
  copiedLabel?: string
  className?: string
  ariaLabel?: string
}

const COPIED_DURATION_MS = 1500

export function CopyButton({
  text,
  label = 'Copy',
  copiedLabel = 'Copied',
  className,
  ariaLabel,
}: CopyButtonProps) {
  const [copied, setCopied] = useState(false)
  const timerRef = useRef<number | null>(null)

  useEffect(() => {
    return () => {
      if (timerRef.current !== null) window.clearTimeout(timerRef.current)
    }
  }, [])

  const onClick = async () => {
    try {
      if (navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(text)
      } else {
        const textarea = document.createElement('textarea')
        textarea.value = text
        textarea.style.position = 'fixed'
        textarea.style.opacity = '0'
        document.body.appendChild(textarea)
        textarea.focus()
        textarea.select()
        document.execCommand('copy')
        document.body.removeChild(textarea)
      }
      setCopied(true)
      if (timerRef.current !== null) window.clearTimeout(timerRef.current)
      timerRef.current = window.setTimeout(() => setCopied(false), COPIED_DURATION_MS)
    } catch {
      // Clipboard unavailable; silently no-op.
    }
  }

  return (
    <button
      type="button"
      onClick={onClick}
      className={`copy-btn ${copied ? 'copied' : ''} ${className ?? ''}`.trim()}
      aria-label={ariaLabel ?? label}
      title={copied ? copiedLabel : label}
    >
      {copied ? <Check size={14} /> : <Copy size={14} />}
    </button>
  )
}

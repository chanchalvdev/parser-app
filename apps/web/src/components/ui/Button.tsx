import { buildClassName } from '@/utils/query'
import type { ButtonHTMLAttributes } from 'react'

export type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  tone?: 'primary' | 'ghost' | 'danger'
}

const toneClasses: Record<NonNullable<ButtonProps['tone']>, string> = {
  primary:
    'bg-blue-500 text-white hover:bg-blue-400 focus:ring-blue-300 disabled:opacity-50 disabled:cursor-not-allowed',
  ghost:
    'border border-slate-500/80 text-slate-100 hover:bg-slate-700/40 focus:ring-slate-400 disabled:opacity-50 disabled:cursor-not-allowed',
  danger:
    'bg-rose-600 text-white hover:bg-rose-500 focus:ring-rose-300 disabled:opacity-50 disabled:cursor-not-allowed',
}

export const Button = ({ tone = 'primary', className, ...props }: ButtonProps) => {
  return (
    <button
      {...props}
      className={buildClassName(
        'inline-flex items-center justify-center rounded-lg px-4 py-2 text-sm font-semibold transition focus:outline-none focus:ring-2 focus:ring-offset-1 disabled:opacity-60',
        toneClasses[tone],
        className,
      )}
    />
  )
}

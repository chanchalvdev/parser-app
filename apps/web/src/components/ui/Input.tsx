import type { InputHTMLAttributes } from 'react'

export const Input = (props: InputHTMLAttributes<HTMLInputElement>) => {
  return <input className="w-full rounded-lg border border-slate-700 bg-slate-900/80 px-3 py-2 text-sm outline-none transition focus:border-blue-400 focus:ring-2 focus:ring-blue-500/40" {...props} />
}

type BadgeTone = 'gray' | 'blue' | 'green' | 'amber' | 'red'

const toneStyles: Record<BadgeTone, string> = {
  gray: 'bg-slate-700/70 text-slate-100 border-slate-500',
  blue: 'bg-blue-900/70 text-blue-100 border-blue-500',
  green: 'bg-emerald-900/70 text-emerald-100 border-emerald-500',
  amber: 'bg-amber-900/70 text-amber-100 border-amber-500',
  red: 'bg-rose-900/70 text-rose-100 border-rose-500',
}

export const Badge = ({ children, tone = 'gray' }: { children: string; tone?: BadgeTone }) => (
  <span className={`chip ${toneStyles[tone]}`}>{children}</span>
)

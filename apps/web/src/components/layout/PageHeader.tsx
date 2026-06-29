export const PageHeader = ({ title, subtitle }: { title: string; subtitle?: string }) => {
  return (
    <div className="mb-4">
      <h1 className="text-2xl font-semibold tracking-wide text-slate-50">{title}</h1>
      {subtitle ? <p className="text-sm text-slate-300">{subtitle}</p> : null}
    </div>
  )
}

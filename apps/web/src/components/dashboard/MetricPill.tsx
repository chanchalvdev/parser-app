export const MetricPill = ({ label, value }: { label: string; value: number | string }) => {
  return (
    <div className="rounded-lg border border-slate-700/80 p-3">
      <div className="text-xs text-slate-300">{label}</div>
      <div className="text-xl font-semibold">{value}</div>
    </div>
  )
}

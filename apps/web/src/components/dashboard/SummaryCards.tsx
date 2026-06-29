import { Card } from '@/components/ui/Card'
import type { DashboardSummary } from '@/types/dashboard'

type SummaryCardsProps = {
  summary?: DashboardSummary
  isLoading?: boolean
}

type SummaryItem = {
  label: string
  value: string | number
}

const formatSummaryValue = (value: number | undefined) => {
  if (value === undefined || Number.isNaN(value)) return '0'
  return value.toLocaleString()
}

const Stat = ({ label, value }: SummaryItem) => (
  <Card>
    <p className="text-xs text-slate-300">{label}</p>
    <p className="mt-2 text-2xl font-bold text-white">{value}</p>
  </Card>
)

export const SummaryCards = ({ summary, isLoading }: SummaryCardsProps) => {
  if (isLoading) {
    return <Card>Loading summary metrics…</Card>
  }

  if (!summary) {
    return <Card>No summary data yet.</Card>
  }

  const metrics: SummaryItem[] = [
    { label: 'Total uploads', value: formatSummaryValue(summary.total_uploads) },
    { label: 'Total files', value: formatSummaryValue(summary.total_files) },
    { label: 'Total extracted files', value: formatSummaryValue(summary.total_extracted_files) },
    { label: 'Total parsed records', value: formatSummaryValue(summary.total_parsed_records) },
    { label: 'Completed jobs', value: formatSummaryValue(summary.completed_jobs) },
    { label: 'Failed jobs', value: formatSummaryValue(summary.failed_jobs) },
    { label: 'Password required', value: formatSummaryValue(summary.password_required_files) },
    { label: 'Quarantined', value: formatSummaryValue(summary.quarantined_files) },
  ]

  return (
    <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
      {metrics.map((metric) => (
        <Stat key={metric.label} {...metric} />
      ))}
    </div>
  )
}

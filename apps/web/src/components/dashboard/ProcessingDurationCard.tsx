import { Card } from '@/components/ui/Card'
import type { DashboardProcessingDuration } from '@/types/dashboard'

type ProcessingDurationCardProps = {
  data?: DashboardProcessingDuration
  isLoading?: boolean
}

const formatMs = (value: number | undefined) => {
  if (!value || Number.isNaN(value)) return '0.0'
  return value.toFixed(1)
}

export const ProcessingDurationCard = ({ data, isLoading }: ProcessingDurationCardProps) => {
  if (isLoading) {
    return (
      <Card title="Processing duration" subtitle="Loading processing duration…">
        <p className="text-sm text-slate-300">Loading timing metrics…</p>
      </Card>
    )
  }

  if (!data) {
    return (
      <Card title="Processing duration" subtitle="No duration samples yet.">
        <p className="text-sm text-slate-300">No completed jobs were measured.</p>
      </Card>
    )
  }

  return (
    <Card title="Processing duration" subtitle={`${data.completed_jobs} completed jobs`}>
      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-6">
        <div>
          <div className="text-xs text-slate-300">Average</div>
          <div className="text-lg font-semibold text-white">{formatMs(data.average_seconds)}s</div>
        </div>
        <div>
          <div className="text-xs text-slate-300">Median</div>
          <div className="text-lg font-semibold text-white">{formatMs(data.median_seconds)}s</div>
        </div>
        <div>
          <div className="text-xs text-slate-300">P95</div>
          <div className="text-lg font-semibold text-white">{formatMs(data.p95_seconds)}s</div>
        </div>
        <div>
          <div className="text-xs text-slate-300">Min</div>
          <div className="text-lg font-semibold text-white">{formatMs(data.min_seconds)}s</div>
        </div>
        <div>
          <div className="text-xs text-slate-300">Max</div>
          <div className="text-lg font-semibold text-white">{formatMs(data.max_seconds)}s</div>
        </div>
        <div>
          <div className="text-xs text-slate-300">Unit</div>
          <div className="text-lg font-semibold text-white">{data.unit || 'seconds'}</div>
        </div>
      </div>
    </Card>
  )
}

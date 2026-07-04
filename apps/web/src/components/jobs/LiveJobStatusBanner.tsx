import type { IngestionJob } from '@/types/job'
import { Card } from '@/components/ui/Card'
import { JobStatusBadge } from './JobStatusBadge'
import { isJobLive } from '@/hooks/apiHooks'

type LiveJobStatusBannerProps = {
  job: IngestionJob
}

const formatStage = (stage: string | null | undefined): string => {
  if (!stage) return 'Waiting to start'
  return stage
    .replace(/_/g, ' ')
    .replace(/\b\w/g, (letter) => letter.toUpperCase())
}

export const LiveJobStatusBanner = ({ job }: LiveJobStatusBannerProps) => {
  const live = isJobLive(job)
  if (!live) {
    return null
  }

  const progress = typeof job.progress_percent === 'number' ? Math.max(0, Math.min(100, job.progress_percent)) : 0

  return (
    <Card title="Live status" subtitle="Updates automatically while the job is running">
      <div className="flex flex-wrap items-center gap-3">
        <span className="inline-flex items-center gap-2 rounded-full bg-emerald-900/30 px-2 py-1 text-xs text-emerald-200">
          <span className="relative inline-flex h-2 w-2">
            <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-75" />
            <span className="relative inline-flex h-2 w-2 rounded-full bg-emerald-300" />
          </span>
          Live
        </span>
        <JobStatusBadge status={job.status} />
        <span className="text-sm text-slate-200">
          <span className="text-slate-400">Stage:</span> {formatStage(job.current_stage)}
        </span>
        <span className="text-sm text-slate-200">
          <span className="text-slate-400">Progress:</span> {progress}%
        </span>
        {typeof job.retry_count === 'number' && job.retry_count > 0 ? (
          <span className="text-sm text-slate-200">
            <span className="text-slate-400">Retries:</span> {job.retry_count}
          </span>
        ) : null}
      </div>
      <div className="mt-3 h-2 w-full overflow-hidden rounded-full bg-slate-800">
        <div
          className="h-full rounded-full bg-blue-500 transition-all duration-500 ease-out"
          style={{ width: `${progress}%` }}
          aria-valuenow={progress}
          aria-valuemin={0}
          aria-valuemax={100}
          role="progressbar"
        />
      </div>
    </Card>
  )
}

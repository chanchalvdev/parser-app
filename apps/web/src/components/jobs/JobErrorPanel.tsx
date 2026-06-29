import type { IngestionJob } from '@/types/job'

type JobErrorPanelProps = {
  job?: Partial<IngestionJob> | null
}

export const JobErrorPanel = ({ job }: JobErrorPanelProps) => {
  if (!job || (!job.error_code && !job.error_message)) {
    return null
  }

  return (
    <section className="rounded-md border border-rose-600 bg-rose-900/20 p-4">
      <h4 className="text-sm font-semibold text-rose-100">Job error</h4>
      <div className="mt-2 space-y-2 text-sm text-rose-100">
        <div>
          <span className="text-rose-200">Code:</span> {job.error_code || 'unknown'}
        </div>
        <div>
          <span className="text-rose-200">Message:</span> {job.error_message}
        </div>
      </div>
    </section>
  )
}

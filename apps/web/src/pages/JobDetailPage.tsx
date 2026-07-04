import { Link, useParams } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useJob, useJobEvents, useFile, isJobLive } from '@/hooks/apiHooks'
import { PageHeader } from '@/components/layout/PageHeader'
import { Card } from '@/components/ui/Card'
import { parseTime } from '@/utils/query'
import { JobStatusBadge } from '@/components/jobs/JobStatusBadge'
import { JobTimeline } from '@/components/jobs/JobTimeline'
import { RetryJobButton } from '@/components/jobs/RetryJobButton'
import { JobErrorPanel } from '@/components/jobs/JobErrorPanel'
import { LiveJobStatusBanner } from '@/components/jobs/LiveJobStatusBanner'
import { PasswordRequiredModal } from '@/components/files/PasswordRequiredModal'

export const JobDetailPage = () => {
  const { jobId = '' } = useParams<{ jobId: string }>()
  const queryClient = useQueryClient()
  const [isPasswordModalOpen, setPasswordModalOpen] = useState(false)

  const jobQuery = useJob(jobId)
  const eventsQuery = useJobEvents(jobId, 1, 100)
  const job = jobQuery.data
  const isPasswordBlocked = job?.status?.toUpperCase() === 'PASSWORD_REQUIRED' || job?.status?.toUpperCase() === 'WRONG_PASSWORD'

  const rootFileQuery = useFile(job?.root_file_id || '', {
    enabled: Boolean(job?.root_file_id),
  })

  const fileStatus = rootFileQuery.data?.processing_status?.toLowerCase() || ''
  const fileNeedsPassword =
    fileStatus === 'password_required' || fileStatus === 'wrong_password'

  const rootFile = rootFileQuery.data

  const blockedFileId = isPasswordBlocked
    ? (() => {
        const events = eventsQuery.data?.events || []
        const blockedEvent = [...events]
          .reverse()
          .find((event) =>
            typeof event.event_type === 'string' &&
            ['password_required', 'wrong_password'].some((needle) =>
              event.event_type.toLowerCase().includes(needle),
            ),
          )
        if (blockedEvent) {
          const details = blockedEvent.event_details as Record<string, unknown> | null
          if (typeof details?.file_id === 'string' && details.file_id.trim().length > 0) {
            return details.file_id
          }
        }

        if (fileNeedsPassword) {
          return rootFile?.id
        }
        return undefined
      })()
    : undefined

  if (jobQuery.isLoading) {
    return <div>Loading job...</div>
  }

  if (jobQuery.isError || !job) {
    return <div className="text-rose-300">Unable to load job details.</div>
  }

  const refreshAfterRetry = () => {
    queryClient.invalidateQueries({ queryKey: ['job', jobId] })
    queryClient.invalidateQueries({ queryKey: ['jobs', jobId, 'events'] })
    queryClient.invalidateQueries({ queryKey: ['jobs'] })
  }

  return (
    <div className="space-y-4">
      <PageHeader title={`Job ${job.id}`} subtitle={`Root file: ${job.root_file_id || '—'}`} />

      <div className="grid gap-4 lg:grid-cols-2">
        <Card title="Job summary">
          <div className="grid gap-2 md:grid-cols-2 text-sm">
            <div>
              <div className="text-slate-400">Status</div>
              <JobStatusBadge status={job.status} />
            </div>
            <div>
              <div className="text-slate-400">Current stage</div>
              <div>{job.current_stage || '—'}</div>
            </div>
            <div>
              <div className="text-slate-400">Progress</div>
              <div>{job.progress_percent}%</div>
            </div>
            <div>
              <div className="text-slate-400">Retry count</div>
              <div>{job.retry_count}</div>
            </div>
            <div>
              <div className="text-slate-400">Started</div>
              <div>{parseTime(job.started_at || job.created_at)}</div>
            </div>
            <div>
              <div className="text-slate-400">Completed</div>
              <div>{parseTime(job.completed_at)}</div>
            </div>
            <div>
              <div className="text-slate-400">Updated</div>
              <div>{parseTime(job.updated_at)}</div>
            </div>
          </div>
        </Card>

      <LiveJobStatusBanner job={job} />

      <Card title="Root file" subtitle={rootFile ? rootFile.original_name : 'Loading root file details'}>
          {rootFileQuery.isLoading ? (
            <div className="text-slate-400">Loading root file...</div>
          ) : rootFile ? (
            <div className="grid gap-2 text-sm">
              <div>
                <span className="text-slate-400">File:</span>{' '}
                <Link className="text-blue-300 underline" to={`/files/${rootFile.id}`}>
                  {rootFile.original_name}
                </Link>
              </div>
              <div>
                <span className="text-slate-400">File ID:</span> {rootFile.id}
              </div>
              <div>
                <span className="text-slate-400">Detected type:</span> {rootFile.detected_file_type || '—'}
              </div>
              <div>
                <span className="text-slate-400">Processing status:</span> {rootFile.processing_status}
              </div>
              {(isPasswordBlocked || fileNeedsPassword) && blockedFileId ? (
                <div>
                  <button
                    className="rounded border border-amber-400/60 px-3 py-1.5 text-xs text-amber-100"
                    onClick={() => setPasswordModalOpen(true)}
                  >
                    Enter archive password
                  </button>
                </div>
              ) : null}
            </div>
          ) : rootFileQuery.isError ? (
            <div className="text-rose-300">Unable to load root file.</div>
          ) : (
            <div className="text-slate-400">Root file no longer available.</div>
          )}
        </Card>
      </div>

      <Card title="Status timeline">
        <JobTimeline events={eventsQuery.data?.events || []} isLoading={eventsQuery.isLoading} />
      </Card>

      <JobErrorPanel job={job} />

      <Card title="Actions">
        <RetryJobButton
          jobId={job.id}
          onRetrySuccess={refreshAfterRetry}
          label={job.status.toLowerCase() === 'failed' ? 'Retry job' : 'Restart job'}
          disabled={isJobLive(job)}
        />
      </Card>

      <PasswordRequiredModal
        fileId={blockedFileId || ''}
        open={isPasswordModalOpen}
        onClose={() => setPasswordModalOpen(false)}
        onSuccess={() => {
          queryClient.invalidateQueries({ queryKey: ['job', jobId] })
          queryClient.invalidateQueries({ queryKey: ['jobs'] })
          queryClient.invalidateQueries({ queryKey: ['job', jobId, 'events'] })
          if (blockedFileId) {
            queryClient.invalidateQueries({ queryKey: ['file', blockedFileId] })
          }
        }}
      />
    </div>
  )
}

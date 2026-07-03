import { useEffect, useState, useCallback } from 'react'
import { AlertCircle, Clock, RefreshCw } from './Icon'
import { Button } from './Button'
import { Card } from './Card'
import { CopyButton } from './CopyButton'
import { EmptyState } from './EmptyState'
import { Modal } from './Modal'
import { Spinner } from './Spinner'
import { StatusBadge } from './StatusBadge'
import { useToast } from './ToastProvider'
import { api } from '../api'
import type { Job } from '../types'

interface JobDetailModalProps {
  open: boolean
  jobId: string | null
  onClose: () => void
}

interface JobEvent {
  id?: string
  event_type: string
  message?: string | null
  created_at: string
  payload?: unknown
}

type RawEvent = Record<string, unknown>

const RELATIVE_THRESHOLDS: Array<[number, string]> = [
  [60, 's'],
  [3600, 'm'],
  [86400, 'h'],
  [86400 * 30, 'd'],
]

function parseDate(iso: string | undefined | null): Date | null {
  if (!iso) return null
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return null
  return d
}

function formatRelative(iso: string | undefined | null): string {
  const d = parseDate(iso)
  if (!d) return '—'
  const diffMs = Date.now() - d.getTime()
  if (diffMs < 0) return 'just now'
  const seconds = Math.floor(diffMs / 1000)
  if (seconds < 5) return 'just now'
  if (seconds >= 86400 * 30) return d.toLocaleDateString()
  if (seconds < 60) return `${seconds}s ago`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`
  return `${Math.floor(seconds / 86400)}d ago`
}

function formatAbsolute(iso: string | undefined | null): string {
  const d = parseDate(iso)
  if (!d) return '—'
  return d.toLocaleString()
}

function normalizeEvent(raw: unknown): JobEvent | null {
  if (!raw || typeof raw !== 'object') return null
  const r = raw as RawEvent
  const created =
    (typeof r.created_at === 'string' && r.created_at) ||
    (typeof r.timestamp === 'string' && r.timestamp) ||
    (typeof r.createdAt === 'string' && r.createdAt) ||
    ''
  const eventType =
    (typeof r.event_type === 'string' && r.event_type) ||
    (typeof r.type === 'string' && r.type) ||
    (typeof r.name === 'string' && r.name) ||
    'event'
  const message =
    (typeof r.message === 'string' && r.message) ||
    (typeof r.detail === 'string' && r.detail) ||
    null
  return {
    id: typeof r.id === 'string' ? r.id : undefined,
    event_type: eventType,
    message,
    created_at: created,
    payload: r.payload,
  }
}

export function JobDetailModal({ open, jobId, onClose }: JobDetailModalProps) {
  const [job, setJob] = useState<Job | null>(null)
  const [events, setEvents] = useState<JobEvent[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [password, setPassword] = useState('')
  const [retrying, setRetrying] = useState(false)
  const toast = useToast()

  const fetchJob = useCallback(async () => {
    if (!jobId) return
    setLoading(true)
    setError(null)
    try {
      const [jobData, eventsRaw] = await Promise.all([
        api.getJob(jobId),
        api.jobEvents(jobId).catch(() => [] as unknown[]),
      ])
      setJob(jobData)
      const list = Array.isArray(eventsRaw) ? eventsRaw : []
      const normalized = list
        .map(normalizeEvent)
        .filter((e): e is JobEvent => e !== null && Boolean(e.created_at))
        .sort((a, b) => (a.created_at < b.created_at ? 1 : -1))
      setEvents(normalized)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load job')
    } finally {
      setLoading(false)
    }
  }, [jobId])

  useEffect(() => {
    if (open && jobId) {
      void fetchJob()
    }
    if (!open) {
      setJob(null)
      setEvents([])
      setError(null)
      setPassword('')
    }
  }, [open, jobId, fetchJob])

  useEffect(() => {
    setPassword('')
  }, [jobId])

  const handleRetry = async () => {
    if (!jobId || !password) return
    setRetrying(true)
    try {
      await api.retryWithPassword(jobId, password)
      toast.success('Retry requested')
      setPassword('')
      void fetchJob()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Retry failed')
    } finally {
      setRetrying(false)
    }
  }

  const footer = (
    <Button variant="ghost" onClick={onClose}>
      Close
    </Button>
  )

  const isFailed = job?.status === 'failed'
  const retryDisabled = !password || !isFailed || retrying

  return (
    <Modal open={open} onClose={onClose} size="lg" title="Job details" footer={footer}>
      {!jobId ? (
        <EmptyState
          icon={AlertCircle}
          title="No job selected"
          description="Pick a job from the list to see details."
        />
      ) : loading ? (
        <div className="stack stack-sm centered-loading">
          <Spinner size="lg" label="Loading job…" />
        </div>
      ) : error ? (
        <EmptyState
          icon={AlertCircle}
          title="Couldn't load job"
          description={error}
          action={
            <Button
              onClick={() => void fetchJob()}
              variant="secondary"
              iconLeft={<RefreshCw size={14} />}
            >
              Retry
            </Button>
          }
        />
      ) : job ? (
        <div className="stack">
          <Card title="Overview">
            <div className="grid-2">
              <div>
                <div className="muted text-sm">Job ID</div>
                <div className="row" style={{ alignItems: 'center' }}>
                  <span className="font-mono text-sm">{job.id}</span>
                  <CopyButton text={job.id} />
                </div>
              </div>
              <div>
                <div className="muted text-sm">Upload ID</div>
                <div className="row" style={{ alignItems: 'center' }}>
                  <span className="font-mono text-sm">{job.upload_id}</span>
                  <CopyButton text={job.upload_id} />
                </div>
              </div>
              <div>
                <div className="muted text-sm">Status</div>
                <StatusBadge status={job.status} kind="job" />
              </div>
              <div>
                <div className="muted text-sm">Stage</div>
                <span>{job.stage || '—'}</span>
              </div>
              <div>
                <div className="muted text-sm">Attempts</div>
                <span>{job.attempt_count}</span>
              </div>
              <div>
                <div className="muted text-sm">Created</div>
                <span title={formatAbsolute(job.created_at)}>
                  {formatRelative(job.created_at)}
                </span>
              </div>
              <div>
                <div className="muted text-sm">Updated</div>
                <span title={formatAbsolute(job.updated_at)}>
                  {formatRelative(job.updated_at)}
                </span>
              </div>
              <div>
                <div className="muted text-sm">Finished</div>
                <span title={formatAbsolute(job.finished_at)}>
                  {job.finished_at ? formatRelative(job.finished_at) : '—'}
                </span>
              </div>
            </div>
          </Card>

          {job.error_message ? (
            <div className="error-banner" role="alert">
              <span className="error-banner-icon" aria-hidden="true">
                <AlertCircle size={16} />
              </span>
              <div>
                <div className="error-banner-title">Error</div>
                <div className="error-banner-message">{job.error_message}</div>
              </div>
            </div>
          ) : null}

          <Card title="Events" subtitle="Most recent first">
            {events.length === 0 ? (
              <EmptyState
                icon={Clock}
                title="No events yet"
                description="The job hasn't produced any events."
              />
            ) : (
              <ol className="timeline">
                {events.map((ev, idx) => (
                  <li className="timeline-item" key={ev.id ?? `${ev.event_type}-${idx}`}>
                    <span className="timeline-dot" aria-hidden="true" />
                    <div className="timeline-body">
                      <div className="row-between">
                        <span className="timeline-type">{ev.event_type}</span>
                        <span className="muted text-sm" title={formatAbsolute(ev.created_at)}>
                          {formatRelative(ev.created_at)}
                        </span>
                      </div>
                      {ev.message ? (
                        <div className="timeline-message">{ev.message}</div>
                      ) : null}
                    </div>
                  </li>
                ))}
              </ol>
            )}
          </Card>

          {isFailed ? (
            <Card title="Retry with password">
              <div className="row" style={{ gap: 8, flexWrap: 'wrap' }}>
                <input
                  type="password"
                  className="input"
                  placeholder="Password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' && !retryDisabled) {
                      e.preventDefault()
                      void handleRetry()
                    }
                  }}
                />
                <Button
                  variant="primary"
                  loading={retrying}
                  disabled={retryDisabled}
                  onClick={() => void handleRetry()}
                >
                  Retry
                </Button>
              </div>
            </Card>
          ) : null}
        </div>
      ) : null}
    </Modal>
  )
}

export default JobDetailModal

import { useCallback, useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api'
import type { Job } from '../types'
import { Button } from '../components/Button'
import { Card } from '../components/Card'
import { CopyButton } from '../components/CopyButton'
import { EmptyState } from '../components/EmptyState'
import { JobDetailModal } from '../components/JobDetailModal'
import { Skeleton } from '../components/Skeleton'
import { StatusBadge } from '../components/StatusBadge'
import {
  Activity,
  AlertCircle,
  Filter,
  Inbox,
  PlayCircle,
  RefreshCw,
  Search,
  UploadCloud,
} from '../components/Icon'

const REFRESH_INTERVAL_MS = 5000

const STATUS_OPTIONS: Array<{ key: string; label: string }> = [
  { key: 'all', label: 'All' },
  { key: 'queued', label: 'Queued' },
  { key: 'running', label: 'Running' },
  { key: 'completed', label: 'Completed' },
  { key: 'failed', label: 'Failed' },
  { key: 'retrying', label: 'Retrying' },
  { key: 'cancelled', label: 'Cancelled' },
]

function formatRelative(iso: string | undefined | null): string {
  if (!iso) return '—'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return '—'
  const diffMs = Date.now() - d.getTime()
  if (diffMs < 0) return 'just now'
  const seconds = Math.floor(diffMs / 1000)
  if (seconds < 5) return 'just now'
  if (seconds < 60) return `${seconds}s ago`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`
  if (seconds < 86400 * 30) return `${Math.floor(seconds / 86400)}d ago`
  return d.toLocaleDateString()
}

export default function JobsPage() {
  const [jobs, setJobs] = useState<Job[] | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [statusFilter, setStatusFilter] = useState<Set<string>>(() => new Set(['all']))
  const [search, setSearch] = useState('')
  const [selectedJobId, setSelectedJobId] = useState<string | null>(null)
  const [tabVisible, setTabVisible] = useState<boolean>(
    typeof document !== 'undefined' ? !document.hidden : true,
  )

  const load = useCallback(async () => {
    try {
      const result = await api.listJobs('?limit=100')
      setJobs(Array.isArray(result) ? result : [])
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load jobs')
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  useEffect(() => {
    const onVisibility = () => setTabVisible(!document.hidden)
    document.addEventListener('visibilitychange', onVisibility)
    return () => document.removeEventListener('visibilitychange', onVisibility)
  }, [])

  useEffect(() => {
    if (!tabVisible) return
    const id = window.setInterval(() => {
      void load()
    }, REFRESH_INTERVAL_MS)
    return () => window.clearInterval(id)
  }, [tabVisible, load])

  const toggleStatus = (key: string) => {
    setStatusFilter((prev) => {
      const next = new Set(prev)
      if (key === 'all') {
        return new Set(['all'])
      }
      next.delete('all')
      if (next.has(key)) {
        next.delete(key)
      } else {
        next.add(key)
      }
      if (next.size === 0) next.add('all')
      return next
    })
  }

  const clearFilters = () => {
    setStatusFilter(new Set(['all']))
    setSearch('')
  }

  const filtered = useMemo(() => {
    const list = jobs ?? []
    const q = search.trim().toLowerCase()
    return list
      .filter((j) => statusFilter.has('all') || statusFilter.has(j.status))
      .filter((j) => {
        if (!q) return true
        return (
          j.id.toLowerCase().includes(q) ||
          j.upload_id.toLowerCase().includes(q)
        )
      })
  }, [jobs, statusFilter, search])

  const total = jobs?.length ?? 0
  const isInitialLoading = jobs === null && error === null
  // Pagination cursor not yet supported by backend
  const loadMore = () => undefined

  const onRowActivate = (jobId: string) => setSelectedJobId(jobId)
  const onRowKeyDown = (e: React.KeyboardEvent<HTMLTableRowElement>, jobId: string) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault()
      setSelectedJobId(jobId)
    }
  }

  const tableBody = (() => {
    if (isInitialLoading) {
      return (
        <div className="stack-sm card-pad-md">
          {Array.from({ length: 8 }).map((_, i) => (
            <Skeleton key={i} height={36} />
          ))}
        </div>
      )
    }

    if (error) {
      return (
        <EmptyState
          icon={AlertCircle}
          title="Couldn't load jobs"
          description={error}
          action={
            <Button onClick={load} variant="secondary" iconLeft={<RefreshCw size={14} />}>
              Retry
            </Button>
          }
        />
      )
    }

    if (total === 0) {
      return (
        <EmptyState
          icon={Inbox}
          title="No jobs yet"
          description="Upload a file to create your first job."
          action={
            <Link to="/upload">
              <Button variant="primary" iconLeft={<UploadCloud size={16} />}>
                Upload a file
              </Button>
            </Link>
          }
        />
      )
    }

    if (filtered.length === 0) {
      return (
        <EmptyState
          icon={Search}
          title="No matching jobs"
          description="No jobs match your filters."
          action={
            <Button onClick={clearFilters} variant="ghost">
              Clear filters
            </Button>
          }
        />
      )
    }

    return (
      <>
        <div className="data-table-wrap">
          <table className="data-table">
            <thead>
              <tr>
                <th>Job</th>
                <th>Upload</th>
                <th>Status</th>
                <th>Stage</th>
                <th>Attempts</th>
                <th>Created</th>
                <th>Updated</th>
                <th>Error</th>
                <th aria-label="Actions" />
              </tr>
            </thead>
            <tbody>
              {filtered.map((job) => (
                <tr
                  key={job.id}
                  className="job-row"
                  tabIndex={0}
                  role="button"
                  aria-label={`Open job ${job.id}`}
                  onClick={() => onRowActivate(job.id)}
                  onKeyDown={(e) => onRowKeyDown(e, job.id)}
                >
                  <td>
                    <span className="row">
                      <code className="font-mono text-sm">{job.id.slice(0, 8)}</code>
                      <CopyButton text={job.id} ariaLabel="Copy job id" />
                    </span>
                  </td>
                  <td>
                    <span className="row muted">
                      <code className="font-mono text-sm">{job.upload_id.slice(0, 8)}</code>
                      <CopyButton text={job.upload_id} ariaLabel="Copy upload id" />
                    </span>
                  </td>
                  <td>
                    <StatusBadge status={job.status} kind="job" />
                  </td>
                  <td className="muted text-sm">{job.stage || '—'}</td>
                  <td>{job.attempt_count}</td>
                  <td className="muted text-sm" title={job.created_at}>
                    {formatRelative(job.created_at)}
                  </td>
                  <td className="muted text-sm" title={job.updated_at}>
                    {formatRelative(job.updated_at)}
                  </td>
                  <td>
                    <span
                      className="truncate muted text-sm"
                      title={job.error_message ?? ''}
                      style={{ maxWidth: 220, display: 'inline-block' }}
                    >
                      {job.error_message || '—'}
                    </span>
                  </td>
                  <td onClick={(e) => e.stopPropagation()}>
                    <Button
                      variant="ghost"
                      size="sm"
                      iconLeft={<PlayCircle size={14} />}
                      onClick={() => setSelectedJobId(job.id)}
                    >
                      Open
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <div className="card-pad-md" style={{ display: 'flex', justifyContent: 'center' }}>
          <Button variant="ghost" onClick={loadMore}>
            Load more
          </Button>
        </div>
      </>
    )
  })()

  return (
    <div className="stack-lg">
      <div className="row-between">
        <div>
          <h2 style={{ margin: 0 }}>Jobs</h2>
          <span className="muted text-sm">
            Track parsing and indexing jobs in real time.
          </span>
        </div>
        <div className="row">
          <span className="row" style={{ gap: 6 }}>
            <span className="status-dot status-dot-info" aria-hidden="true" />
            <span className="text-sm muted">{tabVisible ? 'Live' : 'Paused'}</span>
          </span>
          <Button
            variant="secondary"
            size="sm"
            iconLeft={<RefreshCw size={14} />}
            onClick={load}
          >
            Refresh
          </Button>
        </div>
      </div>

      <Card padding="md">
        <div className="stack">
          <div className="row" style={{ gap: 6, flexWrap: 'wrap' }}>
            <span
              className="row muted text-sm"
              style={{ gap: 4, marginRight: 4 }}
              aria-hidden="true"
            >
              <Filter size={14} />
              Status
            </span>
            {STATUS_OPTIONS.map((opt) => {
              const active = statusFilter.has(opt.key)
              return (
                <button
                  key={opt.key}
                  type="button"
                  className={`chip ${active ? 'chip-active' : ''}`.trim()}
                  aria-pressed={active}
                  onClick={() => toggleStatus(opt.key)}
                >
                  {opt.label}
                </button>
              )
            })}
          </div>
          <div className="row" style={{ gap: 6 }}>
            <Search size={14} className="muted" />
            <input
              className="input"
              type="search"
              placeholder="Filter by job id or upload id…"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              aria-label="Filter jobs by id or upload id"
            />
          </div>
        </div>
      </Card>

      <Card
        title={`Jobs (${filtered.length})`}
        subtitle={`Showing ${filtered.length} of ${total}`}
        padding="none"
        actions={
          <span className="row muted text-sm" style={{ gap: 4 }}>
            <Activity size={14} />
            {tabVisible ? 'Auto-refreshing' : 'Paused'}
          </span>
        }
      >
        {tableBody}
      </Card>

      <JobDetailModal
        open={selectedJobId !== null}
        jobId={selectedJobId}
        onClose={() => setSelectedJobId(null)}
      />
    </div>
  )
}

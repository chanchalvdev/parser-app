import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import {
  Activity,
  AlertCircle,
  FileText,
  Hash,
  Inbox,
  ListChecks,
  Loader2,
  RefreshCw,
  Search,
  UploadCloud,
} from '../components/Icon'
import { Card } from '../components/Card'
import { Button } from '../components/Button'
import { StatusBadge } from '../components/StatusBadge'
import { Skeleton } from '../components/Skeleton'
import { EmptyState } from '../components/EmptyState'
import { JobDetailModal } from '../components/JobDetailModal'
import { api } from '../api'
import type { DashboardSummary, Job } from '../types'

const STATUS_COLOR: Record<string, string> = {
  completed: 'var(--success)',
  running: 'var(--info)',
  queued: 'var(--text-muted)',
  failed: 'var(--danger)',
  retrying: 'var(--warning)',
  cancelled: 'var(--text-dim)',
}

function colorFor(status: string): string {
  return STATUS_COLOR[status.toLowerCase()] ?? 'var(--accent)'
}

function formatRelative(iso: string | undefined | null): string {
  if (!iso) return '—'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return '—'
  const diffMs = Date.now() - d.getTime()
  if (diffMs < 0) return 'just now'
  const seconds = Math.floor(diffMs / 1000)
  if (seconds < 60) return 'just now'
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`
  if (seconds < 86400 * 30) return `${Math.floor(seconds / 86400)}d ago`
  return d.toLocaleDateString()
}

interface StatCardProps {
  icon: typeof UploadCloud
  value: number | string
  label: string
  tint?: string
}

function StatCard({ icon: Icon, value, label, tint = 'var(--accent)' }: StatCardProps) {
  return (
    <Card padding="md">
      <div className="stack-sm">
        <span
          className="stat-icon"
          style={{ background: `var(--accent-soft, ${tint})`, color: tint }}
          aria-hidden="true"
        >
          <Icon size={18} />
        </span>
        <div className="text-3xl" style={{ fontWeight: 700, lineHeight: 1.1 }}>{value}</div>
        <div className="text-sm">{label}</div>
        <div className="muted text-sm">since yesterday</div>
      </div>
    </Card>
  )
}

export default function Dashboard() {
  const [summary, setSummary] = useState<DashboardSummary | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [selectedJobId, setSelectedJobId] = useState<string | null>(null)

  const reload = useCallback(async () => {
    setError(null)
    try {
      const data = await api.dashboardSummary()
      setSummary(data)
    } catch (e) {
      const msg = e instanceof Error ? e.message : 'Failed to fetch dashboard'
      setError(msg)
    }
  }, [])

  useEffect(() => {
    void reload()
  }, [reload])

  const loading = !summary && !error

  const totalJobs = summary
    ? Object.values(summary.jobs_by_status).reduce((acc, n) => acc + (n ?? 0), 0)
    : 0

  const inProgressCount = summary
    ? (summary.jobs_by_status.queued ?? 0) +
      (summary.jobs_by_status.running ?? 0) +
      (summary.jobs_by_status.retrying ?? 0)
    : 0

  return (
    <>
      <div className="stack-lg">
        <header>
          <h2>Dashboard</h2>
          <p className="muted">Overview of recent activity</p>
        </header>

        {error ? (
          <EmptyState
            icon={AlertCircle}
            title="Couldn't load dashboard"
            description={error}
            action={
              <Button
                onClick={() => void reload()}
                variant="secondary"
                iconLeft={<RefreshCw size={14} />}
              >
                Retry
              </Button>
            }
          />
        ) : loading ? (
          <div className="grid-4">
            <Skeleton height={88} />
            <Skeleton height={88} />
            <Skeleton height={88} />
            <Skeleton height={88} />
          </div>
        ) : summary ? (
          <div className="grid-4">
            <StatCard icon={UploadCloud} value={summary.uploads} label="Total uploads" tint="var(--info)" />
            <StatCard icon={FileText} value={summary.files} label="Total files" tint="var(--accent)" />
            <StatCard icon={Hash} value={summary.parsed_records} label="Parsed records" tint="var(--success)" />
            <StatCard icon={Loader2} value={inProgressCount} label="Jobs in progress" tint="var(--warning)" />
          </div>
        ) : null}

        {summary && !error ? (
          <Card title="Job status breakdown" subtitle="Composition across all jobs">
            {totalJobs === 0 ? (
              <EmptyState
                icon={Activity}
                title="No jobs yet"
                description="Job status will appear here once you upload your first file."
              />
            ) : (
              <div className="stack">
                <div
                  className="status-bar"
                  role="img"
                  aria-label="Job status composition"
                >
                  {Object.entries(summary.jobs_by_status).map(([status, count]) => {
                    const pct = totalJobs > 0 ? (count / totalJobs) * 100 : 0
                    if (pct <= 0) return null
                    return (
                      <span
                        key={status}
                        className="status-bar-segment"
                        style={{
                          width: `${pct}%`,
                          background: colorFor(status),
                        }}
                        title={`${status}: ${count}`}
                      />
                    )
                  })}
                </div>
                <div className="status-legend">
                  {Object.entries(summary.jobs_by_status).map(([status, count]) => (
                    <span key={status} className="status-legend-chip">
                      <span
                        className="dot"
                        style={{ background: colorFor(status) }}
                        aria-hidden="true"
                      />
                      <span>{status}</span>
                      <span className="muted">{count}</span>
                    </span>
                  ))}
                </div>
              </div>
            )}
          </Card>
        ) : null}

        {summary && !error ? (
          <Card
            title="Recent jobs"
            subtitle="The five most recent jobs"
            actions={
              <Link to="/jobs">
                <Button variant="ghost" size="sm" iconRight={<ListChecks size={14} />}>
                  View all
                </Button>
              </Link>
            }
          >
            {summary.recent_jobs.length === 0 ? (
              <EmptyState
                icon={Inbox}
                title="No activity yet"
                description="Upload your first file to get started."
                action={
                  <Link to="/upload">
                    <Button variant="primary">Upload a file</Button>
                  </Link>
                }
              />
            ) : (
              <div className="stack-sm" style={{ margin: 'calc(var(--space-3) * -1)' }}>
                {summary.recent_jobs.map((job: Job) => (
                  <button
                    key={job.id}
                    type="button"
                    className="job-row"
                    onClick={() => setSelectedJobId(job.id)}
                  >
                    <span className="muted" aria-hidden="true">
                      <Hash size={14} />
                    </span>
                    <code className="font-mono text-sm">{job.id.slice(0, 8)}</code>
                    <StatusBadge status={job.status} kind="job" />
                    <span className="muted text-sm truncate">{job.stage || '—'}</span>
                    <span className="muted text-sm" style={{ marginLeft: 'auto' }}>
                      {formatRelative(job.updated_at)}
                    </span>
                  </button>
                ))}
              </div>
            )}
          </Card>
        ) : null}

        <Card title="Quick actions" subtitle="Common starting points">
          <div className="row" style={{ flexWrap: 'wrap', gap: 'var(--space-3)' }}>
            <Link to="/upload">
              <Button variant="primary" iconLeft={<UploadCloud size={16} />}>
                Upload a file
              </Button>
            </Link>
            <Link to="/search">
              <Button variant="secondary" iconLeft={<Search size={16} />}>
                Search content
              </Button>
            </Link>
            <Link to="/jobs">
              <Button variant="ghost" iconLeft={<ListChecks size={16} />}>
                View all jobs
              </Button>
            </Link>
          </div>
        </Card>
      </div>

      <JobDetailModal
        open={!!selectedJobId}
        jobId={selectedJobId}
        onClose={() => setSelectedJobId(null)}
      />
    </>
  )
}

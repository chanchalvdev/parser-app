import { Link, useSearchParams } from 'react-router-dom'
import { useMemo } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useUiStore } from '@/stores/uiStore'
import { useJobs, isJobLive } from '@/hooks/apiHooks'
import { DataTable } from '@/components/ui/DataTable'
import { PageHeader } from '@/components/layout/PageHeader'
import { Input } from '@/components/ui/Input'
import { parseTime } from '@/utils/query'
import type { ColumnDef } from '@tanstack/react-table'
import type { IngestionJob } from '@/types/job'
import { JobStatusBadge } from '@/components/jobs/JobStatusBadge'
import { RetryJobButton } from '@/components/jobs/RetryJobButton'

export const JobsPage = () => {
  const tenantId = useUiStore((state) => state.tenantId)
  const queryClient = useQueryClient()
  const [searchParams, setSearchParams] = useSearchParams()
  const page = Number(searchParams.get('page') || '1')
  const pageSize = Number(searchParams.get('page_size') || '25')
  const statusFilter = searchParams.get('status') || ''

  const query = useJobs({
    tenant_id: tenantId,
    status: statusFilter || undefined,
    page,
    page_size: pageSize,
  })

  const columns: ColumnDef<IngestionJob>[] = useMemo(
    () => [
      {
        accessorKey: 'id',
        header: 'Job',
        cell: (info) => {
          const jobId = String(info.getValue())
          return (
            <Link to={`/jobs/${jobId}`} className="text-blue-300 hover:underline">
              {jobId.slice(0, 8)}…
            </Link>
          )
        },
      },
      {
        accessorKey: 'status',
        header: 'Status',
        cell: (info) => <JobStatusBadge status={String(info.getValue())} />,
      },
      {
        accessorKey: 'current_stage',
        header: 'Current stage',
        cell: (info) => String(info.getValue() || '—'),
      },
      {
        accessorKey: 'progress_percent',
        header: 'Progress',
        cell: (info) => `${info.getValue() as number}%`,
      },
      {
        accessorKey: 'retry_count',
        header: 'Retries',
        cell: (info) => info.getValue() as number,
      },
      {
        accessorKey: 'updated_at',
        header: 'Updated',
        cell: (info) => parseTime(info.getValue() as string),
      },
      {
        id: 'actions',
        header: 'Actions',
        cell: (info) => {
          const job = info.row.original
          const live = isJobLive(job)
          return (
            <div className="flex items-center gap-2">
              <Link to={`/jobs/${job.id}`} className="rounded border border-slate-500 px-3 py-1.5 text-xs">
                View
              </Link>
              <RetryJobButton
                jobId={job.id}
                onRetrySuccess={() => queryClient.invalidateQueries({ queryKey: ['jobs'] })}
                label={job.status.toLowerCase() === 'failed' ? 'Retry job' : 'Restart job'}
                disabled={live}
              />
            </div>
          )
        },
      },
    ],
    [queryClient],
  )

  const setPage = (next: number) => {
    const nextParams = new URLSearchParams(searchParams)
    nextParams.set('page', String(next))
    setSearchParams(nextParams)
  }

  const setFilter = (key: string, value: string) => {
    const next = new URLSearchParams(searchParams)
    if (value) {
      next.set(key, value)
    } else {
      next.delete(key)
    }
    next.set('page', '1')
    setSearchParams(next)
  }

  return (
    <div className="space-y-4">
      <PageHeader title="Jobs" subtitle="Inspect extraction jobs and retry failed ones." />
      <div className="panel p-4">
        <div className="grid gap-3 md:grid-cols-2">
          <label className="space-y-1">
            <span className="text-xs text-slate-300">Status</span>
            <Input value={statusFilter} onChange={(event) => setFilter('status', event.target.value)} />
          </label>
          <label className="space-y-1">
            <span className="text-xs text-slate-300">Page size</span>
            <Input
              value={pageSize}
              type="number"
              onChange={(event) => setFilter('page_size', event.target.value)}
            />
          </label>
        </div>
      </div>
      <div className="panel p-4">
        <DataTable data={query.data?.jobs || []} columns={columns} isLoading={query.isLoading} />
      </div>
      <div className="flex justify-end gap-2 text-sm">
        <button
          className="rounded border border-slate-600 px-3 py-1 hover:bg-slate-800"
          disabled={page <= 1}
          onClick={() => setPage(page - 1)}
        >
          Previous
        </button>
        <button
          className="rounded border border-slate-600 px-3 py-1 hover:bg-slate-800"
          disabled={!query.data || page * pageSize >= query.data.total}
          onClick={() => setPage(page + 1)}
        >
          Next
        </button>
      </div>
      {query.isError ? <p className="text-sm text-rose-300">Unable to load jobs.</p> : null}
      <p className="text-xs text-slate-400">Showing page {page}</p>
    </div>
  )
}

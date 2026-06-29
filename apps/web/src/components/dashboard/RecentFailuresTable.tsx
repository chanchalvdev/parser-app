import { useMemo } from 'react'
import { DataTable } from '@/components/ui/DataTable'
import { Card } from '@/components/ui/Card'
import type { DashboardErrorBreakdown } from '@/types/dashboard'
import type { ColumnDef } from '@tanstack/react-table'

type RecentFailuresTableProps = {
  breakdown?: DashboardErrorBreakdown
  isLoading?: boolean
}

type FailureRow = {
  error_code: string
  count: number
  last_seen: string
}

const parseRows = (data?: DashboardErrorBreakdown) => {
  if (!data?.errors) return []

  return [...data.errors]
    .sort((a, b) => {
      const aTime = Date.parse(a.last_seen)
      const bTime = Date.parse(b.last_seen)
      return bTime - aTime
    })
    .map((item) => ({
      error_code: item.error_code || 'unknown',
      count: item.count,
      last_seen: item.last_seen,
    }))
}

export const RecentFailuresTable = ({ breakdown, isLoading }: RecentFailuresTableProps) => {
  const rows: FailureRow[] = useMemo(() => parseRows(breakdown), [breakdown])

  const columns = useMemo<ColumnDef<FailureRow>[]>(
    () => [
      {
        accessorKey: 'error_code',
        header: 'Error code',
      },
      {
        accessorKey: 'count',
        header: 'Count',
      },
      {
        accessorKey: 'last_seen',
        header: 'Last seen',
      },
    ],
    [],
  )

  return (
    <Card title="Recent failures" subtitle="Most recently seen error codes">
      <DataTable data={rows} columns={columns} isLoading={isLoading} />
      {rows.length === 0 && !isLoading ? <p className="mt-2 text-sm text-slate-300">No failures to report.</p> : null}
    </Card>
  )
}

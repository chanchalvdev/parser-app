import { useMemo } from 'react'
import { DataTable } from '@/components/ui/DataTable'
import { Card } from '@/components/ui/Card'
import type { DashboardEntities } from '@/types/dashboard'
import type { ColumnDef } from '@tanstack/react-table'

type TopEntitiesTableProps = {
  data?: DashboardEntities
  limit?: number
  isLoading?: boolean
}

type TopEntityRow = {
  kind: string
  value: string
  count: number
}

const flattenEntities = (entities?: DashboardEntities, limit = 12): TopEntityRow[] => {
  if (!entities) return []

  const buckets: TopEntityRow[] = []

  Object.entries(entities.entities || {}).forEach(([kind, values]) => {
    if (!Array.isArray(values)) return

    values.forEach((entry) => {
      buckets.push({
        kind,
        value: entry.value,
        count: entry.count,
      })
    })
  })

  return buckets.sort((a, b) => b.count - a.count).slice(0, limit)
}

export const TopEntitiesTable = ({ data, limit = 12, isLoading }: TopEntitiesTableProps) => {
  const rows = useMemo(() => flattenEntities(data, limit), [data, limit])
  const columns = useMemo<ColumnDef<TopEntityRow>[]>(
    () => [
      {
        accessorKey: 'kind',
        header: 'Entity type',
      },
      {
        accessorKey: 'value',
        header: 'Value',
      },
      {
        accessorKey: 'count',
        header: 'Count',
        cell: (info) => String(info.getValue()),
      },
    ],
    [],
  )

  return (
    <Card title="Top entities" subtitle="Most frequent extracted entities">
      <DataTable data={rows} columns={columns} isLoading={isLoading} />
    </Card>
  )
}

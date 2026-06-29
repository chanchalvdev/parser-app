import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { DataTable } from '@/components/ui/DataTable'
import { getFileRecords } from '@/services/fileService'
import { Badge } from '@/components/ui/Badge'
import { parseTime } from '@/utils/query'
import type { ParsedRecord } from '@/types/domain'
import type { ColumnDef } from '@tanstack/react-table'

type Props = {
  fileId: string
  page?: number
  pageSize?: number
}

export const FileRecordsPanel = ({ fileId, page = 1, pageSize = 10 }: Props) => {
  const query = useQuery({
    queryKey: ['file-records', fileId, page, pageSize],
    queryFn: () => getFileRecords(fileId, page, pageSize),
    enabled: Boolean(fileId),
  })

  const columns = useMemo<ColumnDef<ParsedRecord>[]>(
    () => [
      { accessorKey: 'record_type', header: 'Type', cell: (info) => info.getValue() || '—' },
      {
        accessorKey: 'line_number',
        header: 'Line',
        cell: (info) => {
          const value = info.getValue() as number | null
          return value || '—'
        },
      },
      {
        accessorKey: 'content_text',
        header: 'Content',
        cell: (info) => {
          const value = (info.getValue() as string | null) || ''
          return <div className="max-w-xl truncate">{value}</div>
        },
      },
      {
        accessorKey: 'event_timestamp',
        header: 'Event time',
        cell: (info) => parseTime(info.getValue() as string),
      },
      {
        id: 'entities',
        header: 'Entities',
        cell: (info) => {
          const row = info.row.original
          const entities = row.extracted_entities || {}
          const values = Object.values(entities).slice(0, 2).join(', ')
          return values ? <Badge tone="blue">{String(values)}</Badge> : '—'
        },
      },
    ],
    [],
  )

  if (query.isError) {
    return <div className="text-rose-300">Error loading records.</div>
  }

  const records = query.data?.records || []

  return (
    <div className="space-y-2">
      <DataTable data={records} columns={columns} isLoading={query.isLoading} caption="Parsed records" />
    </div>
  )
}

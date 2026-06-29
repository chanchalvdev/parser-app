import { Link } from 'react-router-dom'
import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { getFileRecords } from '@/services/filesApi'
import { DataTable } from '@/components/ui/DataTable'
import { parseTime } from '@/utils/query'
import type { ParsedRecord } from '@/types/file'
import type { FileRecordsResponse } from '@/types/file'
import type { ColumnDef } from '@tanstack/react-table'

type ParsedRecordsPreviewProps = {
  fileId: string
  pageSize?: number
  initialData?: FileRecordsResponse
}

export const ParsedRecordsPreview = ({ fileId, pageSize = 5, initialData }: ParsedRecordsPreviewProps) => {
  const query = useQuery({
    queryKey: ['file-records-preview', fileId, pageSize],
    queryFn: () => getFileRecords(fileId, 1, pageSize),
    enabled: !!fileId && !initialData,
  })

  const data = initialData || query.data
  const isLoading = query.isLoading && !initialData

  const columns = useMemo<ColumnDef<ParsedRecord>[]>(
    () => [
      {
        accessorKey: 'record_type',
        header: 'Type',
        cell: (info) => info.getValue() || 'text_line',
      },
      {
        accessorKey: 'line_number',
        header: 'Line',
        cell: (info) => info.getValue() || '—',
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
        accessorKey: 'created_at',
        header: 'Created',
        cell: (info) => parseTime(info.getValue() as string),
      },
    ],
    [],
  )

  if (query.isError) {
    return <div className="text-sm text-rose-300">Unable to load parsed records.</div>
  }

  return (
    <div className="space-y-2">
      <div className="text-sm text-slate-300">
        {data?.total ? <>Recent parsed records: {data.total}</> : 'No parsed records yet.'}
      </div>
      <DataTable data={data?.records || []} columns={columns} isLoading={isLoading} caption="Recent parsed records" />
      {data && data.total > data.records.length ? (
        <Link className="inline-block text-xs text-blue-300 underline" to={`/search?file_id=${encodeURIComponent(fileId)}`}>
          View all records
        </Link>
      ) : null}
    </div>
  )
}

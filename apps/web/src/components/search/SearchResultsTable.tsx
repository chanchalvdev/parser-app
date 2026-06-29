import { ColumnDef } from '@tanstack/react-table'
import { useMemo } from 'react'
import { Link } from 'react-router-dom'
import { DataTable } from '@/components/ui/DataTable'
import { Button } from '@/components/ui/Button'
import { parseTime } from '@/utils/query'
import { EntityBadges } from '@/components/search/EntityBadges'
import type { SearchResult } from '@/types/search'

type SearchResultsTableProps = {
  results: SearchResult[]
  isLoading?: boolean
  onOpenRecord: (result: SearchResult) => void
}

const renderHighlighted = (value: string | undefined | null) => {
  if (!value) return null

  const hasMarkup = /<[^>]+>/.test(value)
  if (!hasMarkup) {
    return <span>{value}</span>
  }

  return <span dangerouslySetInnerHTML={{ __html: value }} />
}

export const SearchResultsTable = ({ results, isLoading, onOpenRecord }: SearchResultsTableProps) => {
  const columns = useMemo<ColumnDef<SearchResult>[]>(
    () => [
      {
        accessorKey: 'record_id',
        header: 'Record',
        cell: (info) => String(info.getValue()).slice(0, 10),
      },
      {
        accessorKey: 'source_file_name',
        header: 'Source file',
        cell: (info) => {
          const row = info.row.original
          return (
            <Link to={`/files/${row.file_id}`} className="text-blue-300 hover:underline">
              {String(info.getValue())}
            </Link>
          )
        },
      },
      {
        accessorKey: 'record_type',
        header: 'Record type',
      },
      {
        accessorKey: 'content_preview',
        header: 'Preview',
        cell: (info) => (
          <div className="max-w-[600px] break-words text-xs text-slate-200">
            {renderHighlighted((info.row.original.highlight || info.getValue()) as string | undefined)}
          </div>
        ),
      },
      {
        accessorKey: 'entities',
        header: 'Entities',
        cell: (info) => <EntityBadges entities={info.getValue() as Record<string, unknown>} />,
      },
      {
        accessorKey: 'created_at',
        header: 'Created',
        cell: (info) => parseTime(info.getValue() as string),
      },
      {
        id: 'actions',
        header: 'Actions',
        cell: (info) => (
          <Button tone="ghost" onClick={() => onOpenRecord(info.row.original)} className="text-xs px-3 py-1">
            Open
          </Button>
        ),
      },
    ],
    [onOpenRecord],
  )

  return (
    <DataTable
      data={results}
      columns={columns}
      isLoading={isLoading}
      caption="Parsed record search results"
    />
  )
}

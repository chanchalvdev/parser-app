import { Link } from 'react-router-dom'
import { useMemo } from 'react'
import { FileStatusBadge } from '@/components/files/FileStatusBadge'
import { FileTypeBadge } from '@/components/files/FileTypeBadge'
import { DataTable } from '@/components/ui/DataTable'
import { parseTime } from '@/utils/query'
import type { FileItem } from '@/types/file'
import type { ColumnDef } from '@tanstack/react-table'

type FileTableProps = {
  files: FileItem[]
  isLoading?: boolean
  showParent?: boolean
}

export const FileTable = ({ files, isLoading, showParent = false }: FileTableProps) => {
  const columns = useMemo<ColumnDef<FileItem>[]>(
    () => [
      {
        accessorKey: 'original_name',
        header: 'File',
        cell: (info) => {
          const row = info.row.original
          return (
            <Link to={`/files/${row.id}`} className="text-blue-300 hover:underline">
              {row.original_name}
            </Link>
          )
        },
      },
      {
        accessorKey: 'extension',
        header: 'Ext',
      },
      {
        accessorKey: 'detected_file_type',
        header: 'Detected type',
        cell: (info) => <FileTypeBadge fileType={info.getValue() as string | null} />,
      },
      {
        accessorKey: 'processing_status',
        header: 'Status',
        cell: (info) => <FileStatusBadge status={info.getValue() as string} />,
      },
      {
        accessorKey: 'created_at',
        header: 'Created',
        cell: (info) => parseTime(info.getValue() as string),
      },
      ...(showParent
        ? [
            {
              id: 'parent',
              header: 'Parent',
              cell: (info: { row: { original: FileItem } }) => {
                const parentId = info.row.original.parent_file_id
                if (!parentId) return '—'
                return (
                  <Link to={`/files/${parentId}`} className="text-blue-300 hover:underline">
                    {parentId}
                  </Link>
                )
              },
            } as ColumnDef<FileItem>,
          ]
        : []),
    ],
    [],
  )

  return (
    <DataTable
      data={files}
      columns={columns}
      isLoading={isLoading}
      caption="File list"
    />
  )
}

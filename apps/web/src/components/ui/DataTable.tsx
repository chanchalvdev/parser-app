import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  useReactTable,
} from '@tanstack/react-table'
import { useMemo } from 'react'

export const DataTable = <T,>({
  data,
  columns,
  isLoading,
  caption,
}: {
  data: T[]
  columns: ColumnDef<T>[]
  isLoading?: boolean
  caption?: string
}) => {
  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
  })

  const renderedRows = useMemo(() => table.getRowModel().rows, [table])

  return (
    <div className="overflow-x-auto scrollbar-thin">
      <table className="w-full min-w-full text-sm text-left">
        {caption ? <caption className="sr-only">{caption}</caption> : null}
        <thead>
          {table.getHeaderGroups().map((headerGroup) => (
            <tr key={headerGroup.id} className="border-b border-slate-700/70">
              {headerGroup.headers.map((header) => (
                <th key={header.id} className="px-3 py-2 text-xs uppercase tracking-wide text-slate-300">
                  {header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}
                </th>
              ))}
            </tr>
          ))}
        </thead>
        <tbody>
          {isLoading ? (
            <tr>
              <td className="px-3 py-5 text-slate-400" colSpan={columns.length}>
                Loading…
              </td>
            </tr>
          ) : renderedRows.length === 0 ? (
            <tr>
              <td className="px-3 py-5 text-slate-400" colSpan={columns.length}>
                No rows found.
              </td>
            </tr>
          ) : (
            renderedRows.map((row) => (
              <tr key={row.id} className="border-b border-slate-800/80 hover:bg-slate-800/60">
                {row.getVisibleCells().map((cell) => (
                  <td key={cell.id} className="px-3 py-2 text-slate-200">
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </td>
                ))}
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  )
}

import { ChangeEvent } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { useUiStore } from '@/stores/uiStore'
import { useFiles } from '@/hooks/apiHooks'
import { Input } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { PageHeader } from '@/components/layout/PageHeader'
import { FileTable } from '@/components/files/FileTable'

export const FilesPage = () => {
  const tenantId = useUiStore((state) => state.tenantId)
  const [searchParams, setSearchParams] = useSearchParams()

  const page = Number(searchParams.get('page') || '1')
  const pageSize = Number(searchParams.get('page_size') || '25')
  const status = searchParams.get('status') || ''
  const extension = searchParams.get('extension') || ''
  const detectedFileType = searchParams.get('detected_file_type') || ''
  const nameQuery = searchParams.get('q') || ''

  const query = useFiles({
    tenant_id: tenantId,
    status: status || undefined,
    extension: extension || undefined,
    detected_file_type: detectedFileType || undefined,
    page,
    page_size: pageSize,
  })

  const files = query.data?.files || []

  const filteredFiles = nameQuery
    ? files.filter((file) => file.original_name.toLowerCase().includes(nameQuery.toLowerCase()))
    : files

  const setFilter = (key: string, value: string) => {
    const next = new URLSearchParams(searchParams)
    if (!value) {
      next.delete(key)
    } else {
      next.set(key, value)
    }
    next.set('page', '1')
    setSearchParams(next)
  }

  const setPage = (nextPage: number) => {
    const next = new URLSearchParams(searchParams)
    next.set('page', String(nextPage))
    setSearchParams(next)
  }

  const onNameInput = (event: ChangeEvent<HTMLInputElement>) => {
    setFilter('q', event.target.value)
  }

  return (
    <div className="space-y-4">
      <PageHeader title="Files" subtitle="Browse files and extracted archive children." />

      <Card>
        <div className="grid gap-3 md:grid-cols-3">
          <label className="space-y-1">
            <span className="text-xs text-slate-300">Search file name</span>
            <Input value={nameQuery} onChange={onNameInput} placeholder="type file name..." />
          </label>
          <label className="space-y-1">
            <span className="text-xs text-slate-300">Status</span>
            <Input value={status} onChange={(event) => setFilter('status', event.target.value)} />
          </label>
          <label className="space-y-1">
            <span className="text-xs text-slate-300">Extension</span>
            <Input value={extension} onChange={(event) => setFilter('extension', event.target.value)} />
          </label>
          <label className="space-y-1">
            <span className="text-xs text-slate-300">Detected type</span>
            <Input
              value={detectedFileType}
              onChange={(event) => setFilter('detected_file_type', event.target.value)}
            />
          </label>
          <label className="space-y-1">
            <span className="text-xs text-slate-300">Page size</span>
            <Input
              value={pageSize}
              type="number"
              onChange={(event) => setFilter('page_size', event.target.value)}
            />
          </label>
          <div className="flex items-end">
            <Button type="button" tone="ghost" onClick={() => setSearchParams(new URLSearchParams(searchParams))}>
              Apply filters
            </Button>
          </div>
        </div>
      </Card>

      <Card>
        <FileTable files={filteredFiles} isLoading={query.isLoading} showParent={false} />
        <div className="mt-4 flex flex-wrap items-center justify-between gap-2 text-sm">
          <span className="text-slate-300">
            Showing {filteredFiles.length} of page {page}
          </span>
          <div className="flex items-center gap-2">
            <button
              className="rounded border border-slate-600 px-3 py-1 hover:bg-slate-800 disabled:opacity-60"
              disabled={page <= 1 || query.isLoading}
              onClick={() => setPage(page - 1)}
            >
              Previous
            </button>
            <span className="text-slate-200">
              Page {page}
            </span>
            <button
              className="rounded border border-slate-600 px-3 py-1 hover:bg-slate-800 disabled:opacity-60"
              disabled={!query.data || page * pageSize >= query.data.total || query.isLoading}
              onClick={() => setPage(page + 1)}
            >
              Next
            </button>
          </div>
        </div>
        {nameQuery ? (
          <p className="mt-2 text-xs text-amber-300">
            Search by file name is filtered on the currently loaded page. Use tighter backend filters for full scans.
          </p>
        ) : null}
        {query.isError ? <p className="mt-2 text-sm text-rose-300">Failed to load files.</p> : null}
      </Card>

      <Card>
        <div className="text-sm">
          <span className="text-slate-300">Need archived detail?</span>{' '}
          <Link className="text-blue-300 underline" to="/files?detected_file_type=archive">
            Show archive-like files
          </Link>
        </div>
      </Card>
    </div>
  )
}

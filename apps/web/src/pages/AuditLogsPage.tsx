import { useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { listAuditLogs } from '@/services/auditService'
import { PageHeader } from '@/components/layout/PageHeader'
import { Card } from '@/components/ui/Card'
import { Input } from '@/components/ui/Input'
import { parseTime } from '@/utils/query'

export const AuditLogsPage = () => {
  const [searchParams, setSearchParams] = useSearchParams()
  const page = Number(searchParams.get('page') || '1')
  const pageSize = Number(searchParams.get('page_size') || '25')

  const [errorMessage, setErrorMessage] = useState<string>('')

  const tenantId = searchParams.get('tenant_id') || undefined

  const query = useQuery({
    queryKey: ['audit-logs', tenantId, page, pageSize],
    queryFn: async () => {
      try {
        setErrorMessage('')
        return await listAuditLogs({ tenantId, page, pageSize })
      } catch (error: any) {
        const message = error?.message || 'Unable to load audit logs.'
        setErrorMessage(message)
        return { total: 0, page: 1, page_size: 25, logs: [] }
      }
    },
  })

  const setPage = (next: number) => {
    const params = new URLSearchParams(searchParams)
    params.set('page', String(next))
    setSearchParams(params)
  }

  return (
    <div className="space-y-4">
      <PageHeader title="Audit Logs" subtitle="Platform activity and action trace." />
      {errorMessage ? <div className="rounded border border-amber-500 bg-amber-900/20 p-3 text-sm text-amber-100">{errorMessage}</div> : null}

      <Card>
        <div className="grid gap-3 md:grid-cols-3">
          <Input
            value={tenantId || ''}
            placeholder="tenant id"
            onChange={(event) => {
              const next = new URLSearchParams(searchParams)
              if (!event.target.value) {
                next.delete('tenant_id')
              } else {
                next.set('tenant_id', event.target.value)
              }
              setSearchParams(next)
            }}
          />
          <Input
            value={pageSize}
            type="number"
            onChange={(event) => {
              const next = new URLSearchParams(searchParams)
              next.set('page_size', event.target.value)
              next.set('page', '1')
              setSearchParams(next)
            }}
          />
        </div>
      </Card>

      <Card>
        <div className="space-y-3">
          {query.data?.logs?.map((item) => (
            <div key={item.id} className="rounded border border-slate-700 p-3 text-sm">
              <div className="text-xs text-slate-300">
                {item.action} · {parseTime(item.created_at)}
              </div>
              <div className="mt-1">tenant: {item.tenant_id}</div>
              <div>entity: {item.entity_type || 'N/A'} / {item.entity_id || 'N/A'}</div>
              <pre className="mt-2 text-xs">{JSON.stringify(item.details)}</pre>
            </div>
          )) || <div className="text-slate-300">No audit events.</div>}
        </div>
      </Card>

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
    </div>
  )
}

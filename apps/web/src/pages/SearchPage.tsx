import { useEffect, useMemo, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useUiStore } from '@/stores/uiStore'
import { useSearch } from '@/hooks/apiHooks'
import { PageHeader } from '@/components/layout/PageHeader'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { SearchBar } from '@/components/search/SearchBar'
import { SearchFilters, type SearchFiltersValues } from '@/components/search/SearchFilters'
import { SearchFacets, type SearchFacetUpdate } from '@/components/search/SearchFacets'
import { SearchResultsTable } from '@/components/search/SearchResultsTable'
import { RecordPreviewDrawer } from '@/components/search/RecordPreviewDrawer'
import type { SearchResponse, SearchResult } from '@/types/search'

const parsePage = (value: string | null, fallback: number): number => {
  const parsed = Number(value || '')
  return Number.isNaN(parsed) || parsed < 1 ? fallback : parsed
}

const parsePageSize = (value: string | null, fallback: number): number => {
  const parsed = Number(value || '')
  return Number.isNaN(parsed) || parsed < 1 ? fallback : parsed
}

const parseSort = (value: string | null): 'relevance' | 'created_at' => {
  return value === 'created_at' ? 'created_at' : 'relevance'
}

export const SearchPage = () => {
  const tenantId = useUiStore((state) => state.tenantId)
  const [searchParams, setSearchParams] = useSearchParams()

  const [selectedRecord, setSelectedRecord] = useState<SearchResult | null>(null)

  const [searchForm, setSearchForm] = useState<SearchFiltersValues>(() => ({
    q: searchParams.get('q') || '',
    detectedFileType: searchParams.get('detected_file_type') || '',
    extension: searchParams.get('extension') || '',
    recordType: searchParams.get('record_type') || '',
    dateFrom: searchParams.get('date_from') || '',
    dateTo: searchParams.get('date_to') || '',
    ip: searchParams.get('ip') || '',
    email: searchParams.get('email') || '',
    domain: searchParams.get('domain') || '',
    sort: parseSort(searchParams.get('sort')),
  }))

  useEffect(() => {
    setSearchForm({
      q: searchParams.get('q') || '',
      detectedFileType: searchParams.get('detected_file_type') || '',
      extension: searchParams.get('extension') || '',
      recordType: searchParams.get('record_type') || '',
      dateFrom: searchParams.get('date_from') || '',
      dateTo: searchParams.get('date_to') || '',
      ip: searchParams.get('ip') || '',
      email: searchParams.get('email') || '',
      domain: searchParams.get('domain') || '',
      sort: parseSort(searchParams.get('sort')),
    })
  }, [searchParams])

  const page = parsePage(searchParams.get('page'), 1)
  const pageSize = parsePageSize(searchParams.get('page_size'), 25)

  const request = useMemo(() => {
    return {
      tenant_id: tenantId,
      q: searchForm.q || undefined,
      extension: searchForm.extension || undefined,
      detected_file_type: searchForm.detectedFileType || undefined,
      record_type: searchForm.recordType || undefined,
      date_from: searchForm.dateFrom || undefined,
      date_to: searchForm.dateTo || undefined,
      ip: searchForm.ip || undefined,
      email: searchForm.email || undefined,
      domain: searchForm.domain || undefined,
      sort: searchForm.sort,
      page,
      page_size: pageSize,
    }
  }, [tenantId, searchForm, page, pageSize])

  const hasCriteria = Boolean(
    request.q ||
      request.extension ||
      request.detected_file_type ||
      request.record_type ||
      request.ip ||
      request.email ||
      request.domain ||
      request.date_from ||
      request.date_to,
  )

  const query = useSearch(request, {
    enabled: hasCriteria,
  })

  const syncToUrl = (next: Partial<SearchFiltersValues> & { page?: number; pageSize?: number }, resetPage = true) => {
    const params = new URLSearchParams(searchParams)

    const setOrClear = (key: string, value?: string | number) => {
      if (value === undefined || value === '' || value === 0) {
        params.delete(key)
      } else {
        params.set(key, String(value))
      }
    }

    setOrClear('q', next.q)
    setOrClear('detected_file_type', next.detectedFileType)
    setOrClear('extension', next.extension)
    setOrClear('record_type', next.recordType)
    setOrClear('date_from', next.dateFrom)
    setOrClear('date_to', next.dateTo)
    setOrClear('ip', next.ip)
    setOrClear('email', next.email)
    setOrClear('domain', next.domain)
    setOrClear('sort', next.sort)

    if (next.pageSize) {
      params.set('page_size', String(next.pageSize))
    }

    if (next.page !== undefined) {
      params.set('page', String(next.page))
    } else if (resetPage) {
      params.set('page', '1')
    }

    setSearchParams(params)
  }

  const updateSearchForm = (next: SearchFiltersValues) => {
    setSearchForm(next)
    syncToUrl(next, true)
  }

  const onFacetUpdate = (update: SearchFacetUpdate) => {
    const next = {
      ...searchForm,
      ...update,
    }

    setSearchForm(next)
    syncToUrl(next, true)
  }

  const resetFilters = () => {
    const cleared: SearchFiltersValues = {
      q: '',
      detectedFileType: '',
      extension: '',
      recordType: '',
      dateFrom: '',
      dateTo: '',
      ip: '',
      email: '',
      domain: '',
      sort: 'relevance',
    }

    setSearchForm(cleared)
    syncToUrl(cleared, true)
  }

  const setPage = (nextPage: number) => {
    syncToUrl(searchForm, false)

    const params = new URLSearchParams(searchParams)
    params.set('page', String(nextPage))
    setSearchParams(params)
  }

  const setPageSize = (nextPageSize: number) => {
    const params = new URLSearchParams(searchParams)
    params.set('page', '1')
    params.set('page_size', String(nextPageSize))
    setSearchParams(params)
  }

  const facets = query.data?.facets

  const renderPaginationInfo = (data: SearchResponse | undefined) => {
    if (!data) return null

    const start = (data.page - 1) * data.page_size + 1
    const end = Math.min(data.page * data.page_size, data.total)

    return (
      <div className="text-xs text-slate-300">
        Showing {start}-{end} of {data.total} results · page {data.page}
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <PageHeader title="Search" subtitle="Search extracted records with filtering and ranking." />

      <SearchBar
        q={searchForm.q}
        sort={searchForm.sort}
        onSubmit={(values) =>
          updateSearchForm({
            ...searchForm,
            ...values,
          })
        }
      />

      <div className="grid gap-4 lg:grid-cols-[2fr_1fr]">
        <SearchFilters
          values={searchForm}
          onApply={updateSearchForm}
          onClear={resetFilters}
        />

        <div className="space-y-4">
          <SearchFacets facets={facets} filters={searchForm} onFacetUpdate={onFacetUpdate} />

          <Card>
            <div className="flex flex-wrap items-center justify-between gap-2 text-sm">
              <div>
                <label htmlFor="page-size" className="mr-2 text-slate-300">
                  Page size:
                </label>
                <input
                  id="page-size"
                  type="number"
                  min={1}
                  value={pageSize}
                  onChange={(event) => setPageSize(Math.max(1, Number(event.target.value || 1)))}
                  className="w-24 rounded border border-slate-700 bg-slate-900/80 px-2 py-1 text-sm"
                />
              </div>
              <div className="text-xs text-slate-300">{renderPaginationInfo(query.data)}</div>
            </div>
          </Card>
        </div>
      </div>

      {!hasCriteria ? (
        <Card>
          <p className="text-sm text-slate-300">Set search terms and filters, then run a search.</p>
        </Card>
      ) : null}

      <Card title="Results">
        <SearchResultsTable
          results={hasCriteria ? query.data?.results || [] : []}
          isLoading={query.isLoading}
          onOpenRecord={setSelectedRecord}
        />

        {query.data ? (
          <div className="mt-4 flex items-center justify-end gap-3 text-sm">
            <Button
              tone="ghost"
              disabled={page <= 1 || query.isLoading}
              onClick={() => setPage(Math.max(1, page - 1))}
            >
              Previous
            </Button>
            <span className="text-slate-300">Page {page}</span>
            <Button
              tone="ghost"
              disabled={!query.data || page * query.data.page_size >= query.data.total || query.isLoading}
              onClick={() => setPage(page + 1)}
            >
              Next
            </Button>
          </div>
        ) : null}

        {query.isError ? <p className="mt-3 text-sm text-rose-300">Search request failed.</p> : null}
      </Card>

      <RecordPreviewDrawer
        record={selectedRecord}
        onClose={() => setSelectedRecord(null)}
      />
    </div>
  )
}

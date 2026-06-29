import { FormEvent, useEffect, useState } from 'react'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'

export type SearchFiltersValues = {
  q: string
  detectedFileType: string
  extension: string
  recordType: string
  dateFrom: string
  dateTo: string
  ip: string
  email: string
  domain: string
  sort: 'relevance' | 'created_at'
}

type SearchFiltersProps = {
  values: SearchFiltersValues
  onApply: (values: SearchFiltersValues) => void
  onClear: () => void
}

export const SearchFilters = ({ values, onApply, onClear }: SearchFiltersProps) => {
  const [draft, setDraft] = useState(values)

  useEffect(() => {
    setDraft(values)
  }, [values])

  const onSubmit = (event: FormEvent) => {
    event.preventDefault()
    onApply(draft)
  }

  const updateField = (field: keyof SearchFiltersValues, value: string) => {
    setDraft((previous) => ({
      ...previous,
      [field]: value,
    }))
  }

  return (
    <form className="panel p-4" onSubmit={onSubmit}>
      <div className="mb-3 grid gap-3 md:grid-cols-2 xl:grid-cols-3">
        <label className="space-y-1">
          <span className="text-xs text-slate-300">File type</span>
          <Input
            value={draft.detectedFileType}
            onChange={(event) => updateField('detectedFileType', event.target.value)}
            placeholder="detected_file_type"
          />
        </label>

        <label className="space-y-1">
          <span className="text-xs text-slate-300">Extension</span>
          <Input
            value={draft.extension}
            onChange={(event) => updateField('extension', event.target.value)}
            placeholder=".json"
          />
        </label>

        <label className="space-y-1">
          <span className="text-xs text-slate-300">Record type</span>
          <Input
            value={draft.recordType}
            onChange={(event) => updateField('recordType', event.target.value)}
            placeholder="record_type"
          />
        </label>

        <label className="space-y-1">
          <span className="text-xs text-slate-300">From date</span>
          <Input
            type="date"
            value={draft.dateFrom}
            onChange={(event) => updateField('dateFrom', event.target.value)}
          />
        </label>

        <label className="space-y-1">
          <span className="text-xs text-slate-300">To date</span>
          <Input type="date" value={draft.dateTo} onChange={(event) => updateField('dateTo', event.target.value)} />
        </label>

        <label className="space-y-1">
          <span className="text-xs text-slate-300">IP</span>
          <Input value={draft.ip} onChange={(event) => updateField('ip', event.target.value)} placeholder="192.168.1.1" />
        </label>

        <label className="space-y-1">
          <span className="text-xs text-slate-300">Email</span>
          <Input value={draft.email} onChange={(event) => updateField('email', event.target.value)} placeholder="user@example.com" />
        </label>

        <label className="space-y-1">
          <span className="text-xs text-slate-300">Domain</span>
          <Input value={draft.domain} onChange={(event) => updateField('domain', event.target.value)} placeholder="example.com" />
        </label>
      </div>

      <div className="flex gap-3">
        <Button type="submit">Apply filters</Button>
        <Button type="button" tone="ghost" onClick={onClear}>
          Clear filters
        </Button>
      </div>
    </form>
  )
}

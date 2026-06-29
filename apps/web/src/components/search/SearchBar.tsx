import { FormEvent, useEffect, useState } from 'react'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'

type SearchBarProps = {
  q: string
  sort: 'relevance' | 'created_at'
  onSubmit: (values: { q: string; sort: 'relevance' | 'created_at' }) => void
}

export const SearchBar = ({ q, sort, onSubmit }: SearchBarProps) => {
  const [keyword, setKeyword] = useState(q)
  const [mode, setMode] = useState(sort)

  useEffect(() => {
    setKeyword(q)
    setMode(sort)
  }, [q, sort])

  const onFormSubmit = (event: FormEvent) => {
    event.preventDefault()
    onSubmit({
      q: keyword.trim(),
      sort: mode,
    })
  }

  return (
    <form onSubmit={onFormSubmit} className="panel p-4">
      <div className="grid gap-3 md:grid-cols-[1fr_auto_auto]">
        <div>
          <label className="mb-1 block text-xs text-slate-300" htmlFor="search-q">
            Keyword
          </label>
          <Input
            id="search-q"
            value={keyword}
            onChange={(event) => setKeyword(event.target.value)}
            placeholder="Search content text..."
          />
        </div>
        <div>
          <label className="mb-1 block text-xs text-slate-300" htmlFor="search-sort">
            Sort
          </label>
          <select
            id="search-sort"
            value={mode}
            onChange={(event) => setMode(event.target.value as 'relevance' | 'created_at')}
            className="w-full rounded border border-slate-700 bg-slate-900/80 px-3 py-2 text-sm"
          >
            <option value="relevance">relevance</option>
            <option value="created_at">created_at</option>
          </select>
        </div>
        <div className="flex items-end">
          <Button type="submit" className="w-full">
            Search
          </Button>
        </div>
      </div>
    </form>
  )
}

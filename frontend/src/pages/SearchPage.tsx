import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import {
  AlertCircle,
  ExternalLink,
  Filter,
  RefreshCw,
  Search,
  Sparkles,
} from '../components/Icon'
import { Card } from '../components/Card'
import { Button } from '../components/Button'
import { Badge } from '../components/Badge'
import { EmptyState } from '../components/EmptyState'
import { Skeleton } from '../components/Skeleton'
import { FileTypeIcon } from '../components/FileTypeIcon'
import { api } from '../api'
import type { SearchItem } from '../types'

const EXAMPLE_QUERIES = ['invoice', 'contract', 'log error', 'customer name']
const FILTER_CHIPS = ['All types', 'Any time', 'Top results']
const SEARCH_DEBOUNCE_MS = 250
const SNIPPET_LIMIT = 200
const SNIPPET_CTX_DEFAULT = 120

function escapeHtml(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

function makeSnippet(
  content: string,
  query: string,
  ctx: number = SNIPPET_CTX_DEFAULT,
): { html: string } {
  const trimmed = query.trim()
  if (trimmed.length === 0) {
    return { html: escapeHtml(content.slice(0, SNIPPET_LIMIT)) }
  }

  const lowerContent = content.toLowerCase()
  const lowerNeedle = trimmed.toLowerCase()
  const matchIdx = lowerContent.indexOf(lowerNeedle)

  if (matchIdx === -1) {
    return { html: escapeHtml(content.slice(0, SNIPPET_LIMIT)) }
  }

  const start = Math.max(0, matchIdx - ctx)
  const end = Math.min(content.length, matchIdx + trimmed.length + ctx)
  const prefix = start > 0 ? '…' : ''
  const suffix = end < content.length ? '…' : ''

  const before = escapeHtml(content.slice(start, matchIdx))
  const matched = escapeHtml(content.slice(matchIdx, matchIdx + trimmed.length))
  const after = escapeHtml(content.slice(matchIdx + trimmed.length, end))

  return {
    html: `${prefix}${before}<mark>${matched}</mark>${after}${suffix}`,
  }
}

export default function SearchPage() {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<SearchItem[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [hasSearched, setHasSearched] = useState(false)

  const doSearch = useCallback(async (q: string): Promise<void> => {
    const trimmed = q.trim()
    if (trimmed.length === 0) {
      setResults([])
      setError(null)
      setHasSearched(false)
      return
    }

    setLoading(true)
    setError(null)
    try {
      const data = await api.search(trimmed, 25)
      setResults(data)
      setHasSearched(true)
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Search failed'
      setError(message)
      setResults([])
      setHasSearched(true)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    const trimmed = query.trim()
    if (trimmed.length === 0) {
      setResults([])
      setError(null)
      setHasSearched(false)
      return
    }

    const handle = window.setTimeout(() => {
      void doSearch(query)
    }, SEARCH_DEBOUNCE_MS)
    return () => {
      window.clearTimeout(handle)
    }
  }, [query, doSearch])

  const clearQuery = useCallback(() => {
    setQuery('')
    setResults([])
    setError(null)
    setHasSearched(false)
  }, [])

  const reload = useCallback(() => {
    void doSearch(query)
  }, [doSearch, query])

  const showInitialEmpty =
    query.length === 0 && results.length === 0 && !hasSearched
  const showLoading = loading && results.length === 0
  const showError = !loading && error !== null
  const showNoResults =
    !loading && error === null && hasSearched && results.length === 0
  const showResults = !loading && error === null && results.length > 0

  return (
    <div className="stack-lg">
      <header>
        <h2>Search parsed content</h2>
        <p className="muted">
          Find invoices, contracts, logs, and more across ingested files.
        </p>
      </header>

      <Card padding="lg">
        <form
          className="stack"
          onSubmit={(e) => {
            e.preventDefault()
            void doSearch(query)
          }}
        >
          <div className="search-hero-input">
            <Search size={20} className="muted" aria-hidden="true" />
            <input
              type="search"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Search parsed content… e.g. invoice number, customer name"
              aria-label="Search parsed content"
            />
          </div>

          <div className="row" style={{ flexWrap: 'wrap', gap: 'var(--space-2)' }}>
            {FILTER_CHIPS.map((label) => (
              <button
                key={label}
                type="button"
                className="chip chip-muted"
                disabled
                aria-disabled="true"
              >
                <Filter size={12} aria-hidden="true" />
                {label}
              </button>
            ))}
          </div>

          <p className="muted text-sm">
            Tip: try <em>invoice</em>, <em>contract</em>, or{' '}
            <em>&ldquo;error 500&rdquo;</em>.
          </p>

          {query.length === 0 ? (
            <div className="row" style={{ flexWrap: 'wrap', gap: 'var(--space-2)' }}>
              <span className="muted text-sm">Try:</span>
              {EXAMPLE_QUERIES.map((example) => (
                <button
                  key={example}
                  type="button"
                  className="chip"
                  onClick={() => setQuery(example)}
                >
                  {example}
                </button>
              ))}
            </div>
          ) : null}
        </form>
      </Card>

      {showInitialEmpty ? (
        <Card>
          <EmptyState
            icon={Sparkles}
            title="Search across your parsed content"
            description="Try searching for invoices, contracts, log errors, or customer names. Results appear as you type."
          />
        </Card>
      ) : null}

      {showLoading ? (
        <Card>
          <div className="stack-sm">
            <Skeleton height={120} />
            <Skeleton height={120} />
            <Skeleton height={120} />
          </div>
        </Card>
      ) : null}

      {showError ? (
        <Card>
          <EmptyState
            icon={AlertCircle}
            title="Search failed"
            description={error ?? 'Unknown error'}
            action={
              <Button
                onClick={reload}
                variant="secondary"
                iconLeft={<RefreshCw size={14} />}
              >
                Retry
              </Button>
            }
          />
        </Card>
      ) : null}

      {showNoResults ? (
        <Card>
          <EmptyState
            icon={Search}
            title={`No matches for "${query}"`}
            description="Try a different keyword or remove filters."
            action={
              <Button onClick={clearQuery} variant="ghost">
                Clear search
              </Button>
            }
          />
        </Card>
      ) : null}

      {showResults ? (
        <div className="stack">
          {results.map((item) => {
            const snippet = makeSnippet(item.content, query)
            const matchPct = `${(item.score * 100).toFixed(0)}% match`
            return (
              <Card
                key={item.index_id}
                className="search-result-card"
                padding="md"
              >
                <div className="stack-sm">
                  <div className="row-between">
                    <div
                      className="row"
                      style={{ minWidth: 0, gap: 'var(--space-3)' }}
                    >
                      <FileTypeIcon filename={item.filename} size={20} />
                      <span className="truncate" style={{ fontWeight: 600 }}>
                        {item.filename}
                      </span>
                    </div>
                    <Badge variant="accent" size="sm">
                      {matchPct}
                    </Badge>
                  </div>

                  <div className="muted font-mono text-sm truncate">
                    {item.path}
                  </div>

                  <p
                    className="search-result-snippet text-sm"
                    dangerouslySetInnerHTML={{ __html: snippet.html }}
                  />

                  <div className="row-between">
                    <span className="muted text-sm">
                      File: {item.file_id.slice(0, 8)}
                    </span>
                    <Link to="/files">
                      <Button
                        variant="ghost"
                        size="sm"
                        iconRight={<ExternalLink size={12} />}
                      >
                        Open in Files
                      </Button>
                    </Link>
                  </div>
                </div>
              </Card>
            )
          })}
        </div>
      ) : null}
    </div>
  )
}

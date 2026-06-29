import { useState } from 'react'
import { api } from '../api'
import type { SearchItem } from '../types'

export default function SearchPage() {
  const [query, setQuery] = useState('invoice')
  const [results, setResults] = useState<SearchItem[]>([])
  const [loading, setLoading] = useState(false)

  const doSearch = async () => {
    setLoading(true)
    try {
      const result = await api.search(query, 50)
      setResults(result as SearchItem[])
    } catch {
      setResults([])
    }
    setLoading(false)
  }

  return (
    <section className="card">
      <h2>Search parsed content</h2>
      <div className="form">
        <input value={query} onChange={(e) => setQuery(e.target.value)} />
        <button onClick={doSearch}>Search</button>
      </div>
      {loading ? <p>Searching...</p> : null}
      <ul>
        {results.map((item) => (
          <li key={item.index_id}>
            <div>
              <strong>{item.filename}</strong> (score: {item.score.toFixed(2)})
            </div>
            <div className="muted">{item.path}</div>
            <p>{item.content.slice(0, 220)}</p>
          </li>
        ))}
      </ul>
    </section>
  )
}


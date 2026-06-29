import { useEffect, useState } from 'react'
import { api } from '../api'
import type { DashboardSummary } from '../types'

export default function Dashboard() {
  const [summary, setSummary] = useState<DashboardSummary | null>(null)
  const [error, setError] = useState<string>('')

  useEffect(() => {
    const fetchData = async () => {
      try {
        setSummary(await api.dashboardSummary())
      } catch (e) {
        setError((e as Error).message || 'failed to fetch dashboard')
      }
    }
    fetchData()
  }, [])

  if (error) return <p className="error">Error: {error}</p>
  if (!summary) return <p>Loading...</p>

  return (
    <section className="card">
      <h2>Dashboard</h2>
      <div className="grid">
        <article>
          <h3>Uploads</h3>
          <div className="metric">{summary.uploads}</div>
        </article>
        <article>
          <h3>Files</h3>
          <div className="metric">{summary.files}</div>
        </article>
        <article>
          <h3>Parsed Records</h3>
          <div className="metric">{summary.parsed_records}</div>
        </article>
      </div>
      <h3>Jobs by status</h3>
      <ul>
        {Object.entries(summary.jobs_by_status).map(([status, count]) => (
          <li key={status}>
            {status}: {count}
          </li>
        ))}
      </ul>
      <h3>Recent jobs</h3>
      <ul>
        {summary.recent_jobs.map((job) => (
          <li key={job.id}>
            {job.id.slice(0, 8)} → {job.status} (attempts: {job.attempt_count})
          </li>
        ))}
      </ul>
    </section>
  )
}


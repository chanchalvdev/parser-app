import { useEffect, useState } from 'react'
import { api } from '../api'
import type { Job } from '../types'

type Row = Job & { upload_id?: string }

export default function JobsPage() {
  const [jobs, setJobs] = useState<Row[]>([])
  const [passwordInput, setPasswordInput] = useState('')
  const [retryJobId, setRetryJobId] = useState('')
  const [message, setMessage] = useState('')

  useEffect(() => {
    const timer = setInterval(async () => {
      try {
        setJobs(await api.listJobs('?limit=100'))
      } catch {
        // ignore
      }
    }, 3000)
    ;(async () => {
      setJobs(await api.listJobs('?limit=100'))
    })()
    return () => clearInterval(timer)
  }, [])

  const retry = async () => {
    try {
      await api.retryWithPassword(retryJobId, passwordInput)
      setMessage('Retry request accepted')
    } catch (e) {
      setMessage((e as Error).message)
    }
  }

  return (
    <section className="card">
      <h2>Job Monitoring</h2>
      <div className="form" style={{ marginBottom: 16 }}>
        <label>
          Job ID
          <input value={retryJobId} onChange={(e) => setRetryJobId(e.target.value)} />
        </label>
        <label>
          Password
          <input value={passwordInput} onChange={(e) => setPasswordInput(e.target.value)} />
        </label>
        <button onClick={retry}>Retry with Password</button>
      </div>
      {message && <p>{message}</p>}
      <table>
        <thead>
          <tr><th>Job</th><th>Upload</th><th>Status</th><th>Stage</th><th>Attempts</th><th>Error</th></tr>
        </thead>
        <tbody>
          {jobs.map((job) => (
            <tr key={job.id}>
              <td>{job.id.slice(0, 8)}</td>
              <td>{job.upload_id?.slice(0, 8)}</td>
              <td>{job.status}</td>
              <td>{job.stage}</td>
              <td>{job.attempt_count}</td>
              <td>{job.error_message || '-'}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  )
}


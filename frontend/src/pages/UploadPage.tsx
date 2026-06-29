import { useState } from 'react'
import { api } from '../api'

export default function UploadPage() {
  const [file, setFile] = useState<File | null>(null)
  const [password, setPassword] = useState('')
  const [result, setResult] = useState('')
  const [error, setError] = useState('')

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!file) {
      setError('select a file')
      return
    }
    try {
      const payload = await api.uploadFile(file, password || undefined)
      setResult(`Upload queued. upload=${payload.upload_id}, job=${payload.job_id}, status=${payload.status}`)
      setError('')
    } catch (e) {
      setError((e as Error).message)
      setResult('')
    }
  }

  return (
    <section className="card">
      <h2>Upload file</h2>
      <form onSubmit={onSubmit} className="form">
        <label>
          File
          <input type="file" onChange={(e) => setFile(e.target.files?.[0] || null)} />
        </label>
        <label>
          Archive password (optional)
          <input value={password} onChange={(e) => setPassword(e.target.value)} type="text" placeholder="zip/rar/7z password" />
        </label>
        <button type="submit">Upload</button>
      </form>
      {result && <p className="success">{result}</p>}
      {error && <p className="error">{error}</p>}
    </section>
  )
}


import { useEffect, useState } from 'react'
import { api } from '../api'

export default function SettingsPage() {
  const [settings, setSettings] = useState<Record<string, any>>({})
  const [maxUpload, setMaxUpload] = useState('')

  useEffect(() => {
    const load = async () => {
      const value = await api.settings()
      setSettings(value)
      setMaxUpload(String(value.max_upload_size_mb || ''))
    }
    load()
  }, [])

  const save = async () => {
    const payload = {
      ...settings,
      max_upload_size_mb: Number(maxUpload || 0),
    }
    const response = await api.updateSettings(payload)
    setSettings(response as Record<string, any>)
  }

  return (
    <section className="card">
      <h2>Admin settings</h2>
      <div className="form">
        <label>
          Max upload size MB
          <input value={maxUpload} onChange={(e) => setMaxUpload(e.target.value)} />
        </label>
        <button onClick={save}>Save</button>
      </div>
      <p className="muted">Placeholder-only admin settings are stored in process memory.</p>
    </section>
  )
}

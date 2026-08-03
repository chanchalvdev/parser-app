import { useCallback, useEffect, useMemo, useRef, useState, type ReactNode } from 'react'
import {
  Link,
  Route,
  Routes,
  useNavigate,
  useParams,
} from 'react-router-dom'

/* ------------------------------------------------------------------ *
 * Minimal tenantless file-ingestion console.
 *
 * Flow:
 *   1. Upload a file (optionally password-protected).
 *   2. A job is created; we poll its progress until it finishes.
 *   3. When done, every parsed record for that job is shown in one
 *      searchable / copyable table.
 * ------------------------------------------------------------------ */

const API_BASE = (
  import.meta.env.VITE_API_BASE_URL || 'http://localhost:8088/api/v1'
).replace(/\/$/, '')

// ---------------------------------------------------------------- types

interface Job {
  id: string
  root_file_id: string
  status: string
  current_stage?: string
  progress_percent: number
  retry_count: number
  error_code?: string
  error_message?: string
  created_at: string
  updated_at?: string
  completed_at?: string
}

interface JobListResponse {
  total: number
  page: number
  page_size: number
  jobs: Job[]
}

interface ParsedRecord {
  id: string
  file_id: string
  record_type?: string
  line_number?: number
  content_text?: string
  extracted_entities?: Record<string, string[]>
  created_at?: string
}

interface JobRecordsResponse {
  job_id: string
  total: number
  page: number
  page_size: number
  search: string
  records: ParsedRecord[]
}

interface FileInfo {
  id: string
  original_name: string
  processing_status: string
  is_password_protected: boolean
  size_bytes: number
}

// ---------------------------------------------------------------- api

async function apiJSON<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { 'Content-Type': 'application/json', ...(init?.headers || {}) },
    ...init,
  })
  if (!res.ok) {
    let detail = res.statusText
    try {
      const body = await res.json()
      detail = body?.error || body?.message || detail
    } catch {
      /* ignore */
    }
    throw new Error(detail)
  }
  if (res.status === 204) return null as T
  return (await res.json()) as T
}

const listJobs = (page = 1, pageSize = 50) =>
  apiJSON<JobListResponse>(`/jobs?page=${page}&page_size=${pageSize}`)

const getJob = (id: string) => apiJSON<Job>(`/jobs/${id}`)

const getFile = (id: string) => apiJSON<FileInfo>(`/files/${id}`)

const getJobRecords = (id: string, q: string, page: number, pageSize: number) =>
  apiJSON<JobRecordsResponse>(
    `/jobs/${id}/records?page=${page}&page_size=${pageSize}&q=${encodeURIComponent(q)}`,
  )

const submitPassword = (fileId: string, password: string) =>
  apiJSON<{ file_id: string; job_id: string; status: string }>(
    `/files/${fileId}/password`,
    { method: 'POST', body: JSON.stringify({ password }) },
  )

// Upload uses initiate -> PUT to presigned URL -> complete.
async function uploadFile(
  file: File,
  passwordProvided: boolean,
  onProgress: (pct: number) => void,
): Promise<{ file_id: string; job_id: string }> {
  const init = await apiJSON<{
    upload_id: string
    upload_url: string
  }>(`/uploads/initiate`, {
    method: 'POST',
    body: JSON.stringify({
      file_name: file.name,
      content_type: file.type || 'application/octet-stream',
      size_bytes: file.size,
      password_provided: passwordProvided,
    }),
  })

  await putWithProgress(init.upload_url, file, onProgress)

  const complete = await apiJSON<{ file_id: string; job_id: string }>(
    `/uploads/complete`,
    { method: 'POST', body: JSON.stringify({ upload_id: init.upload_id }) },
  )
  return complete
}

function putWithProgress(
  url: string,
  file: File,
  onProgress: (pct: number) => void,
): Promise<void> {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest()
    xhr.open('PUT', url)
    xhr.upload.onprogress = (e) => {
      if (e.lengthComputable) onProgress(Math.round((e.loaded / e.total) * 100))
    }
    xhr.onload = () =>
      xhr.status >= 200 && xhr.status < 300
        ? resolve()
        : reject(new Error(`storage upload failed (${xhr.status})`))
    xhr.onerror = () => reject(new Error('storage upload network error'))
    xhr.send(file)
  })
}

// ---------------------------------------------------------------- helpers

const TERMINAL = new Set(['completed', 'failed', 'succeeded'])
const isTerminal = (s: string) => TERMINAL.has(s.toLowerCase())

function statusColor(status: string): string {
  const s = status.toLowerCase()
  if (s === 'completed' || s === 'succeeded') return 'text-emerald-400 border-emerald-500/40 bg-emerald-500/10'
  if (s === 'failed') return 'text-rose-400 border-rose-500/40 bg-rose-500/10'
  if (s.includes('password')) return 'text-amber-400 border-amber-500/40 bg-amber-500/10'
  return 'text-sky-400 border-sky-500/40 bg-sky-500/10'
}

function StatusBadge({ status }: { status: string }) {
  return (
    <span className={`inline-block rounded-full border px-2.5 py-0.5 text-xs font-medium ${statusColor(status)}`}>
      {status}
    </span>
  )
}

function timeAgo(iso?: string): string {
  if (!iso) return '—'
  const d = new Date(iso).getTime()
  if (Number.isNaN(d)) return '—'
  const s = Math.floor((Date.now() - d) / 1000)
  if (s < 60) return `${s}s ago`
  if (s < 3600) return `${Math.floor(s / 60)}m ago`
  if (s < 86400) return `${Math.floor(s / 3600)}h ago`
  return new Date(iso).toLocaleString()
}

async function copyText(text: string) {
  try {
    await navigator.clipboard.writeText(text)
  } catch {
    const ta = document.createElement('textarea')
    ta.value = text
    document.body.appendChild(ta)
    ta.select()
    document.execCommand('copy')
    document.body.removeChild(ta)
  }
}

// ---------------------------------------------------------------- shell

function Shell({ children }: { children: ReactNode }) {
  return (
    <div className="min-h-screen bg-[#05070a] text-slate-200">
      <header className="border-b border-slate-700/40 bg-[#0b0f16]">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
          <Link to="/" className="flex items-center gap-2">
            <span className="text-lg font-bold tracking-tight text-emerald-400" style={{ fontFamily: 'Space Grotesk, sans-serif' }}>
              ⛃ File Ingestion
            </span>
            <span className="text-xs text-slate-500">parse · index · search</span>
          </Link>
          <Link
            to="/"
            className="rounded-md border border-slate-700 px-3 py-1.5 text-sm text-slate-300 hover:border-emerald-500/50 hover:text-emerald-300"
          >
            Jobs
          </Link>
        </div>
      </header>
      <main className="mx-auto max-w-6xl px-6 py-8">{children}</main>
    </div>
  )
}

// ---------------------------------------------------------------- upload

function UploadCard({ onUploaded }: { onUploaded: () => void }) {
  const navigate = useNavigate()
  const [file, setFile] = useState<File | null>(null)
  const [password, setPassword] = useState('')
  const [phase, setPhase] = useState<'idle' | 'uploading' | 'completing'>('idle')
  const [pct, setPct] = useState(0)
  const [error, setError] = useState<string | null>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  const submit = async () => {
    if (!file) return
    setError(null)
    setPhase('uploading')
    setPct(0)
    try {
      const res = await uploadFile(file, password.trim().length > 0, setPct)
      setPhase('completing')
      // Stash the password so the job page can auto-submit it if the
      // archive turns out to be password-protected.
      if (password.trim()) {
        sessionStorage.setItem(`pw:${res.job_id}`, password.trim())
      }
      onUploaded()
      navigate(`/jobs/${res.job_id}`)
    } catch (e) {
      setError((e as Error).message)
      setPhase('idle')
    }
  }

  return (
    <div className="rounded-xl border border-slate-700/50 bg-[#0d1119] p-6">
      <h2 className="mb-1 text-lg font-semibold text-slate-100">Upload a file</h2>
      <p className="mb-5 text-sm text-slate-400">
        Logs, JSON, CSV, XML, PDF, Excel, or archives (.zip / .rar / .tar / .7z) —
        password-protected or not. It is parsed into records you can search.
      </p>

      <div
        onClick={() => inputRef.current?.click()}
        onDrop={(e) => {
          e.preventDefault()
          if (e.dataTransfer.files?.[0]) setFile(e.dataTransfer.files[0])
        }}
        onDragOver={(e) => e.preventDefault()}
        className="mb-4 cursor-pointer rounded-lg border border-dashed border-slate-600 px-4 py-8 text-center hover:border-emerald-500/60"
      >
        <input
          ref={inputRef}
          type="file"
          className="hidden"
          onChange={(e) => setFile(e.target.files?.[0] ?? null)}
        />
        {file ? (
          <div className="text-sm">
            <div className="font-medium text-emerald-300">{file.name}</div>
            <div className="text-slate-500">{(file.size / 1024).toFixed(1)} KB</div>
          </div>
        ) : (
          <div className="text-sm text-slate-400">
            Drop a file here or <span className="text-emerald-400">browse</span>
          </div>
        )}
      </div>

      <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-slate-500">
        Password (optional — for encrypted archives)
      </label>
      <input
        type="password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        placeholder="Leave blank if not protected"
        className="mb-4 w-full rounded-md border border-slate-700 bg-[#10151f] px-3 py-2 text-sm text-slate-200 outline-none focus:border-emerald-500/60"
      />

      {phase === 'uploading' && (
        <div className="mb-4">
          <div className="mb-1 flex justify-between text-xs text-slate-400">
            <span>Uploading…</span>
            <span>{pct}%</span>
          </div>
          <div className="h-2 overflow-hidden rounded-full bg-slate-800">
            <div className="h-full bg-emerald-500 transition-all" style={{ width: `${pct}%` }} />
          </div>
        </div>
      )}

      {error && <div className="mb-4 rounded-md border border-rose-500/40 bg-rose-500/10 px-3 py-2 text-sm text-rose-300">{error}</div>}

      <button
        disabled={!file || phase !== 'idle'}
        onClick={submit}
        className="rounded-md bg-emerald-500 px-4 py-2 text-sm font-semibold text-slate-950 disabled:cursor-not-allowed disabled:opacity-40 hover:bg-emerald-400"
      >
        {phase === 'idle' ? 'Upload & parse' : phase === 'uploading' ? 'Uploading…' : 'Starting…'}
      </button>
    </div>
  )
}

// ---------------------------------------------------------------- jobs list

function JobsList({ refreshKey }: { refreshKey: number }) {
  const [data, setData] = useState<JobListResponse | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let alive = true
    const load = () =>
      listJobs()
        .then((d) => alive && setData(d))
        .catch((e) => alive && setError((e as Error).message))
    load()
    // Poll so in-flight jobs update their status/progress in the list.
    const t = setInterval(load, 2500)
    return () => {
      alive = false
      clearInterval(t)
    }
  }, [refreshKey])

  if (error) return <div className="text-sm text-rose-400">Failed to load jobs: {error}</div>
  if (!data) return <div className="text-sm text-slate-500">Loading jobs…</div>
  if (data.jobs.length === 0)
    return <div className="rounded-lg border border-slate-700/50 bg-[#0d1119] p-6 text-sm text-slate-400">No jobs yet. Upload a file to get started.</div>

  return (
    <div className="overflow-hidden rounded-xl border border-slate-700/50 bg-[#0d1119]">
      <table className="w-full text-sm">
        <thead className="bg-[#10151f] text-left text-xs uppercase tracking-wide text-slate-500">
          <tr>
            <th className="px-4 py-3">Job</th>
            <th className="px-4 py-3">Status</th>
            <th className="px-4 py-3">Progress</th>
            <th className="px-4 py-3">Created</th>
            <th className="px-4 py-3"></th>
          </tr>
        </thead>
        <tbody>
          {data.jobs.map((j) => (
            <tr key={j.id} className="border-t border-slate-800/70 hover:bg-slate-800/20">
              <td className="px-4 py-3 font-mono text-xs text-slate-400">{j.id.slice(0, 8)}</td>
              <td className="px-4 py-3"><StatusBadge status={j.status} /></td>
              <td className="px-4 py-3 w-40">
                <div className="flex items-center gap-2">
                  <div className="h-1.5 flex-1 overflow-hidden rounded-full bg-slate-800">
                    <div className={`h-full ${j.status === 'failed' ? 'bg-rose-500' : 'bg-emerald-500'}`} style={{ width: `${j.progress_percent}%` }} />
                  </div>
                  <span className="w-9 text-right text-xs text-slate-500">{j.progress_percent}%</span>
                </div>
              </td>
              <td className="px-4 py-3 text-slate-400">{timeAgo(j.created_at)}</td>
              <td className="px-4 py-3 text-right">
                <Link to={`/jobs/${j.id}`} className="text-emerald-400 hover:text-emerald-300">View data →</Link>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

// ---------------------------------------------------------------- home

function HomePage() {
  const [refreshKey, setRefreshKey] = useState(0)
  return (
    <div className="grid gap-8 md:grid-cols-[1fr_1.2fr]">
      <UploadCard onUploaded={() => setRefreshKey((k) => k + 1)} />
      <div>
        <h2 className="mb-4 text-lg font-semibold text-slate-100">Jobs</h2>
        <JobsList refreshKey={refreshKey} />
      </div>
    </div>
  )
}

// ---------------------------------------------------------------- job detail

function PasswordPrompt({ fileId, wrong, onSubmitted }: { fileId: string; wrong: boolean; onSubmitted: () => void }) {
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)
  const submit = async () => {
    if (!password.trim()) return
    setBusy(true)
    setError(null)
    try {
      await submitPassword(fileId, password.trim())
      onSubmitted()
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setBusy(false)
    }
  }
  return (
    <div className="rounded-lg border border-amber-500/40 bg-amber-500/10 p-4">
      <div className="mb-2 text-sm font-medium text-amber-300">
        {wrong ? 'Incorrect password — try again' : 'This archive is password-protected'}
      </div>
      <div className="flex gap-2">
        <input
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="Enter archive password"
          className="flex-1 rounded-md border border-slate-700 bg-[#10151f] px-3 py-2 text-sm text-slate-200 outline-none focus:border-amber-500/60"
        />
        <button onClick={submit} disabled={busy} className="rounded-md bg-amber-500 px-4 py-2 text-sm font-semibold text-slate-950 disabled:opacity-40 hover:bg-amber-400">
          {busy ? 'Submitting…' : 'Unlock'}
        </button>
      </div>
      {error && <div className="mt-2 text-xs text-rose-300">{error}</div>}
    </div>
  )
}

function EntityChips({ entities }: { entities?: Record<string, string[]> }) {
  const flat = useMemo(() => {
    if (!entities) return [] as string[]
    const out: string[] = []
    for (const [k, vals] of Object.entries(entities)) {
      for (const v of vals || []) out.push(`${k.replace(/_/g, ' ')}: ${v}`)
    }
    return out.slice(0, 6)
  }, [entities])
  if (flat.length === 0) return <span className="text-slate-600">—</span>
  return (
    <div className="flex flex-wrap gap-1">
      {flat.map((t, i) => (
        <span key={i} className="rounded bg-slate-800 px-1.5 py-0.5 text-[10px] text-slate-400">{t}</span>
      ))}
    </div>
  )
}

const PAGE_SIZE = 100

function JobDetailPage() {
  const { id = '' } = useParams()
  const [job, setJob] = useState<Job | null>(null)
  const [rootFile, setRootFile] = useState<FileInfo | null>(null)
  const [jobError, setJobError] = useState<string | null>(null)

  const [records, setRecords] = useState<JobRecordsResponse | null>(null)
  const [search, setSearch] = useState('')
  const [debounced, setDebounced] = useState('')
  const [page, setPage] = useState(1)
  const autoTriedRef = useRef(false)

  // Poll the job + its root file until the job reaches a terminal state.
  useEffect(() => {
    let alive = true
    let timer: ReturnType<typeof setTimeout>
    const tick = async () => {
      try {
        const j = await getJob(id)
        if (!alive) return
        setJob(j)
        try {
          const f = await getFile(j.root_file_id)
          if (alive) setRootFile(f)
        } catch {
          /* file may not be readable yet */
        }
        // Auto-submit a stashed password once, if the archive needs one.
        const needsPw = j.root_file_id && (await fileNeedsPassword(j.root_file_id))
        if (needsPw && !autoTriedRef.current) {
          autoTriedRef.current = true
          const stashed = sessionStorage.getItem(`pw:${id}`)
          if (stashed) {
            try {
              await submitPassword(j.root_file_id, stashed)
            } catch {
              /* fall through to manual prompt */
            }
          }
        }
        if (!isTerminal(j.status)) timer = setTimeout(tick, 1800)
      } catch (e) {
        if (alive) setJobError((e as Error).message)
      }
    }
    tick()
    return () => {
      alive = false
      clearTimeout(timer)
    }
  }, [id])

  // Debounce the search box.
  useEffect(() => {
    const t = setTimeout(() => {
      setDebounced(search)
      setPage(1)
    }, 300)
    return () => clearTimeout(t)
  }, [search])

  // Load records. Re-run while the job is active so the table fills in live.
  const loadRecords = useCallback(() => {
    getJobRecords(id, debounced, page, PAGE_SIZE)
      .then(setRecords)
      .catch(() => setRecords(null))
  }, [id, debounced, page])

  useEffect(() => {
    loadRecords()
    if (job && !isTerminal(job.status)) {
      const t = setInterval(loadRecords, 3000)
      return () => clearInterval(t)
    }
  }, [loadRecords, job])

  const passwordRequired =
    rootFile?.processing_status?.toLowerCase() === 'password_required' ||
    rootFile?.processing_status?.toLowerCase() === 'wrong_password'

  const copyAll = async () => {
    if (!records) return
    const text = records.records.map((r) => r.content_text ?? '').join('\n')
    await copyText(text)
  }

  const totalPages = records ? Math.max(1, Math.ceil(records.total / PAGE_SIZE)) : 1

  return (
    <div>
      <Link to="/" className="mb-4 inline-block text-sm text-slate-400 hover:text-emerald-300">← All jobs</Link>

      {jobError && <div className="mb-4 rounded-md border border-rose-500/40 bg-rose-500/10 px-3 py-2 text-sm text-rose-300">{jobError}</div>}

      {job && (
        <div className="mb-6 rounded-xl border border-slate-700/50 bg-[#0d1119] p-5">
          <div className="mb-3 flex items-center justify-between">
            <div>
              <div className="text-lg font-semibold text-slate-100">{rootFile?.original_name || 'File'}</div>
              <div className="font-mono text-xs text-slate-500">job {job.id}</div>
            </div>
            <StatusBadge status={job.status} />
          </div>

          <div className="mb-1 flex justify-between text-xs text-slate-400">
            <span>{job.current_stage || job.status}</span>
            <span>{job.progress_percent}%</span>
          </div>
          <div className="h-2 overflow-hidden rounded-full bg-slate-800">
            <div className={`h-full transition-all ${job.status === 'failed' ? 'bg-rose-500' : 'bg-emerald-500'}`} style={{ width: `${job.progress_percent}%` }} />
          </div>

          {job.error_message && (
            <div className="mt-3 rounded-md border border-rose-500/40 bg-rose-500/10 px-3 py-2 text-sm text-rose-300">
              {job.error_message}
            </div>
          )}

          {passwordRequired && (
            <div className="mt-4">
              <PasswordPrompt
                fileId={job.root_file_id}
                wrong={rootFile?.processing_status?.toLowerCase() === 'wrong_password'}
                onSubmitted={() => { autoTriedRef.current = true }}
              />
            </div>
          )}
        </div>
      )}

      {/* Records table */}
      <div className="mb-3 flex items-center justify-between gap-3">
        <input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search parsed data…"
          className="w-full max-w-md rounded-md border border-slate-700 bg-[#10151f] px-3 py-2 text-sm text-slate-200 outline-none focus:border-emerald-500/60"
        />
        <div className="flex items-center gap-3 whitespace-nowrap text-sm text-slate-400">
          <span>{records ? records.total.toLocaleString() : 0} records</span>
          <button onClick={copyAll} className="rounded-md border border-slate-700 px-3 py-1.5 text-xs hover:border-emerald-500/50 hover:text-emerald-300">Copy page</button>
        </div>
      </div>

      <div className="overflow-hidden rounded-xl border border-slate-700/50 bg-[#0d1119]">
        <div className="max-h-[60vh] overflow-auto">
          <table className="w-full text-sm">
            <thead className="sticky top-0 bg-[#10151f] text-left text-xs uppercase tracking-wide text-slate-500">
              <tr>
                <th className="w-14 px-3 py-2">#</th>
                <th className="px-3 py-2">Content</th>
                <th className="w-56 px-3 py-2">Entities</th>
                <th className="w-16 px-3 py-2"></th>
              </tr>
            </thead>
            <tbody>
              {records?.records.map((r, i) => (
                <tr key={r.id} className="border-t border-slate-800/70 align-top hover:bg-slate-800/20">
                  <td className="px-3 py-2 font-mono text-xs text-slate-600">{(page - 1) * PAGE_SIZE + i + 1}</td>
                  <td className="px-3 py-2">
                    <code className="whitespace-pre-wrap break-all font-mono text-xs text-slate-300">{r.content_text}</code>
                  </td>
                  <td className="px-3 py-2"><EntityChips entities={r.extracted_entities} /></td>
                  <td className="px-3 py-2 text-right">
                    <button onClick={() => copyText(r.content_text ?? '')} className="text-xs text-slate-500 hover:text-emerald-300">copy</button>
                  </td>
                </tr>
              ))}
              {records && records.records.length === 0 && (
                <tr>
                  <td colSpan={4} className="px-3 py-8 text-center text-sm text-slate-500">
                    {job && !isTerminal(job.status) ? 'Parsing… records will appear here.' : 'No records match.'}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      {records && records.total > PAGE_SIZE && (
        <div className="mt-3 flex items-center justify-center gap-4 text-sm text-slate-400">
          <button disabled={page <= 1} onClick={() => setPage((p) => p - 1)} className="rounded-md border border-slate-700 px-3 py-1 disabled:opacity-30">Prev</button>
          <span>Page {page} / {totalPages}</span>
          <button disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)} className="rounded-md border border-slate-700 px-3 py-1 disabled:opacity-30">Next</button>
        </div>
      )}
    </div>
  )
}

// A password-required file is signalled through its processing_status.
async function fileNeedsPassword(fileId: string): Promise<boolean> {
  try {
    const f = await getFile(fileId)
    const s = f.processing_status?.toLowerCase()
    return s === 'password_required' || s === 'wrong_password'
  } catch {
    return false
  }
}

// ---------------------------------------------------------------- app

export function App() {
  return (
    <Shell>
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/jobs/:id" element={<JobDetailPage />} />
        <Route path="*" element={<HomePage />} />
      </Routes>
    </Shell>
  )
}

export default App

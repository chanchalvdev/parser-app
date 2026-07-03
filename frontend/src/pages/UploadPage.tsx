import { useCallback, useEffect, useId, useMemo, useRef, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api'
import type { Settings, Upload } from '../types'
import { Button } from '../components/Button'
import { Card } from '../components/Card'
import { CopyButton } from '../components/CopyButton'
import { EmptyState } from '../components/EmptyState'
import { FileTypeIcon } from '../components/FileTypeIcon'
import {
  AlertCircle,
  ExternalLink,
  RefreshCw,
  Sparkles,
  Trash2,
  UploadCloud,
} from '../components/Icon'
import { Skeleton } from '../components/Skeleton'
import { StatusBadge } from '../components/StatusBadge'
import { useToast } from '../components/ToastProvider'

const ARCHIVE_EXTS = ['zip', 'rar', '7z', 'tar', 'gz', 'tgz', 'bz2', 'xz', 'zst'] as const

type PendingFile = { id: string; file: File }
type UploadResult = { filename: string; upload_id: string; job_id: string }

function formatBytes(n: number): string {
  if (!Number.isFinite(n) || n < 0) return '0 B'
  if (n < 1024) return `${n} B`
  const units = ['KB', 'MB', 'GB', 'TB']
  let value = n / 1024
  let i = 0
  while (value >= 1024 && i < units.length - 1) {
    value /= 1024
    i++
  }
  return `${value.toFixed(value >= 100 ? 0 : 1)} ${units[i]}`
}

function formatRelative(iso: string | undefined | null): string {
  if (!iso) return '—'
  const d = new Date(iso)
  const ms = Date.now() - d.getTime()
  if (Number.isNaN(ms)) return '—'
  if (ms < 0) return 'just now'
  const seconds = Math.floor(ms / 1000)
  if (seconds < 5) return 'just now'
  if (seconds >= 86400 * 30) return d.toLocaleDateString()
  if (seconds < 60) return `${seconds}s ago`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`
  return `${Math.floor(seconds / 86400)}d ago`
}

function getExtension(name: string): string {
  const idx = name.lastIndexOf('.')
  if (idx < 0 || idx === name.length - 1) return ''
  return name.slice(idx + 1).toLowerCase()
}

function makeId(): string {
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) {
    return crypto.randomUUID()
  }
  return `f-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`
}

export default function UploadPage() {
  const toast = useToast()
  const [files, setFiles] = useState<PendingFile[]>([])
  const [password, setPassword] = useState('')
  const [uploading, setUploading] = useState(false)
  const [results, setResults] = useState<UploadResult[]>([])
  const [recent, setRecent] = useState<Upload[] | null>(null)
  const [recentError, setRecentError] = useState<string | null>(null)
  const [recentLoading, setRecentLoading] = useState(true)
  const [settings, setSettings] = useState<Settings | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const inputId = useId()

  const loadRecent = useCallback(async () => {
    setRecentLoading(true)
    setRecentError(null)
    try {
      const data = await api.listUploads('?limit=5')
      setRecent(data)
    } catch (e) {
      setRecentError((e as Error).message || 'Failed to load recent uploads')
    } finally {
      setRecentLoading(false)
    }
  }, [])

  useEffect(() => {
    loadRecent()
  }, [loadRecent])

  useEffect(() => {
    let cancelled = false
    api
      .settings()
      .then((s) => {
        if (!cancelled) setSettings(s)
      })
      .catch(() => {
        /* settings are optional for the hint card */
      })
    return () => {
      cancelled = true
    }
  }, [])

  const hasArchive = useMemo(
    () => files.some((f) => ARCHIVE_EXTS.includes(getExtension(f.file.name) as typeof ARCHIVE_EXTS[number])),
    [files],
  )

  const addFiles = useCallback((incoming: FileList | File[] | null) => {
    if (!incoming) return
    const list = Array.from(incoming)
    if (list.length === 0) return
    setFiles((prev) => {
      const seen = new Set(prev.map((p) => `${p.file.name}-${p.file.size}-${p.file.lastModified}`))
      const additions: PendingFile[] = []
      for (const file of list) {
        const key = `${file.name}-${file.size}-${file.lastModified}`
        if (seen.has(key)) continue
        seen.add(key)
        additions.push({ id: makeId(), file })
      }
      return [...prev, ...additions]
    })
  }, [])

  const removeFile = useCallback((id: string) => {
    setFiles((prev) => prev.filter((p) => p.id !== id))
  }, [])

  const onPick = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      addFiles(e.target.files)
      // allow re-picking the same file later
      e.target.value = ''
    },
    [addFiles],
  )

  const onDragOver = useCallback((e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault()
    e.stopPropagation()
  }, [])

  const onDrop = useCallback(
    (e: React.DragEvent<HTMLDivElement>) => {
      e.preventDefault()
      e.stopPropagation()
      if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
        addFiles(e.dataTransfer.files)
      }
    },
    [addFiles],
  )

  const openPicker = useCallback(() => {
    if (uploading) return
    fileInputRef.current?.click()
  }, [uploading])

  const onUpload = useCallback(async () => {
    if (files.length === 0 || uploading) return
    setUploading(true)
    setResults([])
    const nextResults: UploadResult[] = []
    const errors: string[] = []
    const isArchive = hasArchive

    for (const item of files) {
      try {
        const payload = await api.uploadFile(item.file, isArchive ? password || undefined : undefined)
        nextResults.push({
          filename: item.file.name,
          upload_id: payload.upload_id,
          job_id: payload.job_id,
        })
        toast.success(`Uploaded ${item.file.name}`)
      } catch (e) {
        errors.push(item.file.name)
        toast.error(`${item.file.name}: ${(e as Error).message}`)
      }
    }

    if (errors.length > 0 && nextResults.length === 0) {
      toast.error(`${errors.length} file(s) failed`)
    } else if (errors.length > 0) {
      toast.warning(`${errors.length} file(s) failed`)
    }

    setResults(nextResults)
    if (errors.length === 0) {
      setFiles([])
      setPassword('')
    }
    setUploading(false)
    loadRecent()
  }, [files, uploading, hasArchive, password, toast, loadRecent])

  const onRetryRecent = useCallback(() => {
    loadRecent()
  }, [loadRecent])

  const maxSizeLabel = settings?.max_upload_size_mb ?? '—'

  return (
    <div className="stack-lg">
      <header className="stack-sm">
        <h2>Upload files</h2>
        <p className="muted">
          Drop files below or click to browse. Supported formats include zip, pdf, csv, json, and more.
        </p>
      </header>

      <Card title="Add files" subtitle="Select one or more files to ingest" padding="lg">
        <div className="stack">
          <div
            className="drop-zone"
            role="button"
            tabIndex={0}
            aria-label="Drop files here or click to browse"
            aria-disabled={uploading || undefined}
            onClick={openPicker}
            onKeyDown={(e) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault()
                openPicker()
              }
            }}
            onDragOver={onDragOver}
            onDrop={onDrop}
          >
            <div className="drop-zone-icon" aria-hidden="true">
              <UploadCloud size={36} />
            </div>
            <div className="drop-zone-title">Drop a file here</div>
            <div className="drop-zone-subtitle muted">or click to browse</div>
            <input
              ref={fileInputRef}
              id={inputId}
              type="file"
              multiple
              className="sr-only"
              onChange={onPick}
              disabled={uploading}
              aria-label="Choose files to upload"
            />
          </div>

          {files.length > 0 && (
            <div className="stack-sm">
              <div className="row-between">
                <h4>Files to upload</h4>
                <span className="muted text-sm">
                  {files.length} file{files.length === 1 ? '' : 's'}
                </span>
              </div>
              <ul className="stack-sm upload-file-list" aria-label="Files queued for upload">
                {files.map((item) => (
                  <li key={item.id} className="upload-file-row">
                    <div className="row" style={{ minWidth: 0, flex: 1 }}>
                      <FileTypeIcon filename={item.file.name} size={20} />
                      <div className="stack-sm" style={{ minWidth: 0, flex: 1 }}>
                        <span className="truncate" title={item.file.name}>
                          {item.file.name}
                        </span>
                        <span className="muted text-sm">{formatBytes(item.file.size)}</span>
                      </div>
                    </div>
                    <Button
                      variant="ghost"
                      size="sm"
                      iconLeft={<Trash2 size={14} />}
                      onClick={() => removeFile(item.id)}
                      disabled={uploading}
                      aria-label={`Remove ${item.file.name}`}
                    >
                      Remove
                    </Button>
                  </li>
                ))}
              </ul>
            </div>
          )}

          {hasArchive && (
            <label className="label" htmlFor={`${inputId}-password`}>
              Archive password
              <input
                id={`${inputId}-password`}
                type="password"
                className="input"
                placeholder="Password for encrypted archives"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                disabled={uploading}
                autoComplete="off"
              />
              <span className="helper-text">Required for encrypted archives</span>
            </label>
          )}

          <div className="row">
            <Button
              variant="primary"
              size="lg"
              loading={uploading}
              disabled={files.length === 0}
              onClick={onUpload}
              iconLeft={<UploadCloud size={16} />}
            >
              Upload {files.length} file{files.length === 1 ? '' : 's'}
            </Button>
            {uploading && <span className="muted text-sm">Uploading, please wait...</span>}
          </div>
        </div>
      </Card>

      {results.length > 0 && (
        <Card title="Upload results" subtitle="Track each file as it moves through the pipeline">
          <ul className="stack-sm upload-result-list">
            {results.map((r) => (
              <li key={`${r.upload_id}-${r.filename}`} className="upload-result-row">
                <div className="stack-sm" style={{ minWidth: 0, flex: 1 }}>
                  <span className="truncate" title={r.filename}>
                    {r.filename}
                  </span>
                  <span className="muted text-sm row" style={{ gap: 6 }}>
                    <span>Job:</span>
                    <code className="font-mono">{r.job_id.slice(0, 8)}</code>
                    <CopyButton text={r.job_id} ariaLabel={`Copy job id ${r.job_id}`} />
                  </span>
                </div>
                <div className="row" style={{ gap: 8 }}>
                  <StatusBadge status="completed" kind="upload" />
                  <Link to="/jobs">
                    <Button variant="ghost" size="sm" iconRight={<ExternalLink size={12} />}>
                      View job
                    </Button>
                  </Link>
                </div>
              </li>
            ))}
          </ul>
        </Card>
      )}

      <Card
        title="Recent uploads"
        subtitle="Last five uploads across the platform"
        actions={
          <Link to="/files">
            <Button variant="ghost" size="sm">
              Browse files
            </Button>
          </Link>
        }
      >
        {recentLoading ? (
          <div className="stack-sm" aria-label="Loading recent uploads">
            <Skeleton height={48} />
            <Skeleton height={48} />
            <Skeleton height={48} />
            <Skeleton height={48} />
            <Skeleton height={48} />
          </div>
        ) : recentError ? (
          <EmptyState
            icon={AlertCircle}
            title="Couldn't load recent uploads"
            description={recentError}
            action={
              <Button
                onClick={onRetryRecent}
                variant="secondary"
                iconLeft={<RefreshCw size={14} />}
              >
                Retry
              </Button>
            }
          />
        ) : !recent || recent.length === 0 ? (
          <EmptyState
            icon={UploadCloud}
            title="No uploads yet"
            description="Your uploaded files will appear here."
          />
        ) : (
          <ul className="stack-sm upload-recent-list">
            {recent.map((u) => (
              <li key={u.id} className="upload-recent-row">
                <div className="row" style={{ minWidth: 0, flex: 1 }}>
                  <FileTypeIcon
                    filename={u.original_name ?? u.filename}
                    mimeType={u.content_type}
                  />
                  <div className="stack-sm" style={{ minWidth: 0, flex: 1 }}>
                    <span className="truncate" title={u.original_name ?? u.filename}>
                      {u.original_name ?? u.filename}
                    </span>
                    <span className="muted text-sm">
                      {formatBytes(u.size_bytes)} - {formatRelative(u.created_at)}
                    </span>
                  </div>
                </div>
                <StatusBadge status={u.status} kind="upload" />
              </li>
            ))}
          </ul>
        )}
      </Card>

      <Card padding="sm">
        <div className="row">
          <Sparkles size={16} className="muted" aria-hidden="true" />
          <span className="muted text-sm">
            Supported formats: zip, pdf, csv, json, txt, docx, and more. Max size: {maxSizeLabel} MB.
          </span>
        </div>
      </Card>
    </div>
  )
}

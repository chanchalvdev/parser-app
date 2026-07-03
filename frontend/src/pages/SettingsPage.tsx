import { useCallback, useEffect, useState } from 'react'
import { Lock, Save } from 'lucide-react'
import {
  Activity,
  AlertCircle,
  AlertTriangle,
  FileText,
  RefreshCw,
} from '../components/Icon'
import { Card } from '../components/Card'
import { Button } from '../components/Button'
import { Spinner } from '../components/Spinner'
import { Skeleton } from '../components/Skeleton'
import { EmptyState } from '../components/EmptyState'
import { useToast } from '../components/ToastProvider'
import { api } from '../api'
import type { Settings } from '../types'

function formatRelative(iso: string | undefined | null): string {
  if (!iso) return '—'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return '—'
  const diffMs = Date.now() - d.getTime()
  if (diffMs < 0) return 'just now'
  const seconds = Math.floor(diffMs / 1000)
  if (seconds < 60) return 'just now'
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`
  if (seconds < 86400 * 30) return `${Math.floor(seconds / 86400)}d ago`
  return d.toLocaleDateString()
}

interface AuditEntry {
  event_type?: string
  type?: string
  action?: string
  actor?: string
  user?: string
  target?: string
  resource?: string
  created_at?: string
  timestamp?: string
  createdAt?: string
}

function pickField<T = string>(
  entry: Record<string, unknown>,
  keys: string[],
  fallback: T,
): T {
  for (const key of keys) {
    const value = entry[key]
    if (value !== undefined && value !== null && value !== '') {
      return value as T
    }
  }
  return fallback
}

export default function SettingsPage() {
  const toast = useToast()

  const [settings, setSettings] = useState<Settings | null>(null)
  const [settingsError, setSettingsError] = useState<string | null>(null)
  const [loadingSettings, setLoadingSettings] = useState(true)
  const [maxUploadMb, setMaxUploadMb] = useState<string>('')
  const [savingUploads, setSavingUploads] = useState(false)

  const [audit, setAudit] = useState<unknown[] | null>(null)
  const [auditError, setAuditError] = useState<string | null>(null)
  const [auditLoading, setAuditLoading] = useState(true)

  const isValid =
    maxUploadMb.trim().length > 0 && Number(maxUploadMb) > 0

  const loadSettings = useCallback(async () => {
    setLoadingSettings(true)
    setSettingsError(null)
    try {
      const data = await api.settings()
      setSettings(data)
      const initial =
        typeof data.max_upload_size_mb === 'number'
          ? String(data.max_upload_size_mb)
          : ''
      setMaxUploadMb(initial)
    } catch (e) {
      const msg = e instanceof Error ? e.message : 'Failed to load settings'
      setSettingsError(msg)
    } finally {
      setLoadingSettings(false)
    }
  }, [])

  const reloadAudit = useCallback(async () => {
    setAuditLoading(true)
    setAuditError(null)
    try {
      const data = await api.auditLogs()
      setAudit(Array.isArray(data) ? data : [])
    } catch (e) {
      const msg = e instanceof Error ? e.message : 'Failed to load audit log'
      setAuditError(msg)
      setAudit(null)
    } finally {
      setAuditLoading(false)
    }
  }, [])

  useEffect(() => {
    void loadSettings()
    void reloadAudit()
  }, [loadSettings, reloadAudit])

  const saveUploads = async () => {
    if (!isValid) return
    setSavingUploads(true)
    try {
      const result = await api.updateSettings({
        max_upload_size_mb: Number(maxUploadMb),
      })
      setSettings(result)
      toast.success('Upload limits updated')
    } catch (e) {
      const msg = e instanceof Error ? e.message : 'Failed to save settings'
      toast.error(msg)
    } finally {
      setSavingUploads(false)
    }
  }

  const auditEntries: AuditEntry[] = Array.isArray(audit)
    ? (audit as AuditEntry[]).slice(0, 20)
    : []

  return (
    <div className="stack-lg">
      <header className="stack-sm">
        <h2>Settings</h2>
        <p className="muted">Tune platform limits and review audit activity.</p>
      </header>

      <Card
        title="Upload limits"
        subtitle="Control how large an upload can be"
        actions={null}
      >
        <div className="stack">
          {loadingSettings ? (
            <Spinner label="Loading settings…" />
          ) : settingsError ? (
            <div className="inline-alert" role="alert">
              <AlertCircle size={14} />
              <span>{settingsError}</span>
            </div>
          ) : (
            <>
              <div className="setting-row">
                <div className="stack-sm" style={{ minWidth: 0 }}>
                  <label htmlFor="max-upload-mb" className="text-sm" style={{ fontWeight: 600 }}>
                    Max upload size
                  </label>
                  <span className="helper-text">
                    Larger uploads take longer to process.
                  </span>
                </div>
                <div className="row" style={{ gap: 'var(--space-2)' }}>
                  <input
                    id="max-upload-mb"
                    type="number"
                    min={1}
                    max={10240}
                    value={maxUploadMb}
                    onChange={(e) => setMaxUploadMb(e.target.value)}
                    className="input input-inline"
                  />
                  <span className="muted text-sm">MB</span>
                </div>
              </div>

              {!isValid ? (
                <span className="text-sm error">
                  Enter a positive number of MB.
                </span>
              ) : null}

              <div>
                <Button
                  variant="primary"
                  loading={savingUploads}
                  disabled={!isValid}
                  onClick={() => void saveUploads()}
                  iconLeft={<Save size={14} />}
                >
                  Save changes
                </Button>
              </div>
            </>
          )}
        </div>
      </Card>

      <Card
        title="Processing"
        subtitle="Tune parsing and indexing behavior"
      >
        <div className="stack">
          <div className="setting-row">
            <div className="stack-sm" style={{ minWidth: 0 }}>
              <label htmlFor="concurrent-workers" className="text-sm" style={{ fontWeight: 600 }}>
                Concurrent workers
              </label>
              <span className="helper-text">
                These settings are read-only in this build.
              </span>
            </div>
            <div className="row" style={{ gap: 'var(--space-2)' }}>
              <input
                id="concurrent-workers"
                type="number"
                value={4}
                disabled
                className="input input-inline"
              />
              <Lock size={12} className="muted" aria-hidden="true" />
            </div>
          </div>

          <div className="setting-row">
            <div className="stack-sm" style={{ minWidth: 0 }}>
              <label htmlFor="parser-timeout" className="text-sm" style={{ fontWeight: 600 }}>
                Parser timeout (seconds)
              </label>
              <span className="helper-text">
                These settings are read-only in this build.
              </span>
            </div>
            <div className="row" style={{ gap: 'var(--space-2)' }}>
              <input
                id="parser-timeout"
                type="number"
                value={300}
                disabled
                className="input input-inline"
              />
              <Lock size={12} className="muted" aria-hidden="true" />
            </div>
          </div>
        </div>
      </Card>

      <Card
        title="Recent audit activity"
        subtitle="Last 20 events across the platform"
        actions={
          <Button
            variant="ghost"
            size="sm"
            iconLeft={<RefreshCw size={14} />}
            onClick={() => void reloadAudit()}
          >
            Refresh
          </Button>
        }
      >
        {auditLoading ? (
          <div className="stack-sm" aria-hidden="true">
            <Skeleton height={56} />
            <Skeleton height={56} />
            <Skeleton height={56} />
            <Skeleton height={56} />
            <Skeleton height={56} />
          </div>
        ) : auditError ? (
          <EmptyState
            icon={AlertCircle}
            title="Couldn't load audit log"
            description={auditError}
            action={
              <Button
                onClick={() => void reloadAudit()}
                variant="secondary"
                iconLeft={<RefreshCw size={14} />}
              >
                Retry
              </Button>
            }
          />
        ) : auditEntries.length === 0 ? (
          <EmptyState
            icon={Activity}
            title="No audit events"
            description="No platform activity has been recorded yet."
          />
        ) : (
          <div className="stack-sm" style={{ margin: 'calc(var(--space-3) * -1)' }}>
            {auditEntries.map((entry, index) => {
              const record = entry as Record<string, unknown>
              const eventType = pickField<string>(
                record,
                ['event_type', 'type', 'action'],
                'event',
              )
              const actor = pickField<string>(
                record,
                ['actor', 'user'],
                'system',
              )
              const target = pickField<string>(
                record,
                ['target', 'resource'],
                '',
              )
              const when = pickField<string | undefined>(
                record,
                ['created_at', 'timestamp', 'createdAt'],
                undefined,
              )
              const key =
                (record.id as string | undefined) ??
                (record.event_id as string | undefined) ??
                `${eventType}-${actor}-${index}`
              return (
                <div key={key} className="audit-row">
                  <FileText size={14} className="muted" aria-hidden="true" />
                  <div className="stack-sm" style={{ minWidth: 0, flex: 1 }}>
                    <span className="text-sm" style={{ fontWeight: 600 }}>
                      {eventType}
                      {target ? (
                        <span className="muted"> · {target}</span>
                      ) : null}
                    </span>
                    <span className="muted text-sm">by {actor}</span>
                  </div>
                  <span className="muted text-sm">
                    {formatRelative(when)}
                  </span>
                </div>
              )
            })}
          </div>
        )}
      </Card>

      <Card className="card-danger" title="Danger zone" subtitle="Irreversible actions">
        <div className="stack">
          <div className="row" style={{ gap: 'var(--space-3)' }}>
            <AlertTriangle size={16} className="error" aria-hidden="true" />
            <span style={{ fontWeight: 700 }}>Reset all data</span>
          </div>
          <p className="muted">
            Clears all uploads, jobs, and parsed files. This cannot be undone.
          </p>
          <div className="row" style={{ gap: 'var(--space-3)', flexWrap: 'wrap' }}>
            <Button
              variant="danger"
              disabled
              iconLeft={<Lock size={14} />}
            >
              Reset all data
            </Button>
            <span className="helper-text">Disabled in this build.</span>
          </div>
        </div>
      </Card>
    </div>
  )
}

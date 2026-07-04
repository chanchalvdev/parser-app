import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { ApiError } from '@/types/api'
import { Card } from '@/components/ui/Card'
import { PageHeader } from '@/components/layout/PageHeader'
import { SettingsForm } from '@/components/admin/SettingsForm'
import { getAdminSettings, updateAdminSettings } from '@/services/adminApi'
import type { AdminSettings, AdminSettingsUpdate } from '@/types/admin'

const DEFAULT_ADMIN_SETTINGS: AdminSettings = {
  tenant_id: '11111111-1111-1111-1111-111111111001',
  max_upload_size_mb: 5120,
  max_archive_depth: 10,
  max_extracted_files: 50000,
  max_extracted_size_mb: 1024,
  max_expansion_ratio: 20,
  txt_small_file_limit_mb: 10,
  enabled_parsers: ['txt', 'log', 'csv', 'json', 'jsonl', 'xml', 'xlsx', 'pdf', 'text'],
  parser_batch_size: 1000,
  search_index_batch_size: 1000,
}

const AVAILABLE_PARSERS = ['csv', 'json', 'jsonl', 'log', 'pdf', 'text', 'txt', 'xlsx', 'xml']

const buildSettingsPatch = (current: AdminSettings, original: AdminSettings): AdminSettingsUpdate => {
  const patch: AdminSettingsUpdate = {}

  if (current.max_upload_size_mb !== original.max_upload_size_mb) {
    patch.max_upload_size_mb = current.max_upload_size_mb
  }
  if (current.max_archive_depth !== original.max_archive_depth) {
    patch.max_archive_depth = current.max_archive_depth
  }
  if (current.max_extracted_files !== original.max_extracted_files) {
    patch.max_extracted_files = current.max_extracted_files
  }
  if (current.max_extracted_size_mb !== original.max_extracted_size_mb) {
    patch.max_extracted_size_mb = current.max_extracted_size_mb
  }
  if (current.txt_small_file_limit_mb !== original.txt_small_file_limit_mb) {
    patch.txt_small_file_limit_mb = current.txt_small_file_limit_mb
  }
  if (current.max_expansion_ratio !== original.max_expansion_ratio) {
    patch.max_expansion_ratio = current.max_expansion_ratio
  }
  if (current.parser_batch_size !== original.parser_batch_size) {
    patch.parser_batch_size = current.parser_batch_size
  }
  if (current.search_index_batch_size !== original.search_index_batch_size) {
    patch.search_index_batch_size = current.search_index_batch_size
  }

  const sameParsers =
    current.enabled_parsers.length === original.enabled_parsers.length &&
    current.enabled_parsers.every((parser) => original.enabled_parsers.includes(parser))

  if (!sameParsers) {
    patch.enabled_parsers = current.enabled_parsers
  }

  return patch
}

export const AdminSettingsPage = () => {
  const queryClient = useQueryClient()
  const [tenantId, setTenantId] = useState(DEFAULT_ADMIN_SETTINGS.tenant_id)
  const [settings, setSettings] = useState<AdminSettings>(DEFAULT_ADMIN_SETTINGS)
  const [original, setOriginal] = useState<AdminSettings>(DEFAULT_ADMIN_SETTINGS)
  const [formError, setFormError] = useState('')
  const [formMessage, setFormMessage] = useState('')

  const settingsQuery = useQuery({
    queryKey: ['admin-settings', tenantId],
    enabled: !!tenantId,
    queryFn: () => getAdminSettings(tenantId),
    onSuccess: (next) => {
      setSettings(next)
      setOriginal(next)
      setFormMessage('')
      setFormError('')
    },
    onError: (error) => {
      setFormError(error instanceof ApiError ? error.payload?.message || error.message : 'Unable to load admin settings.')
    },
  })

  const mutation = useMutation({
    mutationFn: async () => {
      const payload = buildSettingsPatch(settings, original)
      return updateAdminSettings(payload, tenantId)
    },
    onSuccess: (next) => {
      setOriginal(next)
      setSettings(next)
      setFormMessage('Settings updated and audit event logged.')
      setFormError('')
      queryClient.invalidateQueries({ queryKey: ['admin-settings', tenantId] })
    },
    onError: (error) => {
      setFormError(
        error instanceof ApiError
          ? error.payload?.message || error.message
          : 'Failed to update admin settings.',
      )
    },
  })

  const changeSet = buildSettingsPatch(settings, original)
  const hasChanges = Object.keys(changeSet).length > 0
  const canManageSettings = true

  const onSave = () => {
    if (!canManageSettings) {
      setFormError('You do not have permission to edit settings.')
      return
    }

    if (!hasChanges) {
      setFormMessage('No changes to save.')
      setFormError('')
      return
    }

    setFormMessage('')
    setFormError('')
    mutation.mutate()
  }

  return (
    <div className="space-y-4">
      <PageHeader title="Admin Settings" subtitle={mutation.isPending ? 'Saving…' : `Tenant ${tenantId}`} />

      <Card title="Tenant context">
        <label className="space-y-1 text-sm">
          <span>Tenant ID</span>
          <input
            value={tenantId}
            onChange={(event) => {
              setTenantId(event.target.value)
            }}
            className="w-full rounded border border-slate-700 bg-slate-900/60 px-2 py-1"
            placeholder="UUID"
            disabled={mutation.isPending}
          />
          <div className="text-xs text-slate-300">Update this to manage another tenant while you are in local mode.</div>
        </label>
        <div className="mt-3">
          <button
            type="button"
            className="rounded border border-slate-700 px-2 py-1 text-xs"
            onClick={() => settingsQuery.refetch()}
            disabled={settingsQuery.isFetching || mutation.isPending}
          >
            {settingsQuery.isFetching ? 'Reloading…' : 'Reload settings'}
          </button>
        </div>
      </Card>

      <Card title="Parser settings" subtitle="Values are persisted in PostgreSQL and changes are written to audit_logs.">
        <div className="mb-3 flex gap-3 text-xs text-amber-200">
          <span className="chip">RBAC placeholder: role checks read from X-User-Role</span>
          <Link className="underline" to="/audit-logs">
            View audit logs
          </Link>
        </div>

        {settingsQuery.isError || formError ? (
          <p className="mb-3 text-sm text-rose-200">
            {formError || 'Unable to load admin settings.'}
          </p>
        ) : null}
        {formMessage ? <p className="mb-3 text-sm text-emerald-200">{formMessage}</p> : null}

        <SettingsForm
          values={settings}
          availableParsers={AVAILABLE_PARSERS}
          onValuesChange={setSettings}
          onSubmit={onSave}
          disabled={!canManageSettings || mutation.isPending}
          submitting={mutation.isPending}
        />

        {!hasChanges ? <p className="mt-2 text-xs text-slate-300">No changes since last save.</p> : null}
      </Card>

      <Card title="Current settings payload">
        <pre className="overflow-auto rounded bg-slate-950/60 p-3 text-xs">{JSON.stringify(settings, null, 2)}</pre>
      </Card>
    </div>
  )
}

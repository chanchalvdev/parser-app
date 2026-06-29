import { apiGet } from './api'
import type { AuditLogResponse, AuditLogFilter, SettingsState } from '@/types/domain'

const SETTINGS_KEY = 'enterprise.web.admin.settings'

export const listAuditLogs = async (filter: AuditLogFilter): Promise<AuditLogResponse> => {
  return apiGet<AuditLogResponse>('/audit-logs', {
    tenant_id: filter.tenant_id || filter.tenantId,
    page: filter.page,
    page_size: filter.pageSize,
  })
}

export const loadSettings = async (): Promise<SettingsState> => {
  const raw = localStorage.getItem(SETTINGS_KEY)
  if (raw) {
    return JSON.parse(raw) as SettingsState
  }

  return {
    retention_days: 30,
    max_file_size_mb: 250,
    auto_retry_attempts: 3,
    highlight_snippets: true,
    default_page_size: 25,
  }
}

export const saveSettings = async (settings: SettingsState): Promise<SettingsState> => {
  localStorage.setItem(SETTINGS_KEY, JSON.stringify(settings))
  return settings
}

import { apiClient } from './apiClient'
import type { AdminSettings, AdminSettingsUpdate } from '@/types/admin'

export const getAdminSettings = async (tenantId?: string): Promise<AdminSettings> => {
  return apiClient.get<AdminSettings>('/admin/settings', {
    query: {
      tenant_id: tenantId,
    },
  })
}

export const updateAdminSettings = async (
  payload: AdminSettingsUpdate,
  tenantId?: string,
): Promise<AdminSettings> => {
  return apiClient.put<AdminSettings>('/admin/settings', payload, {
    query: {
      tenant_id: tenantId,
    },
  })
}


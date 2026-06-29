import { apiClient } from './apiClient'
import type {
  DashboardDistribution,
  DashboardEntities,
  DashboardErrorBreakdown,
  DashboardProcessingDuration,
  DashboardSummary,
  DashboardUploadVolume,
} from '@/types/dashboard'

export const getDashboardSummary = async (tenantId?: string): Promise<DashboardSummary> => {
  return apiClient.get<DashboardSummary>('/dashboard/summary', {
    query: {
      tenant_id: tenantId,
    },
  })
}

export const getDashboardFileTypes = async (tenantId?: string, limit = 25): Promise<DashboardDistribution> => {
  return apiClient.get<DashboardDistribution>('/dashboard/file-types', {
    query: {
      tenant_id: tenantId,
      limit,
    },
  })
}

export const getDashboardProcessingStatus = async (
  tenantId?: string,
  limit = 25,
): Promise<DashboardDistribution> => {
  return apiClient.get<DashboardDistribution>('/dashboard/processing-status', {
    query: {
      tenant_id: tenantId,
      limit,
    },
  })
}

export const getDashboardUploadVolume = async (
  tenantId?: string,
  days = 7,
): Promise<DashboardUploadVolume> => {
  return apiClient.get<DashboardUploadVolume>('/dashboard/upload-volume', {
    query: {
      tenant_id: tenantId,
      days,
    },
  })
}

export const getDashboardErrorBreakdown = async (
  tenantId?: string,
  limit = 10,
): Promise<DashboardErrorBreakdown> => {
  return apiClient.get<DashboardErrorBreakdown>('/dashboard/error-breakdown', {
    query: {
      tenant_id: tenantId,
      limit,
    },
  })
}

export const getDashboardEntities = async (
  tenantId?: string,
  limit = 10,
): Promise<DashboardEntities> => {
  return apiClient.get<DashboardEntities>('/dashboard/entities', {
    query: {
      tenant_id: tenantId,
      limit,
    },
  })
}

export const getDashboardProcessingDuration = async (tenantId?: string): Promise<DashboardProcessingDuration> => {
  return apiClient.get<DashboardProcessingDuration>('/dashboard/processing-duration', {
    query: {
      tenant_id: tenantId,
    },
  })
}

import { apiGet } from './api'
import type {
  DashboardSummary,
  DashboardDistribution,
  DashboardUploadVolume,
  DashboardErrorBreakdown,
  DashboardEntities,
  DashboardProcessingDuration,
} from '@/types/domain'

export const getDashboardSummary = async (tenantId?: string): Promise<DashboardSummary> => {
  return apiGet<DashboardSummary>('/dashboard/summary', {
    tenant_id: tenantId,
  })
}

export const getDashboardFileTypes = async (tenantId?: string, limit = 25): Promise<DashboardDistribution> => {
  return apiGet<DashboardDistribution>('/dashboard/file-types', {
    tenant_id: tenantId,
    limit,
  })
}

export const getDashboardProcessingStatus = async (
  tenantId?: string,
  limit = 25,
): Promise<DashboardDistribution> => {
  return apiGet<DashboardDistribution>('/dashboard/processing-status', {
    tenant_id: tenantId,
    limit,
  })
}

export const getDashboardUploadVolume = async (tenantId?: string, days = 7): Promise<DashboardUploadVolume> => {
  return apiGet<DashboardUploadVolume>('/dashboard/upload-volume', {
    tenant_id: tenantId,
    days,
  })
}

export const getDashboardErrorBreakdown = async (
  tenantId?: string,
  limit = 10,
): Promise<DashboardErrorBreakdown> => {
  return apiGet<DashboardErrorBreakdown>('/dashboard/error-breakdown', {
    tenant_id: tenantId,
    limit,
  })
}

export const getDashboardEntities = async (tenantId?: string, limit = 10): Promise<DashboardEntities> => {
  return apiGet<DashboardEntities>('/dashboard/entities', {
    tenant_id: tenantId,
    limit,
  })
}

export const getDashboardProcessingDuration = async (tenantId?: string): Promise<DashboardProcessingDuration> => {
  return apiGet<DashboardProcessingDuration>('/dashboard/processing-duration', {
    tenant_id: tenantId,
  })
}

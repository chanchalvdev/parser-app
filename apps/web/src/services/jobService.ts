import { apiGet, apiPost } from './api'
import type {
  JobListResponse,
  IngestionJob,
  JobEventsResponse,
} from '@/types/domain'

export const listJobs = async (args: { tenantId?: string; page?: number; pageSize?: number }): Promise<JobListResponse> => {
  return apiGet<JobListResponse>('/jobs', {
    tenant_id: args.tenantId,
    page: args.page,
    page_size: args.pageSize,
  })
}

export const getJob = async (jobId: string): Promise<IngestionJob> => {
  return apiGet<IngestionJob>(`/jobs/${jobId}`)
}

export const getJobEvents = async (jobId: string, page = 1, pageSize = 25): Promise<JobEventsResponse> => {
  return apiGet<JobEventsResponse>(`/jobs/${jobId}/events`, {
    page,
    page_size: pageSize,
  })
}

export const retryJob = async (jobId: string): Promise<IngestionJob> => {
  return apiPost<IngestionJob>(`/jobs/${jobId}/retry`, {})
}

import { apiClient } from './apiClient'
import type { IngestionJob, JobEventsResponse, JobListResponse } from '@/types/job'

export type ListJobsParams = {
  tenant_id?: string
  tenantId?: string
  page?: number
  page_size?: number
  pageSize?: number
  status?: string
}

export const listJobs = async (args: ListJobsParams = {}): Promise<JobListResponse> => {
  return apiClient.get<JobListResponse>('/jobs', {
    query: {
      tenant_id: args.tenant_id ?? args.tenantId,
      status: args.status,
      page: args.page,
      page_size: args.page_size ?? args.pageSize,
    },
  })
}

export const getJob = async (jobId: string): Promise<IngestionJob> => {
  return apiClient.get<IngestionJob>(`/jobs/${jobId}`)
}

export const getJobEvents = async (
  jobId: string,
  page = 1,
  pageSize = 25,
): Promise<JobEventsResponse> => {
  return apiClient.get<JobEventsResponse>(`/jobs/${jobId}/events`, {
    query: {
      page,
      page_size: pageSize,
    },
  })
}

export const retryJob = async (jobId: string): Promise<IngestionJob> => {
  return apiClient.post<IngestionJob>(`/jobs/${jobId}/retry`, {})
}

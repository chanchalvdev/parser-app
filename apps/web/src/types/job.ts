import type { ApiPagination } from './api'

export type IngestionJob = {
  id: string
  tenant_id: string
  root_file_id: string
  status: string
  current_stage?: string | null
  progress_percent: number
  retry_count: number
  error_code?: string | null
  error_message?: string | null
  started_at: string | null
  completed_at: string | null
  created_at: string
  updated_at: string
}

export type JobListResponse = ApiPagination & {
  jobs: IngestionJob[]
}

export type JobEvent = {
  id: string
  tenant_id: string
  job_id: string
  event_type: string
  event_message?: string | null
  event_details: Record<string, unknown>
  created_by?: string | null
  created_at: string
}

export type JobEventsResponse = ApiPagination & {
  job_id: string
  events: JobEvent[]
}

export type RetryJobResponse = IngestionJob

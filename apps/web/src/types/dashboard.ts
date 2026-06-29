import type { ApiPagination } from './api'

export type DashboardSummary = {
  tenant_id: string
  total_uploads: number
  total_files: number
  total_extracted_files: number
  total_parsed_records: number
  completed_jobs: number
  failed_jobs: number
  password_required_files: number
  quarantined_files: number
}

export type DashboardBucket = {
  value: string
  count: number
}

export type DashboardDistribution = {
  tenant_id: string
  buckets: DashboardBucket[]
}

export type UploadVolumeBucket = {
  bucket: string
  uploads: number
  files: number
  parsed_records: number
}

export type DashboardUploadVolume = {
  tenant_id: string
  unit: string
  days: number
  from: string
  to: string
  buckets: UploadVolumeBucket[]
}

export type DashboardErrorItem = {
  error_code: string
  count: number
  last_seen: string
}

export type DashboardErrorBreakdown = {
  tenant_id: string
  total: number
  errors: DashboardErrorItem[]
}

export type DashboardEntities = {
  tenant_id: string
  limit: number
  entities: {
    ip_addresses: DashboardBucket[]
    emails: DashboardBucket[]
    urls: DashboardBucket[]
    domains: DashboardBucket[]
    hashes: DashboardBucket[]
  }
}

export type DashboardProcessingDuration = {
  tenant_id: string
  unit: string
  completed_jobs: number
  average_seconds: number
  median_seconds: number
  p95_seconds: number
  min_seconds: number
  max_seconds: number
}

export type DashboardChartsPayload = {
  tenant_id: string
  file_types: DashboardDistribution['buckets']
  processing_status: DashboardDistribution['buckets']
  upload_volume: DashboardUploadVolume['buckets']
  error_breakdown: DashboardErrorBreakdown['errors']
  top_entities: DashboardEntities['entities']
  processing_duration: DashboardProcessingDuration
}

export type DashboardChartsQuery = ApiPagination & {
  payload: DashboardChartsPayload
}

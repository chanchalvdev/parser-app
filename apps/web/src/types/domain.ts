export type ApiPagination = {
  total: number
  page: number
  page_size: number
}

export type FileItem = {
  id: string
  tenant_id: string
  parent_file_id: string | null
  upload_id: string | null
  original_name: string
  normalized_name?: string | null
  extension?: string | null
  detected_mime_type?: string | null
  detected_file_type?: string | null
  storage_path: string
  size_bytes: number
  sha256_hash?: string | null
  depth: number
  is_archive: boolean
  is_password_protected: boolean
  processing_status: string
  created_by?: string | null
  created_at: string
  updated_at: string
}

export type ParsedRecord = {
  id: string
  tenant_id: string
  file_id: string
  job_id: string
  record_type: string | null
  record_number: number | null
  line_number: number | null
  chunk_number: number | null
  start_line: number | null
  end_line: number | null
  content_text: string | null
  structured_data: Record<string, unknown> | null
  extracted_entities: Record<string, unknown> | null
  event_timestamp: string | null
  created_at: string
}

export type FileListResponse = ApiPagination & { files: FileItem[] }

export type FileChildrenResponse = ApiPagination & {
  file_id: string
  children: FileItem[]
}

export type FileTreeNode = {
  file: FileItem
  children: FileTreeNode[]
}

export type FileRecordsResponse = ApiPagination & {
  file_id: string
  records: ParsedRecord[]
}

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

export type SearchResult = {
  record_id: string
  file_id: string
  job_id: string
  source_file_name: string
  record_type: string
  content_preview: string
  highlight: string
  entities: Record<string, unknown>
  created_at: string
}

export type SearchFacetItem = {
  value: string
  count: number
}

export type SearchFacets = {
  record_type: SearchFacetItem[]
  detected_file_type: SearchFacetItem[]
  entities: {
    ip_addresses: SearchFacetItem[]
    emails: SearchFacetItem[]
    domains: SearchFacetItem[]
    urls: SearchFacetItem[]
    hashes: SearchFacetItem[]
  }
}

export type SearchResponse = ApiPagination & {
  results: SearchResult[]
  facets: SearchFacets
}

export type SearchSuggestions = {
  suggestions: {
    record_type: SearchFacetItem[]
    detected_file_type: SearchFacetItem[]
    entities: {
      ip_addresses: SearchFacetItem[]
      emails: SearchFacetItem[]
      domains: SearchFacetItem[]
      urls: SearchFacetItem[]
      hashes: SearchFacetItem[]
    }
  }
}

export type SearchRequest = {
  q?: string
  file_id?: string
  extension?: string
  detected_file_type?: string
  record_type?: string
  date_from?: string
  date_to?: string
  ip?: string
  email?: string
  domain?: string
  job_id?: string
  page?: number
  page_size?: number
  sort?: 'relevance' | 'created_at'
  tenant_id?: string
}

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

export type UploadInitiateRequest = {
  file_name: string
  content_type: string
  size_bytes: number
  password_provided: boolean
}

export type UploadInitiateResponse = {
  upload_id: string
  object_key: string
  upload_url: string
  expires_in_seconds: number
}

export type UploadCompleteRequest = {
  upload_id: string
}

export type UploadCompleteResponse = {
  upload_id: string
  file_id: string
  job_id: string
  status: string
}

export type AuditLog = {
  id: string
  tenant_id: string
  actor_user_id?: string | null
  action: string
  entity_type?: string | null
  entity_id?: string | null
  details: Record<string, unknown>
  ip_address?: string | null
  user_agent?: string | null
  created_at: string
}

export type AuditLogFilter = {
  tenant_id?: string
  tenantId?: string
  page?: number
  page_size?: number
}

export type AuditLogResponse = ApiPagination & {
  logs: AuditLog[]
}

export type SettingsState = {
  retention_days: number
  max_file_size_mb: number
  auto_retry_attempts: number
  highlight_snippets: boolean
  default_page_size: number
}

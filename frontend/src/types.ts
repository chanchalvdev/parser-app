export type Upload = {
  id: string
  filename: string
  original_name: string
  storage_key: string
  uploader_id: string
  status: string
  content_type: string
  size_bytes: number
  created_at: string
  updated_at: string
  error_message?: string
}

export type Job = {
  id: string
  upload_id: string
  status: string
  stage: string
  attempt_count: number
  error_message?: string
  created_at: string
  updated_at: string
  finished_at?: string
}

export type SearchItem = {
  file_id: string
  upload_id: string
  job_id: string
  filename: string
  path: string
  content: string
  score: number
  index_id: string
}

export type FileNode = {
  id: string
  upload_id: string
  job_id: string
  parent_id?: string
  path: string
  filename: string
  kind: string
  mime_type?: string
  size_bytes: number
  sha256?: string
  metadata: unknown
  created_at: string
  children?: FileNode[]
}

export type DashboardSummary = {
  jobs_by_status: Record<string, number>
  uploads: number
  files: number
  parsed_records: number
  recent_jobs: Job[]
}

export type Settings = {
  max_upload_size_mb?: number
  [key: string]: unknown
}


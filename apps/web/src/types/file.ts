import type { ApiPagination } from './api'

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

export type ListFilesParams = {
  tenant_id?: string
  tenantId?: string
  status?: string
  extension?: string
  detected_file_type?: string
  detectedFileType?: string
  page?: number
  page_size?: number
  pageSize?: number
}

export type FileListResponse = ApiPagination & {
  files: FileItem[]
}

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

export type SubmitFilePasswordRequest = {
  password: string
}

export type SubmitFilePasswordResponse = {
  file_id: string
  job_id: string
  status: string
}

export type UploadPagination = ApiPagination

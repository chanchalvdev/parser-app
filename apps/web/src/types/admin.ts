export type AdminSettings = {
  tenant_id: string
  max_upload_size_mb: number
  max_archive_depth: number
  max_extracted_files: number
  max_extracted_size_mb: number
  max_expansion_ratio: number
  txt_small_file_limit_mb: number
  enabled_parsers: string[]
  parser_batch_size: number
  search_index_batch_size: number
}

export type AdminSettingsUpdate = Partial<Omit<AdminSettings, 'tenant_id'>>


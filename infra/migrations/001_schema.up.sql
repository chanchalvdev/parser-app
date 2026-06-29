-- file-platform initial schema
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE tenants (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL UNIQUE,
  slug TEXT NOT NULL UNIQUE,
  status TEXT NOT NULL DEFAULT 'active',
  contact_email TEXT,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  email TEXT NOT NULL,
  password_hash TEXT,
  display_name TEXT,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT users_tenant_email_unique UNIQUE (tenant_id, email)
);

CREATE TABLE roles (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  description TEXT,
  is_system_role BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT roles_tenant_name_unique UNIQUE (tenant_id, name)
);

CREATE TABLE user_roles (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  granted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  granted_by UUID,
  CONSTRAINT user_roles_tenant_user_role_unique UNIQUE (tenant_id, user_id, role_id)
);

CREATE TABLE uploads (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  initiated_by UUID REFERENCES users(id),
  status TEXT NOT NULL DEFAULT 'pending',
  original_source TEXT,
  storage_prefix TEXT,
  requested_by TEXT,
  total_files INTEGER NOT NULL DEFAULT 0,
  total_size_bytes BIGINT NOT NULL DEFAULT 0,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE files (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  parent_file_id UUID REFERENCES files(id) ON DELETE SET NULL,
  upload_id UUID REFERENCES uploads(id) ON DELETE SET NULL,
  original_name TEXT NOT NULL,
  normalized_name TEXT,
  extension TEXT,
  detected_mime_type TEXT,
  detected_file_type TEXT,
  storage_path TEXT NOT NULL,
  size_bytes BIGINT NOT NULL DEFAULT 0,
  sha256_hash TEXT,
  depth INTEGER NOT NULL DEFAULT 0,
  is_archive BOOLEAN NOT NULL DEFAULT FALSE,
  is_password_protected BOOLEAN NOT NULL DEFAULT FALSE,
  processing_status TEXT NOT NULL DEFAULT 'pending',
  created_by UUID REFERENCES users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT files_size_bytes_non_negative CHECK (size_bytes >= 0),
  CONSTRAINT files_depth_non_negative CHECK (depth >= 0)
);

CREATE TABLE ingestion_jobs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  root_file_id UUID NOT NULL REFERENCES files(id) ON DELETE RESTRICT,
  status TEXT NOT NULL DEFAULT 'queued',
  current_stage TEXT,
  progress_percent NUMERIC(5,2) NOT NULL DEFAULT 0,
  retry_count INTEGER NOT NULL DEFAULT 0,
  error_code TEXT,
  error_message TEXT,
  started_at TIMESTAMPTZ,
  completed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT ingestion_jobs_progress_valid CHECK (progress_percent >= 0 AND progress_percent <= 100),
  CONSTRAINT ingestion_jobs_retry_count_non_negative CHECK (retry_count >= 0)
);

CREATE TABLE job_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  job_id UUID NOT NULL REFERENCES ingestion_jobs(id) ON DELETE CASCADE,
  event_type TEXT NOT NULL,
  event_message TEXT,
  event_details JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_by UUID REFERENCES users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE parsed_records (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
  job_id UUID NOT NULL REFERENCES ingestion_jobs(id) ON DELETE CASCADE,
  record_type TEXT,
  record_number BIGINT,
  line_number BIGINT,
  chunk_number BIGINT,
  start_line BIGINT,
  end_line BIGINT,
  content_text TEXT,
  structured_data JSONB NOT NULL DEFAULT '{}'::jsonb,
  extracted_entities JSONB NOT NULL DEFAULT '{}'::jsonb,
  event_timestamp TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE parser_errors (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  job_id UUID NOT NULL REFERENCES ingestion_jobs(id) ON DELETE CASCADE,
  file_id UUID REFERENCES files(id) ON DELETE SET NULL,
  upload_id UUID REFERENCES uploads(id) ON DELETE SET NULL,
  error_code TEXT,
  error_message TEXT NOT NULL,
  error_context JSONB NOT NULL DEFAULT '{}'::jsonb,
  is_retryable BOOLEAN NOT NULL DEFAULT FALSE,
  stack_trace TEXT,
  occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE search_index_status (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
  job_id UUID NOT NULL REFERENCES ingestion_jobs(id) ON DELETE CASCADE,
  index_name TEXT NOT NULL DEFAULT 'parsed_records',
  document_id TEXT,
  status TEXT NOT NULL DEFAULT 'pending',
  attempts INTEGER NOT NULL DEFAULT 0,
  last_error TEXT,
  last_indexed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT search_index_status_attempts_non_negative CHECK (attempts >= 0),
  CONSTRAINT search_index_status_unique UNIQUE (tenant_id, file_id, job_id, index_name)
);

CREATE TABLE archive_password_refs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
  upload_id UUID REFERENCES uploads(id) ON DELETE CASCADE,
  password_ref_hash TEXT NOT NULL,
  algorithm TEXT NOT NULL DEFAULT 'sha256',
  is_valid BOOLEAN NOT NULL DEFAULT FALSE,
  validated BOOLEAN NOT NULL DEFAULT FALSE,
  attempt_count INTEGER NOT NULL DEFAULT 0,
  last_validated_at TIMESTAMPTZ,
  created_by UUID REFERENCES users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT archive_password_refs_attempts_non_negative CHECK (attempt_count >= 0),
  CONSTRAINT archive_password_refs_unique UNIQUE (tenant_id, file_id, password_ref_hash)
);

CREATE TABLE audit_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  actor_user_id UUID REFERENCES users(id),
  action TEXT NOT NULL,
  entity_type TEXT,
  entity_id UUID,
  details JSONB NOT NULL DEFAULT '{}'::jsonb,
  ip_address INET,
  user_agent TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE saved_searches (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  query JSONB NOT NULL DEFAULT '{}'::jsonb,
  filters JSONB NOT NULL DEFAULT '{}'::jsonb,
  is_default BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT saved_searches_user_name_unique UNIQUE (tenant_id, user_id, name)
);

CREATE TABLE system_settings (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  setting_key TEXT NOT NULL,
  setting_value JSONB NOT NULL,
  setting_type TEXT NOT NULL DEFAULT 'json',
  description TEXT,
  is_secret BOOLEAN NOT NULL DEFAULT FALSE,
  updated_by UUID REFERENCES users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT system_settings_tenant_key_unique UNIQUE (tenant_id, setting_key)
);

-- Indexes
CREATE INDEX tenants_slug_idx ON tenants(slug);
CREATE INDEX users_tenant_id_idx ON users(tenant_id);
CREATE INDEX roles_tenant_id_idx ON roles(tenant_id);
CREATE INDEX user_roles_tenant_id_idx ON user_roles(tenant_id);
CREATE INDEX user_roles_user_id_idx ON user_roles(user_id);
CREATE INDEX user_roles_role_id_idx ON user_roles(role_id);
CREATE INDEX uploads_tenant_id_idx ON uploads(tenant_id);
CREATE INDEX files_tenant_id_idx ON files(tenant_id);
CREATE INDEX files_parent_file_id_idx ON files(parent_file_id);
CREATE INDEX files_upload_id_idx ON files(upload_id);
CREATE INDEX files_sha256_hash_hash_idx ON files USING hash (sha256_hash);
CREATE INDEX ingestion_jobs_tenant_id_idx ON ingestion_jobs(tenant_id);
CREATE INDEX ingestion_jobs_root_file_id_idx ON ingestion_jobs(root_file_id);
CREATE INDEX ingestion_jobs_status_idx ON ingestion_jobs(status);
CREATE INDEX job_events_tenant_id_idx ON job_events(tenant_id);
CREATE INDEX job_events_job_id_idx ON job_events(job_id);
CREATE INDEX parsed_records_tenant_id_idx ON parsed_records(tenant_id);
CREATE INDEX parsed_records_file_id_idx ON parsed_records(file_id);
CREATE INDEX parsed_records_job_id_idx ON parsed_records(job_id);
CREATE INDEX parsed_records_record_number_idx ON parsed_records(record_number);
CREATE INDEX parser_errors_tenant_id_idx ON parser_errors(tenant_id);
CREATE INDEX parser_errors_job_id_idx ON parser_errors(job_id);
CREATE INDEX search_index_status_tenant_id_idx ON search_index_status(tenant_id);
CREATE INDEX search_index_status_status_idx ON search_index_status(status);
CREATE INDEX archive_password_refs_tenant_id_idx ON archive_password_refs(tenant_id);
CREATE INDEX archive_password_refs_file_id_idx ON archive_password_refs(file_id);
CREATE INDEX audit_logs_tenant_id_idx ON audit_logs(tenant_id);
CREATE INDEX audit_logs_actor_user_id_idx ON audit_logs(actor_user_id);
CREATE INDEX saved_searches_tenant_id_idx ON saved_searches(tenant_id);
CREATE INDEX saved_searches_user_id_idx ON saved_searches(user_id);
CREATE INDEX system_settings_tenant_id_idx ON system_settings(tenant_id);

-- JSONB indexes
CREATE INDEX parsed_records_structured_data_gin_idx ON parsed_records USING GIN (structured_data);
CREATE INDEX parsed_records_extracted_entities_gin_idx ON parsed_records USING GIN (extracted_entities);
CREATE INDEX saved_searches_query_gin_idx ON saved_searches USING GIN (query);
CREATE INDEX audit_logs_details_gin_idx ON audit_logs USING GIN (details);

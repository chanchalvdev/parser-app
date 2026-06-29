# Product PRD Summary (MVP)

## Purpose

This project is a local-first enterprise ingestion and discovery platform that:

- ingests files from browser, browser via presigned object upload, and worker pipeline
- extracts text/records from TXT, CSV, JSON/JSONL, and archives
- stores parsed records in PostgreSQL
- indexes records and file metadata in OpenSearch
- exposes API and UI surfaces for monitoring, search, and operations

## Primary user workflows

1. Upload
   - User opens UI, provides file + optional archive password
   - API issues presigned URL
   - Browser uploads directly to object storage
   - API marks upload complete and enqueues processing job
2. Process
   - Worker downloads and detects file type
   - Parsers emit one parsed record per object/line
   - Worker writes records to DB and indexes to search
3. Monitor
   - User views files/jobs/file trees/job timelines
   - Operators retry failed jobs and submit archive passwords
4. Search
   - Users run query + filters and navigate to relevant records
5. Admin
   - Settings changes, parser toggles, runtime limits, and batch sizes
   - Settings writes are audited

## MVP requirements covered

- Upload service:
  - presigned upload + complete flow
  - size/type validation
  - max size constraints
- Parser support:
  - `.txt`, `.csv`, `.json`, `.jsonl`
  - JSON array/object/lines modes and malformed JSONL continuation behavior
- Archive processing:
  - recursive extraction with safety controls
  - password required and wrong password recovery workflows
- Records and indexing:
  - Postgres persistence with structured JSON and entity extraction
  - OpenSearch indexing and search API with filters, pagination, sorting, highlights
- APIs:
  - files/jobs/search/dashboard/admin/password/jobs/retry endpoints
  - health/readiness
- Operations:
  - job_events as authoritative timeline
  - audit log coverage for important actions
- UI:
  - upload, files, jobs, search, dashboard, admin settings, detail pages

## Out-of-scope (MVP)

- Full multi-tenant auth with identity provider
- Advanced export, bulk download, and custom roles
- Full threat scanning/AV integration
- Complete production-grade RBAC and secrets platform

## Success criteria

- File ingestion path from UI to search completes end-to-end for common fixtures
- Job state transitions are visible and consistent
- Search returns records with highlights and filter correctness
- Failed archive requiring password can be recovered via password submission flow
- Dashboard shows meaningful totals and recency indicators

## Current risks

- Some endpoints still use placeholder/middleware placeholders for future RBAC
- Search relevance tuning is basic and can evolve with domain-specific scoring
- Some quality gates are placeholders and still being expanded (linting, coverage depth)


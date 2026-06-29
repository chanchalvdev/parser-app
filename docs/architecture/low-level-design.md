# Low-Level Design

## 1. API Layer (`apps/api`)

### Router and middleware
- `internal/http/routes/router.go` defines all REST route wiring.
- Middleware chain: Request ID, auth placeholder, CORS, recovery, structured logging.
- API is stateless and relies on context values (`tenant_id`, roles, request metadata) for trace/audit context.

### Handlers
- `health`, `version`, `upload`, `file`, `job`, `search`, `dashboard`, `admin`.
- All handlers convert service errors to HTTP status using thin wrapper functions in handler files.

### Service layer
- Encapsulates orchestration logic:
  - upload lifecycle (`UploadService`)
  - file inventory (`FileService`)
  - ingestion jobs (`JobService`)
  - search (`SearchService`)
  - dashboard aggregates (`DashboardService`)
  - runtime settings (`AdminSettingsService`)

### Repository layer
- PostgreSQL repositories for uploads/files/jobs/job_events/parsed_records/settings/audit logs.
- Repositories return sorted paginated slices and counts.

### Search client
- Internal OpenSearch client wrapper (`internal/search`) used by search service.
- Supports record search, facet extraction, and file-type resolution.

## 2. Worker Layer (`apps/worker`)

### Orchestrator and parsing
- Worker consumes Redis queue (`WORKER_QUEUE_NAME`).
- Downloads file by queue payload.
- Detects MIME/extension and applies parsers.

### Loader flow
- DB loader inserts parsed records with batch handling.
- Search loader indexes into OpenSearch `parsed-records` and `files`.

### State transition + audit
- Worker updates job status/progress in PostgreSQL.
- Password requests create and consume password refs via `archive_password_refs`.

## 3. Search indexing model

### `parsed-records`
- Contains text content, entity JSON, timestamps, tenancy, identifiers.

### `files`
- Contains file metadata and parsing status for archive tree rendering.

## 4. Frontend (`apps/web`)

- React routes for upload, files, jobs, search, dashboard, admin.
- API client wrappers with typed payloads.
- Pages consume paginated API endpoints and render trees/results.

## 5. Security modules

- Path normalization and extension checks on upload.
- CORS with explicit allowed origins in env (`ALLOWED_ORIGINS`).
- Placeholder auth and role guards (`middleware.RequireRole`).
- Security sensitive operations are audit-logged.

## 6. Data model highlights

- **Upload**: input upload session metadata + object key.
- **File**: all discovered/expanded files and archive child links.
- **Job**: status timeline anchor.
- **ParsedRecord**: parsed output for search.
- **JobEvent**: immutable timeline/events for troubleshooting.

## 7. Default pagination and filtering behavior

- Pagination defaults are enforced by services.
- Maximum page sizes and defaults:
  - file listing: default 25, max 200
  - job listing: default 25, max 200
  - search: default 25, max 100

## 8. Error envelope

Most failures are returned as:

```json
{"error":"message"}
```

Status mapping follows handler mapping and service-defined error classes.


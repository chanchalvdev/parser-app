# REST API Reference

Swagger/OpenAPI is available interactively at `GET /docs` and as JSON at `GET /docs/openapi.json`.

## Base paths
- Service root: `http://localhost:8080`
- API base: `/api/v1`

## Standard response

Most non-list endpoints return JSON. Error responses are:

```json
{"error": "<human readable message>"}
```

## Health

### `GET /health`
- Returns service health and server time.

### `GET /ready`
- Returns readiness, checks DB dependency.
- Returns `503` when dependency checks fail.

## Upload

### `POST /api/v1/uploads/initiate`
- Request: `UploadInitiateRequest`
  - `file_name` (required)
  - `content_type` (required)
  - `size_bytes` (required)
  - `password_provided` (optional)
- Response: `201` with `UploadInitiateResponse`.

### `POST /api/v1/uploads/complete`
- Request: `UploadCompleteRequest`
- Response: `201` with `UploadCompleteResponse`.

### `GET /api/v1/uploads/{upload_id}`
- Returns upload row.

## Files

### `GET /api/v1/files`
- Query: `status|processing_status`, `extension`, `detected_file_type`, `tenant_id`, `page`, `page_size`, `limit` (legacy alias).
- Response: `FileListResponse`

### `GET /api/v1/files/{file_id}`
- Response: `File`

### `POST /api/v1/files/{file_id}/password`
- Request: `SubmitFilePasswordRequest` (`password`).
- Requeues the associated job after storing the encrypted/encoded password reference and sets status to queued when possible.

### `GET /api/v1/files/{file_id}/children`
- Query: `page`, `page_size`
- Response: `FileChildrenResponse`

### `GET /api/v1/files/{file_id}/tree`
- Response: `FileTreeNode` recursive tree view.

### `GET /api/v1/files/{file_id}/records`
- Query: `page`, `page_size`
- Response: `FileRecordsResponse`

## Jobs

### `GET /api/v1/jobs`
- Query: `tenant_id`, `page`, `page_size`
- Response: `JobListResponse`

### `GET /api/v1/jobs/{job_id}`
- Response: `IngestionJob`

### `GET /api/v1/jobs/{job_id}/events`
- Query: `page`, `page_size`
- Response: `JobEventsResponse` (sorted by event time asc)

### `POST /api/v1/jobs/{job_id}/retry`
- Requeues job and increments `retry_count`.
- Returns updated `IngestionJob`.

## Search

### `GET /api/v1/search`
- Query parameters:
  - text: `q`
  - filters: `file_id`, `extension`, `detected_file_type`, `record_type`, `date_from`, `date_to`, `ip`, `email`, `domain`, `job_id`, `tenant_id`
  - paging: `page`, `page_size`
  - `sort`: `relevance` or `created_at`
- Response: `SearchResponse`

### `POST /api/v1/search`
- Request body: `SearchRecordsRequest` with same schema as GET.
- Response: `SearchResponse`

### `GET /api/v1/search/suggestions`
- Query subset (`q`, `extension`, `detected_file_type`).
- Response: `SearchSuggestionResponse`

## Dashboard

### `GET /api/v1/dashboard/summary`
- Returns aggregate counts (`DashboardSummary`).

### `GET /api/v1/dashboard/file-types`
- Optional `limit` (default 25, max 200)
- Returns `DashboardDistribution` of detected file types.

### `GET /api/v1/dashboard/processing-status`
- Optional `limit`.
- Returns `DashboardDistribution` of status distribution.

### `GET /api/v1/dashboard/upload-volume`
- Optional `days` (default 7, max 365)
- Returns `DashboardUploadVolume` with per-day buckets.

### `GET /api/v1/dashboard/error-breakdown`
- Optional `limit`.
- Returns `DashboardErrorBreakdown`.

### `GET /api/v1/dashboard/entities`
- Optional `limit`.
- Returns `DashboardEntities` grouped by type.

### `GET /api/v1/dashboard/processing-duration`
- Returns quantiles and duration metrics.

## Admin

### `GET /api/v1/admin/settings`
- Optional `tenant_id`.
- Returns all configured settings with defaults.

### `PUT /api/v1/admin/settings`
- Placeholder RBAC via `RequireRole("admin")`.
- Body: `UpdateAdminSettingsRequest`.
- Returns updated settings.

## Local docs and examples

- OpenAPI: `GET /docs` and `GET /docs/openapi.json`
- End-to-end examples and troubleshooting: check `docs/runbooks/local-development.md`
- Security model: `docs/security/threat-model.md`


# Plan: Run Enterprise File Ingestion Platform + Enable 5GB Uploads

## Context

The project is a local-first multi-service file ingestion platform (`parser/`) consisting of:

- **Go API** (`apps/api`) on host port `8088` — chi-router HTTP server with presigned-URL upload flow
- **Python worker** (`apps/worker`) — consumes Redis queue, parses uploaded files
- **React frontend** (`apps/web`) on host port `5188`
- **PostgreSQL** (host `5434`), **Redis** (`6379`), **MinIO** (`9000`/`9001`), **OpenSearch** (`9200`), **Dashboards** (`5601`)

Current upload-size policy (read from PostgreSQL `system_settings.max_upload_size_mb`):

| Source | Value |
| --- | --- |
| `infra/migrations/seed.sql` (default row) | `512` MB |
| `apps/api/internal/services/admin_settings_service.go` `defaultAdminMaxUploadSizeMB` | `512` MB |
| `infra/docker/minio-proxy.conf` `client_max_body_size` | `1024m` |
| `apps/web/src/pages/AdminSettingsPage.tsx` UI default | `512` |
| `apps/web/src/services/auditService.ts` local-storage default | `250` MB |
| `.env` `MAX_UPLOAD_SIZE_MB` | `250` (legacy; unused by active API) |

The Go API uses **presigned URL** upload (client PUTs file directly to MinIO via nginx), so the data path is **client → nginx → MinIO**. Only the `InitiateUpload` JSON and `CompleteUpload` JSON cross the Go API, both are tiny — no Go-side body-size limit applies to the file itself. The load-bearing limits are therefore:

1. `max_upload_size_mb` in PostgreSQL (validated by `apps/api/internal/services/upload_service.go::InitiateUpload`)
2. nginx `client_max_body_size` (`infra/docker/minio-proxy.conf`)
3. MinIO default 5 TiB per object — fine for 5 GB single PUT (single PUT works up to 5 GiB; presigned PUT in Go SDK supports up to 5 GiB)

## Goal

Bring the full stack up successfully (`make up`, migrate, seed, search-init, e2e-upload) AND change the system to allow **5 GB (5120 MB)** raw uploads end-to-end without changing anything outside the relevant configuration files.

## Changes (6 files)

### 1. `parser/infra/migrations/seed.sql`
Change `'max_upload_size_mb', to_jsonb(512)` → `'max_upload_size_mb', to_jsonb(5120)` (with description "Maximum raw upload size in MB (5GB)").

### 2. `parser/apps/api/internal/services/admin_settings_service.go`
Change `defaultAdminMaxUploadSizeMB = 512` → `defaultAdminMaxUploadSizeMB = 5120`.

### 3. `parser/infra/docker/minio-proxy.conf`
- `client_max_body_size 1024m;` → `client_max_body_size 6144m;` (6 GB gives 1 GB headroom for multipart overhead / finalisation).
- Add proxy timeouts appropriate for multi-GB PUTs:
  - `proxy_read_timeout 3600s;`
  - `proxy_send_timeout 3600s;`
  - `proxy_connect_timeout 60s;`
  - `client_body_timeout 3600s;`
  - `proxy_request_buffering off;` (stream the request body rather than buffering to disk — required for large uploads).
  - `proxy_buffering off;`
- Update the inline comment to mention 5 GB target.

### 4. `parser/.env` and `parser/.env.example`
Change `MAX_UPLOAD_SIZE_MB=250` → `MAX_UPLOAD_SIZE_MB=5120` (legacy / informational — the active API reads from PostgreSQL, but keeping the value consistent avoids confusion).

### 5. `parser/apps/web/src/pages/AdminSettingsPage.tsx`
Change `max_upload_size_mb: 512` → `max_upload_size_mb: 5120` in `DEFAULT_ADMIN_SETTINGS`.

### 6. `parser/apps/web/src/services/auditService.ts`
Change `max_file_size_mb: 250` → `max_file_size_mb: 5120` in `loadSettings()` default.

## Files NOT touched (deliberate)

- `parser/apps/api/internal/services/upload_service.go` — no code change needed; it already converts MB → bytes (`int64(rawMB * 1024 * 1024)`) and works for any positive value.
- `parser/apps/api/internal/services/admin_settings_service.go::UpdateSettings` — already accepts any positive value.
- `parser/apps/api/internal/http/routes/router.go` — chi does not impose a body limit on JSON routes; the file body never crosses the Go API anyway.
- `parser/apps/worker/app/config.py` `WORKER_MAX_EXTRACTED_SIZE_MB` — this is the **archive extraction** ceiling, not the raw upload ceiling. A 5 GB archive will hit this; raising it is a separate decision and out of scope for "accept a 5 GB file" (the user said "5 GB size file", not "5 GB extracted payload").
- `parser/backend/` — legacy API; not built by `docker-compose.yml` (which uses `apps/api/Dockerfile`).

## Verification

After applying changes:

1. `cd parser && make up` — start the full stack.
2. `cd parser && make migrate` — apply schema migrations.
3. `cd parser && make seed` — seed default rows. After seed, `SELECT setting_value FROM system_settings WHERE setting_key='max_upload_size_mb';` must return `5120`.
4. `cd parser && make search-init` — create OpenSearch indexes.
5. `curl -fsS http://localhost:8088/health` → `{"status":"ok"}`.
6. `curl -fsS http://localhost:9200/_cat/indices?v` → at least `parsed-records` and `files`.
7. `cd parser && make e2e-upload` — local E2E smoke test passes (small dummy file round-trip).
8. **5 GB limit assertion**: `curl -fsS -X POST http://localhost:8088/api/v1/uploads/initiate -H "Content-Type: application/json" -d '{"file_name":"big.bin","content_type":"application/octet-stream","size_bytes":5368709120,"password_provided":false}'` → `200 Created` with `upload_url` (proves the API accepts a 5 GiB `size_bytes`). Then PUT a real 5 GB file to the presigned URL and POST `/uploads/complete` (optional but strongest evidence).

## Risk / rollback

- All changes are config-defaults; production deployments already override settings via the `/api/v1/admin/settings` endpoint.
- The `client_max_body_size 6144m` raise on nginx is local-only (the proxy container is local-only).
- Rollback = `git checkout --` on the touched files.

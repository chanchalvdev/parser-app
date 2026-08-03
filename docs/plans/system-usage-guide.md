# Enterprise File Ingestion Platform — Usage Guide

A practical guide to using the running system. Assumes you have already run
`make up && make migrate && make seed && make search-init` from `parser/`.

---

## 1. Where everything lives

| Surface | URL | Notes |
|---|---|---|
| Web UI (React + Vite dev server) | http://localhost:5188 | The polished dashboard, upload, jobs, search, files, and settings pages |
| API (Go) | http://localhost:8088 | REST + Swagger at `/docs`, OpenAPI at `/docs/openapi.json` |
| API health | http://localhost:8088/health | Always 200 if the process is alive |
| API readiness | http://localhost:8088/ready | 200 only when DB / Redis / OpenSearch / MinIO are all reachable |
| OpenSearch | http://localhost:9200 | Search engine — `parsed-records` and `files` indexes |
| OpenSearch Dashboards | http://localhost:5601 | Browse raw indexed documents |
| MinIO console | http://localhost:9001 | Login with `MINIO_ACCESS_KEY` / `MINIO_SECRET_KEY` from `parser/.env` |
| Postgres | `localhost:5434` | DB credentials in `parser/.env` |
| Redis | `localhost:6379` | Job queue |

> **Port note:** The README quotes `5173` and `8080`, but `docker-compose.yml`
> actually maps host **5188 → container 5173** (web) and **8088 → container 8080**
> (API). Use the host ports above.

---

## 2. The end-to-end flow

```
Upload file  →  presigned PUT to MinIO  →  mark complete
       ↓
Worker picks up job from Redis queue
       ↓
Download from MinIO  →  detect type  →  parse  →  extract records
       ↓
Insert records into Postgres  →  index into OpenSearch
       ↓
Job status transitions: queued → running → completed (or failed / password_required)
```

Every step is observable in the UI and via the API.

---

## 3. Using the Web UI

Open http://localhost:5188 — the left sidebar collapses to a drawer on mobile,
the top bar has a search shortcut, theme toggle, and breadcrumb.

### 3.1 Dashboard (`/`)
At-a-glance tiles for total uploads, files, parsed records, and jobs in
progress, plus a stacked status breakdown, recent jobs (click a row to open
the **Job details** modal — full job info, events timeline, and retry-with-password
form for failed jobs), and three quick-action buttons.

### 3.2 Upload (`/upload`)
- Drag-and-drop a file (or click the zone to browse). Multi-file supported.
- If you drop a `.zip` / `.rar` / `.7z` / `.tar.*`, an **Archive password** field
  appears. Leave blank for unencrypted archives.
- Click **Upload N files**. The button shows a spinner and disables while in
  flight; per-file success or failure appears in the **Upload results** card
  with copy-able job IDs and a link to the job.
- Below that, the **Recent uploads** card lists the last five uploads from the
  platform — useful for jumping back to something you just uploaded.

> **Supported formats:** txt, log, csv, tsv, json, jsonl, xml, xlsx, pdf.
> Archives (zip / tar / tar.gz / 7z / rar) are extracted before parsing.
> Max upload size defaults to 512 MB (configurable on the Settings page).

### 3.3 Jobs (`/jobs`)
- Auto-refreshes every 5 seconds; pauses when the browser tab is hidden
  (look for the **Live** / **Paused** dot in the header).
- Filter by status (click chips), search by job id or upload id.
- Click any row to open the **Job details** modal:
  - Full job metadata with copy buttons for ids
  - Recent events timeline (from `/api/v1/jobs/{id}/events`)
  - Inline **retry with password** form (visible only for failed jobs)

### 3.4 Search (`/search`)
- Hero search box. Typing is debounced 250 ms.
- Result cards show filename, path, score badge, and a highlighted snippet
  (`<mark>` around the matched substring).
- The empty state suggests example queries like `invoice`, `contract`, `log error`.
- Click **Open in Files** to jump to the Files page.

### 3.5 Files (`/files`)
Two-pane browser:
- **Left:** all uploads, filterable. Each row shows a file-type icon, name,
  size, date, and upload status badge.
- **Right:** collapsible tree of files for the selected upload. Click a node to
  see details (filename, kind, MIME, size, sha256 with copy, path). The
  details panel has a button to jump back to Search pre-filled with that
  filename.

### 3.6 Settings (`/settings`)
Three sections:
- **Upload limits** — `max_upload_size_mb` is editable and persists immediately.
  Other fields (`max_archive_depth`, `max_extracted_files`, etc.) are
  read-only in the UI.
- **Processing** — placeholder for parser/worker tuning (read-only in this
  build).
- **Recent audit activity** — last 20 platform events.
- **Danger zone** — reset-all-data button is disabled.

---

## 4. Using the API directly

Base URL: `http://localhost:8088/api/v1`. All endpoints accept an optional
`X-API-Key` header (set `API_KEY` in `parser/.env` if you enabled auth).

### 4.1 Quick checks
```bash
curl -s http://localhost:8088/health | jq
curl -s http://localhost:8088/ready   | jq
curl -s http://localhost:8088/api/v1/version | jq
```

### 4.2 Upload a file (the canonical 3-step flow)

The backend uses presigned uploads to MinIO, so you don't push the file
through the API. The flow is **initiate → PUT to MinIO → complete**:

```bash
# 1) Initiate — ask the API for an upload id and a presigned PUT URL
INIT=$(curl -s -X POST http://localhost:8088/api/v1/uploads/initiate \
  -H 'Content-Type: application/json' \
  -d '{
        "file_name": "sample.txt",
        "content_type": "text/plain",
        "size_bytes": 163,
        "password_provided": false
      }')
UPLOAD_ID=$(echo "$INIT" | jq -r '.upload_id')
UPLOAD_URL=$(echo "$INIT" | jq -r '.upload_url')
echo "upload_id=$UPLOAD_ID"

# 2) PUT the bytes directly to MinIO
curl -s -X PUT -T ./sample.txt "$UPLOAD_URL"

# 3) Mark complete — the API enqueues a parsing job
COMPLETE=$(curl -s -X POST http://localhost:8088/api/v1/uploads/complete \
  -H 'Content-Type: application/json' \
  -d "{\"upload_id\": \"$UPLOAD_ID\"}")
FILE_ID=$(echo "$COMPLETE" | jq -r '.file_id')
JOB_ID=$(echo "$COMPLETE" | jq -r '.job_id')
echo "file_id=$FILE_ID  job_id=$JOB_ID"
```

### 4.3 Watch a job
```bash
curl -s http://localhost:8088/api/v1/jobs/$JOB_ID | jq
# { id, status, current_stage, progress_percent, retry_count, ... }

curl -s http://localhost:8088/api/v1/jobs/$JOB_ID/events | jq
# chronological events emitted by the worker
```

`status` will transition through `queued` → `running` → `completed`. If the
file is encrypted you'll see `password_required` instead of `completed` —
submit the password and the job will resume:

```bash
curl -s -X POST http://localhost:8088/api/v1/jobs/$JOB_ID/retry \
  -H 'Content-Type: application/json' \
  -d '{"password":"hunter2"}'
```

### 4.4 Browse parsed records
```bash
curl -s "http://localhost:8088/api/v1/files/$FILE_ID/records?page=1&page_size=10" | jq
# { total, records: [ { record_id, content_preview, entities, ... } ] }
```

### 4.5 Search
```bash
curl -s "http://localhost:8088/api/v1/search?q=dummy&limit=10" | jq
# { total, results: [ { file_id, source_file_name, content_preview, highlight, entities } ], facets }
```

The `entities` block on each hit lists detected emails, IPs, URLs, hashes,
and domains — useful for pivoting from a hit to its structured metadata.

### 4.6 File tree
```bash
curl -s "http://localhost:8088/api/v1/files/$FILE_ID/tree" | jq
```

### 4.7 Dashboard summary
```bash
curl -s http://localhost:8088/api/v1/dashboard/summary | jq
# { tenant_id, total_uploads, total_files, total_parsed_records,
#   completed_jobs, failed_jobs, password_required_files, quarantined_files }
```

There are also dedicated chart endpoints used by the dashboard:
`/dashboard/file-types`, `/dashboard/processing-status`,
`/dashboard/upload-volume`, `/dashboard/error-breakdown`,
`/dashboard/entities`, `/dashboard/processing-duration`.

### 4.8 Settings
```bash
# Read
curl -s http://localhost:8088/api/v1/admin/settings | jq

# Update
curl -s -X PUT http://localhost:8088/api/v1/admin/settings \
  -H 'Content-Type: application/json' \
  -d '{"max_upload_size_mb": 1024}'
```

### 4.9 Search suggestions (autocomplete)
```bash
curl -s "http://localhost:8088/api/v1/search/suggestions?q=inv&limit=5" | jq
```

---

## 5. One-shot end-to-end smoke test

The fastest way to verify everything works:

```bash
cd parser
make e2e-upload
```

What it does (see `parser/scripts/e2e-upload.sh`):
1. Creates `/tmp/e2e-dummy-upload.txt` with placeholder content
   (`alice@example.com`, `203.0.113.42`, `abc1234-local-test-token`).
2. Initiates an upload session, PUTs the file to MinIO, completes the upload.
3. Polls the job every 2 s (timeout 180 s) until it reaches a terminal status.
4. Fetches parsed records and asserts there is at least one.
5. Runs three search assertions: `dummy-local-test`, `alice@example.com`,
   `203.0.113.42` — each must return ≥1 hit.
6. Prints the dashboard summary.

Override defaults with env vars:
```bash
API_BASE_URL=http://localhost:8088/api/v1 \
DUMMY_FILE_PATH=/path/to/your.txt \
SEARCH_ASSERTIONS="invoice contract 192.168" \
make e2e-upload
```

---

## 6. Generating sample archives for manual testing

```bash
cd parser
python3 infra/scripts/create_sample_archives.py
```

Produces under `tests/fixtures/`:
- `sample.zip`, `sample_nested.zip`, `sample.tar.gz`
- `sample.7z` (only if `py7zr` is installed)
- `password_sample.zip` (only if `pyzipper` is installed)

Then upload any of these from the UI to exercise the archive-extraction path.

---

## 7. Day-to-day commands

```bash
cd parser

# Status of all containers
make ps
docker compose ps

# Follow logs (all services, or pick one)
make logs
docker compose logs -f api
docker compose logs -f worker
docker compose logs -f api worker --tail=200

# Restart a single service
docker compose restart api

# Tear it all down (keeps volumes)
make down

# Nuke everything including data volumes
make clean

# Re-seed after a reset
make migrate && make seed && make search-init

# Inspect what's in the queue
docker compose exec redis redis-cli LLEN ingestion_jobs
```

---

## 8. Troubleshooting

| Symptom | First thing to check |
|---|---|
| API returns 500 / `/ready` fails | `docker compose logs api` and `docker compose logs postgres`; verify `DATABASE_URL` in `.env` |
| Upload button shows error | Confirm `make search-init` ran and MinIO bucket exists (MinIO console at `:9001`) |
| Job stuck on `queued` | `docker compose logs worker`; check the queue isn't empty: `docker compose exec redis redis-cli LLEN ingestion_jobs` |
| Search returns 0 hits | `curl http://localhost:9200/_cat/indices?v` — should list `parsed-records` and `files`; re-run `make search-init` if missing |
| CORS errors in the browser | Confirm `VITE_API_BASE_URL` in `.env` matches `http://localhost:8088` and `ALLOWED_ORIGINS` in the API env allows `http://localhost:5188` |
| OpenSearch OOM | Bump Docker memory limit (Docker Desktop → Settings → Resources); the default `-Xms512m -Xmx512m` JVM heap is set in `docker-compose.yml` |
| Worker boots but never picks up jobs | Verify `WORKER_QUEUE_NAME` in `.env` matches what the API enqueues to; both default to `ingestion_jobs` |

---

## 9. What the system is *not* doing (today)

Useful to know before you assume behavior:
- **No auth UI** — the API key gate is environment-driven. Set `API_KEY` in
  `.env` to enable; the React UI does not surface a login screen.
- **No real upload progress** — the API only reports "complete" or "error".
  The progress bar in the UI is optimistic.
- **Cursor pagination** — Jobs and Files use page-based pagination only;
  "Load more" on Jobs is a stub.
- **Settings page** — only `max_upload_size_mb` is writable through the UI;
  other tunables live in the `.env` file.
- **Production hardening** — no TLS termination, no rate limiting, no
  multi-tenant isolation beyond the seeded single tenant.

# Enterprise File Ingestion Platform

## What this project is

Enterprise File Ingestion Platform is a local-first scaffold for:

- uploading files (presigned direct-to-object-store PUT, multi-GB capable)
- recursively extracting archives (ZIP, RAR, TAR, TAR.GZ/TGZ, 7Z — including nested and password-protected)
- parsing content into structured records (TXT/LOG, CSV, JSON/JSONL, XML, XLSX, PDF, generic text)
- extracting entities/IOCs from every record (IPv4, emails, URLs, domains, timestamps, MD5/SHA1/SHA256, CVE IDs, Bitcoin addresses)
- normalizing infostealer log dumps into unified credential records (secrets hashed, never stored in plaintext)
- indexing parsed records into OpenSearch
- monitoring jobs and files through a Go API + React dashboard
- storing metadata/events in PostgreSQL and passing work through Redis

The project is production-pattern aligned for a single-tenant local deployment and includes APIs, workers, dashboards, search, and security/audit hooks.

## Architecture (text)

```text
Client (React Web)
  -> Go API (:8080 in container, :8088 on host)
     - writes metadata + jobs to PostgreSQL
     - enqueues work in Redis
     - issues presigned upload PUT URLs (MinIO via nginx proxy)
     - exposes REST + /docs (Swagger)
     -> PostgreSQL
     -> Redis
     -> MinIO
     -> OpenSearch

Go API -> Worker Container
  -> downloads uploaded file from MinIO
  -> detects file type (libmagic + extension + null-byte sampling)
  -> extracts archives recursively under depth/count/size/expansion limits
  -> parses and extracts records + entities
  -> inserts parsed records in PostgreSQL (batched)
  -> indexes parsed_records + files metadata in OpenSearch (batched)
  -> emits job events, progress updates, and writes audit records
  -> updates job progress/state
```

```text
Core data stores/services
- PostgreSQL: authoritative metadata, jobs, events, audit logs, settings
- Redis: job queue and work orchestration
- MinIO: raw object payloads (fronted by nginx for large-body PUTs)
- OpenSearch: parsed record + file search surface
```

## Tech stack

- Go 1.23 (HTTP API, queue producer/consumers adapters)
- Python 3.10+ (worker orchestrator/parsers/extractors/loaders)
- React 18 + TypeScript + Vite 6 + Tailwind (UI)
- PostgreSQL 16
- Redis 7
- MinIO (S3-compatible storage) + nginx 1.27 proxy
- OpenSearch 2.13 + Dashboards
- Docker Compose + Make

## Local quick start

```bash
cp .env.example .env
make up
make migrate
make seed
make search-init
```

Open:
- Web UI: `http://localhost:5188`
- API health: `http://localhost:8088/health`
- API docs: `http://localhost:8088/docs`
- OpenSearch: `http://localhost:9200`

## Services and ports

Host port on the left, container port on the right.

| Service | Host | Container |
|---|---|---|
| `postgres` | 5434 | 5432 |
| `redis` | 6379 | 6379 |
| `minio-proxy` (nginx → MinIO S3 API) | 9000 | 9000 |
| `minio` console | 9001 | 9001 |
| `opensearch` | 9200 | 9200 |
| `opensearch-dashboards` | 5601 | 5601 |
| `api` | 8088 | 8080 |
| `worker` | — (no host port) | — |
| `web` | 5188 | 5173 |

The MinIO S3 API is not published directly; all client PUTs go through `minio-proxy`, which sets
`client_max_body_size` and streaming timeouts sized for multi-GB uploads.

## Supported formats

### Archives (recursive, nested)
`.zip` · `.rar` · `.tar` · `.tar.gz` / `.tgz` · `.7z`

Password-protected archives are supported: the job parks in `PASSWORD_REQUIRED`, the UI/API submits
a password via `POST /api/v1/files/{file_id}/password`, and processing resumes. Wrong passwords are
recorded as attempts against `archive_password_refs`.

### Parsers
Selected in order by `ParserRegistry` (`apps/worker/app/parsers/registry.py`):

| Parser | Extensions / hints |
|---|---|
| `txt` | `txt`, `log`, `out`, `err`, `text/plain` |
| `csv` | `csv` |
| `json` | `json`, `jsonl` |
| `xml` | `xml` |
| `excel` | `xlsx` |
| `pdf` | `pdf` |
| `generic_text` | fallback for anything text-like |

Files with no matching parser and no text-like signature fail with `NO_PARSER_FOUND` and are recorded
in `parser_errors` — a child failure does not abort the whole job, only a root failure does.

### Entity / IOC extraction
Every text record is scanned for: `ipv4`, `emails`, `urls`, `domains`, `timestamps`,
`md5_like_hashes`, `sha1_like_hashes`, `sha256_like_hashes`, `cve`, `bitcoin_addresses`.

### Infostealer log handling
`apps/worker/app/parsers/infostealer.py` adds stealer-log awareness on top of the TXT parser:

- **Filename/directory routing** — `Passwords.txt`, `UserInformation.txt`, `InstalledSoftware.txt`,
  `ProcessList.txt`, `Cookies/`, `Autofills/`, `FileGrabber/` etc. are tagged with a
  `stealer_category` in `structured_data`.
- **Credential-block grammar** — repeated `KEY: value` blocks are normalized across stealer families
  (`SOFT`/`URL`/`USER`/`PASS` and their many spellings), including leetspeak folding so `P455W0RD`
  resolves to `password`.
- **Secret safety** — passwords are SHA-256 hashed into `secret_hash`. Plaintext secrets are never
  persisted or indexed.
- Password dumps additionally emit one consolidated `stealer_credential` record per file, because
  small files are parsed line-by-line and multi-line blocks would otherwise never assemble.

## Limits and settings

Runtime limits are read per tenant from `system_settings` (seeded by `make seed`, editable via
`PUT /api/v1/admin/settings`):

| Setting | Seeded default |
|---|---|
| `max_upload_size_mb` | 10240 (10 GB) |
| `max_archive_depth` | 20 |
| `max_extracted_files` | 1000000 |
| `max_extracted_size_mb` | 102400 |
| `max_expansion_ratio` | 100 |
| `txt_small_file_limit_mb` | 10 |
| `parser_batch_size` | 1000 |
| `search_index_batch_size` | 1000 |
| `enabled_parsers` | txt, log, csv, json, jsonl, xml, xlsx, pdf, text |

Expansion-ratio and file-count limits are the zip-bomb guard; exceeding them fails the job with an
`ArchiveLimitsExceededError` at the `archive_extract` stage.

## API endpoints

### Core
- `GET /health`
- `GET /ready`
- `GET /api/v1/version`
- `GET /docs`
- `GET /docs/openapi.json`

### Upload
- `POST /api/v1/uploads/initiate`
- `POST /api/v1/uploads/complete`
- `GET /api/v1/uploads/{upload_id}`

### Files
- `GET /api/v1/files`
- `GET /api/v1/files/{file_id}`
- `POST /api/v1/files/{file_id}/password`
- `GET /api/v1/files/{file_id}/children`
- `GET /api/v1/files/{file_id}/tree`
- `GET /api/v1/files/{file_id}/records`

### Jobs
- `GET /api/v1/jobs`
- `GET /api/v1/jobs/{job_id}`
- `GET /api/v1/jobs/{job_id}/events`
- `GET /api/v1/jobs/{job_id}/records`
- `POST /api/v1/jobs/{job_id}/retry`

### Search
- `GET /api/v1/search`
- `POST /api/v1/search`
- `GET /api/v1/search/suggestions`

### Dashboard
- `GET /api/v1/dashboard/summary`
- `GET /api/v1/dashboard/file-types`
- `GET /api/v1/dashboard/processing-status`
- `GET /api/v1/dashboard/upload-volume`
- `GET /api/v1/dashboard/error-breakdown`
- `GET /api/v1/dashboard/entities`
- `GET /api/v1/dashboard/processing-duration`

### Admin
Both routes require the `admin` role (`middleware.RequireRole("admin")`).
- `GET /api/v1/admin/settings`
- `PUT /api/v1/admin/settings`

## API docs

Interactive API documentation is available at:
- `http://localhost:8088/docs`

OpenAPI JSON is at:
- `http://localhost:8088/docs/openapi.json`

## Development workflow

1. Edit code under `apps/api`, `apps/worker`, or `apps/web`.
2. Rebuild containers when needed:

```bash
make restart
```

3. Review logs for behavior:

```bash
docker compose logs -f api
docker compose logs -f worker
```

4. Refresh fixtures and index initialization when storage/search state changes:

```bash
make seed
make search-init
```

5. Run local tests and checks:

```bash
make test-api      # Go: go test ./...
make test-worker   # Python: pytest -q tests
make test-web      # Frontend: vitest run
make lint
```

## Common commands

```bash
make up            # start stack
make down          # stop stack
make ps            # status
make logs          # follow logs (all services)
make restart       # restart services
make clean         # remove containers + known local cache dirs
make migrate       # apply schema migrations
make seed          # seed default rows
make search-init   # create OpenSearch mappings
make test          # legacy test entry (infra/scripts/test.sh)
make test-api      # Go API tests
make test-worker   # Python worker tests
make test-web      # Frontend tests
make test-all      # api + worker + web
make e2e-upload    # end-to-end upload/parse/search smoke test
make lint          # lint placeholders and check messages
make fmt           # formatting helpers used by existing scripts
```

## End-to-end demo

1. Start services and initialize dependencies:

```bash
make up && make migrate && make seed && make search-init
```

2. Generate sample fixtures:

```bash
python3 infra/scripts/create_sample_archives.py
```

3. Open the UI `http://localhost:5188` and upload `tests/fixtures/sample.txt`.
4. Confirm:
- upload job created
- file status transitions in Job details
- parsed records are searchable
- dashboard summary updates

### Reusable local E2E upload smoke test

Run a local end-to-end upload+parse+search check with a dummy file:

```bash
make e2e-upload
```

Defaults:
- dummy file: `/tmp/e2e-dummy-upload.txt`
- API base: `http://localhost:8080/api/v1`
- timeout: 180s

Environment overrides:
- `API_BASE_URL` (set to `http://localhost:8088/api/v1` to match the published host port)
- `API_KEY` (if API auth is enabled)
- `DUMMY_FILE_PATH`
- `DUMMY_FILE_NAME`
- `SEARCH_ASSERTIONS` (space-separated queries)
- `JOB_TIMEOUT_SECONDS`
- `POLL_INTERVAL_SECONDS`

## Job lifecycle

```text
queued -> running -> completed
                  -> failed              (root parse/extract error)
                  -> PASSWORD_REQUIRED   (encrypted archive, awaiting password)
                  -> WRONG_PASSWORD      (submitted password rejected)
```

Progress reporting: `mark_job_running` sets 5%; parsing occupies the 10%–90% band and approaches 90%
asymptotically as records stream (total record count is unknown mid-stream); `mark_job_completed`
sets 100%. Progress updates are best-effort and never fail a job.

Job events emitted per file: `worker.started`, `worker.detected_file_type`, `worker.extracted_file`,
`PARSING_STARTED`, `PARSING_PROGRESS`, `PARSING_COMPLETED`, `PARSING_ERROR`, `INDEXING_COMPLETED`,
`INDEXING_FAILED`, `worker.completed`, plus the failure/password variants.

## Repository layout

```text
apps/api/          Go HTTP API (handlers -> services -> repositories)
apps/worker/       Python worker (queue consumer, orchestrator, extractors, parsers, loaders)
apps/web/          React + Vite dashboard
infra/migrations/  SQL schema + seed
infra/opensearch/  index mapping bootstrap
infra/docker/      nginx MinIO proxy config
infra/scripts/     migrate / seed / test / fixture helpers
scripts/           e2e-upload.sh
tests/fixtures/    sample TXT/CSV/JSON/JSONL/LOG + nested archive source
docs/plans/        design + implementation plan notes
```

## Documentation

- [System usage guide](docs/plans/system-usage-guide.md)
- [5 GB upload plan](docs/plans/plan-5gb-uploads.md)
- [Live status + eye icon plan](docs/plans/plan-live-status-and-eye-icon.md)
- [UX redesign spec](docs/plans/ux-redesign-spec.md)

## Troubleshooting

- API/Ready failing:
  - `docker compose logs api`
  - `docker compose logs postgres`
  - Check `DATABASE_URL` in `.env`
- Worker idle/no job progress:
  - `docker compose logs worker`
  - confirm queue list not empty (`redis-cli -h 127.0.0.1 -p 6379 LLEN ingestion_jobs`) in container
- Presigned upload failures:
  - ensure the MinIO bucket exists and `minio-proxy` is healthy
  - check `MINIO_*` settings in `.env` — `MINIO_PRESIGN_ENDPOINT` must be reachable from the browser
  - for large files, confirm `client_max_body_size` in `infra/docker/minio-proxy.conf`
- Upload rejected as too large:
  - `max_upload_size_mb` in `system_settings` is the load-bearing limit, not `.env`
  - check via `GET /api/v1/admin/settings`
- Search returns empty:
  - open `http://localhost:9200/_cat/indices?v`
  - confirm search documents indexed and `make search-init` run after `opensearch` restart
- CORS issues:
  - verify frontend `VITE_API_BASE_URL` and API `ALLOWED_ORIGINS` both reference `http://localhost:5188`
  - check browser network console for preflight errors
- Docker memory:
  - reduce concurrent workload, increase Docker memory limit, inspect OOM logs in `docker compose logs`

## Roadmap

- Hardening RBAC for `/admin` and optional secured `audit log` surfaces
- Expand parser coverage and parser-fallback safety checks
- Add worker and API metrics endpoints
- Improve archive parser observability and recovery workflows
- Introduce long-form job-level search filters and export endpoints
- Add full CI test matrix and code-quality gates

# Enterprise File Ingestion Platform

## What this project is

Enterprise File Ingestion Platform is a local-first scaffold for:

- uploading files
- extracting and parsing content (TXT, JSON/JSONL, CSV, archives)
- indexing parsed records into OpenSearch
- monitoring jobs and files through a Go API + React dashboard
- storing metadata/events in PostgreSQL and passing work through Redis

The project is production-pattern aligned for a single-tenant local deployment and includes APIs, workers, dashboards, search, and security/audit hooks.

## Architecture (text)

```text
Client (React Web)
  -> Go API (:8080)
     - writes metadata + jobs to PostgreSQL
     - enqueues work in Redis
     - receives upload object PUT URLs (MinIO)
     - exposes REST + /docs (Swagger)
     -> PostgreSQL
     -> Redis
     -> MinIO
     -> OpenSearch

Go API -> Worker Container
  -> downloads uploaded file from MinIO
  -> parses and extracts records
  -> inserts parsed records in PostgreSQL
  -> indexes parsed_records + files metadata in OpenSearch
  -> emits job events and writes audit records
  -> updates job progress/state
```

```text
Core data stores/services
- PostgreSQL: authoritative metadata, jobs, events, audit logs, settings
- Redis: job queue and work orchestration
- MinIO: raw object payloads
- OpenSearch: parsed record + file search surface
```

## Tech stack

- Go 1.22+ (HTTP API, queue producer/consumers adapters)
- Python 3.10+ (worker orchestrator/parsers/indexers)
- React + TypeScript + Vite + Tailwind (UI)
- PostgreSQL 16
- Redis 7
- MinIO (S3-compatible storage)
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
- Web UI: `http://localhost:5173`
- API health: `http://localhost:8080/health`
- OpenSearch: `http://localhost:9200`

## Services and ports

- `postgres` (5432)
- `redis` (6379)
- `minio` (9000)
- `minio console` (9001)
- `opensearch` (9200)
- `opensearch-dashboards` (5601)
- `api` (8080)
- `worker` (no host port)
- `web` (5173)

## API endpoints (selected)

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
- `GET /api/v1/admin/settings`
- `PUT /api/v1/admin/settings`

## API docs

Interactive API documentation is available at:
- `http://localhost:8080/docs`

OpenAPI JSON is at:
- `http://localhost:8080/docs/openapi.json`

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
make test-api
make test-worker
make test-web
make lint
```

6. Enforce the harness workflow before handoff:

```bash
make harness      # review/design/plan/agree
make harness-full # add test and validation evidence
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
make test          # legacy test entry
make test-api      # Go API tests
make test-worker   # Python worker tests (if tests are present)
make test-web      # Frontend tests/build check
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
python infra/scripts/create_sample_archives.py
```

3. Open the UI `http://localhost:5173/upload` and upload `tests/fixtures/sample.txt`.
4. Confirm:
- upload job created
- file status transitions in Job details
- parsed records are searchable at `/search`
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
- `API_BASE_URL` (e.g. `http://localhost:8080/api/v1`)
- `API_KEY` (if API auth is enabled)
- `DUMMY_FILE_PATH`
- `DUMMY_FILE_NAME`
- `SEARCH_ASSERTIONS` (space-separated queries)
- `JOB_TIMEOUT_SECONDS`
- `POLL_INTERVAL_SECONDS`

## Troubleshooting

- Docs index:
  - [Local Development Runbook](docs/runbooks/local-development.md)
  - [Troubleshooting Runbook](docs/runbooks/troubleshooting.md)
  - [API Reference](docs/api/rest-api.md)
  - [High-Level Architecture](docs/architecture/high-level-architecture.md)

- API/Ready failing:
  - `docker compose logs api`
  - `docker compose logs postgres`
  - Check `DATABASE_URL` in `.env`
- Worker idle/no job progress:
  - `docker compose logs worker`
  - confirm queue list not empty (`redis-cli -h 127.0.0.1 -p 6379 LLEN ingestion_jobs`) in container
- Presigned upload failures:
  - ensure `make search-init` complete and MinIO bucket exists
  - check `MINIO_*` settings in `.env`
- Search returns empty:
  - open `http://localhost:9200/_cat/indices?v`
  - confirm search documents indexed and `make search-init` run after `opensearch` restart
- CORS issues:
  - verify frontend `VITE_API_BASE_URL` and API `ALLOWED_ORIGINS` match `http://localhost:5173`
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

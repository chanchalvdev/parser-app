# Local Development Runbook

This runbook is for getting the full stack running locally and validating the primary ingestion/search flow.

## Prerequisites

Install and verify:

- Docker (Engine)
- Docker Compose
- Go (matching project requirements)
- Python 3.10+ (for the worker and archive fixture script)
- Node.js 18+ (for frontend tooling)

Also ensure you have `make` available in your shell.

## Setup

1. Copy environment defaults and review overrides

```bash
cp .env.example .env
```

2. Start the full stack

```bash
make up
```

This starts PostgreSQL, Redis, MinIO, OpenSearch, OpenSearch Dashboards, Go API, web app, and Python worker.

3. Apply migrations

```bash
make migrate
```

4. Seed local starter rows

```bash
make seed
```

5. Initialize OpenSearch mappings

```bash
make search-init
```

6. Generate sample fixtures and archives

```bash
python infra/scripts/create_sample_archives.py
```

If your environment supports it, this will also generate:

- `tests/fixtures/sample.7z`
- `tests/fixtures/password_sample.zip`

If not, generation still creates:

- `tests/fixtures/sample.zip`
- `tests/fixtures/sample_nested.zip`
- `tests/fixtures/sample.tar.gz`

## Verify the platform

### 1) API service

```bash
curl -sS http://localhost:8080/health | jq
curl -sS http://localhost:8080/ready | jq
```

Expected response status is HTTP 200.

### 2) PostgreSQL connectivity

```bash
docker compose exec postgres psql "$DATABASE_URL" -c "\l"

docker compose exec postgres psql "$DATABASE_URL" -c "\dt"
```

If `DATABASE_URL` is not exported in the current shell, use direct credentials from `.env`.

### 3) MinIO

- MinIO API: `http://localhost:9000`
- MinIO Console: `http://localhost:9001`

Open the console in a browser and confirm the `file-ingestion` bucket exists.

### 4) OpenSearch

```bash
curl -sS http://localhost:9200/_cluster/health?pretty | jq
curl -sS -u "$OPENSEARCH_USER:$OPENSEARCH_PASSWORD" http://localhost:9200/_cat/indices?v | cat
```

Expected to include `parsed-records` and `files` indexes after `make search-init`.

### 5) Web app

Open `http://localhost:5173` and verify routes load (Dashboard, Upload, Files, Jobs, Search).

## End-to-end validation flow

1. Generate sample artifacts

```bash
python infra/scripts/create_sample_archives.py
```

2. Open web UI at `http://localhost:5173/upload`.

3. Upload `tests/fixtures/sample.zip`.

4. Watch job progress on the upload/job detail screen until job status reaches `completed`.

5. Open the file detail page and verify the archive tree is rendered from `/api/v1/files/{file_id}/tree`.

6. Run a search for known text from fixtures:

```bash
curl -G "http://localhost:8080/api/v1/search" \
  --data-urlencode "q=Authentication attempt" \
  --data-urlencode "page=1" \
  --data-urlencode "page_size=10"
```

(Any of `alice`, `auth-fail`, `198.51.100.12`, or `users.csv` also work.)

7. Open Dashboard and confirm cards/charts reflect the newly ingested records.

## Useful operational commands

```bash
docker compose logs api
docker compose logs worker

docker compose exec postgres psql "$DATABASE_URL" -c "SELECT 1;"

curl -sS http://localhost:8080/health
curl -sS http://localhost:8080/ready
```

## Troubleshooting

### API cannot connect to DB

- Confirm Postgres is running:

  ```bash
  docker compose ps postgres
  ```

- Confirm `.env` has correct `DATABASE_URL` and credentials.
- Check API logs:

  ```bash
  docker compose logs api
  ```

- Verify DB schema is present:

  ```bash
  docker compose exec postgres psql "$DATABASE_URL" -c "\dt"
  ```

### Worker not consuming jobs

- Confirm worker container is up:

  ```bash
  docker compose ps worker
  ```

- Check worker logs for queue/activity and errors:

  ```bash
  docker compose logs worker
  ```

- Confirm job exists and is queued in DB:

  ```bash
  docker compose exec postgres psql "$DATABASE_URL" -c "SELECT id,status,current_stage,progress_percent FROM jobs ORDER BY created_at DESC LIMIT 20;"
  ```

- Verify queue is populated:

  ```bash
  docker compose exec redis redis-cli -p 6379 LLEN ingestion_jobs
  ```

### MinIO upload fails

- Validate presigned upload response and content type are non-empty.
- Confirm bucket exists and MinIO credentials are correct.
- Retry with a small file (`sample.txt`), then try archive.
- Inspect API and MinIO logs:

  ```bash
docker compose logs api
  docker compose logs minio
  ```

### OpenSearch index missing

- Ensure initialization has completed:

  ```bash
  make search-init
  ```

- Re-check cluster/index list:

  ```bash
  curl -sS http://localhost:9200/_cat/indices?v | cat
  ```

- Re-run worker/API pipeline after index recovery:

  ```bash
  curl -sS http://localhost:8080/health | cat
  ```

### CORS issue

- Confirm frontend calls the API base URL `http://localhost:8080/api/v1`.
- Confirm frontend and API URLs match environment values (`VITE_API_BASE_URL`, browser origin).
- Restart web and API containers after env changes:

  ```bash
  docker compose restart api web
  ```

### Docker memory issue

- Restart with reduced local pressure and then increase allocated memory for Docker Desktop.
- Verify containers are not OOM-killing:

  ```bash
  docker compose ps -a
  docker compose logs --tail=100 api
  docker compose logs --tail=100 worker
  ```

- Temporarily reduce load:
  - process smaller fixture files only
  - avoid parallel heavy workloads while stabilizing local state

## Cleanup

```bash
make down
```

To reclaim storage and start fresh:

```bash
make clean
```

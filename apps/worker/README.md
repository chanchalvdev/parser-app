# apps/worker/

Python ingestion worker for queue consumption and file processing.

## Responsibilities (MVP)

- Block-pop job messages from Redis list `ingestion_jobs`.
- Parse JSON queue payload.
- Update `ingestion_jobs` lifecycle status.
- Write `job_events`.
- Download object from MinIO into a temporary working directory.
- Compute SHA-256 hash and store it in `files`.
- Minimal file type detection (extension/mime).
- Mark non-archive job as completed in placeholder mode.

## Environment

- `DATABASE_URL`
- `REDIS_ADDR`
- `REDIS_PASSWORD`
- `REDIS_DB`
- `WORKER_QUEUE_NAME` (defaults to `QUEUE_NAME` if unset)
- `WORKER_REDIS_BLOCK_TIMEOUT` (defaults `5`)
- `WORKER_POLL_INTERVAL_SECONDS` (defaults `1`)
- `WORKER_MAX_ARCHIVE_DEPTH` (defaults `10`)
- `WORKER_MAX_EXTRACTED_FILES` (defaults `50000`)
- `WORKER_MAX_EXTRACTED_SIZE_MB` (defaults `1024`)
- `WORKER_MAX_EXPANSION_RATIO` (defaults `20.0`)
- `MINIO_ENDPOINT`
- `MINIO_ACCESS_KEY`
- `MINIO_SECRET_KEY`
- `MINIO_BUCKET`
- `MINIO_USE_SSL`
- `OPENSEARCH_URL` (defaults to `http://opensearch:9200`)
- `OPENSEARCH_USER` (optional)
- `OPENSEARCH_PASSWORD` (optional)
- `WORKER_TEMP_DIR` (defaults `/tmp/file-worker`)
- `WORKER_LOG_LEVEL` (defaults `INFO`)

## Local run options

Docker (preferred):

```bash
docker compose up --build -d worker
docker compose logs -f worker
```

Local Python execution:

```bash
cd apps/worker
python -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
python -m app.main
```

## Job message schema

The worker expects messages produced by API on `QUEUE_NAME`:

```json
{
  "job_id": "...",
  "tenant_id": "...",
  "root_file_id": "...",
  "storage_path": "...",
  "bucket": "file-ingestion",
  "original_name": "archive.zip",
  "depth": 0,
  "created_at": "2026-06-15T12:34:56Z"
}
```

## Archive extraction behavior

- Detects archive type using extension/mime and uses extractor modules per format:
  - ZIP, TAR/TAR.GZ/TGZ, 7Z, RAR.
- Extracts child entries into local temp directories with zip-slip protection.
- Enforces extraction limits:
  - `max_archive_depth`
  - `max_extracted_files`
  - `max_extracted_size_mb`
  - `max_expansion_ratio`
- Uploads extracted children back to MinIO path:
  `extracted/{tenant_id}/{upload_id}/{parent_file_id}/{relative_child_path}`
- Creates `files` rows for children and processes them recursively until depth limit.
- Marks `PASSWORD_REQUIRED` on job when encrypted archives cannot be opened.

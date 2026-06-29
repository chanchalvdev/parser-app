# apps/api/

Go backend API service (clean architecture foundation).

## Responsibilities

- Request orchestration and REST surface for ingestion workflow APIs.
- Health/readiness/version endpoints for service observability.
- Repository/service/handler separation for future growth.

## Current endpoints

- `GET /health`
- `GET /ready` (validates PostgreSQL connectivity)
- `GET /api/v1/version`
- `POST /api/v1/uploads/initiate`
- `POST /api/v1/uploads/complete`
- `GET /api/v1/uploads/{upload_id}`
- `GET /api/v1/search`
- `POST /api/v1/search`
- `GET /api/v1/search/suggestions`
- `GET /api/v1/files`
- `GET /api/v1/files/{file_id}`
- `POST /api/v1/files/{file_id}/password`
- `GET /api/v1/files/{file_id}/children`
- `GET /api/v1/files/{file_id}/tree`
- `GET /api/v1/files/{file_id}/records`
- `GET /api/v1/jobs`
- `GET /api/v1/jobs/{job_id}`
- `GET /api/v1/jobs/{job_id}/events`
- `POST /api/v1/jobs/{job_id}/retry`
- `GET /api/v1/admin/settings`
- `PUT /api/v1/admin/settings`
- `GET /api/v1/dashboard/summary`
- `GET /api/v1/dashboard/file-types`
- `GET /api/v1/dashboard/processing-status`
- `GET /api/v1/dashboard/upload-volume`
- `GET /api/v1/dashboard/error-breakdown`
- `GET /api/v1/dashboard/entities`
- `GET /api/v1/dashboard/processing-duration`

### Upload API curl examples

Initiate:

```bash
curl -X POST http://localhost:8080/api/v1/uploads/initiate \
  -H "Content-Type: application/json" \
  -d '{
    "file_name": "archive.zip",
    "content_type": "application/zip",
    "size_bytes": 123456,
    "password_provided": false
  }'
```

Complete:

```bash
curl -X POST http://localhost:8080/api/v1/uploads/complete \
  -H "Content-Type: application/json" \
  -d '{
    "upload_id": "PASTE_UPLOAD_ID_HERE"
  }'
```

Get:

```bash
curl -X GET "http://localhost:8080/api/v1/uploads/PASTE_UPLOAD_ID_HERE"
```

### Search API curl examples

Search with keyword query and filters:

```bash
curl -G "http://localhost:8080/api/v1/search" \
  --data-urlencode "q=error" \
  --data-urlencode "file_id=PASTE_FILE_ID" \
  --data-urlencode "record_type=text_line" \
  --data-urlencode "ip=192.0.2.1" \
  --data-urlencode "extension=txt" \
  --data-urlencode "detected_file_type=text/plain" \
  --data-urlencode "job_id=PASTE_JOB_ID" \
  --data-urlencode "date_from=2026-06-01T00:00:00Z" \
  --data-urlencode "date_to=2026-06-30T23:59:59Z" \
  --data-urlencode "page=1" \
  --data-urlencode "page_size=25" \
  --data-urlencode "sort=relevance"
```

Search endpoint with JSON body:

```bash
curl -X POST "http://localhost:8080/api/v1/search" \
  -H "Content-Type: application/json" \
  -d '{
    "q": "login attempt",
    "record_type": "text_line",
    "extension": "jsonl",
    "page": 1,
    "page_size": 10,
    "sort": "created_at"
  }'
```

Suggestions endpoint:

```bash
curl -G "http://localhost:8080/api/v1/search/suggestions" \
  --data-urlencode "q=log" \
  --data-urlencode "tenant_id=11111111-1111-1111-1111-111111111001"
```

Search response example:

```json
{
  "total": 123,
  "page": 1,
  "page_size": 25,
  "results": [
    {
      "record_id": "parsed-record-id",
      "file_id": "file-id",
      "job_id": "job-id",
      "source_file_name": "sample.txt",
      "record_type": "text_line",
      "content_preview": "Authentication attempt failed for user...",
      "highlight": "Authentication attempt <em>failed</em> for user...",
      "entities": {
        "ip_addresses": ["192.0.2.1"],
        "emails": ["user@example.com"],
        "urls": ["https://example.com"],
        "domains": ["example.com"],
        "hashes": []
      },
      "created_at": "2026-06-15T12:34:56Z"
    }
  ],
  "facets": {
    "record_type": [
      {"value": "text_line", "count": 42}
    ],
    "detected_file_type": [
      {"value": "text/plain", "count": 21}
    ],
    "entities": {
      "ip_addresses": [
        {"value": "192.0.2.1", "count": 7}
      ],
      "emails": [
        {"value": "user@example.com", "count": 3}
      ],
      "domains": [
        {"value": "example.com", "count": 5}
      ]
    }
  }
}
```

### File APIs

List files with pagination and filters:

```bash
curl -G "http://localhost:8080/api/v1/files" \
  --data-urlencode "tenant_id=11111111-1111-1111-1111-111111111001" \
  --data-urlencode "status=queued" \
  --data-urlencode "extension=jsonl" \
  --data-urlencode "detected_file_type=text/plain" \
  --data-urlencode "page=1" \
  --data-urlencode "page_size=25"
```

Get a file:

```bash
curl -X GET "http://localhost:8080/api/v1/files/PASTE_FILE_ID_HERE"
```

List file children:

```bash
curl -G "http://localhost:8080/api/v1/files/PASTE_FILE_ID_HERE/children" \
  --data-urlencode "page=1" \
  --data-urlencode "page_size=50"
```

Fetch file tree:

```bash
curl -X GET "http://localhost:8080/api/v1/files/PASTE_FILE_ID_HERE/tree"
```

List parsed records for a file:

```bash
curl -G "http://localhost:8080/api/v1/files/PASTE_FILE_ID_HERE/records" \
  --data-urlencode "page=1" \
  --data-urlencode "page_size=25"
```

Submit archive password:

```bash
curl -X POST "http://localhost:8080/api/v1/files/PASTE_FILE_ID_HERE/password" \
  -H "Content-Type: application/json" \
  -d '{
    "password": "secret-password"
  }'
```

```json
{
  "file_id": "PASTE_FILE_ID_HERE",
  "job_id": "PASTE_JOB_ID_HERE",
  "status": "queued"
}
```

### Job APIs

List jobs:

```bash
curl -G "http://localhost:8080/api/v1/jobs" \
  --data-urlencode "tenant_id=11111111-1111-1111-1111-111111111001" \
  --data-urlencode "page=1" \
  --data-urlencode "page_size=25"
```

Get a job:

```bash
curl -X GET "http://localhost:8080/api/v1/jobs/PASTE_JOB_ID_HERE"
```

List job events:

```bash
curl -G "http://localhost:8080/api/v1/jobs/PASTE_JOB_ID_HERE/events" \
  --data-urlencode "page=1" \
  --data-urlencode "page_size=50"
```

Retry a job:

```bash
curl -X POST "http://localhost:8080/api/v1/jobs/PASTE_JOB_ID_HERE/retry"
```

### Admin settings API

Fetch admin settings for tenant:

```bash
curl -G "http://localhost:8080/api/v1/admin/settings" \
  --data-urlencode "tenant_id=11111111-1111-1111-1111-111111111001"
```

Update admin settings:

```bash
curl -X PUT "http://localhost:8080/api/v1/admin/settings?tenant_id=11111111-1111-1111-1111-111111111001" \
  -H "Content-Type: application/json" \
  -d '{
    "max_upload_size_mb": 1024,
    "enabled_parsers": ["txt", "json", "jsonl", "csv"],
    "parser_batch_size": 500,
    "search_index_batch_size": 1000,
    "max_expansion_ratio": 25
  }'
```

### Dashboard APIs

Summary:

```bash
curl -G "http://localhost:8080/api/v1/dashboard/summary" \
  --data-urlencode "tenant_id=11111111-1111-1111-1111-111111111001"
```

File type distribution:

```bash
curl -G "http://localhost:8080/api/v1/dashboard/file-types" \
  --data-urlencode "tenant_id=11111111-1111-1111-1111-111111111001" \
  --data-urlencode "limit=25"
```

Processing status distribution:

```bash
curl -G "http://localhost:8080/api/v1/dashboard/processing-status" \
  --data-urlencode "tenant_id=11111111-1111-1111-1111-111111111001" \
  --data-urlencode "limit=25"
```

Upload volume:

```bash
curl -G "http://localhost:8080/api/v1/dashboard/upload-volume" \
  --data-urlencode "tenant_id=11111111-1111-1111-1111-111111111001" \
  --data-urlencode "days=30"
```

Error breakdown:

```bash
curl -G "http://localhost:8080/api/v1/dashboard/error-breakdown" \
  --data-urlencode "tenant_id=11111111-1111-1111-1111-111111111001" \
  --data-urlencode "limit=10"
```

Top entities:

```bash
curl -G "http://localhost:8080/api/v1/dashboard/entities" \
  --data-urlencode "tenant_id=11111111-1111-1111-1111-111111111001" \
  --data-urlencode "limit=10"
```

Processing duration:

```bash
curl -G "http://localhost:8080/api/v1/dashboard/processing-duration" \
  --data-urlencode "tenant_id=11111111-1111-1111-1111-111111111001"
```

## Environment variables

- `APP_ENV`
- `API_PORT`
- `DATABASE_URL`
- `REDIS_ADDR`
- `REDIS_PASSWORD`
- `REDIS_DB`
- `QUEUE_NAME`
- `MINIO_ENDPOINT`
- `MINIO_ACCESS_KEY`
- `MINIO_SECRET_KEY`
- `MINIO_BUCKET`
- `OPENSEARCH_URL`
- `OPENSEARCH_USER`
- `OPENSEARCH_PASSWORD`
- `OPENSEARCH_PARSED_RECORDS_INDEX`
- `OPENSEARCH_FILES_INDEX`

See `docs/api/job-message-contract.md` for the Redis queue message schema used on upload completion.

## Run (local container)

This service is started through the root `docker-compose.yml` API service and is reachable at `http://localhost:8080`.

# Job Queue Message Contract

## Producer contract: `ingestion_jobs`

The Go API publishes ingestion work to the Redis list specified by `QUEUE_NAME` (default `ingestion_jobs`).
Each message is JSON with this schema:

```json
{
  "job_id": "...",
  "tenant_id": "...",
  "root_file_id": "...",
  "storage_path": "raw/<tenant>/<upload-id>/<file-name>",
  "bucket": "file-ingestion",
  "original_name": "archive.zip",
  "depth": 0,
  "created_at": "2026-06-15T12:34:56Z"
}
```

### Field definitions
- `job_id` (string): UUID of the ingestion job.
- `tenant_id` (string): Tenant UUID owning the job.
- `root_file_id` (string): UUID of the root `files` record to process.
- `storage_path` (string): Object key in MinIO.
- `bucket` (string): S3 bucket where the raw object is stored (`file-ingestion`).
- `original_name` (string): Original filename used when creating the upload.
- `depth` (int): Initial depth for extraction/parsing (root is `0`).
- `created_at` (string): ISO-8601 UTC timestamp when the API enqueued the job.

### Queue behavior
- API enqueues on successful upload completion and file/job creation.
- If enqueue fails, API marks the ingestion job as `failed` and writes a `upload.queue.failed` job event.
- On successful enqueue, API marks job as `queued` and writes a `upload.completed` event.

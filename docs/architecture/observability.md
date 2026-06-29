# Observability (Local MVP)

## Logs

### API (Go)

- Structured JSON logs are emitted via `zap` and include at least:
  - `request_id`
  - request `method`, `path`, `status`, `duration_ms`
  - when available: `remote_addr`, `query`
- Request ID is generated (or reused from `X-Request-ID`) and echoed in responses.
- Panic recovery middleware writes structured `ERROR` logs with `request_id`, method/path, and panic payload.
- Endpoints `GET /health` and `GET /ready` provide service liveness/readiness signals.
- `job_events` is the primary operational timeline for workflow transitions and failures.

### Worker (Python)

- Structured JSON logging is configured in `apps/worker/app/logging_config.py` with timestamp, level, module, line, and custom fields.
- Worker logs include structured metadata for most state-changing actions:
  - `tenant_id`
  - `job_id`
  - `file_id`
  - parser/indexing/extraction counters
- Job status transitions are logged from `WorkerRepository.update_job_status`.
- Parser start/completion and index flush/complete paths log record counters in
  `apps/worker/app/processing/orchestrator.py`.
- OpenSearch indexing logs include index name, indexed count, and tenant/job/file IDs in
  `apps/worker/app/loaders/search_loader.py`.

## Metrics (Planned)

Current MVP focuses on structured logs and SQL/job events for observability.

Planned metrics (Prometheus) endpoints:

- `worker_job_in_progress_total`
- `worker_job_completed_total`
- `worker_job_failed_total`
- `worker_records_parsed_total`
- `worker_records_inserted_total`
- `worker_records_indexed_total`
- `worker_archive_children_total`
- `http_requests_total`
- `http_request_duration_seconds`

If we add metrics in a later iteration, `/metrics` can be exposed from the Go API
and a lightweight `/metrics` exporter can be added to the worker process.

## Traces (Planned)

Suggested tracing rollout:

1. Add OpenTelemetry SDK initialization to both services.
2. Generate a propagation `trace_id` and attach it to logs and job lifecycle events.
3. Add a trace exporter (Jaeger/Tempo) with service names:
   - `api`
   - `worker`

## Dashboards (Planned)

Planned dashboard panels:

- Job outcome by status and stage
- Worker parsing throughput (records/min)
- Archive extraction counts and fail reasons
- OpenSearch indexing health and queue delays
- API latency histograms by endpoint (`p50`, `p95`, `p99`)

## Local debugging commands

- API health:
  - `curl -sf http://localhost:8080/health | jq`
  - `curl -sf http://localhost:8080/ready | jq`
- API error and access logs:
  - `docker compose logs api --follow`
- Worker logs:
  - `docker compose logs worker --follow`
- PostgreSQL heartbeat:
  - `docker compose exec postgres psql "$DATABASE_URL" -c "SELECT 1;"`
- OpenSearch health:
  - `curl -s http://localhost:9200/_cluster/health | jq`
- Search indexing troubleshooting:
  - `docker compose exec postgres psql "$DATABASE_URL" -c "SELECT tenant_id, status, current_stage, progress_percent FROM ingestion_jobs ORDER BY updated_at DESC LIMIT 10;"`
  - `docker compose exec postgres psql "$DATABASE_URL" -c "SELECT tenant_id, event_type, event_message, event_details, created_at FROM job_events ORDER BY created_at DESC LIMIT 50;"`
- Frontend / network errors:
  - `docker compose logs web --follow`

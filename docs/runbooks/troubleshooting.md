# Troubleshooting Runbook

Use this runbook when local flow is broken after making changes.

## 0) Health check baseline

```bash
curl -sS http://localhost:8080/health | jq
curl -sS http://localhost:8080/ready | jq
docker compose logs api --tail=50
docker compose logs worker --tail=50
```

## API cannot connect DB

1. Confirm service status:

```bash
docker compose ps postgres
docker compose exec postgres pg_isready
```

2. Verify `DATABASE_URL` in `.env` and restart API:

```bash
docker compose logs api --tail=80
docker compose restart api
```

3. Validate schema:

```bash
docker compose exec postgres psql "$DATABASE_URL" -c "\dt"
```

## Worker not consuming jobs

1. Confirm queue and worker are running:

```bash
docker compose ps worker redis
docker compose exec redis redis-cli -p 6379 LLEN $REDIS_QUEUE_NAME
```

2. Confirm latest jobs exist:

```bash
docker compose exec postgres psql "$DATABASE_URL" \
  -c "SELECT id,status,current_stage,progress_percent FROM ingestion_jobs ORDER BY created_at DESC LIMIT 20;"
```

3. Confirm queue visibility:

```bash
docker compose logs worker --tail=120
```

## MinIO upload fails

1. Confirm presigned flow:

```bash
curl -I http://localhost:9000
docker compose logs minio --tail=80
```

2. Re-upload small fixture (`sample.txt`) first to validate client + bucket path.
3. Check that upload initiation returns non-empty `upload_url`.

## OpenSearch index missing

```bash
curl -sS http://localhost:9200/_cat/indices?v
```

If `parsed-records` or `files` are missing:

```bash
make search-init
docker compose logs opensearch --tail=120
```

Then re-run the job.

## CORS issue

1. Validate API env `ALLOWED_ORIGINS` includes `http://localhost:5173`.
2. Check `VITE_API_BASE_URL` from web container env.
3. Restart both:

```bash
docker compose restart api web
```

4. Re-check browser preflight errors in devtools.

## CORS/Swagger not loading

1. Confirm frontend calls `GET /api/v1/...` using `VITE_API_BASE_URL`.
2. Ensure `/docs` responds:

```bash
curl -sS http://localhost:8080/docs | head
```

3. Validate OpenAPI JSON:

```bash
curl -sS http://localhost:8080/docs/openapi.json | head -n 20
```

## Password-required archive recovery

1. Re-open file or job detail; status should show `PASSWORD_REQUIRED`.
2. Submit password via modal/API.
3. Confirm status transitions:

```bash
curl -sS http://localhost:8080/api/v1/jobs/{job_id} | jq
```

4. Confirm job event log includes `PASSWORD_SUBMITTED`.

## Docker memory / restart loops

1. Reduce concurrent workloads and restart.
2. Check exit codes:

```bash
docker compose ps -a
docker compose logs --tail=80 api worker
```

3. If needed, run clean restart:

```bash
make down
docker system prune -f
make up
```

## Quick recovery checklist

- `make up`
- `make migrate`
- `make seed`
- `make search-init`
- Re-run upload with `tests/fixtures/sample.txt`
- Re-open `Dashboard` and `Search`


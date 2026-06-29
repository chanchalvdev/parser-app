# High-Level Architecture

## Scope

The platform is a local-first ingestion system that coordinates:

- API orchestration (Go)
- Background processing (Python)
- Persistent storage (PostgreSQL)
- Raw object storage (MinIO/S3)
- Search index (OpenSearch)
- Frontend web app (React)

## Diagram

```text
+------------------+        HTTPS/JSON         +-----------------------+
|  Browser/Web UI  | ------------------------> |      Go API (8080)    |
+------------------+                           +----------+------------+
                                                           |
                                                           v
                                                   +-------+-------+
                                                   | PostgreSQL    |
                                                   | metadata/job  |
                                                   +-------+-------+
                                                           |
                                                           v
+------------------+        Redis                  +--------+---------+
| Worker Service   | <----------------------------> | Redis queue      |
| (Python parser)  |                               +--------+---------+
+---------+--------+                                        |
          |                                                  v
          |                                         +--------+---------+
          +---> MinIO (raw objects)                 | Worker states    |
          |                                         +--------+---------+
          |                                                  |
          |                                     +------------+-----------+
          +----> PostgreSQL inserts parsed_records ------------+---+
          |                                                  |   |
          v                                                  v   |
 +--------+---------+                                  +------+------+ 
 |    MinIO         |                                  | OpenSearch      |
 | object store     |                                  | parsed records  |
 +------------------+                                  +----------------+
```

## Request flow

1. User uploads file via UI.
2. UI calls `POST /api/v1/uploads/initiate`.
3. API returns presigned upload URL and `upload_id`.
4. User uploads to MinIO directly with PUT.
5. UI calls `POST /api/v1/uploads/complete`.
6. API validates and enqueues a job.
7. Worker pops queue, downloads file from MinIO, parses and extracts.
8. Worker writes structured metadata + records into PostgreSQL.
9. Worker bulk indexes into OpenSearch.
10. UI reads job and file events from PostgreSQL and search hits from OpenSearch.

## Operational boundaries

- **API boundary**: request/response, validation, audit, queueing.
- **Worker boundary**: extraction/parsing/indexing and state transitions.
- **Storage boundary**: PostgreSQL (state), MinIO (objects), OpenSearch (analytics/search).

## Observability boundary

- Logs + request IDs from the API and worker
- Primary operational timeline is `job_events` table in PostgreSQL
- Health/readiness endpoints on API


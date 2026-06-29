# Threat Model (MVP)

## Assets

- Uploaded objects in MinIO (raw file payloads)
- File records (`files`) and extraction metadata in PostgreSQL
- Job lifecycle records and events (`jobs`, `job_events`)
- Parsed records (`parsed_records`) and extracted entities
- OpenSearch indexed documents (`parsed-records`, `files`)
- Audit logs (`audit_logs`)
- Archive passwords and password references
- Tenant and queue context (`tenant_id`, `job_id`, `file_id`)

## Trust boundaries

- User browser ↔ API (public HTTP API over localhost)
- API ↔ PostgreSQL
- API ↔ MinIO
- API ↔ Redis
- API/Worker ↔ external search cluster (OpenSearch)
- Worker runtime ↔ temporary filesystem/extracted files
- Workers and API code ↔ filesystem path and logs

## Threats and mitigations

### Upload and file handling

- Threat: Path traversal through file names (zip-slip / relative paths).
  - Mitigation:
    - normalize file names at upload and archive extraction boundary
    - canonical path checks and extraction allowlist directory roots
    - quarantine on uncertain archive entries
- Threat: oversized or zero-byte upload abuse.
  - Mitigation:
    - explicit file size checks in upload initiation and worker detection flow
    - zero-byte rejection
    - max upload size validation via settings.
- Threat: malformed archive causes extraction failures or resource exhaustion.
  - Mitigation:
    - archive safety limits (max depth, max extracted files, max extracted size, expansion ratio checks)
    - corrupt archive handling with explicit file/job event status (`failed` path)
- Threat: password leakage.
  - Mitigation:
    - no logging of password values
    - password references stored through abstracted secret repository placeholder
    - requeue path does not emit credentials.

### API and service boundaries

- Threat: cross-origin request abuse from unauthorized origins.
  - Mitigation:
    - CORS configured to allowed origin(s), default includes local web origin
    - request ID middleware for request correlation
    - placeholder auth middleware and role checks on admin endpoints (`RequireRole("admin")`).
- Threat: unauthorized retry/audit/setting mutation.
  - Mitigation:
    - endpoint-level authorization hooks and dedicated audit event creation for security-relevant actions.
- Threat: missing request traceability.
  - Mitigation:
    - request IDs propagated to logs and responses
    - job_events table as single authoritative state timeline.

### Data integrity and observability

- Threat: parser or search failures cause data loss.
  - Mitigation:
    - parsed records inserted into PostgreSQL with loader counts and failures surfaced in job events
    - search indexing failures update `search_index_status` while preserving DB inserts.
- Threat: malformed content causing crash or silent corruption.
  - Mitigation:
    - parser errors captured in structured fields and parser error events
    - JSONL malformed-line handling continues with per-line error reporting.
- Threat: SQL or index query injection from filter parameters.
  - Mitigation:
    - parameterized repository queries and OpenSearch query builders with safe filter composition.

## Residual risks

- No production authentication provider is yet wired end-to-end (placeholder RBAC remains).
- Secret backend is a local placeholder abstraction; production Key Vault integration is planned but not yet active.
- Full malware/content inspection is not implemented; archive safety is preventive and bounded, not forensic.
- Frontend input sanitization and CSP posture should be hardened before public deployment.
- Long-tail schema/contract drift depends on manual review because migration/versioning is MVP-light for some auxiliary artifacts.

## Review expectations

For any high-risk change touching security-critical paths:
- run `harness/hooks/security-review.md`
- run `harness/hooks/parser-safety-review.md`
- run the troubleshooting flow for the impacted endpoint in `docs/runbooks/troubleshooting.md`


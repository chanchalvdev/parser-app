# Product Goals

## Vision
Deliver a local-first file ingestion, parsing, indexing, and observability platform with predictable recovery workflows and operator-friendly dashboards.

## Must-haves
- Reliable upload with presigned object handling and status tracking.
- Archive extraction with password-aware recovery and child-file lineage.
- Parser coverage for text, CSV, JSON, JSONL.
- Search with keyword and filtered entity-aware retrieval.
- File/job timeline visibility and retry/recovery workflows.
- Admin controls for batch sizes, archive limits, parser enablement.

## Outcomes
- Reduce average time-to-search for uploaded samples.
- Make blocked states (password, parse failure, indexing failure) user-resolvable.
- Keep operator actions understandable with event timeline and status semantics.

## Non-goals (MVP)
- Production-grade OAuth/RBAC beyond placeholder readiness.
- Real-time sockets for all dashboards.
- Multi-region deployment and billing enforcement.

## Acceptance check
- All must-haves are implemented with user-facing evidence in API + UI.
- Recovery workflows can be demonstrated with sample files.

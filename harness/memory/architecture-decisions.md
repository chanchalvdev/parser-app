# Architecture Decisions Log

## ADR-001: Parser-Worker-Search Separation
- **Date:** 2026-06-16
- **Decision:** Keep parser and indexing in worker; API remains orchestration and metadata authority.
- **Rationale:** Clear service boundaries allow independent scaling and isolated failure domains.
- **Alternatives considered:** Indexing in API after DB write; rejected due API bottleneck and tight coupling.
- **Impact:** API remains query-centric; worker owns extraction and search document generation.
- **Status:** Accepted

## ADR-002: Password Reference Abstraction Layer
- **Date:** 2026-06-16
- **Decision:** Use repository-based secret abstraction with local placeholder and clear migration to external vault.
- **Rationale:** Enables local flow today and secure production path later.
- **Impact:** Worker resolves passwords via provider abstraction before archive extraction.
- **Status:** Accepted

## ADR-003: Password Recovery and Retry Strategy
- **Date:** 2026-06-16
- **Decision:** Password-related failures set explicit file/job states and allow explicit password re-submit + requeue.
- **Rationale:** Improves recoverability while preserving already persisted progress.
- **Impact:** Events/audit capture retry action; partial work is not lost.
- **Status:** Accepted

## ADR-004: Search Non-blocking Indexing
- **Date:** 2026-06-16
- **Decision:** Search indexing failures should not block DB persistence.
- **Rationale:** Preserve parsed data while surfacing indexing health separately.
- **Impact:** `search_index_status` explicitly tracks index state and worker raises `INDEXING_*` events.
- **Status:** Accepted

## ADR-005: Harness-Driven Development
- **Date:** 2026-06-19
- **Decision:** Formalize codex-friendly harness with agents, workflows, hooks, and memory updates.
- **Rationale:** Enable repeatable and auditable AI+human collaboration.
- **Impact:** Explicit review/design/plan/agree/test/validation flow in repository docs.
- **Status:** Accepted

- Self-improve run on 2026-06-19 16:44:05 UTC: test harness command
- File touched: harness/memory/

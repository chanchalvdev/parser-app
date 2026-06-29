# Engineering Goals

## Core goals
- Preserve modular service boundaries (API, worker, UI, storage, search).
- Keep contracts typed and stable where feasible.
- Maintain clear state transitions for files/jobs (including password/retry states).
- Keep parsers and loaders idempotent under retries.

## Performance goals
- Default batch processing with bounded memory and clear limits.
- Non-blocking processing for large or archive-heavy workloads.
- Avoid N+1 patterns in listing and dashboard queries.

## Operability goals
- Every major state transition emits job/file events and timestamps.
- Local runbooks include complete bootstrap and diagnostics.
- Observability artifacts (logs/events/metrics placeholders) stay close to code.
- Release and incident paths are deterministic.

## Acceptance check
- Feature changes can be isolated and rolled back per boundary.
- Core paths continue without blocking across service restart.

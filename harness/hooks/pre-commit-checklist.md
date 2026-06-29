# Pre-Commit Checklist

## Trigger
Before handoff or commit for a completed task chunk.

## Checklist
- [ ] Changed code has targeted unit tests and/or fixture updates.
- [ ] Service-level behavior changes include test plan evidence.
- [ ] New frontend flows include API/client contract checks and error-state coverage.
- [ ] Error paths emit structured events, logs, and metrics where required.
- [ ] Security-sensitive areas include validation tests for inputs and secret handling.
- [ ] Migration/schema changes include rollback or restoration notes.
- [ ] `TODO` comments and accidental debug logs removed.
- [ ] Query cache/state invalidation is updated where frontend state may stale.
- [ ] Docs and memory entries updated for any changed behavior.

## Blocking criteria
- Missing/insufficient tests for changed behavior.
- Missing event/audit updates for required actions.
- New security exposure without mitigation.
- Missing migration rollback notes for persistence changes.
- Unreviewed parser safety risks.

## Required evidence
- Modified file list.
- Command evidence of tests executed (or blocked rationale).
- Hook completion status for security/parser as applicable.
- Documentation updates with file references.

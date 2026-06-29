# Release Readiness Hook

## Trigger
Before release tagging, deployment handoff, or major branch merge.

## Checklist
- [ ] All required workflows executed and evidence recorded.
- [ ] `make test-api`, `make test-worker`, `make test-web`, `make test-all` completed or risks accepted.
- [ ] CI placeholder workflow present and aligned with repo commands.
- [ ] Critical security and parsing blockers cleared.
- [ ] Migrations are validated on a clean run.
- [ ] Runbook and setup steps match current stack behavior.
- [ ] Rollback strategy and command set are documented.
- [ ] Known risks updated and owners assigned.

## Blocking criteria
- Unresolved critical functional/ security bugs.
- Missing migration/test evidence.
- Unverified startup sequence or environment assumptions.
- Documentation mismatch for startup, ports, or command flow.

## Required evidence
- Command log with pass/fail for test gates.
- Release summary with owners.
- Smoke test checklist and risk closure notes.

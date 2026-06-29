# Release Workflow

## Purpose
Prepare a controlled release handoff with stable runtime, rollback posture, and complete validation artifacts.

## Phases

1. Review
   - Check changed boundaries and final scope.
   - Confirm all blockers closed or explicitly accepted.
2. Design
   - Validate migration and bootstrap order.
   - Define rollback checkpoints.
3. Plan
   - Collect evidence for required commands and test coverage.
   - Freeze dependency and config assumptions.
4. Agree
   - Confirm release owner approvals and sign-off from architect/security/qa.
5. Execute
   - Run full test gates and lint/quality checks.
   - Validate service startup and smoke flows.
6. Test
   - Execute release checklist in a clean environment.
7. Validation
   - Verify dashboards/logs and major user flows still function.
8. Self-improve
   - Update runbooks if new recovery or bootstrap steps were needed.

## Exit criteria

- All release blockers resolved or risk-reviewed.
- Migrations and bootstrap validated.
- Test evidence and smoke checks are complete.
- Rollback and known limitations are documented.

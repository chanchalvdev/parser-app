# Feature Development Workflow

## Purpose
Develop a feature from requirements to merge-ready implementation without breaking service boundaries or contracts.

## Phases

1. Review
   - Read impacted goals and architecture notes.
   - Confirm scope and define explicit out-of-scope boundaries.
   - Identify required migrations, index changes, hooks, and test surfaces.
2. Design
   - Produce HLD/LLD deltas and ownership map with Architect.
   - Draft contract and migration impact notes.
3. Plan
   - Break work into owner-specific streams (API, worker, UI, DB, security, docs).
   - Define acceptance criteria per stream.
   - Add blocking risks and rollback conditions.
4. Agree
   - Obtain owner sign-off from affected primary agents.
   - Freeze implementation sequence and confirm non-goal behaviors.
5. Execute
   - Implement one ownership boundary at a time.
   - Respect repository conventions, tenant constraints, and existing event semantics.
6. Test
   - Unit tests and integration scaffolds where practical.
   - Update validation evidence for each changed behavior.
7. Validate
   - End-to-end smoke scenario.
   - Dashboard/search/file/job behavior checks if applicable.
8. Self-improve
   - Update memory/ADR, risks, and backlog.

## Exit criteria

- Scope and assumptions are accepted and documented.
- Required hooks and workflow-specific checklists executed.
- No blocking security/parser risks remain.
- Owners have sign-off and evidence references.
- Docs and runbooks updated if operator-visible behavior changed.

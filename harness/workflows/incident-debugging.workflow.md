# Incident Debugging Workflow

## Purpose
Resolve production-like incidents through a structured, low-variance sequence.

## Phases

1. Review
   - Capture failing identifiers and reproducer details.
   - Identify blast radius and impacted user workflows.
2. Design
   - Form a narrow and testable failure hypothesis.
   - Define minimal reversible fix path.
3. Plan
   - Assign component owner (API/worker/search/infra).
   - Capture rollback condition before patching.
4. Agree
   - Confirm owner acceptance for workaround vs permanent fix.
5. Execute
   - Apply focused fix.
   - Add or update regression test coverage.
6. Test
   - Reproduce from smallest sample.
   - Validate normal path and failure path.
7. Validation
   - Update runbook with remediation steps.
8. Self-improve
   - Add incident lessons to `harness/memory/known-risks.md` and backlog.

## Exit criteria

- Reproducible scenario fixed and verified.
- Root cause and prevention notes recorded.
- Evidence log includes failing IDs, logs, and successful retest.

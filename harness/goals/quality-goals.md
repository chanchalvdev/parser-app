# Quality Goals

## Test and verification goals
- Deterministic tests for new behavior where practical.
- Critical paths include failure and recovery cases.
- Regression cases for parser and contract-sensitive API paths.
- Manual smoke validation documented for end-to-end flows.

## Code quality goals
- Consistent error envelopes and event recording.
- Types-first API and client contract alignment.
- Focused, low-friction changes without speculative rewrites.
- Hook discipline enforced before handoff.

## Process quality goals
- Every feature includes:
  - review/design/plan/agree checkpoints,
  - test plan,
  - validation evidence,
  - self-improvement memory updates.
- Risks blocked by known issues are explicitly documented.

## Acceptance check
- No high-risk path is changed without at least one matching test or documented blocker.

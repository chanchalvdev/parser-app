# Testing Skill

## When to use
To define and execute verification from unit tests to integration smoke tests.

## Procedure
1. Build a matrix for success/failure paths.
2. Select practical fixture sets.
3. Add unit tests for isolated logic, and integration checks for cross-service paths.
4. Capture command evidence for each evidence step.
5. Record blocked or deferred areas with owner and date.

## Checklist
- Happy path and failure paths covered per requirement.
- Parser and security edge cases included where high-risk.
- Cross-service path validated with sample data and logs.
- Test output is reproducible from documented commands.

## Guardrails
- Don’t overfit with flaky tests as sole coverage.
- Keep fixture tests deterministic and small.
- Never mark a test gate as passed without evidence.

## Example output
- `make test-api` passes handler and service tests for upload/search/file endpoints.
- `make test-worker` validates parser and loader resilience.
- `make test-web` includes API client smoke or component checks.

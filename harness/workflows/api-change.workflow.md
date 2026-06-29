# API Change Workflow

## Purpose
Introduce API changes with safe contracts, migration awareness, and client compatibility.

## Phases

1. Review
   - Collect impacted request/response schemas.
   - Determine consumer changes (frontend/mobile/CLI).
2. Design
   - Define versioning and migration plan for breaking/non-breaking changes.
   - Add contract notes for pagination/filter/sorting behavior.
3. Plan
   - Assign service, repository, and migration updates to agents.
   - Define test matrix for positive/negative API paths.
4. Agree
   - Get Security/Docs owner confirmation for public-facing behavior.
5. Execute
   - Implement routes, handlers, DTOs, and errors.
   - Update service/repository orchestration.
   - Add request validation and explicit audit hooks.
6. Test
   - Add/adjust handler tests and service tests.
   - Include failure-path tests (422, 404, forbidden, bad payload).
7. Validation
   - Confirm openAPI/README examples match actual response.
8. Self-improve
   - Update `docs/api/rest-api.md` and `docs/harness/how-to-use-harness.md` if contract patterns change.

## Exit criteria

- Contract changes are mapped and documented.
- Backward compatibility or migration impact is explicit.
- Error semantics are stable and tested.
- API docs include representative request/response examples.

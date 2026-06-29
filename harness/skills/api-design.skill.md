# API Design Skill

## When to use
When adding or modifying HTTP endpoints, DTOs, pagination, filters, sorting, auth, and response contracts.

## Procedure
1. Define request schema with validation constraints.
2. Define response schema with deterministic pagination/filter/sort behavior.
3. Update handler, service, and repository layers in boundary-respecting manner.
4. Define audit/security side effects.
5. Update API docs and frontend type expectations.

## Checklist
- Tenant scoping exists where applicable.
- Input validation rejects malformed and oversized values.
- Error responses are deterministic and machine-readable.
- Pagination defaults and bounds are explicit.
- Audit/event side effects documented and implemented.

## Guardrails
- Avoid breaking existing clients without migration.
- Keep route semantics consistent across versions.
- Do not return secret or raw infrastructure values.

## Example output
- `POST /api/v1/files/{file_id}/password` implemented with validation, status transitions, and event creation.
- Endpoint includes request/response examples and non-2xx behavior.
- Frontend client model updated in `apps/web/src/types`.

# Docker Compose Skill

## When to use
When services, ports, startup dependencies, or local bootstrap steps change.

## Procedure
1. Update compose services and environment wiring.
2. Confirm dependency graph and healthcheck ordering.
3. Keep service names and networking stable where scripts depend on them.
4. Add/adjust command shortcuts and one-shot bootstrap containers.
5. Validate local startup and recovery behavior.

## Checklist
- `make up/down` works from a clean workspace.
- Startup depends on readiness of DB, Redis, and storage/search where required.
- Ports and URLs match docs and frontend/API assumptions.
- Logs and restart behavior are easy to inspect.

## Guardrails
- Do not introduce breaking host port changes without docs updates.
- Keep local defaults safe and explicitly documented.

## Example output
- `make search-init` command wired into local flow and verified in runbook.
- MinIO and OpenSearch startup checks included for first-run bootstrap.

# Security Review Hook

## Trigger
Before merge for API, parser, upload, auth, secret, or privileged-worker changes.

## Checklist
- [ ] Tenant checks enforced on mutable queries and mutations.
- [ ] Secrets/password values are not logged and not returned to clients.
- [ ] Input type, size, and format boundaries are explicit.
- [ ] Archive path traversal, expansion, and extraction limits are enforced.
- [ ] Event/audit data excludes secret material.
- [ ] Known authentication/authorization placeholders are documented with upgrade plan.

## Blocking criteria
- Unbounded input parsing with known DoS risk.
- Missing tenant check on sensitive mutation/query path.
- Secret material visible in logs, events, or responses.
- Missing threat-model update for increased attack surface.

## Required evidence
- Threat model entry or security note in `memory/known-risks.md`.
- Mitigation owner and follow-up date for any accepted risk.
- Evidence from tests validating critical security constraints.

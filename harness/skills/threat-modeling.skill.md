# Threat Modeling Skill

## When to use
Before shipping changes to input handling, auth boundaries, secrets, or high-risk parser/worker paths.

## Procedure
1. Identify trust boundaries and data flows.
2. Model attacks with severity/likelihood and attack scenarios.
3. Assign mitigations and owners.
4. Add accepted risk notes for deferred actions.
5. Update docs and hooks with review evidence.

## Checklist
- Tenant and user context validated at boundaries.
- Secret retrieval/logging behavior reviewed.
- Path traversal/size/ratio/rate limits defined.
- Error responses avoid leaking secrets.

## Guardrails
- Don’t skip risk modeling for parser, upload, or archive changes.
- Never document unresolved critical risks as “done.”

## Example output
- Threat model entry in `docs/security/threat-model.md` and `harness/memory/known-risks.md` for encrypted archive handling.

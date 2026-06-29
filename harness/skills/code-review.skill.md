# Code Review Skill

## When to use
For final review prior to handoff where behavior, security, and quality are tightly coupled.

## Procedure
1. Compare patch against requirements and acceptance criteria.
2. Review contracts, status semantics, and migration impact.
3. Validate tests and evidence links.
4. Confirm hooks and self-improvement updates were executed.
5. Flag and categorize blockers clearly.

## Checklist
- Contract compatibility and migration safety verified.
- Error handling and retries are deterministic.
- Security-sensitive paths and logs reviewed.
- Docs and runbooks updated where visible behavior changed.
- Memory files updated for non-trivial ADR/risk updates.

## Guardrails
- Keep reviewer recommendations actionable with priority and owner.
- Don’t approve without required hooks and test evidence.

## Example output
- Structured review note with PASS/CONCERN lists, risk ratings, and required follow-up.

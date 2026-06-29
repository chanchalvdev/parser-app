# Pre-Implementation Checklist

## Trigger
Before starting code changes for any feature, schema update, parser change, or major refactor.

## Checklist
- [ ] Requirements, scope, and out-of-scope clearly documented.
- [ ] Relevant goals reviewed (product/engineering/security/quality).
- [ ] Primary owner and all secondary owners selected.
- [ ] Architecture, contract, and migration impact understood.
- [ ] Test impact and rollback plan defined.
- [ ] Security-sensitive boundaries and validation requirements listed.
- [ ] Existing evidence log or ADRs consulted for similar work.

## Blocking criteria
- Missing owner/agent assignment.
- Unclear status transitions or lifecycle states.
- No migration/rollback strategy for persistence-impacting changes.
- Security implications not identified.
- Missing acceptance criteria or confidence on test plan.

## Required evidence
- Scope note with impacted files.
- Draft acceptance criteria.
- Initial owner map and sign-off status.
- Decision on rollback and risk boundaries before edits.

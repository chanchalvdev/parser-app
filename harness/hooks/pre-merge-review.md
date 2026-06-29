# Pre-Merge Review Hook

## Trigger
Before merge/hand-off of a feature branch to mainline.

## Checklist
- [ ] Acceptance criteria mapped and validated for each implemented area.
- [ ] API/service/worker/frontend tests added or existing evidence linked.
- [ ] Security and parser hooks passed, or explicit waivers with mitigation plan.
- [ ] Data migration/rollback path documented.
- [ ] Docs and runbooks updated for behavior changes.
- [ ] Memory files updated for new ADRs, risks, and backlog actions.
- [ ] Owners have provided explicit sign-off.

## Blocking criteria
- Any unresolved high-severity security finding.
- Missing evidence for changed behavior or incomplete migration notes.
- Unclear ownership of contract changes.
- Documentation mismatch with actual API/behavior.
- Blocking test gaps without explicit risk acceptance.

## Required evidence
- QA summary with passed/blocked matrix.
- Hook and workflow completion checklist.
- Links to changed files and migration notes.
- Security and architecture owner decisions.

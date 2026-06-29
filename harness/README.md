# Internal Engineering Harness

This folder is the operational handbook for human and AI contributors.

It provides a complete, reusable setup for planning, implementation, verification, and self-improvement across this project.

## Purpose

- Keep implementation consistent across Go API, Python worker, frontend, DB, search, and infrastructure.
- Provide explicit ownership and guardrails for every class of change.
- Maintain a reviewable quality system with required hooks and evidence.
- Capture lessons in `harness/memory/**` so the process improves over time.

## Layout

- `harness.yaml` — machine-readable orchestrator (agents, hooks, workflows, gates, commands).
- `harness/commands.md` — operational command catalog.
- `harness/run.sh` — command entrypoint.
- `harness/scripts/` — executable helper scripts.
- `harness/artifacts/` — generated evidence files.
- `goals/` — product, engineering, security, and quality objectives.
- `agents/` — primary owners for sustained code areas.
- `subagents/` — task-oriented specialists.
- `skills/` — reusable procedures and checklists.
- `hooks/` — mandatory process and safety checkpoints.
- `workflows/` — staged playbooks for features, API changes, incidents, and releases.
- `memory/` — durable ADRs, risks, backlog, glossary.

## Using this harness with Codex

1. Read goals to establish scope and constraints.
2. Select ownership:
   - Primary agent for principal code path.
   - Secondary agents for contracts, security, and test impact.
3. Ask the primary agent to execute:
   - **Review** prompt.
   - **Design** prompt (if architecture or cross-service impact exists).
4. Execute the workflow template matching your task:
   - Feature → `workflows/feature-development.workflow.md`
   - Parser-focused → `workflows/parser-addition.workflow.md`
   - API contract changes → `workflows/api-change.workflow.md`
   - Incident → `workflows/incident-debugging.workflow.md`
   - Release → `workflows/release.workflow.md`
5. Run required hooks before coding (`pre-implementation`, safety/security hooks when relevant).
6. Implement in owned directories from the selected agent definition.
7. Apply verification steps and close hooks with evidence.
8. Perform the self-improvement step (memory/risk/backlog updates).
9. Final review with merge/readiness hooks and documentation updates.

## Every coding batch should pass through harness

Use this as the standard local loop for each code change:

1. `make harness` for lightweight planning and guardrail evidence.
2. `make harness-full` when you also want test + validation evidence in the same cycle.

Equivalent direct commands:

- `bash harness/run.sh pipeline --scope "<scope>" --agent "<agent>" --mode light`
- `bash harness/run.sh pipeline --scope "<scope>" --agent "<agent>" --mode full`

## End-to-end operating model

This harness enforces:

- **Review**: confirm intent, constraints, and blast radius.
- **Design**: architecture and contract updates with rollback.
- **Plan**: task decomposition + acceptance criteria.
- **Agree**: explicit owner sign-off before implementation.
- **Test**: unit, integration, and workflow checks.
- **Validation**: end-to-end smoke with sample data.

### Agreement and handoff expectations

- For each implemented feature, include owner signoff from affected major lanes in task notes.
- Security, parser, and migration-sensitive work requires explicit `security-review` and/or `parser-safety-review` pass.
- Merge/hand-off waits for `pre-merge-review` and `release-readiness` evidence.

## Self-improving loop

After each completion cycle:

- If architecture changed, update `harness/memory/architecture-decisions.md`.
- If a new risk appears, update `harness/memory/known-risks.md`.
- If follow-up scope emerges, update `harness/memory/backlog.md`.
- If terminology changes, update `harness/memory/glossary.md`.
- Keep command evidence linked to changed behavior in PR/task notes.

## Executable command flow

- Use the command catalog: `harness/commands.md`.
- Run as one-liners: `bash harness/run.sh <command> [args]`.
  - Example: `bash harness/run.sh review "harness completion"`.

## Custom command catalog (logical)

The harness defines command-style operations in `harness.yaml` as process shortcuts:

- `review` — scope review + risk register check.
- `design` — HLD/LLD + contract and migration proposal.
- `plan` — task and milestone decomposition.
- `agree` — ownership and rollback acknowledgment.
- `test` — run test-all and evidence collection.
- `validation` — end-to-end smoke pass and dashboard proof.
- `self_improve` — memory and docs hardening.
- `release` — release package summary.
- `harness/commands.md` contains practical mapped shell equivalents and expected outputs for each command.
- `pipeline` — full flow command (`light` or `full` mode).
- `harness` — alias for pipeline.

## Primary ownership map

- **API and orchestration**: Backend Go Agent.
- **Ingestion/parsers/loaders/recursive workflows**: Worker Python Agent.
- **UI/workflows**: Frontend React Agent.
- **Index/search contracts/queries**: Search Agent.
- **Schema and migrations**: Database Agent.
- **Infrastructure/local bootstrap**: DevOps Agent.
- **Security posture/tooling**: Security Agent.
- **Validation and test strategy**: QA Agent.
- **Canonical docs/runbooks**: Docs Agent.
- **Cross-service architecture**: Architect Agent.

## Related docs

- `docs/harness/how-to-use-harness.md` for a shorter practical path.
- `docs/architecture/**`, `docs/api/**`, `docs/runbooks/**` for implementation details.
- `harness/commands.md` and `harness/run.sh` for execution.

## Optional local enforcement

Create a lightweight pre-commit hook when you want Codex work to always include harness review artifacts:

```bash
mkdir -p .githooks
cat > .githooks/pre-commit <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

bash "$(pwd)/harness/run.sh" harness --scope "pre-commit guardrail" --agent "architect" --mode light
EOF
chmod +x .githooks/pre-commit
git config core.hooksPath .githooks
```

Keep this command lightweight unless your environment is ready for full test/validation in pre-commit.

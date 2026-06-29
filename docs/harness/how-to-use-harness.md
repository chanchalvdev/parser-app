# How to use the harness with Codex

This guide is the practical command flow for using the harness artifacts in day-to-day development.

## 1) Start with context

1. Open `harness/goals/product-goals.md`, `harness/goals/engineering-goals.md`, and security/quality goals.
2. Review existing behavior in API/worker/UI docs if your change is contract-sensitive.
3. Read the nearest workflow template in `harness/workflows/`.

## 2) Pick owners and agree

1. Choose one primary owning agent from `harness/agents/`.
2. Add secondary agents when a change touches multiple boundaries.
3. Copy the reusable prompt from the chosen agent and request:
   - scope review
   - contract impact
   - acceptance criteria
4. Record “agree” with owner expectations before editing.

## 3) Execute workflow sequence

- **Features**: `feature-development.workflow.md`
- **Parser changes**: `parser-addition.workflow.md`
- **API contract changes**: `api-change.workflow.md`
- **Incident work**: `incident-debugging.workflow.md`
- **Release prep**: `release.workflow.md`

Each workflow should include hook execution points and closure evidence.

## 4) Guardrail execution order

1. `harness/hooks/pre-implementation-checklist.md`
2. Run design/safety hooks based on scope:
   - `security-review.md` for API, upload, auth, secrets, archive handling.
   - `parser-safety-review.md` for parser or extractor work.
3. Build/test and quality gates by affected domains.
4. `harness/hooks/pre-commit-checklist.md`
5. `harness/hooks/pre-merge-review.md`
6. `harness/hooks/release-readiness.md` for release handoff.

## 5) Validate and document

- Use command evidence:
  - `make test-api`, `make test-worker`, `make test-web`, `make test-all`.
  - local development runbook checks.
- Update:
  - `docs/runbooks/**` for behavior path changes.
  - `harness/memory/**` for architecture risk decisions and residual risks.
  - `harness/commands.md` and `harness/memory` when process expectations changed.

## 6) Use harness scripts

You can run harness commands directly:

- `bash harness/run.sh review "<scope>"`
- `bash harness/run.sh design <agent>`
- `bash harness/run.sh plan "<scope>"`
- `bash harness/run.sh agree`
- `bash harness/run.sh test`
- `bash harness/run.sh validate`
- `bash harness/run.sh self_improve "<note>"`
- `bash harness/run.sh release`
- `bash harness/run.sh pipeline --scope "<scope>" --agent "<agent>" --mode light`
- `bash harness/run.sh pipeline --scope "<scope>" --agent "<agent>" --mode full`

Evidence is emitted under `harness/artifacts/`.

For local cadence:

- `make harness` (default light mode)
- `make harness-full` (light + tests + validation)
- `make harness HARNESS_SCOPE="parser update" HARNESS_AGENT="worker-python"`

### Optional pre-commit wiring

If you want to gate each local commit by harness guardrails:

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

## 7) Self-improvement loop

After each major task, run a short process update:

1. Note what changed in `harness/memory/backlog.md`.
2. Add/confirm ADRs in `harness/memory/architecture-decisions.md`.
3. Capture unresolved risks and next actions in `harness/memory/known-risks.md`.
4. Keep the agent and hook evidence in PR notes for future automation.

## 8) Why this exists

The goal is to make AI and human collaboration repeatable and auditable:
- stable ownership,
- predictable workflows,
- explicit security gates,
- measurable validation,
- continuous process improvement.

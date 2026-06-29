# Harness Command Catalog

Use these commands as process operators (run from repository root).

## `harness:review <scope>`
- **Shell command:** `bash harness/scripts/review.sh <scope>`
- **Intent:** map impacted surfaces, boundaries, and owners.
- **Output:** timestamped `harness/artifacts/review-*.md`.

## `harness:design <agent>`
- **Shell command:** `bash harness/scripts/design.sh <agent>`
- **Intent:** define contract and migration deltas + rollback strategy.
- **Output:** `harness/artifacts/design-*.md`.

## `harness:plan <scope>`
- **Shell command:** `bash harness/scripts/plan.sh <scope>`
- **Intent:** sequence work by owners and acceptance criteria.
- **Output:** `harness/artifacts/plan-*.md`.

## `harness:agree`
- **Shell command:** `bash harness/scripts/agree.sh`
- **Intent:** capture explicit owner sign-off checkpoints before implementation.
- **Output:** `harness/artifacts/agree-*.md`.

## `harness:test`
- **Shell command:** `bash harness/scripts/test-run.sh`
- **Intent:** collect verification evidence.
- **Output:** `harness/artifacts/test-*.md` (and make command logs in `/tmp`).

## `harness:validate`
- **Shell command:** `bash harness/scripts/validate.sh`
- **Intent:** run end-to-end smoke evidence checks.
- **Output:** `harness/artifacts/validate-*.md`.

## `harness:pipeline` / `harness` (`--mode` light or full)
- **Shell command:** `bash harness/run.sh pipeline --scope "<scope>" --agent "<agent>" --mode <light|full>`
- **Intent:** run the full workflow in one step.
- **Light mode output:** `review`, `design`, `plan`, and `agree` evidence.
- **Full mode output:** `review`, `design`, `plan`, `agree`, `test`, and `validate` evidence.
- **Default mode:** `light`.
- **Output:** `harness/artifacts/pipeline-*.md` plus generated artifacts from each stage.

## `harness:self_improve "<message>"`
- **Shell command:** `bash harness/scripts/self-improve.sh "<message>"`
- **Intent:** append process-learning notes and backlog updates.
- **Output:** updates to `harness/memory/{backlog,known-risks,architecture-decisions}.md`.

## `harness:release`
- **Shell command:** `bash harness/scripts/release.sh`
- **Intent:** assemble release evidence package.
- **Output:** `harness/artifacts/release-*.md`.

## Optional unified entrypoint

- `bash harness/run.sh <command> [args]`
  - Commands: `review`, `design`, `plan`, `agree`, `test`, `validate`, `pipeline`, `harness`, `self_improve`, `release`.
  - Example: `bash harness/run.sh review "parser changes"`
- `make harness` and `make harness-full` are convenience wrappers around `bash harness/scripts/pipeline.sh` for local developer flow.

### Make wrappers

- `make harness` -> `bash harness/scripts/pipeline.sh --scope "$(HARNESS_SCOPE)" --agent "$(HARNESS_AGENT)" --mode "$(HARNESS_MODE)"`
- `make harness-full` -> `bash harness/scripts/pipeline.sh --scope "$(HARNESS_SCOPE)" --agent "$(HARNESS_AGENT)" --mode full`

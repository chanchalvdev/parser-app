#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
ARTIFACT_DIR="$ROOT/harness/artifacts"
mkdir -p "$ARTIFACT_DIR"

SCOPE="implementation"
AGENT="architect"
MODE="light"
SELF_IMPROVE_NOTE=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --scope)
      SCOPE="${2:-implementation}"
      shift 2
      ;;
    --agent)
      AGENT="${2:-architect}"
      shift 2
      ;;
    --mode)
      MODE="${2:-light}"
      if [[ "$MODE" != "light" && "$MODE" != "full" ]]; then
        echo "Invalid mode '$MODE'. Expected light or full."
        exit 1
      fi
      shift 2
      ;;
    --self-improve)
      SELF_IMPROVE_NOTE="${2:-}"
      shift 2
      ;;
    -h|--help)
      cat <<'USAGE'
Usage: bash harness/scripts/pipeline.sh [--scope "<scope>"] [--agent "<agent>"] [--mode <light|full>] [--self-improve "<note>"]

Modes:
  light  - review -> design -> plan -> agree
  full   - review -> design -> plan -> agree -> test -> validate

Example:
  bash harness/scripts/pipeline.sh --scope "parser enhancement" --agent "backend-go" --mode full
USAGE
      exit 0
      ;;
    *)
      echo "Unknown argument: $1"
      exit 1
      ;;
  esac
done

TS="$(date +%Y%m%d-%H%M%S)"
OUT="$ARTIFACT_DIR/pipeline-${TS}.md"
PIPELINE_STEPS=()

{
  echo "# Harness Pipeline"
  echo "- Timestamp: $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
  echo "- Scope: ${SCOPE}"
  echo "- Agent: ${AGENT}"
  echo "- Mode: ${MODE}"
  echo
} > "$OUT"

run_step() {
  local script="$1"
  local step_name="$2"
  local arg="$3"

  if [[ -n "$arg" ]]; then
    "$script" "$arg" | tee -a "$OUT"
  else
    "$script" | tee -a "$OUT"
  fi

  PIPELINE_STEPS+=("$step_name")
}

run_step "$SCRIPT_DIR/review.sh" "review" "$SCOPE"
run_step "$SCRIPT_DIR/design.sh" "design" "$AGENT"
run_step "$SCRIPT_DIR/plan.sh" "plan" "$SCOPE"
run_step "$SCRIPT_DIR/agree.sh" "agree" ""

if [[ "$MODE" == "full" ]]; then
  run_step "$SCRIPT_DIR/test-run.sh" "test" ""
  run_step "$SCRIPT_DIR/validate.sh" "validate" ""
fi

if [[ -n "$SELF_IMPROVE_NOTE" ]]; then
  "$SCRIPT_DIR/self-improve.sh" "$SELF_IMPROVE_NOTE" | tee -a "$OUT"
  if [[ -f "$ROOT/harness/memory/backlog.md" ]]; then
    echo "- self-improve: backlog updated" | tee -a "$OUT"
  else
    echo "- self-improve: skipped (memory file missing)" | tee -a "$OUT"
  fi
fi

{
  echo "## Executed steps"
  for step in "${PIPELINE_STEPS[@]}"; do
    echo "- ${step}: completed"
  done
  if [[ -n "$SELF_IMPROVE_NOTE" ]]; then
    echo "- self-improve: completed"
  fi
} >> "$OUT"

echo "Created pipeline evidence: $OUT"
cat "$OUT"

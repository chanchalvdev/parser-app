#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
ARTIFACT_DIR="$ROOT/harness/artifacts"
mkdir -p "$ARTIFACT_DIR"

SCOPE="${1:-unspecified}"
TS="$(date +%Y%m%d-%H%M%S)"
OUT="$ARTIFACT_DIR/review-${TS}.md"
{
  echo "# Harness Review"
  echo "- Timestamp: $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
  echo "- Scope: ${SCOPE}"
  echo
  echo "## Context"
  echo "- Read: harness/goals/product-goals.md, harness/goals/engineering-goals.md"
  echo "- Read: harness/goals/security-goals.md, harness/goals/quality-goals.md"
  echo "- Read: harness/agents/*.agent.md and harness/workflows/*.workflow.md"
  echo
  echo "## Required checks"
  echo "- Confirm ownership map in harness/README.md"
  echo "- Identify impacted boundaries (API/worker/UI/DB/search/infra)"
  echo "- Check existing hooks in harness/hooks/*"
  echo
  echo "## Suggested impacted files"
  if [[ -n "$SCOPE" && -d "$ROOT/$SCOPE" ]]; then
    rg -n "" "$ROOT/$SCOPE" | head -n 5 | sed 's#^#- #' || true
  else
    echo "- api: harness/agents/backend-go.agent.md"
    echo "- worker: harness/agents/worker-python.agent.md"
    echo "- ui: harness/agents/frontend-react.agent.md"
    echo "- db/search: harness/agents/database.agent.md / harness/agents/search.agent.md"
  fi
  echo
  echo "## Owner sign-off candidates"
  echo "- architect"
  echo "- backend-go"
  echo "- worker-python"
  echo "- security"
  echo "- qa"
} > "$OUT"

echo "Created review evidence: $OUT"
cat "$OUT"

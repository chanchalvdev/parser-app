#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
ARTIFACT_DIR="$ROOT/harness/artifacts"
mkdir -p "$ARTIFACT_DIR"

SCOPE="${1:-implementation}"
TS="$(date +%Y%m%d-%H%M%S)"
OUT="$ARTIFACT_DIR/plan-${TS}.md"

{
  echo "# Harness Execution Plan"
  echo "- Timestamp: $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
  echo "- Scope: ${SCOPE}"
  echo
  echo "## Workstream decomposition"
  echo "1. Contract and scope freeze"
  echo "2. Implementation by owned boundaries"
  echo "3. Validation tests"
  echo "4. Hook closure and evidence collection"
  echo "5. Documentation and self-improvement updates"
  echo
  echo "## Suggested tasks"
  echo "- [ ] API changes (routes, handlers, services, repos)"
  echo "- [ ] Worker changes (parser/orchestrator/loader)"
  echo "- [ ] Frontend hooks and page updates"
  echo "- [ ] DB/index migrations if needed"
  echo "- [ ] Security + parser hooks"
  echo "- [ ] Documentation updates"
  echo
  echo "## Acceptance criteria"
  echo "- Required hooks executed"
  echo "- Test commands executed or risk-blocked with rationale"
  echo "- Evidence logs attached"
} > "$OUT"

echo "Created plan evidence: $OUT"
cat "$OUT"

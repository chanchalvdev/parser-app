#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
ARTIFACT_DIR="$ROOT/harness/artifacts"
mkdir -p "$ARTIFACT_DIR"

TS="$(date +%Y%m%d-%H%M%S)"
OUT="$ARTIFACT_DIR/agree-${TS}.md"

{
  echo "# Harness Agreement"
  echo "- Timestamp: $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
  echo
  echo "## Owner approvals"
  echo "- [ ] Architect"
  echo "- [ ] Backend Go"
  echo "- [ ] Worker Python"
  echo "- [ ] Security"
  echo "- [ ] QA"
  echo
  echo "## Blocking risks"
  echo "- [ ] Risk statements captured in harness/memory/known-risks.md"
  echo "- [ ] Rollback strategy approved"
  echo
  echo "## Non-goals / exclusions"
  echo "- Document any excluded changes explicitly"
  echo
  echo "## Signature"
  echo "- Added by: ${USER:-unknown}"
  echo "- Date: $(date -u '+%Y-%m-%d')"
} > "$OUT"

echo "Created agreement evidence: $OUT"
cat "$OUT"

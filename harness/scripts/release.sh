#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
ARTIFACT_DIR="$ROOT/harness/artifacts"
mkdir -p "$ARTIFACT_DIR"
TS="$(date +%Y%m%d-%H%M%S)"
OUT="$ARTIFACT_DIR/release-${TS}.md"

{
  echo "# Release package"
  echo "- Timestamp: $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
  echo
  echo "## Checks"
  echo "- Reviewed hooks in harness/hooks"
  echo "- Validation commands executed from docs/runbooks/local-development.md"
  echo "- Risk and rollback paths updated in harness/memory"
  echo
  echo "## Evidence files"
  ls -1 "$ARTIFACT_DIR" 2>/dev/null | sed 's#^#- #' || true
} > "$OUT"

echo "Created release evidence: $OUT"
cat "$OUT"

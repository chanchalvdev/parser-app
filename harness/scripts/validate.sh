#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
ARTIFACT_DIR="$ROOT/harness/artifacts"
mkdir -p "$ARTIFACT_DIR"
TS="$(date +%Y%m%d-%H%M%S)"
OUT="$ARTIFACT_DIR/validate-${TS}.md"

{
  echo "# Harness Validation"
  echo "- Timestamp: $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
  echo "- Scope: local smoke e2e"
  echo
  echo "## Commands"
  echo "- make -C $ROOT up"
  echo "- API health: curl http://localhost:8080/health"
  echo "- Worker/API runbook checks"
  echo
  echo "## Notes"
  echo "Run this after the full stack is up and sample data is available."
} > "$OUT"

if command -v curl >/dev/null 2>&1; then
  if curl -sSf http://localhost:8080/health >/tmp/harness_health.log 2>&1; then
    echo "- health check: PASS" >> "$OUT"
  else
    echo "- health check: FAIL (no running API)" >> "$OUT"
  fi
else
  echo "- health check: SKIPPED (curl not available)" >> "$OUT"
fi

echo "Created validation evidence: $OUT"
cat "$OUT"

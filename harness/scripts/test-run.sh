#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
ARTIFACT_DIR="$ROOT/harness/artifacts"
mkdir -p "$ARTIFACT_DIR"
TS="$(date +%Y%m%d-%H%M%S)"
OUT="$ARTIFACT_DIR/test-${TS}.md"

{
  echo "# Harness Test Run"
  echo "- Timestamp: $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
  echo "- Target: test-all"
  echo
  echo "## Commands"
} > "$OUT"

set +e
if make -C "$ROOT" test-api > /tmp/harness_test_api.log 2>&1; then
  echo "- test-api: PASS" >> "$OUT"
else
  echo "- test-api: FAIL (see /tmp/harness_test_api.log)" >> "$OUT"
fi
if make -C "$ROOT" test-worker > /tmp/harness_test_worker.log 2>&1; then
  echo "- test-worker: PASS" >> "$OUT"
else
  echo "- test-worker: FAIL (see /tmp/harness_test_worker.log)" >> "$OUT"
fi
if make -C "$ROOT" test-web > /tmp/harness_test_web.log 2>&1; then
  echo "- test-web: PASS" >> "$OUT"
else
  echo "- test-web: FAIL (see /tmp/harness_test_web.log)" >> "$OUT"
fi
set -e

echo "Created test evidence: $OUT"
cat "$OUT"

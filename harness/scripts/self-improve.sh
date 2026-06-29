#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
BACKLOG="$ROOT/harness/memory/backlog.md"
RISKS="$ROOT/harness/memory/known-risks.md"
ADR="$ROOT/harness/memory/architecture-decisions.md"
MSG="${1:-manual self-improve step run}"
TS="$(date +%Y%m%d-%H%M%S)"

{
  echo "## Self-improvement update ($TS)"
  echo "- $MSG"
  echo
} >> "$BACKLOG"

{
  echo
  echo "- Added from harness self-improve run: $MSG"
} >> "$RISKS"

{
  echo
  echo "- Self-improve run on $(date -u '+%Y-%m-%d %H:%M:%S UTC'): $MSG"
  echo "- File touched: harness/memory/"
} >> "$ADR"

echo "Self-improvement notes appended to backlog, risks, ADR."

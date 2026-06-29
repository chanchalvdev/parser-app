#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
ARTIFACT_DIR="$ROOT/harness/artifacts"
mkdir -p "$ARTIFACT_DIR"

AGENT="${1:-architect}"
TS="$(date +%Y%m%d-%H%M%S)"
OUT="$ARTIFACT_DIR/design-${TS}.md"

{
  echo "# Harness Design"
  echo "- Timestamp: $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
  echo "- Primary agent: ${AGENT}"
  echo
  echo "## HLD/LLD impact"
  echo "- Service boundaries touched:"
  echo "  - API (routes/services/repositories)"
  echo "  - Worker (parser/orchestrator/loaders)"
  echo "  - Search (mappings/queries)"
  echo "  - UI (types/services/pages)"
  echo "  - DB (schema/queries/indices)"
  echo
  echo "## Migration and rollback"
  echo "- Record migration requirements in docs/runbooks/local-development.md"
  echo "- Add rollback note to harness/memory/architecture-decisions.md if risky"
  echo
  echo "## Contracts"
  echo "- Request/response deltas for API and frontend contracts"
  echo "- Status/state machine changes"
  echo "- Event and audit impact"
  echo
  echo "## Validation plan"
  echo "- Contract tests: API + frontend model compatibility"
  echo "- Smoke path: upload/parse/index/search"
} > "$OUT"

echo "Created design evidence: $OUT"
cat "$OUT"

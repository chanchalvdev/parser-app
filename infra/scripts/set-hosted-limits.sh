#!/usr/bin/env bash
# Tighten processing limits for the hosted deployment (run after migrate + seed).
# Large files (> 10 MB) are meant to be processed on a local stack, not here.
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -a; . ./.env; set +a
DC="docker compose -f docker-compose.yml -f docker-compose.prod.yml"
$DC exec -T postgres psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -v ON_ERROR_STOP=1 <<'SQL'
UPDATE system_settings s SET setting_value = v.val, updated_at = NOW()
FROM (VALUES
  ('max_upload_size_mb',    to_jsonb(10)),
  ('max_extracted_size_mb', to_jsonb(200)),
  ('max_extracted_files',   to_jsonb(5000)),
  ('max_archive_depth',     to_jsonb(5)),
  ('max_expansion_ratio',   to_jsonb(20))
) AS v(key, val)
WHERE s.setting_key = v.key;
SELECT setting_key, setting_value FROM system_settings WHERE setting_key LIKE 'max_%' ORDER BY 1;
SQL

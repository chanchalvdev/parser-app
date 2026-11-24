#!/usr/bin/env bash
# One-time setup of a hosted database (e.g. Neon) from your machine.
# Usage: DATABASE_URL='postgres://...sslmode=require' ./infra/scripts/remote-setup.sh
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../.."
: "${DATABASE_URL:?set DATABASE_URL}"
for f in $(ls infra/migrations/*.up.sql | sort); do
  echo "applying $f"; psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$f"
done
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f infra/migrations/seed.sql
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 <<'SQL'
UPDATE system_settings s SET setting_value = v.val, updated_at = NOW()
FROM (VALUES
  ('max_upload_size_mb',    to_jsonb(10)),
  ('max_extracted_size_mb', to_jsonb(200)),
  ('max_extracted_files',   to_jsonb(5000)),
  ('max_archive_depth',     to_jsonb(5)),
  ('max_expansion_ratio',   to_jsonb(20))
) AS v(key, val) WHERE s.setting_key = v.key;
SQL

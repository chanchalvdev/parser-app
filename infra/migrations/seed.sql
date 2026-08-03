BEGIN;

WITH tenant_seed AS (
  INSERT INTO tenants (id, name, slug, status, created_at, updated_at)
  VALUES (
    '00000000-0000-0000-0000-000000000001'::uuid,
    'Default Tenant',
    'default',
    'active',
    NOW(),
    NOW()
  )
  ON CONFLICT (slug) DO UPDATE
  SET name = EXCLUDED.name,
      status = EXCLUDED.status,
      updated_at = NOW()
  RETURNING id
),
selected_tenant AS (
  SELECT id FROM tenant_seed
  UNION ALL
  SELECT id FROM tenants WHERE slug = 'default' AND NOT EXISTS (SELECT 1 FROM tenant_seed)
  LIMIT 1
),
admin_user AS (
  INSERT INTO users (
    id,
    tenant_id,
    email,
    password_hash,
    display_name,
    is_active,
    created_at,
    updated_at
  )
  SELECT
    '11111111-1111-1111-1111-111111111100'::uuid,
    (SELECT id FROM selected_tenant),
    'admin@fileplatform.local',
    '$2a$12$C5j4D4wQW7h4V2g5l2N7ieHf5cV8VQ0x2h7kJ6v7f9p6x6xw8kA1q',
    'Platform Admin',
    TRUE,
    NOW(),
    NOW()
  WHERE NOT EXISTS (
    SELECT 1
    FROM users
    WHERE tenant_id = (SELECT id FROM selected_tenant)
      AND email = 'admin@fileplatform.local'
  )
  RETURNING id, tenant_id
)

INSERT INTO system_settings (tenant_id, setting_key, setting_value, description, is_secret, created_at, updated_at)
SELECT
  s_tenant.id,
  s_key,
  s_value,
  s_desc,
  FALSE,
  NOW(),
  NOW()
FROM (
  SELECT
    (SELECT id FROM selected_tenant) AS id
) AS s_tenant
CROSS JOIN LATERAL (
  VALUES
    ('max_archive_depth', to_jsonb(20), 'Maximum nesting depth for archive extraction'),
    ('max_extracted_files', to_jsonb(1000000), 'Maximum number of files extracted per upload'),
    ('max_extracted_size_mb', to_jsonb(102400), 'Maximum extracted payload size in MB'),
    ('txt_small_file_limit_mb', to_jsonb(10), 'Maximum MB limit for small-text optimization'),
    ('max_upload_size_mb', to_jsonb(10240), 'Maximum raw upload size in MB (default 10 GB)'),
    ('enabled_parsers', to_jsonb(ARRAY['txt', 'log', 'csv', 'json', 'jsonl', 'xml', 'xlsx', 'pdf', 'text']), 'Comma-separable list of enabled parsers'),
    ('max_expansion_ratio', to_jsonb(100), 'Maximum extracted payload expansion ratio allowed'),
    ('parser_batch_size', to_jsonb(1000), 'Parser batch size for worker bulk DB writes'),
    ('search_index_batch_size', to_jsonb(1000), 'Search index batch size for worker bulk requests')
) AS s(s_key, s_value, s_desc)
ON CONFLICT (tenant_id, setting_key) DO UPDATE
SET setting_value = EXCLUDED.setting_value,
    description = EXCLUDED.description,
    updated_at = NOW();

COMMIT;

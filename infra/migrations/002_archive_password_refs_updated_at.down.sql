-- Rollback for 002_archive_password_refs_updated_at.
ALTER TABLE archive_password_refs
  DROP COLUMN IF EXISTS updated_at;

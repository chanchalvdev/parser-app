-- Add updated_at to archive_password_refs (was missing in 001).
-- Worker code (apps/worker/app/db/repositories.py::get_latest_archive_password_ref)
-- and API code (apps/api/internal/repositories/archive_password_refs_repository.go)
-- both SELECT updated_at and the API also writes it via ON CONFLICT DO UPDATE.
ALTER TABLE archive_password_refs
  ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

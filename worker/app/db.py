import json
from dataclasses import dataclass
from typing import Any, Optional

import psycopg

from .config import settings


@dataclass
class FileRecord:
    id: str
    upload_id: str
    job_id: str
    parent_id: Optional[str]
    path: str
    filename: str
    kind: str
    mime_type: str
    size_bytes: int
    sha256: str
    metadata: Any


class Database:
    def __init__(self) -> None:
        self.conn = psycopg.connect(settings.database_url)
        self.conn.autocommit = True

    def set_upload_status(self, upload_id: str, status: str, error_message: str = "") -> None:
        with self.conn.cursor() as cur:
            cur.execute(
                """
                UPDATE uploads SET status=%s, error_message=%s, updated_at=NOW()
                WHERE id=%s
                """,
                (status, error_message, upload_id),
            )

    def set_job_status(self, job_id: str, status: str, stage: str, error_message: str = "") -> None:
        with self.conn.cursor() as cur:
            cur.execute(
                """
                UPDATE jobs
                SET status=%s, stage=%s, error_message=%s, updated_at=NOW(),
                    attempt_count=attempt_count + CASE WHEN %s IN ('failed', 'password_required') THEN 1 ELSE 0 END,
                    finished_at=CASE WHEN %s IN ('completed','failed','password_required') THEN NOW() ELSE finished_at END
                WHERE id=%s
                """,
                (status, stage, error_message, status, status, job_id),
            )

    def add_file_record(self, *, upload_id: str, job_id: str, parent_id: Optional[str], file_path: str, filename: str, kind: str,
                       mime_type: str, size_bytes: int, sha256: str, metadata: dict[str, Any]) -> str:
        with self.conn.cursor() as cur:
            cur.execute(
                """
                INSERT INTO files (
                  id, upload_id, job_id, parent_id, path, filename, kind, mime_type, size_bytes, sha256, metadata
                )
                VALUES (gen_random_uuid(), %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
                RETURNING id
                """,
                (upload_id, job_id, parent_id, file_path, filename, kind, mime_type, size_bytes, sha256, json.dumps(metadata) if metadata else None),
            )
            return str(cur.fetchone()[0])

    def add_parsed_record(self, *, file_id: str, upload_id: str, job_id: str, file_path: str, source_format: str,
                          content: str, metadata: dict[str, Any]) -> None:
        with self.conn.cursor() as cur:
            cur.execute(
                """
                INSERT INTO parsed_records (
                  id, file_id, upload_id, job_id, file_path, source_format, content, metadata
                )
                VALUES (gen_random_uuid(), %s, %s, %s, %s, %s, %s, %s)
                """,
                (file_id, upload_id, job_id, file_path, source_format, content, json.dumps(metadata) if metadata else None),
            )

    def add_audit_log(self, upload_id: str, job_id: str, actor: str, event: str, details: dict[str, Any]) -> None:
        with self.conn.cursor() as cur:
            cur.execute(
                """
                INSERT INTO audit_logs (upload_id, job_id, actor, event, details, created_at)
                VALUES (%s, %s, %s, %s, %s, NOW())
                """,
                (upload_id, job_id, actor, event, json.dumps(details)),
            )

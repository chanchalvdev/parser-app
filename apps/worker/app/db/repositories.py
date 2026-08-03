from __future__ import annotations

from datetime import datetime, timezone
from typing import Any, Iterable, Mapping, Optional
import logging

from psycopg2.extras import Json, execute_values

from app.db.connection import DatabaseConnection


class WorkerRepositoryError(RuntimeError):
    pass


class NotFoundError(WorkerRepositoryError):
    pass


class WorkerRepository:
    def __init__(self, connection: DatabaseConnection):
        self._connection = connection
        self._logger = logging.getLogger("worker.repository")

    def _log_and_raise(self, action: str, error: Exception, context: Mapping[str, Any]) -> None:
        if isinstance(error, WorkerRepositoryError):
            raise error

        self._logger.exception(
            "worker repository action failed",
            extra={"action": action, **{k: v for k, v in context.items() if v is not None}},
        )
        raise WorkerRepositoryError(f"{action} failed: {error}") from error

    @staticmethod
    def _timestamp_now() -> datetime:
        return datetime.now(timezone.utc)

    def get_file(self, tenant_id: str, file_id: str) -> dict[str, Any]:
        if not tenant_id or not file_id:
            raise WorkerRepositoryError("tenant_id and file_id are required")
        try:
            with self._connection.connect() as connection:
                with connection.cursor() as cursor:
                    cursor.execute(
                        """
                        SELECT
                          id,
                          tenant_id,
                          parent_file_id,
                          upload_id,
                          original_name,
                          normalized_name,
                          extension,
                          detected_mime_type,
                          detected_file_type,
                          storage_path,
                          size_bytes,
                          sha256_hash,
                          depth,
                          is_archive,
                          is_password_protected,
                          processing_status,
                          created_by,
                          created_at,
                          updated_at
                        FROM files
                        WHERE id = %s AND tenant_id = %s
                        """,
                        (file_id, tenant_id),
                    )
                    row = cursor.fetchone()
                    if row is None:
                        raise NotFoundError(f"file not found for tenant {tenant_id} and file {file_id}")
                    columns = [desc[0] for desc in cursor.description]
                    return dict(zip(columns, row))
        except Exception as exc:
            self._log_and_raise("get_file", exc, {"tenant_id": tenant_id, "file_id": file_id})

    def get_setting(self, tenant_id: str, setting_key: str, default: Optional[Any] = None) -> Optional[Any]:
        if not tenant_id:
            raise WorkerRepositoryError("tenant_id is required")
        if not setting_key:
            raise WorkerRepositoryError("setting_key is required")
        try:
            with self._connection.connect() as connection:
                with connection.cursor() as cursor:
                    cursor.execute(
                        """
                        SELECT setting_value
                        FROM system_settings
                        WHERE tenant_id = %s AND setting_key = %s
                        """,
                        (tenant_id, setting_key),
                    )
                    row = cursor.fetchone()
                    if row is None:
                        return default
                    return row[0]
        except Exception as exc:
            self._log_and_raise(
                "get_setting",
                exc,
                {"tenant_id": tenant_id, "setting_key": setting_key},
            )

    def get_latest_archive_password_ref(self, tenant_id: str, file_id: str) -> dict[str, Any] | None:
        if not tenant_id or not file_id:
            raise WorkerRepositoryError("tenant_id and file_id are required")
        try:
            with self._connection.connect() as connection:
                with connection.cursor() as cursor:
                    cursor.execute(
                        """
                        SELECT
                          id,
                          tenant_id,
                          file_id,
                          upload_id,
                          password_ref_hash,
                          algorithm,
                          is_valid,
                          validated,
                          attempt_count,
                          last_validated_at,
                          created_by,
                          created_at,
                          updated_at
                        FROM archive_password_refs
                        WHERE tenant_id = %s AND file_id = %s
                        ORDER BY created_at DESC
                        LIMIT 1
                        """,
                        (tenant_id, file_id),
                    )
                    row = cursor.fetchone()
                    if row is None:
                        return None

                    columns = [desc[0] for desc in cursor.description]
                    password_ref = dict(zip(columns, row))
                    return password_ref
        except Exception as exc:
            self._log_and_raise(
                "get_latest_archive_password_ref",
                exc,
                {"tenant_id": tenant_id, "file_id": file_id},
            )

    def update_archive_password_ref_status(
        self,
        tenant_id: str,
        file_id: str,
        password_ref_hash: str,
        is_valid: bool,
        validated: bool,
        increment_attempt_count: bool = False,
    ) -> bool:
        if not tenant_id or not file_id:
            raise WorkerRepositoryError("tenant_id and file_id are required")
        if not password_ref_hash:
            raise WorkerRepositoryError("password_ref_hash is required")

        attempt_expr = "attempt_count + 1" if increment_attempt_count else "attempt_count"
        query = f"""
            UPDATE archive_password_refs
            SET is_valid = %s,
                validated = %s,
                attempt_count = {attempt_expr},
                last_validated_at = %s,
                updated_at = NOW()
            WHERE tenant_id = %s
              AND file_id = %s
              AND password_ref_hash = %s
        """

        try:
            with self._connection.connect() as connection:
                with connection.cursor() as cursor:
                    cursor.execute(
                        query,
                        (
                            is_valid,
                            validated,
                            self._timestamp_now(),
                            tenant_id,
                            file_id,
                            password_ref_hash,
                        ),
                    )
                    return cursor.rowcount > 0
        except Exception as exc:
            self._log_and_raise(
                "update_archive_password_ref_status",
                exc,
                {
                    "tenant_id": tenant_id,
                    "file_id": file_id,
                    "password_ref_hash": password_ref_hash,
                },
            )

    def update_file_status(self, tenant_id: str, file_id: str, status: str) -> None:
        if not tenant_id or not file_id:
            raise WorkerRepositoryError("tenant_id and file_id are required")
        if not status:
            raise WorkerRepositoryError("status is required")
        try:
            with self._connection.connect() as connection:
                with connection.cursor() as cursor:
                    cursor.execute(
                        """
                        UPDATE files
                        SET processing_status = %s,
                            updated_at = NOW()
                        WHERE id = %s AND tenant_id = %s
                        """,
                        (status, file_id, tenant_id),
                    )
                    if cursor.rowcount == 0:
                        raise NotFoundError(f"file not found for tenant {tenant_id} and file {file_id}")
        except Exception as exc:
            self._log_and_raise(
                "update_file_status",
                exc,
                {"tenant_id": tenant_id, "file_id": file_id, "status": status},
            )

    def update_file_hash(self, tenant_id: str, file_id: str, sha256_hash: str) -> None:
        if not tenant_id or not file_id:
            raise WorkerRepositoryError("tenant_id and file_id are required")
        if not sha256_hash:
            raise WorkerRepositoryError("sha256_hash is required")
        try:
            with self._connection.connect() as connection:
                with connection.cursor() as cursor:
                    cursor.execute(
                        """
                        UPDATE files
                        SET sha256_hash = %s,
                            updated_at = NOW()
                        WHERE id = %s AND tenant_id = %s
                        """,
                        (sha256_hash, file_id, tenant_id),
                    )
                    if cursor.rowcount == 0:
                        raise NotFoundError(f"file not found for tenant {tenant_id} and file {file_id}")
        except Exception as exc:
            self._log_and_raise(
                "update_file_hash",
                exc,
                {"tenant_id": tenant_id, "file_id": file_id},
            )

    def create_child_file(
        self,
        tenant_id: str,
        parent_file_id: Optional[str],
        upload_id: Optional[str],
        original_name: str,
        storage_path: str,
        size_bytes: int,
        *,
        depth: int = 1,
        normalized_name: Optional[str] = None,
        extension: Optional[str] = None,
        detected_mime_type: Optional[str] = None,
        detected_file_type: Optional[str] = None,
        is_archive: bool = False,
        is_password_protected: bool = False,
        processing_status: str = "queued",
        created_by: Optional[str] = None,
    ) -> dict[str, Any]:
        if not original_name or not storage_path:
            raise WorkerRepositoryError("original_name and storage_path are required")
        if not tenant_id:
            raise WorkerRepositoryError("tenant_id is required")
        if size_bytes < 0:
            raise WorkerRepositoryError("size_bytes must be non-negative")
        if depth < 0:
            raise WorkerRepositoryError("depth must be non-negative")

        try:
            with self._connection.connect() as connection:
                with connection.cursor() as cursor:
                    cursor.execute(
                        """
                        INSERT INTO files (
                          tenant_id,
                          parent_file_id,
                          upload_id,
                          original_name,
                          normalized_name,
                          extension,
                          detected_mime_type,
                          detected_file_type,
                          storage_path,
                          size_bytes,
                          depth,
                          is_archive,
                          is_password_protected,
                          processing_status,
                          created_by
                        )
                        VALUES (%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s)
                        RETURNING id, tenant_id, parent_file_id, upload_id, original_name, storage_path
                        """,
                        (
                            tenant_id,
                            parent_file_id,
                            upload_id,
                            original_name,
                            normalized_name,
                            extension,
                            detected_mime_type,
                            detected_file_type,
                            storage_path,
                            size_bytes,
                            depth,
                            is_archive,
                            is_password_protected,
                            processing_status,
                            created_by,
                        ),
                    )
                    file_id_row = cursor.fetchone()
                    if file_id_row is None:
                        raise WorkerRepositoryError("failed to create child file")
                    columns = ["id", "tenant_id", "parent_file_id", "upload_id", "original_name", "storage_path"]
                    return dict(zip(columns, file_id_row))
        except Exception as exc:
            self._log_and_raise(
                "create_child_file",
                exc,
                {"tenant_id": tenant_id, "parent_file_id": parent_file_id, "original_name": original_name},
            )

    def create_job_event(
        self,
        tenant_id: str,
        job_id: str,
        event_type: str,
        stage: str,
        message: str,
        metadata: Optional[dict[str, Any]] = None,
    ) -> dict[str, Any]:
        payload: dict[str, Any] = dict(metadata or {})
        payload["stage"] = stage
        try:
            with self._connection.connect() as connection:
                with connection.cursor() as cursor:
                    cursor.execute(
                        """
                        INSERT INTO job_events (
                          tenant_id,
                          job_id,
                          event_type,
                          event_message,
                          event_details
                        )
                        VALUES (%s, %s, %s, %s, %s)
                        RETURNING id, tenant_id, job_id, event_type, event_message, event_details, created_at
                        """,
                        (
                            tenant_id,
                            job_id,
                            event_type,
                            message,
                            Json(payload),
                        ),
                    )
                    row = cursor.fetchone()
                    if row is None:
                        raise WorkerRepositoryError("job event insert returned no row")
                    return {
                        "id": row[0],
                        "tenant_id": row[1],
                        "job_id": row[2],
                        "event_type": row[3],
                        "event_message": row[4],
                        "event_details": row[5],
                        "created_at": row[6],
                    }
        except Exception as exc:
            self._log_and_raise(
                "create_job_event",
                exc,
                {"tenant_id": tenant_id, "job_id": job_id, "event_type": event_type},
            )

    def update_job_status(
        self,
        tenant_id: str,
        job_id: str,
        status: str,
        current_stage: str,
        progress_percent: float,
        error_code: Optional[str] = None,
        error_message: Optional[str] = None,
        *,
        started_at: Optional[datetime] = None,
        completed_at: Optional[datetime] = None,
    ) -> dict[str, Any]:
        if not tenant_id or not job_id:
            raise WorkerRepositoryError("tenant_id and job_id are required")
        if not status:
            raise WorkerRepositoryError("status is required")
        if status not in {"pending", "running", "queued", "completed", "failed", "retrying"}:
            # allow platform-specific custom statuses, but still record
            self._logger.warning("update_job_status received unknown status", extra={"status": status, "job_id": job_id})
        if progress_percent is None:
            raise WorkerRepositoryError("progress_percent is required")
        if progress_percent < 0 or progress_percent > 100:
            raise WorkerRepositoryError("progress_percent must be between 0 and 100")
        if not current_stage:
            current_stage = status

        event_context = {
            "status": status,
            "current_stage": current_stage,
            "progress_percent": progress_percent,
            "error_code": error_code,
            "error_message": error_message,
        }

        try:
            with self._connection.connect() as connection:
                with connection.cursor() as cursor:
                    cursor.execute(
                        """
                        UPDATE ingestion_jobs
                        SET status = %s,
                            current_stage = %s,
                            progress_percent = %s,
                            error_code = %s,
                            error_message = %s,
                            started_at = COALESCE(%s, started_at),
                            completed_at = COALESCE(%s, completed_at),
                            updated_at = NOW()
                        WHERE id = %s AND tenant_id = %s
                        RETURNING id, tenant_id, root_file_id, status, current_stage, progress_percent, retry_count
                        """,
                        (
                            status,
                            current_stage,
                            progress_percent,
                            error_code,
                            error_message,
                            started_at,
                            completed_at,
                            job_id,
                            tenant_id,
                        ),
                    )
                    row = cursor.fetchone()
                    if row is None:
                        raise NotFoundError(f"job not found for tenant {tenant_id} and job {job_id}")

                    cursor.execute(
                        """
                        INSERT INTO job_events (
                          tenant_id,
                          job_id,
                          event_type,
                          event_message,
                          event_details
                        ) VALUES (%s, %s, %s, %s, %s)
                        """,
                        (
                            tenant_id,
                            job_id,
                            "job.status.updated",
                            f"job status changed to {status}",
                            Json(event_context),
                        ),
                    )
                    result = {
                        "id": row[0],
                        "tenant_id": row[1],
                        "root_file_id": row[2],
                        "status": row[3],
                        "current_stage": row[4],
                        "progress_percent": row[5],
                        "retry_count": row[6],
                    }

                    self._logger.info(
                        "job status transition",
                        extra={
                            "tenant_id": tenant_id,
                            "job_id": job_id,
                            "file_id": row[2],
                            "status": status,
                            "current_stage": current_stage,
                            "progress_percent": progress_percent,
                            "error_code": error_code,
                            "error_message": error_message,
                        },
                    )
                    return result
        except Exception as exc:
            self._log_and_raise(
                "update_job_status",
                exc,
                {
                    "tenant_id": tenant_id,
                    "job_id": job_id,
                    "status": status,
                    "current_stage": current_stage,
                },
            )

    def insert_parsed_records_batch(self, records: Iterable[dict[str, Any]]) -> list[dict[str, Any]]:
        normalized_records: list[tuple[Any, ...]] = []
        for record in records:
            tenant_id = record.get("tenant_id")
            file_id = record.get("file_id")
            job_id = record.get("job_id")
            if not tenant_id or not file_id or not job_id:
                raise WorkerRepositoryError("tenant_id, file_id, and job_id are required for parsed records")

            structured = record.get("structured_data")
            entities = record.get("extracted_entities")
            normalized_records.append(
                (
                    tenant_id,
                    file_id,
                    job_id,
                    record.get("record_type"),
                    record.get("record_number"),
                    record.get("line_number"),
                    record.get("chunk_number"),
                    record.get("start_line"),
                    record.get("end_line"),
                    record.get("content_text"),
                    Json(structured or {}),
                    Json(entities or {}),
                    record.get("event_timestamp"),
                )
            )

        if not normalized_records:
            return []

        query = """
            INSERT INTO parsed_records (
                tenant_id, file_id, job_id, record_type, record_number, line_number,
                chunk_number, start_line, end_line, content_text, structured_data,
                extracted_entities, event_timestamp
            )
            VALUES %s
            RETURNING id, tenant_id, file_id, job_id, record_type, record_number, created_at
        """
        try:
            with self._connection.connect() as connection:
                with connection.cursor() as cursor:
                    execute_values(cursor, query, normalized_records, page_size=500)
                    rows = cursor.fetchall()
                    return [
                        {
                            "id": row[0],
                            "tenant_id": row[1],
                            "file_id": row[2],
                            "job_id": row[3],
                            "record_type": row[4],
                            "record_number": row[5],
                            "created_at": row[6],
                        }
                        for row in rows
                    ]
        except Exception as exc:
            self._log_and_raise("insert_parsed_records_batch", exc, {"batch_size": len(normalized_records)})

    def insert_parser_error(
        self,
        tenant_id: str,
        job_id: str,
        error_message: str,
        *,
        file_id: Optional[str] = None,
        upload_id: Optional[str] = None,
        error_code: Optional[str] = None,
        error_context: Optional[dict[str, Any]] = None,
        is_retryable: bool = False,
        stack_trace: Optional[str] = None,
        occurred_at: Optional[datetime] = None,
    ) -> dict[str, Any]:
        if not error_message:
            raise WorkerRepositoryError("error_message is required")
        try:
            with self._connection.connect() as connection:
                with connection.cursor() as cursor:
                    cursor.execute(
                        """
                        INSERT INTO parser_errors (
                            tenant_id,
                            job_id,
                            file_id,
                            upload_id,
                            error_code,
                            error_message,
                            error_context,
                            is_retryable,
                            stack_trace,
                            occurred_at
                        )
                        VALUES (%s,%s,%s,%s,%s,%s,%s,%s,%s,%s)
                        RETURNING id, tenant_id, job_id, file_id, upload_id, error_code, error_message, is_retryable, occurred_at
                        """,
                        (
                            tenant_id,
                            job_id,
                            file_id,
                            upload_id,
                            error_code,
                            error_message,
                            Json(error_context or {}),
                            is_retryable,
                            stack_trace,
                            occurred_at or self._timestamp_now(),
                        ),
                    )
                    row = cursor.fetchone()
                    if row is None:
                        raise WorkerRepositoryError("parser error insert returned no row")
                    return {
                        "id": row[0],
                        "tenant_id": row[1],
                        "job_id": row[2],
                        "file_id": row[3],
                        "upload_id": row[4],
                        "error_code": row[5],
                        "error_message": row[6],
                        "is_retryable": row[7],
                        "occurred_at": row[8],
                    }
        except Exception as exc:
            self._log_and_raise(
                "insert_parser_error",
                exc,
                {"tenant_id": tenant_id, "job_id": job_id},
            )

    def update_search_index_status(
        self,
        tenant_id: str,
        file_id: str,
        job_id: str,
        index_name: str,
        status: str,
        *,
        attempts: int = 0,
        last_error: Optional[str] = None,
        last_indexed_at: Optional[datetime] = None,
        document_id: Optional[str] = None,
    ) -> dict[str, Any]:
        if not status:
            raise WorkerRepositoryError("status is required")
        if not tenant_id or not file_id or not job_id or not index_name:
            raise WorkerRepositoryError("tenant_id, file_id, job_id, and index_name are required")
        try:
            with self._connection.connect() as connection:
                with connection.cursor() as cursor:
                    cursor.execute(
                        """
                        INSERT INTO search_index_status (
                            tenant_id, file_id, job_id, index_name, status,
                            attempts, last_error, last_indexed_at, document_id
                        )
                        VALUES (%s,%s,%s,%s,%s,%s,%s,%s,%s)
                        ON CONFLICT (tenant_id, file_id, job_id, index_name)
                        DO UPDATE SET
                          status = EXCLUDED.status,
                          attempts = EXCLUDED.attempts,
                          last_error = EXCLUDED.last_error,
                          last_indexed_at = EXCLUDED.last_indexed_at,
                          document_id = COALESCE(EXCLUDED.document_id, search_index_status.document_id),
                          updated_at = NOW()
                        RETURNING id, tenant_id, file_id, job_id, index_name, status, attempts, last_error, document_id
                        """,
                        (
                            tenant_id,
                            file_id,
                            job_id,
                            index_name,
                            status,
                            attempts,
                            last_error,
                            last_indexed_at,
                            document_id,
                        ),
                    )
                    row = cursor.fetchone()
                    if row is None:
                        raise WorkerRepositoryError("search_index_status upsert returned no row")
                    return {
                        "id": row[0],
                        "tenant_id": row[1],
                        "file_id": row[2],
                        "job_id": row[3],
                        "index_name": row[4],
                        "status": row[5],
                        "attempts": row[6],
                        "last_error": row[7],
                        "document_id": row[8],
                    }
        except Exception as exc:
            self._log_and_raise(
                "update_search_index_status",
                exc,
                {
                    "tenant_id": tenant_id,
                    "file_id": file_id,
                    "job_id": job_id,
                    "index_name": index_name,
                },
            )

    def get_job_context(self, job_id: str) -> tuple[str, str]:
        try:
            with self._connection.connect() as connection:
                with connection.cursor() as cursor:
                    cursor.execute(
                        """
                        SELECT tenant_id, root_file_id
                        FROM ingestion_jobs
                        WHERE id = %s
                        """,
                        (job_id,),
                    )
                    row = cursor.fetchone()
                    if row is None:
                        raise NotFoundError(f"job not found: {job_id}")
                    return row[0], row[1]
        except Exception as exc:
            self._log_and_raise("get_job_context", exc, {"job_id": job_id})

    def mark_job_running(self, job_id: str, tenant_id: str) -> dict[str, Any]:
        return self.update_job_status(
            tenant_id=tenant_id,
            job_id=job_id,
            status="running",
            current_stage="running",
            progress_percent=5.0,
            started_at=self._timestamp_now(),
        )

    def mark_job_completed(self, job_id: str, tenant_id: str) -> dict[str, Any]:
        return self.update_job_status(
            tenant_id=tenant_id,
            job_id=job_id,
            status="completed",
            current_stage="completed",
            progress_percent=100.0,
            completed_at=self._timestamp_now(),
        )

    def mark_job_failed(
        self,
        job_id: str,
        tenant_id: str,
        stage: str,
        error_code: str,
        message: str,
    ) -> dict[str, Any]:
        return self.update_job_status(
            tenant_id=tenant_id,
            job_id=job_id,
            status="failed",
            current_stage=stage,
            progress_percent=0.0,
            error_code=error_code,
            error_message=message,
            completed_at=self._timestamp_now(),
        )

    def update_file_metadata(
        self,
        file_id: str,
        tenant_id: str,
        sha256_hash: str,
        detected_mime_type: str,
        detected_file_type: str,
        is_archive: bool,
        processing_status: str,
        *,
        normalized_name: Optional[str] = None,
        extension: Optional[str] = None,
        is_password_protected: Optional[bool] = None,
    ) -> None:
        if not tenant_id or not file_id:
            raise WorkerRepositoryError("tenant_id and file_id are required")
        try:
            with self._connection.connect() as connection:
                with connection.cursor() as cursor:
                    columns = [
                        "sha256_hash = %s",
                        "detected_mime_type = %s",
                        "detected_file_type = %s",
                        "is_archive = %s",
                        "processing_status = %s",
                    ]
                    values: list[Any] = [
                        sha256_hash,
                        detected_mime_type,
                        detected_file_type,
                        is_archive,
                        processing_status,
                    ]

                    if normalized_name is not None:
                        columns.append("normalized_name = %s")
                        values.append(normalized_name)
                    if extension is not None:
                        columns.append("extension = %s")
                        values.append(extension)
                    if is_password_protected is not None:
                        columns.append("is_password_protected = %s")
                        values.append(is_password_protected)

                    query = f"""
                        UPDATE files
                        SET {", ".join(columns)},
                            updated_at = NOW()
                        WHERE id = %s AND tenant_id = %s
                        """
                    values.extend((file_id, tenant_id))

                    cursor.execute(
                        query,
                        tuple(values),
                    )
                    if cursor.rowcount == 0:
                        raise NotFoundError(f"file not found for update: {file_id}")
        except Exception as exc:
            self._log_and_raise(
                "update_file_metadata",
                exc,
                {"tenant_id": tenant_id, "file_id": file_id},
            )

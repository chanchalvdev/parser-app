from __future__ import annotations

import logging
from typing import Any, Iterable, Mapping

from app.db.repositories import WorkerRepository


class DatabaseLoadError(RuntimeError):
    """Raised when a parsed record load fails."""


class ParsedRecordLoader:
    """Batch-load parsed records into Postgres."""

    def __init__(
        self,
        repository: WorkerRepository,
        *,
        tenant_id: str,
        job_id: str,
        file_id: str,
        batch_size: int = 1000,
        parser_name: str | None = None,
    ):
        if batch_size <= 0:
            raise ValueError("batch_size must be a positive integer")

        self._repository = repository
        self._tenant_id = tenant_id
        self._job_id = job_id
        self._file_id = file_id
        self._parser_name = parser_name or "unknown"
        self._batch_size = batch_size
        self._buffer: list[dict[str, Any]] = []
        self._inserted_records = 0
        self._logger = logging.getLogger("worker.loaders.postgres")

    @property
    def batch_size(self) -> int:
        return self._batch_size

    @property
    def inserted_count(self) -> int:
        return self._inserted_records

    def add_record(self, record: Mapping[str, Any]) -> int:
        self._buffer.append(dict(record))
        if len(self._buffer) >= self._batch_size:
            return self.flush()
        return 0

    def add_records(self, records: Iterable[Mapping[str, Any]]) -> int:
        inserted = 0
        for record in records:
            inserted += self.add_record(record)
        return inserted

    def flush(self) -> int:
        if not self._buffer:
            return 0

        batch = self._buffer
        try:
            inserted_rows = self._repository.insert_parsed_records_batch(batch)
            inserted = len(inserted_rows)
            self._inserted_records += inserted
            self._buffer = []
            return inserted
        except Exception as exc:
            self._log_load_error(len(batch), exc)
            raise DatabaseLoadError(
                f"failed to insert parsed records batch for file {self._file_id} "
                f"with parser {self._parser_name}"
            ) from exc

    def _log_load_error(self, batch_size: int, error: Exception) -> None:
        error_message = f"{type(error).__name__}: {error}"
        try:
            self._repository.create_job_event(
                tenant_id=self._tenant_id,
                job_id=self._job_id,
                event_type="PARSER_LOAD_ERROR",
                stage="parsing",
                message="failed to persist parsed records",
                metadata={
                    "file_id": self._file_id,
                    "parser": self._parser_name,
                    "batch_size": batch_size,
                    "error_code": "PARSER_LOAD_ERROR",
                    "error_message": error_message,
                },
            )
        except Exception:
            self._logger.exception(
                "failed to record load error event",
                extra={"tenant_id": self._tenant_id, "job_id": self._job_id, "file_id": self._file_id},
            )

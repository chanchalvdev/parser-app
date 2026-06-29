from __future__ import annotations

from collections.abc import Mapping
from datetime import datetime
from typing import Any, Iterable

import logging

from opensearchpy import OpenSearch, helpers

from app.db.repositories import WorkerRepository


PARSED_RECORDS_INDEX = "parsed-records"
FILES_INDEX = "files"


class SearchIndexError(RuntimeError):
    """Raised when OpenSearch indexing fails."""


class SearchIndexLoader:
    """Bulk index parsed records and file metadata into OpenSearch."""

    def __init__(
        self,
        repository: WorkerRepository,
        *,
        opensearch_url: str,
        tenant_id: str,
        job_id: str,
        file_id: str,
        batch_size: int = 1000,
        parsed_records_index: str = PARSED_RECORDS_INDEX,
        files_index: str = FILES_INDEX,
        username: str | None = None,
        password: str | None = None,
    ):
        if batch_size <= 0:
            raise ValueError("batch_size must be a positive integer")
        if not opensearch_url:
            raise ValueError("opensearch_url is required")

        self._repository = repository
        self._tenant_id = tenant_id
        self._job_id = job_id
        self._file_id = file_id
        self._batch_size = batch_size
        self._parsed_records_index = parsed_records_index
        self._files_index = files_index
        self._logger = logging.getLogger("worker.loaders.search")

        self._client = OpenSearch(
            hosts=[opensearch_url],
            basic_auth=(username, password) if username and password else None,
            http_compress=True,
        )

    @property
    def batch_size(self) -> int:
        return self._batch_size

    @property
    def parsed_records_index(self) -> str:
        return self._parsed_records_index

    @property
    def files_index(self) -> str:
        return self._files_index

    def index_parsed_records(self, records: Iterable[Mapping[str, Any]]) -> int:
        records_list = list(records)
        if not records_list:
            return 0
        actions = [self._build_parsed_action(record) for record in records_list]
        return self._bulk_actions(actions, self._parsed_records_index)

    def index_file_record(self, file_record: Mapping[str, Any], source_name: str | None = None) -> int:
        file_id = str(file_record.get("id") or "")
        if not file_id:
            raise SearchIndexError("file_record must include id")
        action = self._build_file_action(file_record, source_name=source_name)
        return self._bulk_actions([action], self._files_index)

    def _bulk_actions(self, actions: list[dict[str, Any]], index_name: str) -> int:
        try:
            count, errors = helpers.bulk(
                self._client,
                actions,
                chunk_size=self._batch_size,
                max_retries=3,
                raise_on_error=False,
                raise_on_exception=False,
                request_timeout=30,
            )
            if errors:
                if index_name == self._parsed_records_index:
                    label = "parsed records"
                elif index_name == self._files_index:
                    label = "file metadata"
                else:
                    label = f"index {index_name}"
                raise SearchIndexError(f"bulk indexing failed for {label}; errors: {errors[:3]}")
            self._logger.info(
                "opensearch bulk indexed",
                extra={
                    "tenant_id": self._tenant_id,
                    "job_id": self._job_id,
                    "file_id": self._file_id,
                    "index_name": index_name,
                    "action_count": len(actions),
                    "indexed_count": count,
                },
            )
            return count
        except SearchIndexError:
            raise
        except Exception as exc:
            self._logger.exception(
                "failed to index records",
                extra={"index_name": index_name, "action_count": len(actions)},
            )
            raise SearchIndexError(f"{type(exc).__name__}: {exc}") from exc

    def _build_parsed_action(self, record: Mapping[str, Any]) -> dict[str, Any]:
        tenant_id = str(record.get("tenant_id") or "")
        file_id = str(record.get("file_id") or "")
        job_id = str(record.get("job_id") or "")
        if not tenant_id or not file_id or not job_id:
            raise SearchIndexError("parsed record missing tenant_id/file_id/job_id")

        record_number = record.get("record_number")
        chunk_number = record.get("chunk_number")
        line_number = record.get("line_number")
        start_line = record.get("start_line")
        end_line = record.get("end_line")
        source_file_name = record.get("source_file_name") or record.get("original_name")
        archive_path = record.get("archive_path")

        doc_id = self._build_parsed_document_id(
            tenant_id=tenant_id,
            file_id=file_id,
            job_id=job_id,
            record_type=str(record.get("record_type") or "unknown"),
            record_number=record_number,
            line_number=line_number,
            chunk_number=chunk_number,
            start_line=start_line,
            end_line=end_line,
        )

        entities = self._normalize_entities(record.get("extracted_entities"))
        return {
            "_op_type": "index",
            "_index": self._parsed_records_index,
            "_id": doc_id,
            "_source": {
                "tenant_id": tenant_id,
                "file_id": file_id,
                "job_id": job_id,
                "source_file_name": source_file_name,
                "archive_path": archive_path,
                "record_type": record.get("record_type"),
                "content": record.get("content_text"),
                "structured_data": record.get("structured_data") or {},
                "entities": {
                    "ip_addresses": entities["ip_addresses"],
                    "emails": entities["emails"],
                    "urls": entities["urls"],
                    "domains": entities["domains"],
                    "hashes": entities["hashes"],
                },
                "event_timestamp": self._serialize_datetime(record.get("event_timestamp")),
                "created_at": self._serialize_datetime(datetime.utcnow().isoformat()),
            },
        }

    def _build_file_action(
        self,
        file_record: Mapping[str, Any],
        *,
        source_name: str | None = None,
    ) -> dict[str, Any]:
        tenant_id = str(file_record.get("tenant_id") or "")
        file_id = str(file_record.get("id") or "")
        if not tenant_id or not file_id:
            raise SearchIndexError("file record missing tenant_id/id")

        return {
            "_op_type": "index",
            "_index": self._files_index,
            "_id": file_id,
            "_source": {
                "tenant_id": tenant_id,
                "file_id": file_id,
                "parent_file_id": self._nullable_uuid(file_record.get("parent_file_id")),
                "original_name": source_name or file_record.get("original_name"),
                "extension": file_record.get("extension"),
                "detected_file_type": file_record.get("detected_file_type"),
                "processing_status": file_record.get("processing_status"),
                "size_bytes": file_record.get("size_bytes"),
                "sha256_hash": file_record.get("sha256_hash"),
                "created_at": self._serialize_datetime(file_record.get("created_at")),
            },
        }

    def _build_parsed_document_id(
        self,
        tenant_id: str,
        file_id: str,
        job_id: str,
        record_type: str,
        record_number: Any,
        line_number: Any,
        chunk_number: Any,
        start_line: Any,
        end_line: Any,
    ) -> str:
        parts = [
            tenant_id,
            file_id,
            job_id,
            record_type,
            str(record_number or ""),
            str(line_number or ""),
            str(chunk_number or ""),
            str(start_line or ""),
            str(end_line or ""),
        ]
        return ":".join(parts)

    @staticmethod
    def _normalize_entities(raw_entities: Any) -> dict[str, list[str]]:
        if not isinstance(raw_entities, Mapping):
            return {
                "ip_addresses": [],
                "emails": [],
                "urls": [],
                "domains": [],
                "hashes": [],
            }

        def as_list(value: Any) -> list[str]:
            if value is None:
                return []
            if isinstance(value, str):
                return [value]
            return [str(v) for v in value if isinstance(v, (str, int, float))]

        ip_addresses = as_list(raw_entities.get("ipv4")) + as_list(raw_entities.get("ip_addresses"))
        emails = as_list(raw_entities.get("emails"))
        urls = as_list(raw_entities.get("urls"))
        domains = as_list(raw_entities.get("domains"))
        hashes = (
            as_list(raw_entities.get("hashes"))
            + as_list(raw_entities.get("md5_like_hashes"))
            + as_list(raw_entities.get("sha256_like_hashes"))
        )

        return {
            "ip_addresses": _dedupe(ip_addresses),
            "emails": _dedupe(emails),
            "urls": _dedupe(urls),
            "domains": _dedupe(domains),
            "hashes": _dedupe(hashes),
        }

    @staticmethod
    def _serialize_datetime(value: Any) -> str | None:
        if isinstance(value, datetime):
            return value.isoformat()
        if isinstance(value, str) and value:
            return value
        return None

    @staticmethod
    def _nullable_uuid(value: Any) -> str | None:
        if value is None:
            return None
        text = str(value)
        return text if text else None


def _dedupe(values: list[str]) -> list[str]:
    seen: set[str] = set()
    result: list[str] = []
    for item in values:
        key = str(item)
        if key in seen:
            continue
        seen.add(key)
        result.append(item)
    return result

from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
import logging
import os
import tempfile
import time
from datetime import datetime, timezone

from app.db.repositories import WorkerRepository
from app.extractors import (
    ArchiveExtractor,
    ArchiveLimits,
    ArchiveLimitsExceededError,
    CorruptArchiveError,
    ExtractedFile,
    PasswordRequiredError,
    UnsupportedArchiveError,
    WrongPasswordError,
)
from app.parsers import ParserRegistry, UnsupportedParserError, create_default_parser_registry
from app.processing.file_detector import FileDetectionResult, detect_file_type
from app.processing.hashing import compute_sha256
from app.queue.redis_consumer import IngestionJobMessage
from app.loaders import (
    DatabaseLoadError,
    ParsedRecordLoader,
    SearchIndexError,
    SearchIndexLoader,
)
from app.secrets import PasswordSecretProvider, LocalArchivePasswordSecretProvider
from app.storage.minio_client import MinioClient


UNKNOWN_UPLOAD_FALLBACK = "unknown-upload"


@dataclass(frozen=True)
class FileArtifact:
    sha256_hash: str
    detected_mime_type: str
    detected_file_type: str
    extension: str
    is_text_like: bool
    is_binary: bool
    is_archive: bool
    parser_hint: str


@dataclass(frozen=True)
class ArchiveLimitsConfig:
    max_archive_depth: int
    max_extracted_files: int
    max_extracted_size_mb: int
    max_expansion_ratio: float


@dataclass(frozen=True)
class ExtractedChild:
    local_path: str
    original_name: str
    relative_path: str
    size_bytes: int
    depth: int
    storage_path: str
    extension: str | None
    detected_mime_type: str | None
    detected_file_type: str
    is_archive: bool


class IngestionOrchestrator:
    def __init__(
        self,
        repository: WorkerRepository,
        storage: MinioClient,
        logger: logging.Logger,
        temp_directory: str | None = None,
        archive_extractor: ArchiveExtractor | None = None,
        parser_registry: ParserRegistry | None = None,
        parser_batch_size: int = 1000,
        max_archive_depth: int = 10,
        max_extracted_files: int = 50000,
        max_extracted_size_mb: int = 1024,
        max_expansion_ratio: float = 20.0,
        password_secret_provider: PasswordSecretProvider | None = None,
        opensearch_url: str = "http://opensearch:9200",
        opensearch_user: str | None = None,
        opensearch_password: str | None = None,
        opensearch_parsed_records_index: str = "parsed-records",
        opensearch_files_index: str = "files",
    ):
        self.repository = repository
        self.storage = storage
        self.logger = logger
        self.temp_directory = temp_directory
        self.archive_extractor = archive_extractor or ArchiveExtractor()
        self.parser_registry = parser_registry or create_default_parser_registry()
        self.password_secret_provider = password_secret_provider or LocalArchivePasswordSecretProvider(repository)
        self.parser_batch_size = parser_batch_size
        self.archive_limits_config = ArchiveLimitsConfig(
            max_archive_depth=max_archive_depth,
            max_extracted_files=max_extracted_files,
            max_extracted_size_mb=max_extracted_size_mb,
            max_expansion_ratio=max_expansion_ratio,
        )
        self.opensearch_url = opensearch_url
        self.opensearch_user = opensearch_user
        self.opensearch_password = opensearch_password
        self.opensearch_parsed_records_index = opensearch_parsed_records_index
        self.opensearch_files_index = opensearch_files_index

    def process(self, message: IngestionJobMessage) -> None:
        self.repository.mark_job_running(job_id=message.job_id, tenant_id=message.tenant_id)
        self.repository.create_job_event(
            tenant_id=message.tenant_id,
            job_id=message.job_id,
            event_type="worker.started",
            stage="running",
            message="worker started processing job",
            metadata={
                "storage_path": message.storage_path,
                "bucket": message.bucket,
                "original_name": message.original_name,
                "depth": message.depth,
            },
        )
        self.logger.info(
            "job started",
            extra={
                "tenant_id": message.tenant_id,
                "job_id": message.job_id,
                "file_id": message.root_file_id,
                "storage_path": message.storage_path,
                "bucket": message.bucket,
                "depth": message.depth,
            },
        )

        limits = self._load_limits(message.tenant_id)
        self.archive_limits_config = limits

        root_file = self.repository.get_file(tenant_id=message.tenant_id, file_id=message.root_file_id)
        upload_id = str(root_file.get("upload_id") or UNKNOWN_UPLOAD_FALLBACK)

        try:
            self._process_file_record(
                tenant_id=message.tenant_id,
                job_id=message.job_id,
                file_id=message.root_file_id,
                storage_path=message.storage_path,
                original_name=message.original_name,
                depth=message.depth,
                upload_id=upload_id,
                is_root=True,
            )
            self.repository.mark_job_completed(job_id=message.job_id, tenant_id=message.tenant_id)
        except PasswordRequiredError as exc:
            raise
        except Exception as exc:
            if isinstance(exc, WrongPasswordError):
                raise

            if isinstance(exc, (UnsupportedArchiveError, CorruptArchiveError, ArchiveLimitsExceededError)):
                self._handle_job_failure(
                    tenant_id=message.tenant_id,
                    file_id=message.root_file_id,
                    job_id=message.job_id,
                    stage="archive_extract",
                    error_code=getattr(exc, "error_code", "WORKER_PROCESSING_ERROR"),
                    message=str(exc),
                )
                raise

            self._handle_job_failure(
                tenant_id=message.tenant_id,
                file_id=message.root_file_id,
                job_id=message.job_id,
                stage="processing",
                error_code="WORKER_PROCESSING_ERROR",
                message=f"{type(exc).__name__}: {exc}",
            )
            raise

        self.logger.info(
            "job completed",
            extra={
                "job_id": message.job_id,
                "tenant_id": message.tenant_id,
                "file_id": message.root_file_id,
                "root_upload_id": upload_id,
            },
        )

    def _process_file_record(
        self,
        tenant_id: str,
        job_id: str,
        file_id: str,
        storage_path: str,
        original_name: str,
        depth: int,
        upload_id: str,
        is_root: bool = False,
    ) -> None:
        with tempfile.TemporaryDirectory(dir=self.temp_directory) as temp_dir:
            local_path = str(Path(temp_dir) / self._safe_filename(original_name))
            self.storage.download_object(storage_path, local_path)

            search_loader = SearchIndexLoader(
                repository=self.repository,
                opensearch_url=self.opensearch_url,
                tenant_id=tenant_id,
                job_id=job_id,
                file_id=file_id,
                batch_size=self.parser_batch_size,
                parsed_records_index=self.opensearch_parsed_records_index,
                files_index=self.opensearch_files_index,
                username=self.opensearch_user,
                password=self.opensearch_password,
            )

            artifact = self._classify_file(local_path, original_name)
            self.repository.update_file_status(tenant_id=tenant_id, file_id=file_id, status="processing")
            self.repository.update_file_hash(
                tenant_id=tenant_id,
                file_id=file_id,
                sha256_hash=artifact.sha256_hash,
            )
            self.repository.update_file_metadata(
                file_id=file_id,
                tenant_id=tenant_id,
                sha256_hash=artifact.sha256_hash,
                detected_mime_type=artifact.detected_mime_type,
                detected_file_type=artifact.detected_file_type,
                is_archive=artifact.is_archive,
                processing_status="processing",
                normalized_name=original_name,
                extension=artifact.extension,
                is_password_protected=False,
            )
            self.repository.create_job_event(
                tenant_id=tenant_id,
                job_id=job_id,
                event_type="worker.detected_file_type",
                stage="detection",
                message="file detection completed",
                metadata={
                    "file_id": file_id,
                    "extension": artifact.extension,
                    "mime_type": artifact.detected_mime_type,
                    "detected_file_type": artifact.detected_file_type,
                    "parser_hint": artifact.parser_hint,
                    "is_text_like": artifact.is_text_like,
                    "is_binary": artifact.is_binary,
                    "is_archive": artifact.is_archive,
                    "depth": depth,
                    "size_bytes": os.path.getsize(local_path),
                },
            )
            self.logger.info(
                "file classified",
                extra={
                    "tenant_id": tenant_id,
                    "job_id": job_id,
                    "file_id": file_id,
                    "extension": artifact.extension,
                    "mime_type": artifact.detected_mime_type,
                    "detected_file_type": artifact.detected_file_type,
                    "is_archive": artifact.is_archive,
                    "size_bytes": os.path.getsize(local_path),
                },
            )

            if not artifact.is_archive:
                parse_success = self._parse_non_archive_file(
                    tenant_id=tenant_id,
                    job_id=job_id,
                    file_id=file_id,
                    local_path=local_path,
                    original_name=original_name,
                    detection_result=FileDetectionResult(
                        extension=artifact.extension,
                        mime_type=artifact.detected_mime_type,
                        detected_file_type=artifact.detected_file_type,
                        is_archive=artifact.is_archive,
                        is_text_like=artifact.is_text_like,
                        is_binary=artifact.is_binary,
                        parser_hint=artifact.parser_hint,
                    ),
                    is_root=is_root,
                    size_bytes=os.path.getsize(local_path),
                    search_loader=search_loader,
                )

                if parse_success:
                    self._finalize_file(
                        tenant_id=tenant_id,
                        job_id=job_id,
                        file_id=file_id,
                        completed_stage="completed",
                        search_loader=search_loader,
                    )
                return

            if depth >= self.archive_limits_config.max_archive_depth:
                self.repository.create_job_event(
                    tenant_id=tenant_id,
                    job_id=job_id,
                    event_type="worker.archive_depth_limit",
                    stage="archive",
                    message="archive depth limit reached; skipping child extraction",
                    metadata={
                        "file_id": file_id,
                        "depth": depth,
                        "max_archive_depth": self.archive_limits_config.max_archive_depth,
                    },
                )
                self.logger.info(
                    "archive depth limit reached for file",
                    extra={
                        "tenant_id": tenant_id,
                        "job_id": job_id,
                        "file_id": file_id,
                        "depth": depth,
                        "max_archive_depth": self.archive_limits_config.max_archive_depth,
                    },
                )
                self._finalize_file(
                    tenant_id=tenant_id,
                    job_id=job_id,
                    file_id=file_id,
                    completed_stage="completed",
                    search_loader=search_loader,
                    details={"depth_limit_hit": True},
                )
                return

            detection = detect_file_type(local_path, original_name)
            extracted_children = self._extract_children(
                tenant_id=tenant_id,
                job_id=job_id,
                file_id=file_id,
                local_file_path=local_path,
                detection=detection,
                depth=depth,
                upload_id=upload_id,
            )
            self.logger.info(
                "archive extraction complete",
                extra={
                    "tenant_id": tenant_id,
                    "job_id": job_id,
                    "file_id": file_id,
                    "extracted_count": len(extracted_children),
                },
            )

            for child in extracted_children:
                child_file = self.repository.create_child_file(
                    tenant_id=tenant_id,
                    parent_file_id=file_id,
                    upload_id=upload_id,
                    original_name=child.original_name,
                    storage_path=child.storage_path,
                    size_bytes=child.size_bytes,
                    depth=child.depth,
                    normalized_name=child.relative_path,
                    extension=child.extension,
                    detected_mime_type=child.detected_mime_type,
                    detected_file_type=child.detected_file_type,
                    is_archive=child.is_archive,
                    is_password_protected=False,
                    processing_status="queued",
                )
                child_file_id = child_file["id"]
                try:
                    self._process_file_record(
                        tenant_id=tenant_id,
                        job_id=job_id,
                        file_id=child_file_id,
                        storage_path=child.storage_path,
                        original_name=child.original_name,
                        depth=child.depth,
                        upload_id=upload_id,
                        is_root=False,
                    )
                except (PasswordRequiredError, WrongPasswordError):
                    raise
                except Exception as exc:
                    if isinstance(exc, RuntimeError) and str(exc).startswith("PARSE_FAILED:"):
                        self.logger.warning(
                            "child parse failed; continuing job orchestration",
                            extra={"job_id": job_id, "tenant_id": tenant_id, "file_id": child_file_id},
                        )
                        continue
                    self._handle_job_failure(
                        tenant_id=tenant_id,
                        file_id=child_file_id,
                        job_id=job_id,
                        stage="archive_child_processing",
                        error_code=getattr(exc, "error_code", "WORKER_PROCESSING_ERROR"),
                        message=f"{type(exc).__name__}: {exc}",
                    )
                    raise

            self._finalize_file(
                tenant_id=tenant_id,
                job_id=job_id,
                file_id=file_id,
                completed_stage="completed",
                search_loader=search_loader,
            )

    def _extract_children(
        self,
        tenant_id: str,
        job_id: str,
        file_id: str,
        local_file_path: str,
        detection: FileDetectionResult,
        depth: int,
        upload_id: str,
    ) -> list[ExtractedChild]:
        self.logger.info(
            "archive extraction started",
            extra={
                "tenant_id": tenant_id,
                "job_id": job_id,
                "file_id": file_id,
                "source_path": local_file_path,
                "detected_file_type": detection.detected_file_type,
            },
        )
        if not self.archive_extractor.can_extract(local_file_path, detection):
            raise UnsupportedArchiveError("no archive extractor available for detected format")

        limits = ArchiveLimits(
            max_extracted_files=self.archive_limits_config.max_extracted_files,
            max_extracted_size_mb=self.archive_limits_config.max_extracted_size_mb,
            max_expansion_ratio=self.archive_limits_config.max_expansion_ratio,
        )
        extracted_files: list[ExtractedFile] = []
        extracted_children: list[ExtractedChild] = []
        with tempfile.TemporaryDirectory(dir=self.temp_directory) as temp_dir:
            password_ref = self.password_secret_provider.get_archive_password(
                tenant_id=tenant_id,
                file_id=file_id,
            )
            archive_password = self.password_secret_provider.resolve_password(
                tenant_id=tenant_id,
                file_id=file_id,
            )

            try:
                extracted_files = self.archive_extractor.extract(
                    file_path=local_file_path,
                    detection_result=detection,
                    destination_dir=temp_dir,
                    password=archive_password,
                    limits=limits,
                )
            except PasswordRequiredError as exc:
                self._handle_password_required(
                    tenant_id=tenant_id,
                    job_id=job_id,
                    file_id=file_id,
                    error_code=exc.error_code,
                    message=str(exc),
                )
                raise
            except WrongPasswordError as exc:
                if password_ref is not None:
                    self.password_secret_provider.record_password_attempt(
                        tenant_id=tenant_id,
                        file_id=file_id,
                        password_ref_hash=password_ref.password_ref_hash,
                        is_valid=False,
                        increment_attempt=True,
                    )
                self._handle_wrong_password(
                    tenant_id=tenant_id,
                    job_id=job_id,
                    file_id=file_id,
                    error_code=exc.error_code,
                    message=str(exc),
                )
                raise

            if not extracted_files:
                self.repository.create_job_event(
                    tenant_id=tenant_id,
                    job_id=job_id,
                    event_type="worker.archive_empty",
                    stage="archive",
                    message="archive contained no extractable files",
                    metadata={"parent_file_id": file_id},
                )
                self.logger.info(
                    "archive extraction produced no files",
                    extra={
                        "tenant_id": tenant_id,
                        "job_id": job_id,
                        "file_id": file_id,
                        "detected_file_type": detection.detected_file_type,
                    },
                )
                return []

            if password_ref is not None:
                self.password_secret_provider.record_password_attempt(
                    tenant_id=tenant_id,
                    file_id=file_id,
                    password_ref_hash=password_ref.password_ref_hash,
                    is_valid=True,
                    increment_attempt=False,
                )

            for extracted in extracted_files:
                child_storage_path = self._build_extracted_storage_key(
                    tenant_id=tenant_id,
                    upload_id=upload_id,
                    parent_file_id=file_id,
                    relative_path=extracted.relative_path,
                )
                self.storage.upload_object(extracted.local_path, child_storage_path)
                child_detection = detect_file_type(extracted.local_path, extracted.original_name)
                child_depth = depth + extracted.depth

                child = ExtractedChild(
                    local_path=extracted.local_path,
                    original_name=extracted.original_name,
                    relative_path=extracted.relative_path,
                    size_bytes=extracted.size_bytes,
                    depth=child_depth,
                    storage_path=child_storage_path,
                    extension=child_detection.extension or None,
                    detected_mime_type=child_detection.mime_type,
                    detected_file_type=child_detection.detected_file_type,
                    is_archive=child_detection.is_archive,
                )
                self.repository.create_job_event(
                    tenant_id=tenant_id,
                    job_id=job_id,
                    event_type="worker.extracted_file",
                    stage="archive",
                    message="archive child extracted and uploaded",
                    metadata={
                        "parent_file_id": file_id,
                        "child_file_id_hint": child.storage_path,
                        "child_depth": child_depth,
                        "mime_type": child.detected_mime_type,
                        "detected_file_type": child.detected_file_type,
                        "storage_path": child_storage_path,
                        "is_archive": child.is_archive,
                    },
                )
                extracted_children.append(child)
            self.logger.info(
                "archive child prepared",
                extra={
                    "tenant_id": tenant_id,
                    "job_id": job_id,
                    "file_id": file_id,
                    "parent_file_id": file_id,
                    "child_file_count": len(extracted_children),
                    "last_child_name": child.original_name,
                    "last_child_size": child.size_bytes,
                },
            )

            self.logger.info(
                "archive extraction completed",
                extra={
                    "tenant_id": tenant_id,
                    "job_id": job_id,
                    "file_id": file_id,
                    "extracted_count": len(extracted_children),
                    "total_extracted_size_bytes": sum(child.size_bytes for child in extracted_children),
                    "detected_file_type": detection.detected_file_type,
                },
            )

        return extracted_children

    def _finalize_file(
        self,
        tenant_id: str,
        job_id: str,
        file_id: str,
        completed_stage: str,
        search_loader: SearchIndexLoader,
        details: dict[str, object] | None = None,
    ) -> None:
        self.repository.update_file_status(tenant_id=tenant_id, file_id=file_id, status="completed")
        self.repository.create_job_event(
            tenant_id=tenant_id,
            job_id=job_id,
            event_type="worker.completed",
            stage=completed_stage,
            message="file processing completed",
            metadata={"file_id": file_id, **(details or {})},
        )

        try:
            file_record = self.repository.get_file(tenant_id=tenant_id, file_id=file_id)
            search_loader.index_file_record(file_record, source_name=str(file_record.get("original_name") or ""))
            self._handle_search_index_completed(
                tenant_id=tenant_id,
                file_id=file_id,
                job_id=job_id,
                index_name=search_loader.files_index,
                attempts=0,
                indexed_count=1,
                metadata={"file_id": file_id, "index_target": "files"},
            )
        except SearchIndexError as exc:
            self._handle_search_index_failure(
                tenant_id=tenant_id,
                file_id=file_id,
                job_id=job_id,
                index_name=search_loader.files_index,
                attempts=1,
                error=f"{type(exc).__name__}: {exc}",
            )

    def _parse_non_archive_file(
        self,
        tenant_id: str,
        job_id: str,
        file_id: str,
        local_path: str,
        original_name: str,
        detection_result: FileDetectionResult,
        is_root: bool,
        size_bytes: int,
        search_loader: SearchIndexLoader,
    ) -> bool:
        self.repository.create_job_event(
            tenant_id=tenant_id,
            job_id=job_id,
            event_type="PARSING_STARTED",
            stage="parsing",
            message="parser started",
            metadata={
                "file_id": file_id,
                "file": original_name,
                "detected_file_type": detection_result.detected_file_type,
            },
        )

        parser_metadata: dict[str, object] = {
            "tenant_id": tenant_id,
            "file_id": file_id,
            "job_id": job_id,
            "size_bytes": size_bytes,
            "original_name": original_name,
        }

        try:
            parser = self.parser_registry.select_parser(parser_metadata, detection_result)
        except UnsupportedParserError as exc:
            self._handle_parser_error(
                tenant_id=tenant_id,
                file_id=file_id,
                job_id=job_id,
                error_code="NO_PARSER_FOUND",
                message=f"{type(exc).__name__}: {exc}",
                detection_result=detection_result,
                is_root=is_root,
                parse_error=exc,
            )
            if is_root:
                raise RuntimeError(f"PARSE_FAILED: {type(exc).__name__}: {exc}") from exc
            return False

        records_parsed = 0
        records_inserted = 0
        search_records_pending: list[dict[str, Any]] = []
        search_indexed_count = 0
        search_attempts = 0
        search_failed = False
        search_last_error: str | None = None
        parse_started_at = time.perf_counter()
        loader = ParsedRecordLoader(
            repository=self.repository,
            tenant_id=tenant_id,
            job_id=job_id,
            file_id=file_id,
            batch_size=self.parser_batch_size,
            parser_name=parser.name,
        )

        self.logger.info(
            "parser started",
            extra={
                "tenant_id": tenant_id,
                "job_id": job_id,
                "file_id": file_id,
                "parser": parser.name,
                "parser_batch_size": self.parser_batch_size,
                "size_bytes": size_bytes,
            },
        )

        try:
            for parsed_record in parser.parse(
                file_path=local_path,
                file_metadata=parser_metadata,
                detection_result=detection_result,
            ):
                normalized = parser.normalize(parsed_record)
                if normalized.record_type == "parser_error":
                    error_context = normalized.structured_data if isinstance(normalized.structured_data, dict) else {}
                    error_code = str(error_context.get("error_code", "PARSER_ERROR"))
                    error_message = normalized.content_text or "parser error"
                    self.repository.insert_parser_error(
                        tenant_id=tenant_id,
                        job_id=job_id,
                        file_id=file_id,
                        upload_id=None,
                        error_code=error_code,
                        error_message=error_message,
                        error_context=error_context,
                        is_retryable=False,
                    )
                    continue

                records_parsed += 1
                normalized_payload = {
                    "tenant_id": normalized.tenant_id,
                    "file_id": normalized.file_id,
                    "job_id": normalized.job_id,
                    "source_file_name": original_name,
                    "archive_path": None,
                    "record_type": normalized.record_type,
                    "record_number": normalized.record_number,
                    "line_number": normalized.line_number,
                    "chunk_number": normalized.chunk_number,
                    "start_line": normalized.start_line,
                    "end_line": normalized.end_line,
                    "content_text": normalized.content_text,
                    "structured_data": normalized.structured_data,
                    "extracted_entities": normalized.extracted_entities,
                    "event_timestamp": normalized.event_timestamp,
                }
                search_records_pending.append(dict(normalized_payload))

                records_inserted += loader.add_record(
                    normalized_payload
                )

                if records_inserted > 0 and records_inserted % self.parser_batch_size == 0:
                    if search_records_pending:
                        try:
                            search_indexed_count += search_loader.index_parsed_records(search_records_pending)
                            self.logger.info(
                                "search indexing flushed from parser",
                                extra={
                                    "tenant_id": tenant_id,
                                    "job_id": job_id,
                                    "file_id": file_id,
                                    "parser": parser.name,
                                    "search_batch_size": len(search_records_pending),
                                    "indexed_count": search_indexed_count,
                                },
                            )
                            search_records_pending = []
                            search_last_error = None
                        except SearchIndexError as exc:
                            search_failed = True
                            search_attempts += 1
                            search_last_error = str(exc)
                            self._handle_search_index_failure(
                                tenant_id=tenant_id,
                                file_id=file_id,
                                job_id=job_id,
                                index_name=search_loader.parsed_records_index,
                                attempts=search_attempts,
                                error=search_last_error,
                            )
                    self.repository.create_job_event(
                        tenant_id=tenant_id,
                        job_id=job_id,
                        event_type="PARSING_PROGRESS",
                        stage="parsing",
                        message="parser progress",
                        metadata={
                            "file_id": file_id,
                            "parser": parser.name,
                            "records_parsed": records_parsed,
                            "records_inserted": records_inserted,
                            "size_bytes": size_bytes,
                            "parser_duration_ms": int((time.perf_counter() - parse_started_at) * 1000),
                            "parsed_count": records_parsed,
                            "inserted_count": records_inserted,
                        },
                    )

            records_inserted += loader.flush()
            if search_records_pending:
                try:
                    search_indexed_count += search_loader.index_parsed_records(search_records_pending)
                    self.logger.info(
                        "search indexing flushed from parser tail",
                        extra={
                            "tenant_id": tenant_id,
                            "job_id": job_id,
                            "file_id": file_id,
                            "parser": parser.name,
                            "search_batch_size": len(search_records_pending),
                            "indexed_count": search_indexed_count,
                        },
                    )
                    search_records_pending = []
                    search_last_error = None
                except SearchIndexError as exc:
                    search_failed = True
                    search_attempts += 1
                    search_last_error = str(exc)
                    self._handle_search_index_failure(
                        tenant_id=tenant_id,
                        file_id=file_id,
                        job_id=job_id,
                        index_name=search_loader.parsed_records_index,
                        attempts=search_attempts,
                        error=search_last_error,
                    )

            if not search_records_pending:
                self._handle_search_index_completed(
                    tenant_id=tenant_id,
                    file_id=file_id,
                    job_id=job_id,
                    index_name=search_loader.parsed_records_index,
                    attempts=search_attempts,
                    indexed_count=search_indexed_count,
                    metadata={
                        "file_id": file_id,
                        "parser": parser.name,
                        "records_parsed": records_parsed,
                        "records_inserted": records_inserted,
                    },
                )
            elif search_failed:
                self._handle_search_index_failure(
                    tenant_id=tenant_id,
                    file_id=file_id,
                    job_id=job_id,
                    index_name=search_loader.parsed_records_index,
                    attempts=search_attempts,
                    error=search_last_error or "unindexed records remain",
                )
            parser_duration_ms = int((time.perf_counter() - parse_started_at) * 1000)
            self.logger.info(
                "parser completed",
                extra={
                    "tenant_id": tenant_id,
                    "job_id": job_id,
                    "file_id": file_id,
                    "parser": parser.name,
                    "records_parsed": records_parsed,
                    "records_inserted": records_inserted,
                    "parser_duration_ms": parser_duration_ms,
                },
            )

            self.repository.create_job_event(
                tenant_id=tenant_id,
                job_id=job_id,
                event_type="PARSING_COMPLETED",
                stage="parsing",
                message="parser completed",
                metadata={
                    "file_id": file_id,
                    "parser": parser.name,
                    "records_parsed": records_parsed,
                    "records_inserted": records_inserted,
                    "parser_duration_ms": parser_duration_ms,
                    "parsed_count": records_parsed,
                    "inserted_count": records_inserted,
                },
            )
            return True
        except SearchIndexError:
            parser_duration_ms = int((time.perf_counter() - parse_started_at) * 1000)
            self.logger.warning(
                "search indexing failed while parsing",
                extra={
                    "tenant_id": tenant_id,
                    "job_id": job_id,
                    "file_id": file_id,
                    "parser": parser.name,
                    "parser_duration_ms": parser_duration_ms,
                },
            )
            return True
        except Exception as exc:
            parser_duration_ms = int((time.perf_counter() - parse_started_at) * 1000)
            self.logger.warning(
                "parser failed",
                extra={
                    "tenant_id": tenant_id,
                    "job_id": job_id,
                    "file_id": file_id,
                    "parser": parser.name,
                    "records_parsed": records_parsed,
                    "records_inserted": records_inserted,
                    "parser_duration_ms": parser_duration_ms,
                    "error_type": type(exc).__name__,
                },
            )
            if isinstance(exc, DatabaseLoadError):
                self._handle_parser_error(
                    tenant_id=tenant_id,
                    file_id=file_id,
                    job_id=job_id,
                    error_code="PARSER_LOAD_ERROR",
                    message=f"{type(exc).__name__}: {exc}",
                    detection_result=detection_result,
                    is_root=is_root,
                    parse_error=exc,
                )
                if is_root:
                    raise RuntimeError(f"PARSE_FAILED: {type(exc).__name__}: {exc}") from exc
                return False

            self._handle_parser_error(
                tenant_id=tenant_id,
                file_id=file_id,
                job_id=job_id,
                error_code="PARSER_ERROR",
                message=f"{type(exc).__name__}: {exc}",
                detection_result=detection_result,
                is_root=is_root,
                parse_error=exc,
            )
            if is_root:
                raise RuntimeError(f"PARSE_FAILED: {type(exc).__name__}: {exc}") from exc
            return False

    def _handle_search_index_completed(
        self,
        tenant_id: str,
        file_id: str,
        job_id: str,
        index_name: str,
        attempts: int,
        indexed_count: int,
        metadata: dict[str, Any],
    ) -> None:
        try:
            self.repository.update_search_index_status(
                tenant_id=tenant_id,
                file_id=file_id,
                job_id=job_id,
                index_name=index_name,
                status="COMPLETED",
                attempts=attempts,
                last_error=None,
                last_indexed_at=datetime.now(timezone.utc),
            )
        except Exception:
            self.logger.warning(
                "failed to update search_index_status to COMPLETED",
                extra={
                    "tenant_id": tenant_id,
                    "job_id": job_id,
                    "file_id": file_id,
                    "index_name": index_name,
                },
            )
        self.logger.info(
            "search indexing completed",
            extra={
                "tenant_id": tenant_id,
                "job_id": job_id,
                "file_id": file_id,
                "index_name": index_name,
                "attempts": attempts,
                "indexed_count": indexed_count,
                "metadata": metadata,
            },
        )
        try:
            self.repository.create_job_event(
                tenant_id=tenant_id,
                job_id=job_id,
                event_type="INDEXING_COMPLETED",
                stage="indexing",
                message="open search indexing completed",
                metadata={"file_id": file_id, "index_name": index_name, "attempts": attempts, "indexed_count": indexed_count, **metadata},
            )
        except Exception:
            self.logger.warning(
                "failed to create INDEXING_COMPLETED event",
                extra={"tenant_id": tenant_id, "job_id": job_id, "file_id": file_id, "index_name": index_name},
            )

    def _handle_search_index_failure(
        self,
        tenant_id: str,
        file_id: str,
        job_id: str,
        index_name: str,
        attempts: int,
        error: str,
    ) -> None:
        try:
            self.repository.update_search_index_status(
                tenant_id=tenant_id,
                file_id=file_id,
                job_id=job_id,
                index_name=index_name,
                status="FAILED",
                attempts=attempts,
                last_error=error,
                last_indexed_at=datetime.now(timezone.utc),
            )
        except Exception:
            self.logger.warning(
                "failed to update search_index_status to FAILED",
                extra={
                    "tenant_id": tenant_id,
                    "job_id": job_id,
                    "file_id": file_id,
                    "index_name": index_name,
                },
            )
        self.logger.warning(
            "search indexing failed",
            extra={
                "tenant_id": tenant_id,
                "job_id": job_id,
                "file_id": file_id,
                "index_name": index_name,
                "attempts": attempts,
                "error": error,
            },
        )
        try:
            self.repository.create_job_event(
                tenant_id=tenant_id,
                job_id=job_id,
                event_type="INDEXING_FAILED",
                stage="indexing",
                message="open search indexing failed",
                metadata={
                    "file_id": file_id,
                    "index_name": index_name,
                    "error": error,
                    "attempts": attempts,
                },
            )
        except Exception:
            self.logger.warning(
                "failed to create INDEXING_FAILED event",
                extra={"tenant_id": tenant_id, "job_id": job_id, "file_id": file_id, "index_name": index_name},
            )

    def _classify_file(self, local_path: str, original_name: str) -> FileArtifact:
        detection = detect_file_type(local_path, original_name)
        sha256_hash = compute_sha256(local_path)

        return FileArtifact(
            sha256_hash=sha256_hash,
            detected_mime_type=detection.mime_type,
            detected_file_type=detection.detected_file_type,
            extension=detection.extension,
            is_text_like=detection.is_text_like,
            is_binary=detection.is_binary,
            is_archive=detection.is_archive,
            parser_hint=detection.parser_hint,
        )

    def _build_extracted_storage_key(
        self,
        tenant_id: str,
        upload_id: str | None,
        parent_file_id: str,
        relative_path: str,
    ) -> str:
        safe_upload_id = upload_id or UNKNOWN_UPLOAD_FALLBACK
        safe_parent = self._safe_filename(parent_file_id or UNKNOWN_UPLOAD_FALLBACK)
        safe_relative_path = relative_path.replace("\\", "/")
        return f"extracted/{tenant_id}/{safe_upload_id}/{safe_parent}/{safe_relative_path}"

    def _safe_filename(self, value: str) -> str:
        return os.path.basename(value).replace("/", "_").replace("\\", "_")[:255] or "upload.bin"

    def _load_limits(self, tenant_id: str) -> ArchiveLimitsConfig:
        max_archive_depth = self._read_int_setting(
            tenant_id=tenant_id,
            setting_key="max_archive_depth",
            default=self.archive_limits_config.max_archive_depth,
        )
        max_extracted_files = self._read_int_setting(
            tenant_id=tenant_id,
            setting_key="max_extracted_files",
            default=self.archive_limits_config.max_extracted_files,
        )
        max_extracted_size_mb = self._read_int_setting(
            tenant_id=tenant_id,
            setting_key="max_extracted_size_mb",
            default=self.archive_limits_config.max_extracted_size_mb,
        )
        max_expansion_ratio = self._read_float_setting(
            tenant_id=tenant_id,
            setting_key="max_expansion_ratio",
            default=self.archive_limits_config.max_expansion_ratio,
        )

        return ArchiveLimitsConfig(
            max_archive_depth=max_archive_depth,
            max_extracted_files=max_extracted_files,
            max_extracted_size_mb=max_extracted_size_mb,
            max_expansion_ratio=max_expansion_ratio,
        )

    def _read_int_setting(self, tenant_id: str, setting_key: str, default: int) -> int:
        value = self.repository.get_setting(tenant_id=tenant_id, setting_key=setting_key, default=default)
        if value is None:
            return default
        if isinstance(value, bool):
            return default
        try:
            return int(value)
        except (TypeError, ValueError):
            return default

    def _read_float_setting(self, tenant_id: str, setting_key: str, default: float) -> float:
        value = self.repository.get_setting(tenant_id=tenant_id, setting_key=setting_key, default=default)
        if value is None:
            return default
        try:
            return float(value)
        except (TypeError, ValueError):
            return default

    def _handle_job_failure(
        self,
        tenant_id: str,
        file_id: str,
        job_id: str,
        stage: str,
        error_code: str,
        message: str,
    ) -> None:
        self.repository.update_file_status(tenant_id=tenant_id, file_id=file_id, status="failed")
        self.repository.create_job_event(
            tenant_id=tenant_id,
            job_id=job_id,
            event_type="worker.processing_failed",
            stage=stage,
            message="file processing failed",
            metadata={"file_id": file_id, "error_code": error_code, "message": message},
        )
        self.repository.mark_job_failed(
            job_id=job_id,
            tenant_id=tenant_id,
            stage=stage,
            error_code=error_code,
            message=message,
        )

    def _handle_parser_error(
        self,
        tenant_id: str,
        file_id: str,
        job_id: str,
        error_code: str,
        message: str,
        detection_result: FileDetectionResult,
        is_root: bool,
        parse_error: Exception,
    ) -> None:
        del parse_error
        self.repository.update_file_status(tenant_id=tenant_id, file_id=file_id, status="parse_failed")
        self.repository.insert_parser_error(
            tenant_id=tenant_id,
            job_id=job_id,
            file_id=file_id,
            upload_id=None,
            error_code=error_code,
            error_message=message,
            error_context={
                "file": file_id,
                "tenant_id": tenant_id,
                "detected_file_type": detection_result.detected_file_type,
                "is_root": is_root,
            },
            is_retryable=False,
        )
        self.repository.create_job_event(
            tenant_id=tenant_id,
            job_id=job_id,
            event_type="PARSING_ERROR",
            stage="parsing",
            message="parser failed for file",
            metadata={
                "file_id": file_id,
                "is_root": is_root,
                "error_code": error_code,
                "error_message": message,
            },
        )
        if is_root:
            self.repository.create_job_event(
                tenant_id=tenant_id,
                job_id=job_id,
                event_type="worker.parsing_failed",
                stage="parsing",
                message="root file parsing failed; aborting job",
                metadata={"file_id": file_id, "error_code": error_code},
            )

    def _handle_password_required(
        self,
        tenant_id: str,
        job_id: str,
        file_id: str,
        error_code: str,
        message: str,
    ) -> None:
        file_record = self.repository.get_file(tenant_id=tenant_id, file_id=file_id)
        self.repository.update_file_metadata(
            file_id=file_id,
            tenant_id=tenant_id,
            sha256_hash=file_record.get("sha256_hash") or "",
            detected_mime_type=file_record.get("detected_mime_type") or "unknown",
            detected_file_type=file_record.get("detected_file_type") or "unknown",
            is_archive=True,
            processing_status="password_required",
            is_password_protected=True,
        )
        self.repository.update_file_status(tenant_id=tenant_id, file_id=file_id, status="password_required")
        self.repository.create_job_event(
            tenant_id=tenant_id,
            job_id=job_id,
            event_type="worker.password_required",
            stage="password_required",
            message="password required for archive content",
            metadata={"file_id": file_id, "error_code": error_code, "message": message},
        )
        self.repository.update_job_status(
            tenant_id=tenant_id,
            job_id=job_id,
            status="PASSWORD_REQUIRED",
            current_stage="password_required",
            progress_percent=0.0,
            error_code=error_code,
            error_message=message,
        )

    def _handle_wrong_password(
        self,
        tenant_id: str,
        job_id: str,
        file_id: str,
        error_code: str,
        message: str,
    ) -> None:
        file_record = self.repository.get_file(tenant_id=tenant_id, file_id=file_id)
        self.repository.update_file_metadata(
            file_id=file_id,
            tenant_id=tenant_id,
            sha256_hash=file_record.get("sha256_hash") or "",
            detected_mime_type=file_record.get("detected_mime_type") or "unknown",
            detected_file_type=file_record.get("detected_file_type") or "unknown",
            is_archive=True,
            processing_status="wrong_password",
            is_password_protected=True,
        )
        self.repository.update_file_status(tenant_id=tenant_id, file_id=file_id, status="wrong_password")
        self.repository.create_job_event(
            tenant_id=tenant_id,
            job_id=job_id,
            event_type="worker.wrong_password",
            stage="wrong_password",
            message="archive password is incorrect",
            metadata={"file_id": file_id, "error_code": error_code, "message": message},
        )
        self.repository.update_job_status(
            tenant_id=tenant_id,
            job_id=job_id,
            status="WRONG_PASSWORD",
            current_stage="wrong_password",
            progress_percent=0.0,
            error_code=error_code,
            error_message=message,
        )

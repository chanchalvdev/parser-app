from __future__ import annotations

import logging
import time
import os

import redis

from app.extractors import PasswordRequiredError, WrongPasswordError
from app.config import load_config
from app.db.connection import DatabaseConnection
from app.db.repositories import WorkerRepository
from app.logging_config import setup_logging
from app.processing.orchestrator import IngestionOrchestrator
from app.secrets import LocalArchivePasswordSecretProvider
from app.queue.redis_consumer import IngestionJobMessage, RedisIngestionQueueConsumer
from app.storage.minio_client import MinioClient


def _split_redis_addr(addr: str) -> tuple[str, int]:
    host = addr
    port = 6379
    if ":" in addr:
        host, port_text = addr.rsplit(":", 1)
        port = int(port_text)
    return host, port


def _error_payload(message: IngestionJobMessage | None, repository: WorkerRepository, logger: logging.Logger, err: Exception) -> None:
    if message is None:
        return
    try:
        if isinstance(err, (PasswordRequiredError, WrongPasswordError)):
            return
        # If the orchestrator already recorded a precise failure for this job,
        # do not overwrite its error_code/stage — this is only a fallback for
        # unexpected crashes that escaped the orchestrator's own handling.
        if getattr(err, "_job_failure_recorded", False):
            return
        repository.mark_job_failed(
            job_id=message.job_id,
            tenant_id=message.tenant_id,
            stage="failed",
            error_code=getattr(err, "error_code", "WORKER_PROCESSING_ERROR"),
            message=str(err),
        )
    except Exception:
        logger.exception(
            "failed to persist worker processing error",
            extra={
                "job_id": message.job_id,
                "tenant_id": message.tenant_id,
                "root_file_id": message.root_file_id,
            },
        )


def main() -> None:
    config = load_config()
    setup_logging(config.log_level)
    logger = logging.getLogger("worker")

    if config.temp_directory:
        os.makedirs(config.temp_directory, exist_ok=True)

    redis_host, redis_port = _split_redis_addr(config.redis_addr)
    redis_client = redis.Redis(
        host=redis_host,
        port=redis_port,
        password=config.redis_password or None,
        db=config.redis_db,
        decode_responses=False,
    )
    redis_client.ping()

    repository = WorkerRepository(DatabaseConnection(config.database_url))
    minio_client = MinioClient(
        endpoint=config.minio_endpoint,
        access_key=config.minio_access_key,
        secret_key=config.minio_secret_key,
        bucket=config.minio_bucket,
        secure=config.minio_use_ssl,
    )
    queue_consumer = RedisIngestionQueueConsumer(
        redis_client=redis_client,
        queue_name=config.queue_name,
        block_timeout_seconds=config.redis_block_timeout,
    )
    orchestrator = IngestionOrchestrator(
        repository=repository,
        storage=minio_client,
        logger=logger,
        temp_directory=config.temp_directory,
        password_secret_provider=LocalArchivePasswordSecretProvider(repository=repository),
        max_archive_depth=config.max_archive_depth,
        max_extracted_files=config.max_extracted_files,
        max_extracted_size_mb=config.max_extracted_size_mb,
        max_expansion_ratio=config.max_expansion_ratio,
        opensearch_url=config.opensearch_url,
        opensearch_user=config.opensearch_user or None,
        opensearch_password=config.opensearch_password or None,
        opensearch_parsed_records_index=config.opensearch_parsed_records_index,
        opensearch_files_index=config.opensearch_files_index,
    )

    logger.info("worker loop started", extra={"queue": config.queue_name, "redis": config.redis_addr})

    while True:
        message: IngestionJobMessage | None = None
        try:
            message = queue_consumer.pop()
            if message is None:
                time.sleep(config.poll_interval_seconds)
                continue

            payload = message.to_log_payload()
            logger.info(
                "received job message",
                extra={
                    "job_id": message.job_id,
                    "tenant_id": message.tenant_id,
                    "root_file_id": message.root_file_id,
                    **payload,
                },
            )
            orchestrator.process(message)
        except Exception as exc:
            if message is not None:
                _error_payload(message, repository, logger, exc)
            if getattr(exc, "_job_failure_recorded", False):
                # Already logged with full context + traceback by the orchestrator's
                # "job failed" record; avoid a duplicate, misleading traceback here.
                logger.info(
                    "job failed (already recorded); skipping duplicate error handling",
                    extra={
                        **(message.to_log_payload() if message is not None else {}),
                        "error_code": getattr(exc, "error_code", "WORKER_PROCESSING_ERROR"),
                    },
                )
            else:
                logger.exception(
                    "error while processing queue message",
                    extra=message.to_log_payload() if message is not None else {},
                )
            time.sleep(config.poll_interval_seconds)


if __name__ == "__main__":
    main()

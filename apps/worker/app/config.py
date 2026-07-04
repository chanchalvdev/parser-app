from __future__ import annotations

from dataclasses import dataclass
from typing import Optional
import os


def _env(key: str, default: str = "") -> str:
    value = os.getenv(key)
    return value if value is not None and value != "" else default


def _env_int(key: str, default: int) -> int:
    value = os.getenv(key)
    if value is None or value == "":
        return default
    try:
        return int(value)
    except ValueError:
        return default


def _env_float(key: str, default: float) -> float:
    value = os.getenv(key)
    if value is None or value == "":
        return default
    try:
        return float(value)
    except (TypeError, ValueError):
        return default


def _env_bool(key: str, default: bool) -> bool:
    value = os.getenv(key)
    if value is None or value == "":
        return default
    return value.strip().lower() in {"1", "true", "t", "yes", "y"}


def _normalize_host(value: str) -> str:
    return value.replace("http://", "").replace("https://", "")


@dataclass(frozen=True)
class WorkerConfig:
    app_env: str
    database_url: str
    redis_addr: str
    redis_password: str
    redis_db: int
    queue_name: str
    redis_block_timeout: int

    minio_endpoint: str
    minio_access_key: str
    minio_secret_key: str
    minio_bucket: str
    minio_use_ssl: bool

    log_level: str
    poll_interval_seconds: int
    max_archive_depth: int
    max_extracted_files: int
    max_extracted_size_mb: int
    max_expansion_ratio: float
    temp_directory: Optional[str]
    opensearch_url: str
    opensearch_user: str
    opensearch_password: str
    opensearch_parsed_records_index: str
    opensearch_files_index: str


def load_config() -> WorkerConfig:
    return WorkerConfig(
        app_env=_env("APP_ENV", "local"),
        database_url=_env(
            "DATABASE_URL",
            "postgres://file_user:file_password@postgres:5432/file_platform?sslmode=disable",
        ),
        redis_addr=_env("REDIS_ADDR", "redis:6379"),
        redis_password=_env("REDIS_PASSWORD", ""),
        redis_db=_env_int("REDIS_DB", 0),
        queue_name=_env("WORKER_QUEUE_NAME", _env("QUEUE_NAME", "ingestion_jobs")),
        redis_block_timeout=_env_int("WORKER_REDIS_BLOCK_TIMEOUT", 5),
        max_archive_depth=_env_int("WORKER_MAX_ARCHIVE_DEPTH", 10),
        max_extracted_files=_env_int("WORKER_MAX_EXTRACTED_FILES", 50000),
        max_extracted_size_mb=_env_int("WORKER_MAX_EXTRACTED_SIZE_MB", 1024),
        max_expansion_ratio=_env_float("WORKER_MAX_EXPANSION_RATIO", 20.0),
        minio_endpoint=_normalize_host(_env("MINIO_ENDPOINT", "minio:9000")),
        minio_access_key=_env("MINIO_ACCESS_KEY", "minioadmin"),
        minio_secret_key=_env("MINIO_SECRET_KEY", "minioadmin"),
        minio_bucket=_env("MINIO_BUCKET", "file-ingestion"),
        minio_use_ssl=_env_bool("MINIO_USE_SSL", False),
        log_level=_env("WORKER_LOG_LEVEL", "INFO").upper(),
        poll_interval_seconds=_env_int("WORKER_POLL_INTERVAL_SECONDS", 1),
        temp_directory=_env("WORKER_TEMP_DIR", "/tmp/file-worker"),
        opensearch_url=_env("OPENSEARCH_URL", "http://opensearch:9200"),
        opensearch_user=_env("OPENSEARCH_USER", ""),
        opensearch_password=_env("OPENSEARCH_PASSWORD", ""),
        opensearch_parsed_records_index=_env("OPENSEARCH_PARSED_RECORDS_INDEX", "parsed-records"),
        opensearch_files_index=_env("OPENSEARCH_FILES_INDEX", "files"),
    )

from __future__ import annotations

from app.config import WorkerConfig
from app.storage.base import ObjectStorage


def build_storage(config: WorkerConfig) -> ObjectStorage:
    backend = (config.storage_backend or "minio").strip().lower()
    if backend == "minio":
        from app.storage.minio_client import MinioClient

        return MinioClient(
            endpoint=config.minio_endpoint,
            access_key=config.minio_access_key,
            secret_key=config.minio_secret_key,
            bucket=config.minio_bucket,
            secure=config.minio_use_ssl,
        )
    if backend == "azure":
        from app.storage.azure_blob_client import AzureBlobClient

        return AzureBlobClient(
            account_name=config.azure_storage_account,
            account_key=config.azure_storage_key,
            bucket=config.minio_bucket,
            endpoint=config.azure_storage_endpoint,
        )
    raise ValueError(f"unknown STORAGE_BACKEND {config.storage_backend!r} (want 'minio' or 'azure')")

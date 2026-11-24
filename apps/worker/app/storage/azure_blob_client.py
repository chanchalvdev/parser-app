from __future__ import annotations

from azure.core.exceptions import ResourceExistsError
from azure.storage.blob import BlobServiceClient, ContentSettings


class AzureBlobClient:
    """Azure Blob Storage backend. `bucket` is the blob container name.

    Large files are streamed in blocks by the SDK on both download and upload,
    so memory stays flat regardless of archive size.
    """

    def __init__(
        self,
        *,
        account_name: str,
        account_key: str,
        bucket: str,
        endpoint: str = "",
    ) -> None:
        if not account_name or not account_key:
            raise ValueError("AZURE_STORAGE_ACCOUNT and AZURE_STORAGE_KEY are required when STORAGE_BACKEND=azure")
        account_url = (endpoint or f"https://{account_name}.blob.core.windows.net").rstrip("/")
        self.bucket = bucket
        self._service = BlobServiceClient(
            account_url=account_url,
            credential={"account_name": account_name, "account_key": account_key},
        )
        self._container = self._service.get_container_client(bucket)

    def download_object(self, object_name: str, local_path: str) -> None:
        downloader = self._container.get_blob_client(object_name).download_blob(max_concurrency=4)
        with open(local_path, "wb") as handle:
            downloader.readinto(handle)

    def upload_object(
        self,
        local_path: str,
        object_name: str,
        *,
        content_type: str | None = None,
    ) -> None:
        with open(local_path, "rb") as handle:
            self._container.get_blob_client(object_name).upload_blob(
                handle,
                overwrite=True,
                max_concurrency=4,
                content_settings=ContentSettings(content_type=content_type or "application/octet-stream"),
            )

    def ensure_bucket(self) -> None:
        try:
            self._container.create_container()
        except ResourceExistsError:
            pass

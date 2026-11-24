from __future__ import annotations

from typing import Protocol


class ObjectStorage(Protocol):
    """What the worker needs from an object store (MinIO/S3 or Azure Blob)."""

    bucket: str

    def download_object(self, object_name: str, local_path: str) -> None: ...

    def upload_object(
        self,
        local_path: str,
        object_name: str,
        *,
        content_type: str | None = None,
    ) -> None: ...

    def ensure_bucket(self) -> None: ...

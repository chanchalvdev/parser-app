from __future__ import annotations

from dataclasses import dataclass

from minio import Minio


@dataclass(frozen=True)
class MinioClient:
    endpoint: str
    access_key: str
    secret_key: str
    bucket: str
    secure: bool = False

    def __post_init__(self) -> None:
        object.__setattr__(
            self,
            "_client",
            Minio(
                self.endpoint,
                access_key=self.access_key,
                secret_key=self.secret_key,
                secure=self.secure,
            ),
        )

    @property
    def client(self) -> Minio:
        return self._client

    def download_object(self, object_name: str, local_path: str) -> None:
        self.client.fget_object(self.bucket, object_name, local_path)

    def upload_object(
        self,
        local_path: str,
        object_name: str,
        *,
        content_type: str | None = None,
    ) -> None:
        self.client.fput_object(
            bucket_name=self.bucket,
            object_name=object_name,
            file_path=local_path,
            content_type=content_type or "application/octet-stream",
        )

    def ensure_bucket(self) -> None:
        if self.client.bucket_exists(self.bucket):
            return
        self.client.make_bucket(self.bucket)

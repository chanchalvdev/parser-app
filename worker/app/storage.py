import boto3
import os
import tempfile
from botocore.config import Config as BotoConfig

from .config import settings


class S3LikeStorage:
    def __init__(self) -> None:
        scheme = "https" if settings.minio_use_ssl else "http"
        self.client = boto3.client(
            "s3",
            endpoint_url=f"{scheme}://{settings.minio_endpoint}",
            aws_access_key_id=settings.minio_access_key,
            aws_secret_access_key=settings.minio_secret_key,
            config=BotoConfig(
                signature_version="s3v4",
                s3={"addressing_style": "path"},
            ),
        )
        self.bucket = settings.minio_bucket

    def download_to_file(self, key: str) -> str:
        suffix = os.path.basename(key)
        fd, path = tempfile.mkstemp(suffix=suffix)
        os.close(fd)
        self.client.download_file(self.bucket, key, path)
        return path

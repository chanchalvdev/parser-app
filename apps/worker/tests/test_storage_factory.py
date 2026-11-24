from __future__ import annotations

import dataclasses

import pytest

from app.config import load_config
from app.storage.azure_blob_client import AzureBlobClient
from app.storage.factory import build_storage
from app.storage.minio_client import MinioClient

AZURITE_KEY = "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw=="


def test_default_backend_is_minio(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.delenv("STORAGE_BACKEND", raising=False)
    config = load_config()
    assert config.storage_backend == "minio"
    assert isinstance(build_storage(config), MinioClient)


def test_azure_backend_from_env(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("STORAGE_BACKEND", "azure")
    monkeypatch.setenv("AZURE_STORAGE_ACCOUNT", "stparsertest")
    monkeypatch.setenv("AZURE_STORAGE_KEY", AZURITE_KEY)
    monkeypatch.setenv("MINIO_BUCKET", "file-ingestion")
    config = load_config()
    storage = build_storage(config)
    assert isinstance(storage, AzureBlobClient)
    assert storage.bucket == "file-ingestion"
    assert storage._container.url == "https://stparsertest.blob.core.windows.net/file-ingestion"


def test_azure_endpoint_override(monkeypatch: pytest.MonkeyPatch) -> None:
    storage = AzureBlobClient(
        account_name="devstoreaccount1",
        account_key=AZURITE_KEY,
        bucket="file-ingestion",
        endpoint="http://azurite:10000/devstoreaccount1/",
    )
    assert storage._container.url == "http://azurite:10000/devstoreaccount1/file-ingestion"


def test_azure_requires_credentials() -> None:
    with pytest.raises(ValueError):
        AzureBlobClient(account_name="", account_key="", bucket="b")


def test_unknown_backend_rejected(monkeypatch: pytest.MonkeyPatch) -> None:
    config = dataclasses.replace(load_config(), storage_backend="gcs")
    with pytest.raises(ValueError):
        build_storage(config)


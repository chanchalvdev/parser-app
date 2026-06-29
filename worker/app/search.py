from typing import Any
from opensearchpy import OpenSearch

from .config import settings


class Indexer:
    def __init__(self) -> None:
        self.enabled = False
        self.client = None
        try:
            self.client = OpenSearch(
                [settings.opensearch_url],
                http_auth=(settings.opensearch_user, settings.opensearch_pass),
                use_ssl=False,
                verify_certs=False,
            )
            if not self.client.indices.exists(index=settings.opensearch_index):
                self.client.indices.create(index=settings.opensearch_index, body={"mappings": {"properties": {
                    "upload_id": {"type": "keyword"},
                    "job_id": {"type": "keyword"},
                    "file_id": {"type": "keyword"},
                    "filename": {"type": "text"},
                    "path": {"type": "text"},
                    "content": {"type": "text"},
                    "source_format": {"type": "keyword"},
                }}})
            self.enabled = True
        except Exception:
            self.enabled = False

    def index(self, doc_id: str, payload: dict[str, Any]) -> None:
        if not self.enabled or not self.client:
            return
        self.client.index(index=settings.opensearch_index, id=doc_id, body=payload)


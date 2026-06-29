from __future__ import annotations

import json
import logging
from dataclasses import dataclass
from typing import Any, Optional

import redis


@dataclass(frozen=True)
class IngestionJobMessage:
    job_id: str
    tenant_id: str
    root_file_id: str
    storage_path: str
    bucket: str
    original_name: str
    depth: int
    created_at: str

    @classmethod
    def from_payload(cls, payload: Any) -> "IngestionJobMessage":
        if not isinstance(payload, dict):
            raise TypeError("job message must be a json object")

        required = [
            "job_id",
            "tenant_id",
            "root_file_id",
            "storage_path",
            "bucket",
            "original_name",
            "depth",
            "created_at",
        ]
        missing = [name for name in required if name not in payload or payload[name] in (None, "")]
        if missing:
            raise ValueError(f"missing job message fields: {', '.join(missing)}")

        return cls(
            job_id=str(payload["job_id"]),
            tenant_id=str(payload["tenant_id"]),
            root_file_id=str(payload["root_file_id"]),
            storage_path=str(payload["storage_path"]),
            bucket=str(payload["bucket"]),
            original_name=str(payload["original_name"]),
            depth=int(payload["depth"]),
            created_at=str(payload["created_at"]),
        )

    def to_log_payload(self) -> dict[str, str]:
        return {
            "job_id": self.job_id,
            "tenant_id": self.tenant_id,
            "root_file_id": self.root_file_id,
            "storage_path": self.storage_path,
            "bucket": self.bucket,
        }


class RedisIngestionQueueConsumer:
    def __init__(self, redis_client: redis.Redis, queue_name: str, block_timeout_seconds: int = 5):
        self.redis_client = redis_client
        self.queue_name = queue_name
        self.block_timeout_seconds = block_timeout_seconds
        self.logger = logging.getLogger("worker.queue")

    def pop(self) -> Optional[IngestionJobMessage]:
        result = self.redis_client.brpop(self.queue_name, timeout=self.block_timeout_seconds)
        if result is None:
            return None

        _, payload = result
        if isinstance(payload, bytes):
            payload = payload.decode("utf-8")

        if not isinstance(payload, str):
            self.logger.error("invalid job payload type", extra={"payload_type": type(payload).__name__})
            return None

        try:
            parsed = json.loads(payload)
        except json.JSONDecodeError:
            self.logger.exception("invalid json in job payload", extra={"payload": payload[:2000]})
            return None

        return IngestionJobMessage.from_payload(parsed)

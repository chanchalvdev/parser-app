from __future__ import annotations

import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from app.queue.redis_consumer import RedisIngestionQueueConsumer


class FakeRedisList:
    """Just enough of redis-py's list commands to model queue ordering."""

    def __init__(self) -> None:
        self.items: list[bytes] = []

    def rpush(self, _key: str, value: str) -> None:
        self.items.append(value.encode("utf-8"))

    def blpop(self, key: str, timeout: int = 0):
        return (key.encode(), self.items.pop(0)) if self.items else None

    def brpop(self, key: str, timeout: int = 0):
        return (key.encode(), self.items.pop()) if self.items else None


def _message(job_id: str) -> str:
    return json.dumps(
        {
            "job_id": job_id,
            "tenant_id": "t1",
            "root_file_id": f"file-{job_id}",
            "storage_path": f"raw/{job_id}",
            "bucket": "file-ingestion",
            "original_name": f"{job_id}.txt",
            "depth": 0,
            "created_at": "2026-01-01T00:00:00Z",
        }
    )


def test_consumer_processes_jobs_in_enqueue_order():
    # The Go producer uses RPUSH; jobs must come back oldest-first.
    redis_list = FakeRedisList()
    for job_id in ("first", "second", "third"):
        redis_list.rpush("ingestion_jobs", _message(job_id))

    consumer = RedisIngestionQueueConsumer(redis_list, "ingestion_jobs", block_timeout_seconds=0)

    assert [consumer.pop().job_id for _ in range(3)] == ["first", "second", "third"]
    assert consumer.pop() is None

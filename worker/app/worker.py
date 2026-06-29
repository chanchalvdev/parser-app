import contextlib
import json
import os
import logging
from dataclasses import dataclass

import redis

from .config import settings
from .db import Database
from .extract import extract_archive, is_archive, cleanup, tempdir
from .parsers import detect_kind, parse, sha256_file
from .search import Indexer
from .storage import S3LikeStorage

logger = logging.getLogger("worker")
logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")


@dataclass
class JobPayload:
    job_id: str
    upload_id: str
    object_key: str
    filename: str
    password: str = ""
    attempts: int = 0


class Worker:
    def __init__(self) -> None:
        self.redis = redis.from_url(f"redis://:{settings.redis_password}@{settings.redis_addr}/{settings.redis_db}" if settings.redis_password else f"redis://{settings.redis_addr}/{settings.redis_db}")
        self.db = Database()
        self.storage = S3LikeStorage()
        self.indexer = Indexer()

    def run(self) -> None:
        queue_name = settings.queue_name
        while True:
            item = self.redis.blpop(queue_name, timeout=5)
            if not item:
                continue
            payload = json.loads(item[1])
            try:
                self.process_job(JobPayload(**payload))
            except Exception as exc:
                logger.exception("failed to process job %s: %s", payload.get("job_id"), exc)

    def process_job(self, job: JobPayload) -> None:
        self.db.set_job_status(job.job_id, "processing", "processing", "")
        self.db.set_upload_status(job.upload_id, "processing", "")
        self.db.add_audit_log(job.upload_id, job.job_id, "worker", "job.processing", {"filename": job.filename})
        raw_path = None
        try:
            raw_path = self.storage.download_to_file(job.object_key)
            self._process_path(
                raw_path,
                rel_prefix=job.filename or os.path.basename(raw_path),
                upload_id=job.upload_id,
                job_id=job.job_id,
                parent_id=None,
                password=job.password or None
            )
            self.db.set_job_status(job.job_id, "completed", "completed", "")
            self.db.set_upload_status(job.upload_id, "completed", "")
            self.db.add_audit_log(job.upload_id, job.job_id, "worker", "job.completed", {})
        except RuntimeError as exc:
            message = str(exc)
            if "password required" in message.lower():
                self.db.set_job_status(job.job_id, "password_required", "password_required", message)
                self.db.set_upload_status(job.upload_id, "password_required", message)
                self.db.add_audit_log(job.upload_id, job.job_id, "worker", "job.password_required", {"message": message})
                logger.warning("job %s needs password", job.job_id)
            else:
                self.db.set_job_status(job.job_id, "failed", "failed", message)
                self.db.set_upload_status(job.upload_id, "failed", message)
                self.db.add_audit_log(job.upload_id, job.job_id, "worker", "job.failed", {"message": message})
                logger.exception("job failed: %s", message)
        finally:
        if raw_path:
                with contextlib.suppress(FileNotFoundError):
                    os.remove(raw_path)

    def _process_path(self, path: str, rel_prefix: str, upload_id: str, job_id: str, parent_id: str | None, password: str | None = None) -> None:
        rel_name = rel_prefix or os.path.basename(path)
        is_archive_input = is_archive(path) or is_archive(rel_name)
        if is_archive_input:
            archive_file_id = self.db.add_file_record(
                upload_id=upload_id,
                job_id=job_id,
                parent_id=parent_id,
                file_path=rel_prefix or rel_name,
                filename=os.path.basename(rel_name),
                kind="archive",
                mime_type="application/octet-stream",
                size_bytes=os.path.getsize(path),
                sha256=sha256_file(path),
                metadata={"path": rel_name},
            )
            workspace = tempdir("extract_")
            try:
                extracted_files = extract_archive(path, workspace, password=password)
                for child in extracted_files:
                    child_abs = child if os.path.isabs(child) else os.path.join(workspace, child)
                    child_rel = os.path.join(rel_name, os.path.relpath(child_abs, workspace))
                    self._process_path(child_abs, child_rel, upload_id, job_id, archive_file_id, password)
            finally:
                cleanup(workspace)
            return

        kind = detect_kind(rel_name)
        mime_type = "application/octet-stream"
        parsed = parse(path, kind)
        file_id = self.db.add_file_record(
            upload_id=upload_id,
            job_id=job_id,
            parent_id=parent_id,
            file_path=rel_name,
            filename=os.path.basename(rel_name),
            kind=kind or "file",
            mime_type=mime_type,
            size_bytes=os.path.getsize(path),
            sha256=sha256_file(path),
            metadata=parsed["metadata"],
        )
        self.db.add_parsed_record(
            file_id=file_id,
            upload_id=upload_id,
            job_id=job_id,
            file_path=rel_name,
            source_format=kind or "txt",
            content=parsed["content"],
            metadata=parsed["metadata"],
        )
        self.indexer.index(file_id, {
            "upload_id": upload_id,
            "job_id": job_id,
            "file_id": file_id,
            "filename": os.path.basename(rel_name),
            "path": rel_name,
            "content": parsed["content"],
            "source_format": kind,
            "metadata": parsed["metadata"],
        })

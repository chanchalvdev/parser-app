from __future__ import annotations

from pathlib import Path
import shutil
import tarfile

from app.extractors.base_extractor import (
    ArchiveLimits,
    BaseArchiveExtractor,
    CorruptArchiveError,
    ExtractedFile,
    safe_relative_path,
)
from app.processing.file_detector import FileDetectionResult


class TarExtractor(BaseArchiveExtractor):
    archive_type = "tar"

    def can_extract(self, file_path: str, detection_result: FileDetectionResult) -> bool:
        if detection_result.detected_file_type == "tar":
            return True
        if detection_result.extension in {"tar", "tar.gz", "tgz", "tar.bz2", "tar.xz"}:
            return True
        return detection_result.mime_type in {"application/x-tar", "application/gzip", "application/x-gzip"}

    def extract(
        self,
        file_path: str,
        destination_dir: str,
        *,
        password: str | None = None,
        limits: ArchiveLimits | None = None,
    ) -> list[ExtractedFile]:
        if password:
            raise CorruptArchiveError("tar archives do not support password-protected extraction in this worker")

        extracted: list[ExtractedFile] = []
        extracted_size = 0
        archive_size = Path(file_path).stat().st_size

        try:
            with tarfile.open(file_path, "r:*") as tar:
                for member in tar:
                    if member.isdir():
                        continue

                    member_name = safe_relative_path(member.name)
                    if not member_name:
                        continue

                    destination = Path(destination_dir) / member_name
                    destination.parent.mkdir(parents=True, exist_ok=True)

                    with tar.extractfile(member) as source:
                        if source is None:
                            continue
                        with open(destination, "wb") as target:
                            shutil.copyfileobj(source, target)

                    size_bytes = destination.stat().st_size
                    extracted_size += size_bytes
                    self.enforce_limits(
                        file_count=len(extracted) + 1,
                        extracted_bytes=extracted_size,
                        archive_size_bytes=archive_size,
                        limits=limits,
                    )
                    extracted.append(
                        ExtractedFile(
                            original_name=Path(member_name).name,
                            local_path=str(destination),
                            size_bytes=size_bytes,
                            relative_path=member_name,
                            depth=1,
                        )
                    )

            return extracted
        except tarfile.TarError as exc:
            raise CorruptArchiveError("tar archive is corrupt") from exc
        except Exception as exc:
            raise CorruptArchiveError(f"tar extraction failed: {exc}") from exc

from __future__ import annotations

from abc import ABC, abstractmethod
from dataclasses import dataclass
from pathlib import Path

from app.processing.file_detector import FileDetectionResult


class ExtractionError(RuntimeError):
    error_code = "EXTRACTION_ERROR"


class UnsupportedArchiveError(ExtractionError):
    error_code = "UNSUPPORTED_ARCHIVE"


class CorruptArchiveError(ExtractionError):
    error_code = "CORRUPT_ARCHIVE"


class PasswordRequiredError(ExtractionError):
    error_code = "PASSWORD_REQUIRED"


class WrongPasswordError(ExtractionError):
    error_code = "WRONG_PASSWORD"


class ArchiveLimitsExceededError(ExtractionError):
    error_code = "ARCHIVE_LIMIT_EXCEEDED"


@dataclass(frozen=True)
class ExtractedFile:
    original_name: str
    local_path: str
    size_bytes: int
    relative_path: str
    depth: int


@dataclass(frozen=True)
class ArchiveLimits:
    max_extracted_files: int
    max_extracted_size_mb: int
    max_expansion_ratio: float

    def max_size_bytes(self) -> int:
        return int(self.max_extracted_size_mb * 1024 * 1024)


def safe_relative_path(candidate_path: str) -> str:
    """Return a safe archive-relative path; raise on traversal/absolute paths."""

    normalized = candidate_path.replace("\\", "/").strip()
    if not normalized:
        return ""
    if normalized.startswith("/"):
        raise CorruptArchiveError(f"unsafe archive member path: {candidate_path}")

    parts = [part for part in Path(normalized).parts if part not in {".", ""}]
    if not parts:
        return ""
    if ".." in parts:
        raise CorruptArchiveError(f"unsafe archive member path: {candidate_path}")
    if ":" in normalized:
        raise CorruptArchiveError(f"unsafe archive member path: {candidate_path}")

    return "/".join(parts)


class BaseArchiveExtractor(ABC):
    archive_type: str = "generic"

    @abstractmethod
    def can_extract(self, file_path: str, detection_result: FileDetectionResult) -> bool:
        ...

    @abstractmethod
    def extract(
        self,
        file_path: str,
        destination_dir: str,
        *,
        password: str | None = None,
        limits: ArchiveLimits | None = None,
    ) -> list[ExtractedFile]:
        ...

    @staticmethod
    def enforce_limits(
        file_count: int,
        extracted_bytes: int,
        archive_size_bytes: int,
        limits: ArchiveLimits | None,
    ) -> None:
        if limits is None:
            return

        if file_count > limits.max_extracted_files:
            raise ArchiveLimitsExceededError(
                f"extracted file count {file_count} exceeds max_extracted_files {limits.max_extracted_files}"
            )
        if extracted_bytes > limits.max_size_bytes():
            raise ArchiveLimitsExceededError(
                f"extracted size {extracted_bytes} exceeds max_extracted_size_mb {limits.max_extracted_size_mb}"
            )
        if archive_size_bytes <= 0:
            return

        ratio = extracted_bytes / float(archive_size_bytes)
        if ratio > limits.max_expansion_ratio:
            raise ArchiveLimitsExceededError(
                f"archive expansion ratio {ratio:.2f} exceeds max_expansion_ratio {limits.max_expansion_ratio}"
            )

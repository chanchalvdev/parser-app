from __future__ import annotations

from pathlib import Path

from app.extractors.base_extractor import (
    ArchiveLimits,
    BaseArchiveExtractor,
    CorruptArchiveError,
    ExtractedFile,
    UnsupportedArchiveError,
    PasswordRequiredError,
    WrongPasswordError,
    safe_relative_path,
)
from app.processing.file_detector import FileDetectionResult

try:
    import py7zr
except Exception:  # pragma: no cover - optional dependency in local dev
    py7zr = None


class SevenZipExtractor(BaseArchiveExtractor):
    archive_type = "7z"

    def can_extract(self, file_path: str, detection_result: FileDetectionResult) -> bool:
        if detection_result.detected_file_type == "7z":
            return True
        if detection_result.extension == "7z":
            return True
        return detection_result.mime_type in {"application/x-7z-compressed", "application/x-7z-compressed"}

    def extract(
        self,
        file_path: str,
        destination_dir: str,
        *,
        password: str | None = None,
        limits: ArchiveLimits | None = None,
    ) -> list[ExtractedFile]:
        if py7zr is None:
            raise UnsupportedArchiveError("7z extraction is not available: py7zr dependency is not installed")

        extracted: list[ExtractedFile] = []
        extracted_size = 0
        archive_size = Path(file_path).stat().st_size

        try:
            with py7zr.SevenZipFile(file_path, mode="r", password=password) as archive:
                members = [entry for entry in archive.list() if not getattr(entry, "is_directory", False)]
                if not members:
                    return []

                for entry in members:
                    # py7zr exposes `is_directory` and `filename` for each item in the archive list.
                    member_name = safe_relative_path(getattr(entry, "filename", ""))
                    if not member_name:
                        continue

                    destination = Path(destination_dir) / member_name
                    destination.parent.mkdir(parents=True, exist_ok=True)
                    archive.extract(path=destination_dir, targets=[member_name])

                    size_bytes = destination.stat().st_size if destination.exists() else 0
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
        except py7zr.Bad7zFile as exc:
            # py7zr may hide a few extraction problems in a generic error message; surface clearly.
            if "password" in str(exc).lower():
                if password is None:
                    raise PasswordRequiredError("7z archive password is required") from exc
                raise WrongPasswordError("7z archive password is incorrect") from exc
            raise CorruptArchiveError("7z archive is corrupt") from exc
        except (PasswordRequiredError, WrongPasswordError):
            raise
        except Exception as exc:
            message = str(exc).lower()
            if "password" in message:
                if password is None:
                    raise PasswordRequiredError("7z archive password is required") from exc
                raise WrongPasswordError("7z archive password is incorrect") from exc
            if "unsupported" in message or "not implemented" in message:
                raise UnsupportedArchiveError(f"7z extraction unsupported: {exc}") from exc
            if "encrypted" in message and password is None:
                raise PasswordRequiredError(f"7z archive requires a password: {exc}") from exc
            raise CorruptArchiveError(f"7z extraction failed: {exc}") from exc

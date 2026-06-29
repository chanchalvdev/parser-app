from __future__ import annotations

from pathlib import Path

from app.extractors.base_extractor import (
    ArchiveLimits,
    BaseArchiveExtractor,
    CorruptArchiveError,
    ExtractedFile,
    PasswordRequiredError,
    UnsupportedArchiveError,
    WrongPasswordError,
    safe_relative_path,
)
from app.processing.file_detector import FileDetectionResult

try:
    import rarfile
except Exception:  # pragma: no cover - optional dependency in local dev
    rarfile = None


class RarExtractor(BaseArchiveExtractor):
    archive_type = "rar"

    def can_extract(self, file_path: str, detection_result: FileDetectionResult) -> bool:
        if detection_result.detected_file_type == "rar":
            return True
        if detection_result.extension == "rar":
            return True
        return detection_result.mime_type == "application/x-rar-compressed"

    def extract(
        self,
        file_path: str,
        destination_dir: str,
        *,
        password: str | None = None,
        limits: ArchiveLimits | None = None,
    ) -> list[ExtractedFile]:
        if rarfile is None:
            raise UnsupportedArchiveError("rar extraction is not available: rarfile dependency is not installed")

        extracted: list[ExtractedFile] = []
        extracted_size = 0
        archive_size = Path(file_path).stat().st_size

        rar_no_exec_error = None
        if rarfile is not None:
            rar_no_exec_error = getattr(rarfile, "NoRarExec", None)
            if rar_no_exec_error is None:
                rar_no_exec_error = getattr(rarfile, "RarCannotExec", None)

        try:
            with rarfile.RarFile(file_path) as archive:
                members = [member for member in archive.infolist() if not member.isdir()]
                if not members:
                    return []

                requires_password = any(member.needs_password() for member in members if hasattr(member, "needs_password"))
                if requires_password and not password:
                    raise PasswordRequiredError("rar archive is password-protected")
                
                for member in members:
                    member_name = safe_relative_path(member.filename)
                    if not member_name:
                        continue

                    destination = Path(destination_dir) / member_name
                    destination.parent.mkdir(parents=True, exist_ok=True)

                    with archive.open(member, pwd=password.encode("utf-8") if password else None) as source:
                        with open(destination, "wb") as target:
                            target.write(source.read())

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
                            original_name=member.filename.rsplit("/", 1)[-1],
                            local_path=str(destination),
                            size_bytes=size_bytes,
                            relative_path=member_name,
                            depth=1,
                        )
                    )
                
                return extracted
        except getattr(rarfile, "BadRarFile", RuntimeError) as exc:
            raise CorruptArchiveError("rar archive is corrupt") from exc
        except Exception as exc:
            # Handle optional runtime dependency absence as unsupported archive extraction.
            if rar_no_exec_error is not None and isinstance(exc, rar_no_exec_error):
                raise UnsupportedArchiveError("rar extraction is unavailable: missing unrar/rar system dependency") from exc

            message = str(exc).lower()
            if "password" in message or "encrypted" in message:
                if password is None:
                    raise PasswordRequiredError(f"rar archive requires a password: {exc}") from exc
                raise WrongPasswordError(f"rar archive password is incorrect: {exc}") from exc

            if "unsupported" in message or "not implemented" in message:
                raise UnsupportedArchiveError(f"rar extraction unsupported: {exc}") from exc
            if "no such file" in message:
                raise UnsupportedArchiveError(f"rar archive could not be read: {exc}") from exc
            if "corrupt" in message:
                raise CorruptArchiveError(f"rar archive is corrupt: {exc}") from exc

            raise CorruptArchiveError(f"rar extraction failed: {exc}") from exc

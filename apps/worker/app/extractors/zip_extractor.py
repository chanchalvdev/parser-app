from __future__ import annotations

from pathlib import Path
import zipfile

from app.extractors.base_extractor import (
    ArchiveLimits,
    BaseArchiveExtractor,
    CorruptArchiveError,
    ExtractedFile,
    PasswordRequiredError,
    WrongPasswordError,
    safe_relative_path,
)
from app.processing.file_detector import FileDetectionResult


class ZipExtractor(BaseArchiveExtractor):
    archive_type = "zip"

    def can_extract(self, file_path: str, detection_result: FileDetectionResult) -> bool:
        if detection_result.detected_file_type == "zip":
            return True
        if detection_result.extension == "zip":
            return True
        if detection_result.mime_type in {"application/zip", "application/x-zip-compressed"}:
            return True
        return False

    def extract(
        self,
        file_path: str,
        destination_dir: str,
        *,
        password: str | None = None,
        limits: ArchiveLimits | None = None,
    ) -> list[ExtractedFile]:
        extracted: list[ExtractedFile] = []
        extracted_size = 0
        archive_size = Path(file_path).stat().st_size
        try:
            with zipfile.ZipFile(file_path, "r") as zf:
                entries = [entry for entry in zf.infolist() if not entry.is_dir()]
                if not entries:
                    return []

                has_encrypted = any(entry.flag_bits & 0x1 for entry in entries)
                if has_encrypted and not password:
                    raise PasswordRequiredError("zip archive is password-protected")

                pwd = password.encode("utf-8") if password else None

                for entry in entries:
                    member_name = safe_relative_path(entry.filename)
                    if not member_name:
                        continue

                    destination = Path(destination_dir) / member_name
                    destination.parent.mkdir(parents=True, exist_ok=True)

                    try:
                        raw = zf.read(entry, pwd=pwd)
                    except RuntimeError as exc:
                        message = str(exc).lower()
                        if "password" in message or "bad password" in message:
                            raise WrongPasswordError("zip archive password is incorrect") from exc
                        raise

                    with open(destination, "wb") as target:
                        target.write(raw)

                    size_bytes = len(raw)
                    extracted_size += size_bytes
                    self.enforce_limits(
                        file_count=len(extracted) + 1,
                        extracted_bytes=extracted_size,
                        archive_size_bytes=archive_size,
                        limits=limits,
                    )
                    extracted.append(
                        ExtractedFile(
                            original_name=entry.filename.rsplit("/", 1)[-1],
                            local_path=str(destination),
                            size_bytes=size_bytes,
                            relative_path=member_name,
                            depth=1,
                        )
                    )

                return extracted
        except zipfile.BadZipFile as exc:
            raise CorruptArchiveError("zip archive is corrupt") from exc
        except WrongPasswordError:
            raise
        except PasswordRequiredError:
            raise
        except Exception as exc:
            raise CorruptArchiveError(f"zip extraction failed: {exc}") from exc

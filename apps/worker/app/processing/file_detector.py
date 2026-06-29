from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
import mimetypes

try:
    import magic  # type: ignore
except Exception:  # pragma: no cover
    magic = None


ARCHIVE_EXTENSIONS = (
    ".zip",
    ".rar",
    ".tar",
    ".tar.gz",
    ".tgz",
    ".7z",
)

TEXT_LIKE_EXTENSIONS = (
    ".txt",
    ".log",
    ".out",
    ".err",
)

BINARY_SIGNATURE_EXTENSIONS = (
    ".pdf",
    ".jpg",
    ".jpeg",
    ".png",
    ".gif",
    ".docx",
    ".xlsx",
    ".pptx",
)


@dataclass(frozen=True)
class FileDetectionResult:
    extension: str
    mime_type: str
    detected_file_type: str
    is_archive: bool
    is_text_like: bool
    is_binary: bool
    parser_hint: str


def _resolve_extension(original_name: str) -> str:
    filename = (original_name or "").strip().lower()
    for archive_suffix in (".tar.gz", ".tgz"):
        if filename.endswith(archive_suffix):
            return archive_suffix[1:]

    suffix = Path(filename).suffix
    return suffix[1:] if suffix else ""


def _mime_by_magic(file_path: str) -> str | None:
    if magic is None or not hasattr(magic, "from_file"):
        return None
    try:
        mime = magic.from_file(file_path, mime=True)
    except Exception:
        return None
    if isinstance(mime, str):
        return mime
    return None


def _mime_by_filename(name: str) -> str:
    mime_type, _ = mimetypes.guess_type(name.strip())
    return mime_type or "application/octet-stream"


def _has_null_bytes(file_path: str, sample_size: int = 4096) -> bool:
    with open(file_path, "rb") as handle:
        sample = handle.read(sample_size)
    return b"\x00" in sample


def _is_binary_like(file_ext: str, mime_type: str, file_path: str) -> tuple[bool, str]:
    if _has_null_bytes(file_path):
        return True, "binary_content"

    if f".{file_ext}" in BINARY_SIGNATURE_EXTENSIONS:
        return True, file_ext

    if mime_type == "application/pdf":
        return True, "pdf"

    return False, ""


def _is_text_like(file_ext: str, mime_type: str) -> bool:
    if f".{file_ext}" in TEXT_LIKE_EXTENSIONS:
        return True

    if file_ext in {"csv", "json", "jsonl", "xml"}:
        return True

    return mime_type == "text/plain" or mime_type.startswith("text/")


def _resolve_detection(file_ext: str, mime_type: str, is_archive: bool, is_binary: bool) -> tuple[str, str]:
    if is_archive:
        return "archive", file_ext if file_ext else "archive"

    if file_ext in {"txt", "log", "out", "err"}:
        return "text", "text"

    if file_ext == "csv":
        return "csv", "csv"
    if file_ext == "json":
        return "json", "json"
    if file_ext == "jsonl":
        return "jsonl", "jsonl"
    if file_ext == "xml":
        return "xml", "xml"
    if file_ext == "xlsx":
        return "xlsx", "xlsx"
    if file_ext == "pdf":
        return "pdf", "pdf"

    if is_binary:
        return "binary", "binary"

    if mime_type == "application/pdf":
        return "pdf", "pdf"

    if mime_type in {"text/plain", "application/json", "application/xml"}:
        return "text", "text"

    return "unknown", file_ext if file_ext else "unknown"


def detect_file_type(file_path: str, original_name: str) -> FileDetectionResult:
    filename = original_name.strip() if original_name else Path(file_path).name
    file_name = filename.strip().lower()
    extension = _resolve_extension(file_name)

    mime_type = _mime_by_magic(file_path) or _mime_by_filename(file_name)
    is_archive = any(file_name.endswith(suffix) for suffix in ARCHIVE_EXTENSIONS)

    is_binary, binary_hint = _is_binary_like(extension, mime_type, file_path)
    is_text_like = False if is_binary else _is_text_like(extension, mime_type)
    detected_file_type, parser_hint = _resolve_detection(extension, mime_type, is_archive, is_binary)

    if is_binary and binary_hint:
        parser_hint = binary_hint
    elif not is_binary and is_text_like:
        parser_hint = parser_hint or "text"
    elif parser_hint == "unknown":
        parser_hint = "binary" if is_binary else "generic"

    return FileDetectionResult(
        extension=extension,
        mime_type=mime_type,
        detected_file_type=detected_file_type,
        is_archive=is_archive,
        is_text_like=is_text_like,
        is_binary=is_binary,
        parser_hint=parser_hint,
    )

from __future__ import annotations

import os
import re
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Iterator, Mapping

from app.parsers.base_parser import BaseParser, ParsedRecord, UnsupportedParserError
from app.processing.file_detector import FileDetectionResult


LARGE_FILE_THRESHOLD_BYTES = 10 * 1024 * 1024
CHUNK_SIZE_LINES = 500
BINARY_SAMPLE_BYTES = 65536
PRINTABLE_BYTES = set(range(9, 13)) | set(range(32, 127))

IPV4_RE = re.compile(r"\b(?:(?:25[0-5]|2[0-4]\d|1?\d?\d)\.){3}(?:25[0-5]|2[0-4]\d|1?\d?\d)\b")
EMAIL_RE = re.compile(r"\b[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}\b")
URL_RE = re.compile(r"\bhttps?://[^\s<>\"]+")
DOMAIN_RE = re.compile(r"\b(?:(?:[A-Za-z0-9-]+\.)+[A-Za-z]{2,})\b")
TIMESTAMP_RE = re.compile(r"\b\d{4}-\d{2}-\d{2}[T\s]\d{2}:\d{2}:\d{2}(?:\.\d{1,6})?(?:Z|[+-]\d{2}:\d{2})?\b|\b\d{1,2}/\d{1,2}/\d{2,4}\b")
MD5_RE = re.compile(r"\b[a-fA-F0-9]{32}\b")
SHA256_RE = re.compile(r"\b[a-fA-F0-9]{64}\b")


def _detect_encoding(file_path: str) -> str:
    """Detect text encoding using optional charset-normalizer, with required fallbacks."""
    candidates = ["utf-8"]

    try:
        from charset_normalizer import from_bytes

        with open(file_path, "rb") as handle:
            sample = handle.read(131072)
        charset_matches = from_bytes(sample).best()
        if charset_matches and getattr(charset_matches, "encoding", None):
            candidate = charset_matches.encoding.lower()
            if candidate and candidate not in candidates:
                candidates.append(candidate)
    except Exception:
        pass

    if "latin-1" not in candidates:
        candidates.append("latin-1")

    for encoding in candidates:
        try:
            with open(file_path, "r", encoding=encoding) as handle:
                # sample read validates decode path without loading entire file.
                handle.read(2048)
            return encoding
        except Exception:
            continue

    # Always supported for all bytes as a final fallback.
    return "latin-1"


def _looks_like_binary(file_path: str) -> bool:
    with open(file_path, "rb") as handle:
        sample = handle.read(BINARY_SAMPLE_BYTES)
    if not sample:
        return False

    if b"\x00" in sample:
        return True

    non_printable = 0
    for value in sample:
        if value not in PRINTABLE_BYTES:
            non_printable += 1
    return non_printable / max(len(sample), 1) > 0.35


def _extract_entities(text: str) -> dict[str, list[str]]:
    def dedupe(values: list[str]) -> list[str]:
        normalized: list[str] = []
        seen: set[str] = set()
        for value in values:
            key = value.lower()
            if key in seen:
                continue
            seen.add(key)
            normalized.append(value)
        return normalized

    domains: list[str] = []
    for match in DOMAIN_RE.finditer(text):
        value = match.group(0)
        start, end = match.span()
        next_char = text[end] if end < len(text) else ""
        if next_char in {":", "[", "]", ")"}:
            continue
        domains.append(value)

    return {
        "ipv4": dedupe(IPV4_RE.findall(text)),
        "emails": dedupe(EMAIL_RE.findall(text)),
        "urls": dedupe(URL_RE.findall(text)),
        "domains": dedupe(domains),
        "timestamps": dedupe(TIMESTAMP_RE.findall(text)),
        "md5_like_hashes": dedupe(MD5_RE.findall(text)),
        "sha256_like_hashes": dedupe(SHA256_RE.findall(text)),
    }


def _is_line_empty(line: str) -> bool:
    return line.strip() == ""


def _line_record(tenant_id: str, file_id: str, job_id: str, record_number: int, line_number: int, line: str) -> ParsedRecord:
    entities = _extract_entities(line)
    return ParsedRecord(
        tenant_id=tenant_id,
        file_id=file_id,
        job_id=job_id,
        record_type="text_line",
        record_number=record_number,
        line_number=line_number,
        chunk_number=1,
        start_line=line_number,
        end_line=line_number,
        content_text=line,
        structured_data={"line_number": line_number},
        extracted_entities=entities,
        event_timestamp=datetime.now(timezone.utc),
    )


def _chunk_record(
    tenant_id: str,
    file_id: str,
    job_id: str,
    record_number: int,
    chunk_number: int,
    chunk_start: int,
    chunk_end: int,
    lines: list[str],
) -> ParsedRecord:
    content = "\n".join(lines)
    entities = _extract_entities(content)
    return ParsedRecord(
        tenant_id=tenant_id,
        file_id=file_id,
        job_id=job_id,
        record_type="text_chunk",
        record_number=record_number,
        line_number=None,
        chunk_number=chunk_number,
        start_line=chunk_start,
        end_line=chunk_end,
        content_text=content,
        structured_data={
            "chunk_size": len(lines),
            "line_start": chunk_start,
            "line_end": chunk_end,
            "chunk_number": chunk_number,
        },
        extracted_entities=entities,
        event_timestamp=datetime.now(timezone.utc),
    )


class TxtParser(BaseParser):
    @property
    def name(self) -> str:
        return "txt"

    @property
    def supported_extensions(self) -> set[str]:
        return {"txt", "log", "out", "err"}

    def can_parse(self, file_metadata: Mapping[str, Any], detection_result: FileDetectionResult) -> bool:
        if detection_result.extension in self.supported_extensions:
            return True
        if detection_result.mime_type == "text/plain":
            return True
        if detection_result.parser_hint == "text":
            return True
        return False

    def parse(
        self,
        file_path: str,
        file_metadata: Mapping[str, Any],
        detection_result: FileDetectionResult,
    ) -> Iterator[ParsedRecord]:
        del detection_result
        tenant_id = str(file_metadata.get("tenant_id", ""))
        file_id = str(file_metadata.get("file_id", ""))
        job_id = str(file_metadata.get("job_id", ""))

        if not tenant_id or not file_id or not job_id:
            raise ValueError("file_metadata must include tenant_id, file_id, and job_id")
        if _looks_like_binary(file_path):
            raise UnsupportedParserError("txt parser skipped file due to binary signature")

        path = Path(file_path)
        if not path.is_file():
            raise FileNotFoundError(file_path)

        encoding = _detect_encoding(file_path)
        file_size = os.path.getsize(file_path)
        stream_as_chunks = file_size >= LARGE_FILE_THRESHOLD_BYTES
        record_number = 0

        if stream_as_chunks:
            chunk_lines: list[str] = []
            chunk_number = 0
            chunk_start_line: int | None = None
        with path.open("r", encoding=encoding, errors="replace") as handle:
            for line_number, line in enumerate(handle, start=1):
                normalized_line = line.rstrip("\r\n")
                if _is_line_empty(normalized_line):
                    continue

                if not stream_as_chunks:
                    record_number += 1
                    yield _line_record(tenant_id, file_id, job_id, record_number, line_number, normalized_line)
                    continue

                if chunk_start_line is None:
                    chunk_start_line = line_number
                chunk_lines.append(normalized_line)
                if len(chunk_lines) >= CHUNK_SIZE_LINES:
                    chunk_number += 1
                    record_number += 1
                    yield _chunk_record(
                        tenant_id=tenant_id,
                        file_id=file_id,
                        job_id=job_id,
                        record_number=record_number,
                        chunk_number=chunk_number,
                        chunk_start=chunk_start_line,
                        chunk_end=line_number,
                        lines=chunk_lines,
                    )
                    chunk_lines = []
                    chunk_start_line = None

            if stream_as_chunks and chunk_lines:
                chunk_number += 1
                record_number += 1
                yield _chunk_record(
                    tenant_id=tenant_id,
                    file_id=file_id,
                    job_id=job_id,
                    record_number=record_number,
                    chunk_number=chunk_number,
                    chunk_start=chunk_start_line if chunk_start_line is not None else 0,
                    chunk_end=line_number if line_number else 0,
                    lines=chunk_lines,
                )

    def normalize(self, raw_record: ParsedRecord) -> ParsedRecord:
        if raw_record.content_text is None:
            return raw_record

        return ParsedRecord(
            tenant_id=raw_record.tenant_id,
            file_id=raw_record.file_id,
            job_id=raw_record.job_id,
            record_type=raw_record.record_type,
            record_number=raw_record.record_number,
            line_number=raw_record.line_number,
            chunk_number=raw_record.chunk_number,
            start_line=raw_record.start_line,
            end_line=raw_record.end_line,
            content_text=raw_record.content_text.strip(),
            structured_data=dict(raw_record.structured_data),
            extracted_entities=dict(raw_record.extracted_entities),
            event_timestamp=raw_record.event_timestamp,
        )

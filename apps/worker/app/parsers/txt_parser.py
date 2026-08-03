from __future__ import annotations

import os
import re
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Iterator, Mapping

from app.parsers.base_parser import BaseParser, ParsedRecord, UnsupportedParserError
from app.parsers.infostealer import (
    BTC_RE,
    CVE_RE,
    SHA1_RE,
    parse_credential_blocks,
    stealer_category,
)
from app.processing.file_detector import FileDetectionResult


LARGE_FILE_THRESHOLD_BYTES = 10 * 1024 * 1024
CHUNK_SIZE_LINES = 500
# Bound how much stealer password text is buffered for credential-block assembly.
CREDENTIAL_BUFFER_CHAR_CAP = 8 * 1024 * 1024
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


def _extract_entities(text: str) -> dict[str, Any]:
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

    entities: dict[str, list[str]] = {
        "ipv4": dedupe(IPV4_RE.findall(text)),
        "emails": dedupe(EMAIL_RE.findall(text)),
        "urls": dedupe(URL_RE.findall(text)),
        "domains": dedupe(domains),
        "timestamps": dedupe(TIMESTAMP_RE.findall(text)),
        "md5_like_hashes": dedupe(MD5_RE.findall(text)),
        "sha1_like_hashes": dedupe(SHA1_RE.findall(text)),
        "sha256_like_hashes": dedupe(SHA256_RE.findall(text)),
        "cve": dedupe(CVE_RE.findall(text)),
        "bitcoin_addresses": dedupe(BTC_RE.findall(text)),
    }
    # Infostealer credential blocks (SOFT/URL/USER/PASS). Secrets are hashed,
    # never stored in plaintext. Stored under a distinct key so IOC consumers
    # (search/dashboard) that read string-list keys are unaffected.
    credentials = parse_credential_blocks(text)
    if credentials:
        entities["credentials"] = credentials
    return entities


def _is_line_empty(line: str) -> bool:
    return line.strip() == ""


def _tagged(structured_data: dict[str, Any], category: str | None) -> dict[str, Any]:
    if category:
        structured_data = {**structured_data, "stealer_category": category}
    return structured_data


def _line_record(
    tenant_id: str,
    file_id: str,
    job_id: str,
    record_number: int,
    line_number: int,
    line: str,
    category: str | None = None,
) -> ParsedRecord:
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
        structured_data=_tagged({"line_number": line_number}, category),
        extracted_entities=entities,
        event_timestamp=datetime.now(timezone.utc),
    )


def _stealer_credential_record(
    tenant_id: str,
    file_id: str,
    job_id: str,
    record_number: int,
    category: str,
    credentials: list[dict[str, Any]],
) -> ParsedRecord:
    """Consolidated credential record for a stealer password file.

    Small stealer logs are parsed line-by-line, so multi-line SOFT/URL/USER/PASS
    blocks never assemble in the per-line path. This record carries the blocks
    parsed from the whole file. Secrets are already hashed by the grammar.
    """
    return ParsedRecord(
        tenant_id=tenant_id,
        file_id=file_id,
        job_id=job_id,
        record_type="stealer_credential",
        record_number=record_number,
        line_number=None,
        chunk_number=1,
        start_line=None,
        end_line=None,
        content_text=None,
        structured_data=_tagged({"credential_count": len(credentials)}, category),
        extracted_entities={"credentials": credentials},
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
    category: str | None = None,
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
        structured_data=_tagged(
            {
                "chunk_size": len(lines),
                "line_start": chunk_start,
                "line_end": chunk_end,
                "chunk_number": chunk_number,
            },
            category,
        ),
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

        # Infostealer filename routing (Passwords.txt, Cookies/, UserInformation.txt, ...).
        original_name = str(file_metadata.get("original_name", ""))
        category = stealer_category(original_name)
        # Password dumps are parsed line-by-line, which breaks multi-line
        # credential blocks; buffer their text to reassemble blocks at the end.
        collect_credentials = category == "stealer_password"
        cred_buffer: list[str] = []
        cred_buffer_chars = 0

        if stream_as_chunks:
            chunk_lines: list[str] = []
            chunk_number = 0
            chunk_start_line: int | None = None
        with path.open("r", encoding=encoding, errors="replace") as handle:
            for line_number, line in enumerate(handle, start=1):
                normalized_line = line.rstrip("\r\n")
                if _is_line_empty(normalized_line):
                    continue

                if collect_credentials and cred_buffer_chars < CREDENTIAL_BUFFER_CHAR_CAP:
                    cred_buffer.append(normalized_line)
                    cred_buffer_chars += len(normalized_line) + 1

                if not stream_as_chunks:
                    record_number += 1
                    yield _line_record(
                        tenant_id, file_id, job_id, record_number, line_number, normalized_line, category
                    )
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
                        category=category,
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
                    category=category,
                )

        if collect_credentials and cred_buffer:
            credentials = parse_credential_blocks("\n".join(cred_buffer))
            if credentials:
                record_number += 1
                yield _stealer_credential_record(
                    tenant_id, file_id, job_id, record_number, category, credentials
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

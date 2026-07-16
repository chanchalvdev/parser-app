from __future__ import annotations

import csv
import os
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Iterator, Mapping

from app.parsers.base_parser import BaseParser, ParsedRecord, UnsupportedParserError
from app.parsers.txt_parser import _detect_encoding, _extract_entities, _looks_like_binary
from app.processing.file_detector import FileDetectionResult


CSV_ROW_RECORD_TYPE = "csv_row"
# Allow very long single fields (e.g. embedded JSON / base64 blobs) without
# tripping Python's default 128 KiB csv field size limit on large exports.
_MAX_FIELD_BYTES = 16 * 1024 * 1024
_SNIFF_SAMPLE_BYTES = 65536


def _raise_field_limit() -> None:
    # csv.field_size_limit is a process-wide setting; raise it high enough for
    # large enterprise exports but keep it bounded to avoid unbounded memory.
    try:
        if csv.field_size_limit() < _MAX_FIELD_BYTES:
            csv.field_size_limit(_MAX_FIELD_BYTES)
    except OverflowError:  # pragma: no cover - platform dependent
        csv.field_size_limit(_MAX_FIELD_BYTES)


def _sniff_dialect(sample: str) -> type[csv.Dialect] | csv.Dialect:
    if not sample:
        return csv.excel
    try:
        return csv.Sniffer().sniff(sample, delimiters=",;\t|")
    except csv.Error:
        return csv.excel


def _clean_header(raw_header: list[str]) -> list[str]:
    header: list[str] = []
    seen: dict[str, int] = {}
    for index, name in enumerate(raw_header):
        candidate = (name or "").strip()
        if not candidate:
            candidate = f"column_{index + 1}"
        if candidate in seen:
            seen[candidate] += 1
            candidate = f"{candidate}_{seen[candidate]}"
        else:
            seen[candidate] = 0
        header.append(candidate)
    return header


def _row_to_mapping(header: list[str], row: list[str]) -> dict[str, Any]:
    mapping: dict[str, Any] = {}
    for index, value in enumerate(row):
        key = header[index] if index < len(header) else f"column_{index + 1}"
        mapping[key] = value
    # Columns declared in the header but absent from a short row are recorded as
    # empty strings so downstream consumers get a stable schema.
    for index in range(len(row), len(header)):
        mapping[header[index]] = ""
    return mapping


def _content_from_mapping(mapping: Mapping[str, Any]) -> str:
    return "\n".join(f"{key}: {value}" for key, value in mapping.items())


class CsvParser(BaseParser):
    @property
    def name(self) -> str:
        return "csv"

    @property
    def supported_extensions(self) -> set[str]:
        return {"csv"}

    def can_parse(self, file_metadata: Mapping[str, Any], detection_result: FileDetectionResult) -> bool:
        del file_metadata
        return detection_result.extension in self.supported_extensions or detection_result.detected_file_type == "csv"

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

        path = Path(file_path)
        if not path.is_file():
            raise FileNotFoundError(file_path)

        if os.path.getsize(file_path) == 0:
            raise UnsupportedParserError("CSV file is empty")

        if _looks_like_binary(file_path):
            raise UnsupportedParserError("csv parser skipped file due to binary signature")

        _raise_field_limit()
        encoding = _detect_encoding(file_path)

        with path.open("r", encoding=encoding, errors="replace", newline="") as sniff_handle:
            sample = sniff_handle.read(_SNIFF_SAMPLE_BYTES)
        dialect = _sniff_dialect(sample)

        # The file handle is streamed row-by-row; the whole file is never loaded
        # into memory, so arbitrarily large CSV exports parse in constant memory.
        with path.open("r", encoding=encoding, errors="replace", newline="") as handle:
            reader = csv.reader(handle, dialect)

            try:
                raw_header = next(reader)
            except StopIteration:
                raise UnsupportedParserError("CSV file has no rows")

            header = _clean_header(raw_header)
            record_number = 0

            for row in reader:
                if not row or all((cell or "").strip() == "" for cell in row):
                    continue

                record_number += 1
                # csv line numbers are 1-based and include the header row.
                line_number = reader.line_num
                mapping = _row_to_mapping(header, row)
                content_text = _content_from_mapping(mapping)
                entities = _extract_entities(content_text)

                structured_data: dict[str, Any] = dict(mapping)
                structured_data["_row_number"] = record_number
                structured_data["_line_number"] = line_number

                yield ParsedRecord(
                    tenant_id=tenant_id,
                    file_id=file_id,
                    job_id=job_id,
                    record_type=CSV_ROW_RECORD_TYPE,
                    record_number=record_number,
                    line_number=line_number,
                    chunk_number=1,
                    start_line=line_number,
                    end_line=line_number,
                    content_text=content_text,
                    structured_data=structured_data,
                    extracted_entities=entities,
                    event_timestamp=datetime.now(timezone.utc),
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

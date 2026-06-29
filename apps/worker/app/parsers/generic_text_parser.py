from __future__ import annotations

from pathlib import Path
from typing import Any, Iterator, Mapping

from app.parsers.base_parser import BaseParser, ParsedRecord, UnsupportedParserError
from app.processing.file_detector import FileDetectionResult


class GenericTextParser(BaseParser):
    @property
    def name(self) -> str:
        return "generic_text"

    @property
    def supported_extensions(self) -> set[str]:
        return set()

    def can_parse(self, file_metadata: Mapping[str, Any], detection_result: FileDetectionResult) -> bool:
        return bool(detection_result.is_text_like)

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

        record_number = 0
        try:
            with path.open("r", encoding="utf-8", errors="replace") as handle:
                for line_number, line in enumerate(handle, start=1):
                    record_number += 1
                    normalized_line = line.rstrip("\r\n")
                    yield ParsedRecord(
                        tenant_id=tenant_id,
                        file_id=file_id,
                        job_id=job_id,
                        record_type="generic_text_line",
                        record_number=record_number,
                        line_number=line_number,
                        chunk_number=1,
                        start_line=line_number,
                        end_line=line_number,
                        content_text=normalized_line,
                        structured_data={"text": normalized_line},
                        extracted_entities={},
                        event_timestamp=None,
                    )
        except UnicodeDecodeError as exc:
            raise UnsupportedParserError(f"file is not parseable as UTF-8 text: {exc}") from exc

    def normalize(self, raw_record: ParsedRecord) -> ParsedRecord:
        if raw_record.content_text is None:
            return raw_record

        normalized_text = raw_record.content_text.strip()
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
            content_text=normalized_text,
            structured_data=dict(raw_record.structured_data),
            extracted_entities=dict(raw_record.extracted_entities),
            event_timestamp=raw_record.event_timestamp,
        )

from __future__ import annotations

from typing import Any, Iterator, Mapping

from app.parsers.base_parser import BaseParser, ParsedRecord, UnsupportedParserError
from app.processing.file_detector import FileDetectionResult


class ExcelParser(BaseParser):
    @property
    def name(self) -> str:
        return "excel"

    @property
    def supported_extensions(self) -> set[str]:
        return {"xlsx"}

    def can_parse(self, file_metadata: Mapping[str, Any], detection_result: FileDetectionResult) -> bool:
        del file_metadata
        return detection_result.extension == "xlsx"

    def parse(
        self,
        file_path: str,
        file_metadata: Mapping[str, Any],
        detection_result: FileDetectionResult,
    ) -> Iterator[ParsedRecord]:
        del file_path, file_metadata, detection_result
        raise UnsupportedParserError("Excel parser is scaffolded only; implementation is intentionally deferred.")

    def normalize(self, raw_record: ParsedRecord) -> ParsedRecord:
        return raw_record

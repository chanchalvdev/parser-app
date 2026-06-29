from __future__ import annotations

from abc import ABC, abstractmethod
from dataclasses import dataclass
from datetime import datetime
from typing import Any, Iterator, Mapping, Optional

from app.processing.file_detector import FileDetectionResult


@dataclass(frozen=True)
class ParsedRecord:
    tenant_id: str
    file_id: str
    job_id: str
    record_type: str
    record_number: int
    line_number: Optional[int]
    chunk_number: Optional[int]
    start_line: Optional[int]
    end_line: Optional[int]
    content_text: Optional[str]
    structured_data: dict[str, Any]
    extracted_entities: dict[str, Any]
    event_timestamp: Optional[datetime]


class UnsupportedParserError(RuntimeError):
    """Raised when a parser can not parse the supplied file type."""


class BaseParser(ABC):
    @property
    @abstractmethod
    def name(self) -> str:
        ...

    @property
    @abstractmethod
    def supported_extensions(self) -> set[str]:
        ...

    @abstractmethod
    def can_parse(self, file_metadata: Mapping[str, Any], detection_result: FileDetectionResult) -> bool:
        ...

    @abstractmethod
    def parse(
        self,
        file_path: str,
        file_metadata: Mapping[str, Any],
        detection_result: FileDetectionResult,
    ) -> Iterator[ParsedRecord]:
        ...

    @abstractmethod
    def normalize(self, raw_record: ParsedRecord) -> ParsedRecord:
        ...

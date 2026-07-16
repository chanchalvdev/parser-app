from __future__ import annotations

import json
from collections.abc import Mapping
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Iterator

from app.parsers.base_parser import BaseParser, ParsedRecord, UnsupportedParserError
from app.parsers.txt_parser import _extract_entities
from app.processing.file_detector import FileDetectionResult


try:
    import ijson  # type: ignore
except Exception:  # pragma: no cover - optional dependency
    ijson = None


JSONL_RECORD_TYPE = "jsonl_line"
JSON_OBJECT_RECORD_TYPE = "json_object"
JSON_ARRAY_RECORD_TYPE = "json_array_item"
JSON_PARSER_ERROR_RECORD_TYPE = "parser_error"

# Keep fallback reads bounded for this MVP. Install ijson for streaming arrays above this size.
MAX_JSON_ARRAY_FALLBACK_BYTES = 16 * 1024 * 1024


def _is_empty(value: Any) -> bool:
    return value is None or value == "" or value == {}


def _serialize_value(value: Any) -> str:
    if isinstance(value, str):
        return value
    if value is None:
        return "null"
    return json.dumps(value, ensure_ascii=False)


def _flatten_values(value: Any, prefix: str, output: list[str]) -> None:
    if isinstance(value, Mapping):
        for key, child in value.items():
            key_name = str(key)
            path = key_name if not prefix else f"{prefix}.{key_name}"
            _flatten_values(child, path, output)
    elif isinstance(value, list):
        for index, child in enumerate(value):
            path = f"[{index}]" if not prefix else f"{prefix}[{index}]"
            _flatten_values(child, path, output)
    else:
        output.append(f"{prefix}: {_serialize_value(value)}" if prefix else _serialize_value(value))


def _flatten_content(value: Any) -> str:
    pieces: list[str] = []
    _flatten_values(value, "", pieces)
    return "\n".join(pieces)


def _read_metadata_from_file(file_path: str) -> str:
    path = Path(file_path)
    with path.open("r", encoding="utf-8", errors="replace") as handle:
        while True:
            chunk = handle.read(4096)
            if not chunk:
                return ""
            for char in chunk:
                if char.isspace():
                    continue
                return char
    return ""


def _to_structured_data(payload: Any) -> dict[str, Any]:
    if isinstance(payload, dict):
        return dict(payload)
    return {"value": payload}


def _content_and_entities(payload: Any) -> tuple[str, dict[str, Any]]:
    content_text = _flatten_content(payload)
    if _is_empty(content_text):
        content_text = _serialize_value(payload)
    return content_text, _extract_entities(content_text)


class JsonParser(BaseParser):
    def __init__(self) -> None:
        self.parser_errors: list[dict[str, Any]] = []

    @property
    def name(self) -> str:
        return "json"

    @property
    def supported_extensions(self) -> set[str]:
        return {"json", "jsonl"}

    def can_parse(self, file_metadata: Mapping[str, Any], detection_result: FileDetectionResult) -> bool:
        del file_metadata
        return detection_result.extension in self.supported_extensions or detection_result.detected_file_type in {"json", "jsonl"}

    def _object_record(
        self,
        tenant_id: str,
        file_id: str,
        job_id: str,
        record_number: int,
        payload: Mapping[str, Any],
        *,
        line_number: int | None = None,
        start_line: int | None = None,
        end_line: int | None = None,
        record_type: str,
    ) -> ParsedRecord:
        content_text, entities = _content_and_entities(payload)
        structured_data = dict(payload)
        return ParsedRecord(
            tenant_id=tenant_id,
            file_id=file_id,
            job_id=job_id,
            record_type=record_type,
            record_number=record_number,
            line_number=line_number,
            chunk_number=1,
            start_line=start_line,
            end_line=end_line,
            content_text=content_text,
            structured_data=structured_data,
            extracted_entities=entities,
            event_timestamp=datetime.now(timezone.utc),
        )

    def _error_record(
        self,
        tenant_id: str,
        file_id: str,
        job_id: str,
        record_number: int,
        line_number: int,
        line_value: str,
        error: str,
    ) -> ParsedRecord:
        error_context = {
            "error_code": "JSONL_MALFORMED_LINE",
            "line_number": line_number,
            "line": line_value,
            "error": error,
        }
        self.parser_errors.append(error_context)
        return ParsedRecord(
            tenant_id=tenant_id,
            file_id=file_id,
            job_id=job_id,
            record_type=JSON_PARSER_ERROR_RECORD_TYPE,
            record_number=record_number,
            line_number=line_number,
            chunk_number=1,
            start_line=line_number,
            end_line=line_number,
            content_text=f"Malformed JSONL at line {line_number}: {error}",
            structured_data=error_context,
            extracted_entities={},
            event_timestamp=datetime.now(timezone.utc),
        )

    def _iter_json_array_streaming(self, file_path: str) -> Iterator[Any]:
        # ijson streams from a binary handle (text mode is deprecated in ijson 3.x),
        # so arbitrarily large JSON arrays are parsed in constant memory.
        with Path(file_path).open("rb") as handle:
            yield from ijson.items(handle, "item")  # type: ignore[misc]

    def _iter_json_array_fallback(self, file_path: str) -> Iterator[Any]:
        path = Path(file_path)
        if path.stat().st_size > MAX_JSON_ARRAY_FALLBACK_BYTES:
            raise UnsupportedParserError(
                "JSON array parsing in fallback mode requires reading the full file; "
                f"array file is larger than {MAX_JSON_ARRAY_FALLBACK_BYTES} bytes. "
                "Install ijson for streaming JSON array parsing."
            )

        with path.open("r", encoding="utf-8", errors="replace") as handle:
            raw = handle.read()
            if not raw.strip():
                raise UnsupportedParserError("Empty JSON array file")

        payload = json.loads(raw)
        if not isinstance(payload, list):
            raise UnsupportedParserError("Expected JSON array but parsed file content was not an array")

        yield from payload

    def _parse_json_array(self, file_path: str) -> Iterator[Any]:
        if ijson is not None:
            try:
                yield from self._iter_json_array_streaming(file_path)
                return
            except Exception:
                # Fall through to fallback parser for non-array / malformed cases.
                pass

        yield from self._iter_json_array_fallback(file_path)

    def parse(
        self,
        file_path: str,
        file_metadata: Mapping[str, Any],
        detection_result: FileDetectionResult,
    ) -> Iterator[ParsedRecord]:
        self.parser_errors = []
        tenant_id = str(file_metadata.get("tenant_id", ""))
        file_id = str(file_metadata.get("file_id", ""))
        job_id = str(file_metadata.get("job_id", ""))
        if not tenant_id or not file_id or not job_id:
            raise ValueError("file_metadata must include tenant_id, file_id, and job_id")

        path = Path(file_path)
        if not path.is_file():
            raise FileNotFoundError(file_path)

        is_jsonl = detection_result.extension == "jsonl" or detection_result.detected_file_type == "jsonl"

        record_number = 0

        if is_jsonl:
            with path.open("r", encoding="utf-8", errors="replace") as handle:
                for line_number, line in enumerate(handle, start=1):
                    normalized = line.strip()
                    if normalized == "":
                        continue
                    record_number += 1
                    try:
                        payload = json.loads(normalized)
                    except Exception as exc:
                        yield self._error_record(
                            tenant_id=tenant_id,
                            file_id=file_id,
                            job_id=job_id,
                            record_number=record_number,
                            line_number=line_number,
                            line_value=normalized,
                            error=f"{type(exc).__name__}: {exc}",
                        )
                        continue

                    payload_dict = _to_structured_data(payload)
                    yield self._object_record(
                        tenant_id=tenant_id,
                        file_id=file_id,
                        job_id=job_id,
                        record_number=record_number,
                        payload=payload_dict,
                        line_number=line_number,
                        start_line=line_number,
                        end_line=line_number,
                        record_type=JSONL_RECORD_TYPE,
                    )
            return

        first_non_ws = _read_metadata_from_file(file_path)
        if first_non_ws == "{":
            raw = path.read_text(encoding="utf-8", errors="replace")
            if not raw.strip():
                raise UnsupportedParserError("Empty JSON file")
            try:
                payload = json.loads(raw)
            except json.JSONDecodeError as exc:
                raise UnsupportedParserError(f"Failed to parse JSON object: {exc}") from exc
            if not isinstance(payload, Mapping):
                raise UnsupportedParserError("Expected a single JSON object in non-array JSON file")

            yield self._object_record(
                tenant_id=tenant_id,
                file_id=file_id,
                job_id=job_id,
                record_number=1,
                payload=payload,
                record_type=JSON_OBJECT_RECORD_TYPE,
            )
            return

        if first_non_ws == "[":
            for payload in self._parse_json_array(file_path):
                if not isinstance(payload, Mapping):
                    payload = _to_structured_data(payload)
                else:
                    payload = dict(payload)
                record_number += 1
                yield self._object_record(
                    tenant_id=tenant_id,
                    file_id=file_id,
                    job_id=job_id,
                    record_number=record_number,
                    payload=payload,
                    record_type=JSON_ARRAY_RECORD_TYPE,
                )
            return

        raise UnsupportedParserError(f"Unsupported JSON shape: first token is {first_non_ws!r}")

    def normalize(self, raw_record: ParsedRecord) -> ParsedRecord:
        if raw_record.content_text is None:
            return raw_record

        normalized_content = raw_record.content_text.strip()
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
            content_text=normalized_content,
            structured_data=dict(raw_record.structured_data),
            extracted_entities=dict(raw_record.extracted_entities),
            event_timestamp=raw_record.event_timestamp,
        )

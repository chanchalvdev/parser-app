from __future__ import annotations

import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from app.processing.file_detector import FileDetectionResult
from app.parsers.json_parser import JsonParser


def _detection_result(kind: str) -> FileDetectionResult:
    return FileDetectionResult(
        extension=kind,
        mime_type="application/json",
        detected_file_type=kind,
        is_archive=False,
        is_text_like=True,
        is_binary=False,
        parser_hint=kind,
    )


def _metadata() -> dict[str, str]:
    return {
        "tenant_id": "tenant-123",
        "file_id": "file-abc",
        "job_id": "job-xyz",
    }


def test_json_parser_single_object(tmp_path: Path) -> None:
    parser = JsonParser()
    payload = '{"id": 1, "nested": {"name": "alpha", "tags": ["x", "y"]}}'
    source = tmp_path / "sample.json"
    source.write_text(payload, encoding="utf-8")

    records = list(parser.parse(str(source), _metadata(), _detection_result("json")))

    assert len(records) == 1
    assert records[0].record_type == "json_object"
    assert records[0].structured_data == {"id": 1, "nested": {"name": "alpha", "tags": ["x", "y"]}}
    assert "id" in records[0].content_text
    assert "nested.name" in records[0].content_text
    assert records[0].extracted_entities == {
        "ipv4": [],
        "emails": [],
        "urls": [],
        "domains": [],
        "timestamps": [],
        "md5_like_hashes": [],
        "sha1_like_hashes": [],
        "sha256_like_hashes": [],
        "cve": [],
        "bitcoin_addresses": [],
    }


def test_json_parser_array(tmp_path: Path) -> None:
    parser = JsonParser()
    payload = '[{"a": 1}, {"a": 2, "nested": {"label": "B"}}]'
    source = tmp_path / "sample.json"
    source.write_text(payload, encoding="utf-8")

    records = list(parser.parse(str(source), _metadata(), _detection_result("json")))

    assert len(records) == 2
    assert records[0].record_type == "json_array_item"
    assert records[0].structured_data == {"a": 1}
    assert records[1].structured_data == {"a": 2, "nested": {"label": "B"}}


def test_json_parser_jsonl(tmp_path: Path) -> None:
    parser = JsonParser()
    payload = '{"a": 1}\n{"b": 2}\n{"c": "value"}'
    source = tmp_path / "sample.jsonl"
    source.write_text(payload, encoding="utf-8")

    records = list(parser.parse(str(source), _metadata(), _detection_result("jsonl")))
    assert len(records) == 3
    assert {r.record_type for r in records} == {"jsonl_line"}
    assert records[2].structured_data["c"] == "value"


def test_json_parser_jsonl_malformed_line(tmp_path: Path) -> None:
    parser = JsonParser()
    payload = '{"a": 1}\n{bad: 2}\n{"c": 3}'
    source = tmp_path / "broken.jsonl"
    source.write_text(payload, encoding="utf-8")

    records = list(parser.parse(str(source), _metadata(), _detection_result("jsonl")))

    parsed_records = [record for record in records if record.record_type == "jsonl_line"]
    error_records = [record for record in records if record.record_type == "parser_error"]

    assert len(parsed_records) == 2
    assert len(error_records) == 1
    assert error_records[0].structured_data["line_number"] == 2
    assert parser.parser_errors

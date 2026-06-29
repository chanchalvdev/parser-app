import json
import tempfile
from pathlib import Path

from app.parsers import parse_json, parse_jsonl


def test_parse_json_object():
    payload = '{"name":"sample","count":1}'
    with tempfile.TemporaryDirectory() as d:
        path = Path(d) / "sample.json"
        path.write_text(payload, encoding="utf-8")
        parsed = parse_json(str(path))

    assert json.loads(parsed["content"])["name"] == "sample"
    assert parsed["metadata"]["json"] is True


def test_parse_json_invalid_payload_falls_back_to_text():
    with tempfile.TemporaryDirectory() as d:
        path = Path(d) / "bad.json"
        path.write_text("{invalid}", encoding="utf-8")
        parsed = parse_json(str(path))

    assert parsed["content"] == "{invalid}"


def test_parse_jsonl_extracts_lines():
    with tempfile.TemporaryDirectory() as d:
        path = Path(d) / "sample.jsonl"
        path.write_text('{"a":1}\\n{"b":2}\\n', encoding="utf-8")
        parsed = parse_jsonl(str(path))

    assert '"a": 1' in parsed["content"]
    assert '"b": 2' in parsed["content"]
    assert parsed["metadata"]["lines"] == 2


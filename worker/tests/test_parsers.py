import tempfile
from pathlib import Path

from app.parsers import detect_kind, parse_jsonl, _is_texty


def test_detect_kind():
    assert detect_kind("foo.TXT") == "txt"
    assert detect_kind("archive.tar.gz") == "gz"


def test_parse_jsonl():
    payload = '{"a":1}\n{"b":2}\n'
    with tempfile.TemporaryDirectory() as d:
        path = Path(d) / "sample.jsonl"
        path.write_text(payload, encoding="utf-8")
        parsed = parse_jsonl(str(path))
        assert "a" in parsed["content"]


def test_texty_detection():
    with tempfile.NamedTemporaryFile(mode="w", delete=False) as f:
        f.write("hello")
        path = f.name
    assert _is_texty(path)


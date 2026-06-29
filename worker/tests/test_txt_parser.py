import tempfile
from pathlib import Path

from app.parsers import parse_txt


def test_parse_txt():
    with tempfile.TemporaryDirectory() as d:
        path = Path(d) / "sample.txt"
        path.write_text("hello world\nline 2", encoding="utf-8")

        parsed = parse_txt(str(path))

        assert parsed["content"] == "hello world\nline 2"
        assert parsed["metadata"]["char_count"] == 13


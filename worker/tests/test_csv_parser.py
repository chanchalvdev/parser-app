import tempfile
from pathlib import Path

from app.parsers import parse_csv


def test_parse_csv():
    with tempfile.TemporaryDirectory() as d:
        path = Path(d) / "sample.csv"
        lines = "\n".join([f"col{i},value{i}" for i in range(3)])
        path.write_text(lines, encoding="utf-8")

        parsed = parse_csv(str(path))

        assert "col0,value0" in parsed["content"]
        assert parsed["metadata"]["rows"] == 3


def test_parse_csv_limits_rows():
    with tempfile.TemporaryDirectory() as d:
        path = Path(d) / "sample.csv"
        lines = "\n".join([f"a,b{i}" for i in range(700)])
        path.write_text(lines, encoding="utf-8")

        parsed = parse_csv(str(path))

        assert parsed["metadata"]["rows"] <= 501


from __future__ import annotations

import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from app.processing.file_detector import FileDetectionResult
from app.parsers.csv_parser import CsvParser


def _detection_result() -> FileDetectionResult:
    return FileDetectionResult(
        extension="csv",
        mime_type="text/csv",
        detected_file_type="csv",
        is_archive=False,
        is_text_like=True,
        is_binary=False,
        parser_hint="csv",
    )


def _metadata() -> dict[str, str]:
    return {"tenant_id": "tenant-123", "file_id": "file-abc", "job_id": "job-xyz"}


def test_csv_parser_emits_row_per_record(tmp_path: Path) -> None:
    source = tmp_path / "sample.csv"
    source.write_text(
        "id,user,src_ip\n1,alice@corp.com,10.0.0.5\n2,bob@corp.com,10.0.0.6\n",
        encoding="utf-8",
    )

    records = list(CsvParser().parse(str(source), _metadata(), _detection_result()))

    assert len(records) == 2
    assert records[0].record_type == "csv_row"
    assert records[0].structured_data["user"] == "alice@corp.com"
    assert records[0].structured_data["src_ip"] == "10.0.0.5"
    assert "10.0.0.5" in records[0].extracted_entities["ipv4"]
    assert "alice@corp.com" in records[0].extracted_entities["emails"]


def test_csv_parser_streams_large_files_without_row_cap(tmp_path: Path) -> None:
    source = tmp_path / "big.csv"
    rows = 5000
    with source.open("w", newline="") as handle:
        handle.write("id,value\n")
        for i in range(rows):
            handle.write(f"{i},v{i}\n")

    count = sum(1 for _ in CsvParser().parse(str(source), _metadata(), _detection_result()))

    assert count == rows  # no truncation / row cap


def test_csv_parser_skips_blank_rows_and_pads_short_rows(tmp_path: Path) -> None:
    source = tmp_path / "ragged.csv"
    source.write_text("a,b,c\n1,2,3\n\n4,5\n", encoding="utf-8")

    records = list(CsvParser().parse(str(source), _metadata(), _detection_result()))

    assert len(records) == 2  # blank row skipped
    assert records[1].structured_data["a"] == "4"
    assert records[1].structured_data["c"] == ""  # missing column padded

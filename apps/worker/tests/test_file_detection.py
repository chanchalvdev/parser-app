from __future__ import annotations

import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from app.processing.file_detector import detect_file_type


def _write(tmp_path: Path, name: str, data: bytes) -> str:
    path = tmp_path / name.replace("/", "_")
    path.write_bytes(data)
    return str(path)


def test_detect_txt_file(tmp_path):
    path = _write(tmp_path, "report.TXT", b"plain text content\n")
    result = detect_file_type(path, "report.TXT")
    assert result.extension == "txt"
    assert result.is_text_like
    assert not result.is_archive
    assert not result.is_binary


def test_detect_zip_archive(tmp_path):
    import zipfile

    path = tmp_path / "bundle.zip"
    with zipfile.ZipFile(path, "w") as zf:
        zf.writestr("a.txt", "x")
    result = detect_file_type(str(path), "bundle.zip")
    assert result.is_archive
    assert result.detected_file_type in {"zip", "archive"}


def test_detect_tar_gz_archive(tmp_path):
    import gzip

    path = tmp_path / "bundle.tar.gz"
    path.write_bytes(gzip.compress(b"payload"))
    result = detect_file_type(str(path), "bundle.tar.gz")
    assert result.is_archive


def test_detect_binary_file(tmp_path):
    path = _write(tmp_path, "blob.bin", b"\x00\x01\x02\x03binary\x00data")
    result = detect_file_type(path, "blob.bin")
    assert result.is_binary
    assert not result.is_text_like


def test_detect_file_without_extension(tmp_path):
    path = _write(tmp_path, "no_ext", b"just text\n")
    result = detect_file_type(path, "no_ext")
    assert result.extension == ""
    assert not result.is_archive

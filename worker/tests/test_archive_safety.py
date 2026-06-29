import tempfile
import zipfile
from pathlib import Path

import pytest

from app.extract import extract_archive, is_archive, safe_extract_path


def test_safe_extract_path_blocks_traversal():
    with tempfile.TemporaryDirectory() as d:
        with pytest.raises(ValueError):
            safe_extract_path(d, "../evil.txt")
        with pytest.raises(ValueError):
            safe_extract_path(d, "/etc/passwd")


def test_safe_extract_path_allows_normal_paths():
    with tempfile.TemporaryDirectory() as d:
        normalized = safe_extract_path(d, "nested/file.txt")
        assert normalized.startswith(d)


def test_extract_zip_blocks_traversal_members():
    with tempfile.TemporaryDirectory() as d:
        zip_path = Path(d) / "malicious.zip"
        with zipfile.ZipFile(zip_path, "w") as zf:
            zf.writestr("../evil.txt", "bad")

        with pytest.raises(ValueError):
            extract_archive(str(zip_path), Path(d) / "out")


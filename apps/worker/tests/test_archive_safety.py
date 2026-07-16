from __future__ import annotations

import sys
import zipfile
from pathlib import Path

import pytest

ROOT = Path(__file__).resolve().parents[2]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from app.extractors.base_extractor import (
    ArchiveLimits,
    ArchiveLimitsExceededError,
    BaseArchiveExtractor,
    CorruptArchiveError,
    ExtractionError,
    safe_relative_path,
)
from app.extractors.zip_extractor import ZipExtractor


def test_safe_relative_path_blocks_traversal():
    with pytest.raises(CorruptArchiveError):
        safe_relative_path("../evil.txt")
    with pytest.raises(CorruptArchiveError):
        safe_relative_path("nested/../../evil.txt")


def test_safe_relative_path_blocks_absolute_paths():
    with pytest.raises(CorruptArchiveError):
        safe_relative_path("/etc/passwd")


def test_safe_relative_path_blocks_windows_drive_paths():
    with pytest.raises(CorruptArchiveError):
        safe_relative_path("C:\\windows\\system32\\evil.dll")


def test_safe_relative_path_normalizes_safe_members():
    assert safe_relative_path("dir/./file.txt") == "dir/file.txt"
    assert safe_relative_path("dir\\sub\\file.txt") == "dir/sub/file.txt"
    assert safe_relative_path("") == ""
    assert safe_relative_path(".") == ""


def test_enforce_limits_rejects_too_many_files():
    limits = ArchiveLimits(max_extracted_files=2, max_extracted_size_mb=10, max_expansion_ratio=100.0)
    with pytest.raises(ArchiveLimitsExceededError):
        BaseArchiveExtractor.enforce_limits(
            file_count=3, extracted_bytes=10, archive_size_bytes=10, limits=limits
        )


def test_enforce_limits_rejects_oversized_output():
    limits = ArchiveLimits(max_extracted_files=100, max_extracted_size_mb=1, max_expansion_ratio=1000.0)
    with pytest.raises(ArchiveLimitsExceededError):
        BaseArchiveExtractor.enforce_limits(
            file_count=1,
            extracted_bytes=2 * 1024 * 1024,
            archive_size_bytes=2 * 1024 * 1024,
            limits=limits,
        )


def test_enforce_limits_rejects_zip_bomb_expansion_ratio():
    limits = ArchiveLimits(max_extracted_files=100, max_extracted_size_mb=100, max_expansion_ratio=5.0)
    with pytest.raises(ArchiveLimitsExceededError):
        BaseArchiveExtractor.enforce_limits(
            file_count=1, extracted_bytes=1000, archive_size_bytes=10, limits=limits
        )


def test_enforce_limits_allows_within_bounds():
    limits = ArchiveLimits(max_extracted_files=10, max_extracted_size_mb=10, max_expansion_ratio=20.0)
    BaseArchiveExtractor.enforce_limits(
        file_count=1, extracted_bytes=100, archive_size_bytes=100, limits=limits
    )


def test_zip_extractor_rejects_traversal_member(tmp_path):
    archive_path = tmp_path / "evil.zip"
    with zipfile.ZipFile(archive_path, "w") as zf:
        zf.writestr("../escape.txt", "malicious")

    destination = tmp_path / "out"
    destination.mkdir()

    with pytest.raises(ExtractionError):
        ZipExtractor().extract(str(archive_path), str(destination))

    assert not (tmp_path / "escape.txt").exists()


def test_zip_extractor_extracts_safe_members(tmp_path):
    archive_path = tmp_path / "safe.zip"
    with zipfile.ZipFile(archive_path, "w") as zf:
        zf.writestr("dir/hello.txt", "hello world")

    destination = tmp_path / "out"
    destination.mkdir()

    extracted = ZipExtractor().extract(str(archive_path), str(destination))

    assert len(extracted) == 1
    assert extracted[0].relative_path == "dir/hello.txt"
    assert (destination / "dir" / "hello.txt").read_text() == "hello world"

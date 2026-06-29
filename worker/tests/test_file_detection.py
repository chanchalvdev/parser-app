import pytest

from app.parsers import detect_kind
from app.extract import is_archive


def test_detect_kind_by_extension():
    assert detect_kind("report.TXT") == "txt"
    assert detect_kind("archive.tar.gz") == "gz"
    assert detect_kind("no_ext") == ""
    assert detect_kind("photo.jpeg") == "jpeg"


def test_is_archive_extensions():
    assert is_archive("bundle.zip")
    assert is_archive("bundle.tar")
    assert is_archive("bundle.tar.gz")
    assert is_archive("bundle.7z")
    assert not is_archive("document.txt")


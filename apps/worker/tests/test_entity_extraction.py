from __future__ import annotations

import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from app.parsers.txt_parser import _extract_entities


def test_extract_entities_returns_expected_groups():
    text = (
        "Contact admin@example.com and backup@example.net. "
        "Connect 192.168.10.12 and 10.0.0.5. "
        "See https://example.com/report?x=1 at 2026-07-15 10:30:00 "
        "with MD5 " + "d" * 32 + " and SHA256 " + "a" * 64
    )

    entities = _extract_entities(text)

    assert "admin@example.com" in entities["emails"]
    assert "backup@example.net" in entities["emails"]
    assert "192.168.10.12" in entities["ipv4"]
    assert "10.0.0.5" in entities["ipv4"]
    assert any(url.startswith("https://example.com/report") for url in entities["urls"])
    assert "2026-07-15 10:30:00" in entities["timestamps"]
    assert "d" * 32 in entities["md5_like_hashes"]
    assert "a" * 64 in entities["sha256_like_hashes"]


def test_extract_entities_dedupes_case_insensitively():
    text = "ping Admin@Example.com then admin@example.com again"
    entities = _extract_entities(text)
    assert len(entities["emails"]) == 1


def test_extract_entities_empty_text():
    entities = _extract_entities("")
    assert entities["emails"] == []
    assert entities["ipv4"] == []
    assert entities["urls"] == []

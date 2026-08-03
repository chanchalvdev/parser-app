from __future__ import annotations

import hashlib
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from app.parsers.infostealer import parse_credential_blocks, stealer_category
from app.parsers.txt_parser import TxtParser, _extract_entities
from app.processing.file_detector import FileDetectionResult


def _txt_detection() -> FileDetectionResult:
    return FileDetectionResult(
        extension="txt",
        mime_type="text/plain",
        detected_file_type="txt",
        is_archive=False,
        is_text_like=True,
        is_binary=False,
        parser_hint="text",
    )


def test_stealer_category_by_filename_and_dir():
    assert stealer_category("logs/Passwords.txt") == "stealer_password"
    assert stealer_category("PC-01/UserInformation.txt") == "system_info"
    assert stealer_category("PC-01/Cookies/chrome.txt") == "cookies"
    assert stealer_category("PC-01/Autofills/data.txt") == "autofill"
    assert stealer_category("random/notes.txt") is None


def test_credential_blocks_normalize_varied_prefixes_and_leetspeak():
    text = (
        "SOFT: Chrome\nURL: https://mail.example.com\nUSER: alice\nPASS: hunter2\n"
        "\n"
        "Soft: Firefox\nHost: https://bank.example.com\nLogin: bob@example.com\nPassword: s3cr3t\n"
        "\n"
        "Browser: Edge\nHostname: https://shop.example.com\nUsername: carol\nP455W0RD: pa55!\n"
    )
    recs = parse_credential_blocks(text)
    assert len(recs) == 3
    assert recs[0]["application"] == "Chrome"
    assert recs[0]["url"] == "https://mail.example.com"
    assert recs[0]["username"] == "alice"
    assert recs[2]["username"] == "carol"
    # Leetspeak "P455W0RD" recognized as password and hashed, not plaintext.
    assert recs[2]["secret_hash"] == hashlib.sha256(b"pa55!").hexdigest()
    for r in recs:
        assert "secret_hash" not in r or len(r["secret_hash"]) == 64


def test_credential_blocks_split_without_blank_lines():
    text = (
        "URL: https://a.com\nLogin: u1\nPassword: p1\n"
        "URL: https://b.com\nLogin: u2\nPassword: p2\n"
    )
    recs = parse_credential_blocks(text)
    assert len(recs) == 2
    assert {r["url"] for r in recs} == {"https://a.com", "https://b.com"}


def test_extract_entities_adds_new_ioc_types_backward_compatible():
    text = (
        "hash1 " + "b" * 40 + " CVE-2021-44228 wallet 1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2 "
        "md5 " + "d" * 32 + " sha256 " + "a" * 64
    )
    ents = _extract_entities(text)
    # New keys present
    assert "b" * 40 in ents["sha1_like_hashes"]
    assert "CVE-2021-44228" in ents["cve"]
    assert "1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2" in ents["bitcoin_addresses"]
    # Existing keys still present and correct (no regression)
    assert "d" * 32 in ents["md5_like_hashes"]
    assert "a" * 64 in ents["sha256_like_hashes"]
    assert set(["ipv4", "emails", "urls", "domains", "timestamps"]).issubset(ents.keys())


def test_extract_entities_captures_credentials_hashed():
    text = "SOFT: Chrome\nURL: https://x.com\nUSER: dave\nPASS: topsecret\n"
    ents = _extract_entities(text)
    assert "credentials" in ents
    assert ents["credentials"][0]["username"] == "dave"
    assert ents["credentials"][0]["secret_hash"] == hashlib.sha256(b"topsecret").hexdigest()


def test_txt_parser_wires_stealer_routing_and_assembles_blocks(tmp_path):
    # Small line-mode Passwords.txt: blocks only assemble via the wired path.
    source = tmp_path / "Passwords.txt"
    source.write_text(
        "SOFT: Chrome\nURL: https://mail.example.com\nUSER: alice\nPASS: hunter2\n"
        "SOFT: Firefox\nURL: https://bank.example.com\nUSER: bob\nPASS: s3cret\n",
        encoding="utf-8",
    )
    metadata = {
        "tenant_id": "t1",
        "file_id": "f1",
        "job_id": "j1",
        "original_name": "Passwords.txt",
    }
    records = list(TxtParser().parse(str(source), metadata, _txt_detection()))

    # Every record is tagged with the routed category.
    assert all(r.structured_data.get("stealer_category") == "stealer_password" for r in records)

    # A consolidated credential record assembled the multi-line blocks.
    cred_records = [r for r in records if r.record_type == "stealer_credential"]
    assert len(cred_records) == 1
    creds = cred_records[0].extracted_entities["credentials"]
    assert len(creds) == 2
    assert {c["username"] for c in creds} == {"alice", "bob"}
    assert all(len(c["secret_hash"]) == 64 for c in creds)


def test_txt_parser_non_stealer_file_no_credential_record(tmp_path):
    source = tmp_path / "notes.txt"
    source.write_text("just some log line\nanother line\n", encoding="utf-8")
    metadata = {"tenant_id": "t1", "file_id": "f1", "job_id": "j1", "original_name": "notes.txt"}
    records = list(TxtParser().parse(str(source), metadata, _txt_detection()))
    assert all(r.record_type != "stealer_credential" for r in records)
    assert all("stealer_category" not in r.structured_data for r in records)

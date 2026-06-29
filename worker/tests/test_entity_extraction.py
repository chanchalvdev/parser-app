from app.parsers import extract_entities


def test_extract_entities_returns_groups():
    text = (
        "Contact admin@example.com and backup@example.net. "
        "Connect 192.168.10.12 and 10.0.0.5. "
        "See https://example.com/report?x=1 and file hash deadbeefdeadbeefdeadbeefdeadbeefab and "
        "malware SHA1 deadbeefdeadbeefdeadbeefdeadbeefdeadbeefde"
    )

    entities = extract_entities(text)

    assert "admin@example.com" in entities["emails"]
    assert "192.168.10.12" in entities["ip_addresses"]
    assert "https://example.com/report?x=1" in entities["urls"]
    assert "example.com" in entities["domains"]
    assert "deadbeefdeadbeefdeadbeefdeadbeefab" in entities["hashes"]


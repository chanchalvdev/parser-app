# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project uses
[Semantic Versioning](https://semver.org/).

## [1.0.0] - 2026-10-24

First tagged release.

### Added
- Presigned direct-to-MinIO uploads (multi-GB) through an nginx proxy tuned for large bodies.
- Recursive archive extraction for ZIP, RAR4/RAR5, TAR, TAR.GZ/TGZ and 7Z, with password
  support (`PASSWORD_REQUIRED` → unlock → resume) and depth, count, size and expansion-ratio guards.
- Parsers for TXT/LOG, CSV, JSON/JSONL and generic text. XML, XLSX and PDF are registered scaffolds.
- IOC extraction: IPv4, emails, URLs, domains, timestamps, MD5/SHA1/SHA256, CVE IDs, Bitcoin addresses.
- Infostealer log handling: filename/directory routing, a credential-block grammar across
  stealer families, and SHA-256 password hashing.
- Categorized job view in the UI (Credentials, Cookies, Autofill, System Info, …).
- OpenSearch indexing of records and files, the search API, a dashboard API and the admin settings API.
- Architecture diagram, demo recording, MIT license and this changelog.

### Fixed
- Stealer-log passwords are hashed again; one commit had stored them in plaintext.
- Raw `PASS:` lines in password dumps are redacted before they are persisted or indexed.
- Every multi-file `.7z` failed CRC checks after its first member.
- AES-encrypted ZIPs (the 7-Zip/WinZip default) failed after the correct password was given.
- RAR extraction never worked in the worker image: no `unar`/`unrar` backend was installed.
- The ingestion queue was consumed LIFO (`RPUSH` + `BRPOP`) and is now FIFO.
- Credential entities rendered as `[object Object]` in the UI.
- `.env.example` pointed the UI at an unpublished API port, so a fresh clone could not connect.
- `create_sample_archives.py` resolved fixtures under `infra/` and always failed.
- CI only triggered on `main`; the default branch is `master`.

[1.0.0]: https://github.com/chanchalvdev/parser-app/releases/tag/v1.0.0

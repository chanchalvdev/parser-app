# Test Fixtures

This directory contains sample payloads and local archive fixtures for manual and parser testing.

## Fixture files

- `sample.txt`
- `sample.log`
- `sample.csv`
- `sample.json`
- `sample.jsonl`
- `nested_archive_source/readme.txt`
- `nested_archive_source/logs/app.log`
- `nested_archive_source/data/users.csv`

## Generate sample archives

The following command creates archive fixtures in this directory:

```bash
python infra/scripts/create_sample_archives.py
```

Optional flags:

- `--fixtures-dir` custom fixture directory (defaults to `tests/fixtures`)
- `--password` password for `password_sample.zip` (default: `changeme`)

Generated outputs:

- `sample.zip`
- `sample_nested.zip`
- `sample.tar.gz`
- `sample.7z` (created when `py7zr` is available)
- `password_sample.zip` (created when a zip-password backend is available, e.g. `pyzipper`)

## Upload sample files from the UI

1. Start the local stack and open the Upload page: `http://localhost:5173/upload`.
2. Choose a file from this directory.
3. Optional: enable **password provided** when uploading `password_sample.zip`.
4. Click upload, wait for completion, then open the returned Job page.

## Upload sample files via curl

1) Initiate upload:

```bash
SIZE_BYTES=$(wc -c < tests/fixtures/sample.zip)
INIT_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/uploads/initiate \
  -H "Content-Type: application/json" \
  -d "{\"file_name\": \"sample.zip\", \"content_type\": \"application/zip\", \"size_bytes\": ${SIZE_BYTES}, \"password_provided\": false}")
```

2) Extract upload metadata:

```bash
UPLOAD_ID=$(python - <<'PY'
import json,sys
print(json.load(sys.stdin)['upload_id'])
PY <<< "$INIT_RESPONSE")
UPLOAD_URL=$(python - <<'PY'
import json,sys
print(json.load(sys.stdin)['upload_url'])
PY <<< "$INIT_RESPONSE")
```

3) Upload bytes directly to presigned URL:

```bash
curl -X PUT "$UPLOAD_URL" --upload-file tests/fixtures/sample.zip
```

4) Complete upload in API:

```bash
curl -X POST http://localhost:8080/api/v1/uploads/complete \
  -H "Content-Type: application/json" \
  -d "{\"upload_id\": \"$UPLOAD_ID\"}"
```

Repeat using other fixture files (`sample.log`, `sample.csv`, `sample.json`, `sample.jsonl`) and archive files as needed.

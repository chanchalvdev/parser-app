# infra/scripts/

Operational scripts for local environment bootstrap and reset flows.

Examples:
- Local data seeding
- Local cluster reset helpers
- Debug utilities

## Sample fixture generator

Generate parser sample archives from `tests/fixtures`:

```bash
python infra/scripts/create_sample_archives.py
```

Optional args:

- `--fixtures-dir` to target a different fixture directory.
- `--password` to set the password for `password_sample.zip` (default: `changeme`).

Creates:
- `tests/fixtures/sample.zip`
- `tests/fixtures/sample_nested.zip`
- `tests/fixtures/sample.tar.gz`
- `tests/fixtures/sample.7z` (requires `py7zr`)
- `tests/fixtures/password_sample.zip` (requires password-capable zip backend, e.g. `pyzipper`)

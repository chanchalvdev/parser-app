# Glossary

- **Archive Password Reference**: Stored password metadata used by worker to retry protected archives.
- **ParsedRecord**: Normalized record produced by worker parser and indexed for search.
- **Search Index Status**: Per-file/job indicator for indexing lifecycle.
- **Parser Error**: Recoverable record-level parsing issue recorded without failing entire job.
- **Requeue**: Re-adding a job message to queue for retry without altering original lineage.
- **Wrong Password**: Password supplied for archive extraction but decryption/test failed.
- **Self-improving harness**: Process loop that updates memory files and workflows after each implementation/release.

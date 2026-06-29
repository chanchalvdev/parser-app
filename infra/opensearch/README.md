# OpenSearch Bootstrap

This directory contains local OpenSearch bootstrap scripts.

Run:

```bash
./create-indexes.sh
```

Required tool: `curl`.

Optional environment overrides (read from `.env` if present):

- `OPENSEARCH_URL` (default: `http://localhost:9200`)
- `OPENSEARCH_SERVICE_URL` (container service URL for API/worker, typically `http://opensearch:9200`)
- `OPENSEARCH_USER` / `OPENSEARCH_PASSWORD` (for protected clusters)

Created indexes:

- `parsed-records`
- `files`

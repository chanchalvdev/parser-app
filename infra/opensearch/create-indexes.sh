#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ENV_FILE="$ROOT_DIR/.env"

if [ -f "$ENV_FILE" ]; then
  set -a
  # shellcheck source=/dev/null
  . "$ENV_FILE"
  set +a
fi

OPENSEARCH_URL="${OPENSEARCH_URL:-http://localhost:9200}"
OPENSEARCH_USER="${OPENSEARCH_USER:-}"
OPENSEARCH_PASSWORD="${OPENSEARCH_PASSWORD:-}"
ACTIVE_OPENSEARCH_URL=""

CURL_AUTH=()
if [ -n "${OPENSEARCH_USER}" ] && [ -n "${OPENSEARCH_PASSWORD}" ]; then
  CURL_AUTH=("-u" "${OPENSEARCH_USER}:${OPENSEARCH_PASSWORD}")
fi

HAS_JQ=0
if command -v jq >/dev/null 2>&1; then
  HAS_JQ=1
fi

require_endpoint() {
  local target_url="$1"
  local status
  local attempt=1
  local max_attempts=24

  while [ "$attempt" -le "$max_attempts" ]; do
    status=$(curl -sS -m 5 -o /dev/null -w "%{http_code}" "${CURL_AUTH[@]}" "$target_url/" || true)
    if [ "$status" = "200" ] || [ "$status" = "401" ] || [ "$status" = "403" ]; then
      echo "$target_url"
      return 0
    fi

    if [ "$attempt" -ge "$max_attempts" ]; then
      break
    fi

    echo "Waiting for OpenSearch at $target_url (attempt $attempt/$max_attempts, HTTP $status)..."
    sleep 2
    attempt=$((attempt + 1))
  done

  echo "OpenSearch is not reachable at $target_url (HTTP $status)." >&2
  return 1
}

resolve_opensearch_url() {
  local candidates=("$1")
  local candidate
  local idx

  if [[ "$1" == *"opensearch"* && "$1" != *"localhost"* && "$1" != *"127.0.0.1"* ]]; then
    candidates+=("http://localhost:9200" "http://127.0.0.1:9200")
  fi

  for idx in "${!candidates[@]}"; do
    candidate="${candidates[$idx]}"
    if require_endpoint "$candidate"; then
      ACTIVE_OPENSEARCH_URL="$candidate"
      if [ "$candidate" != "$1" ]; then
        echo "Using alternate OpenSearch endpoint for local bootstrap: $candidate"
      fi
      return 0
    fi
  done

  echo "No reachable OpenSearch endpoint found."
  return 1
}

if ! resolve_opensearch_url "$OPENSEARCH_URL"; then
  exit 1
fi

create_or_skip_index() {
  local index_name="$1"
  local payload="$2"
  local status
  local existing_mapping

  status=$(curl -sS -m 5 -o /dev/null -w "%{http_code}" "${CURL_AUTH[@]}" -X GET "$ACTIVE_OPENSEARCH_URL/$index_name" || true)
  if [ "$status" = "200" ]; then
    existing_mapping=$(curl -sS -m 5 "${CURL_AUTH[@]}" "$ACTIVE_OPENSEARCH_URL/$index_name/_mapping" || true)
    if [ "$HAS_JQ" -eq 1 ]; then
      file_id_type=$(echo "$existing_mapping" | jq -r ".[\"$index_name\"].mappings.properties.file_id.type // \"\"")
      record_type_type=$(echo "$existing_mapping" | jq -r ".[\"$index_name\"].mappings.properties.record_type.type // \"\"")
      if [ "$file_id_type" = "text" ] || [ "$record_type_type" = "text" ]; then
        echo "Index $index_name has legacy text mapping; recreating for compatibility."
        curl -sS -m 5 -o /dev/null "${CURL_AUTH[@]}" -X DELETE "$ACTIVE_OPENSEARCH_URL/$index_name"
        status="404"
      else
        echo "Index already exists: $index_name"
        return 0
      fi
    else
      normalized_mapping=$(echo "$existing_mapping" | tr -d '\n')
      if echo "$normalized_mapping" | grep -Eq '\"file_id\"[[:space:]]*:[[:space:]]*\{[[:space:]]*\"type\"[[:space:]]*:[[:space:]]*\"text\"|\"record_type\"[[:space:]]*:[[:space:]]*\{[[:space:]]*\"type\"[[:space:]]*:[[:space:]]*\"text\"'; then
        echo "Index $index_name has legacy text mapping; recreating for compatibility."
        curl -sS -m 5 -o /dev/null "${CURL_AUTH[@]}" -X DELETE "$ACTIVE_OPENSEARCH_URL/$index_name"
        status="404"
      else
        echo "Index already exists: $index_name"
        return 0
      fi
    fi
  fi

  if [ "$status" != "404" ]; then
    echo "Could not check index $index_name (HTTP $status)." >&2
    exit 1
  fi

  echo "Creating index: $index_name"
  status=$(curl -sS -o /dev/null -w "%{http_code}" \
    "${CURL_AUTH[@]}" \
    -X PUT "$ACTIVE_OPENSEARCH_URL/$index_name" \
    -H "Content-Type: application/json" \
    -d "$payload")
  if [ "$status" != "200" ] && [ "$status" != "201" ]; then
    echo "Failed to create index $index_name (HTTP $status)." >&2
    exit 1
  fi
}

PARSED_RECORDS_MAPPING='
{
  "mappings": {
    "properties": {
      "tenant_id": { "type": "keyword" },
      "file_id": { "type": "keyword" },
      "job_id": { "type": "keyword" },
      "source_file_name": { "type": "keyword" },
      "archive_path": { "type": "keyword" },
      "record_type": { "type": "keyword" },
      "content": { "type": "text" },
      "structured_data": {
        "type": "object",
        "enabled": true
      },
      "entities": {
        "properties": {
          "ip_addresses": { "type": "ip" },
          "emails": { "type": "keyword" },
          "urls": { "type": "keyword" },
          "domains": { "type": "keyword" },
          "hashes": { "type": "keyword" }
        }
      },
      "event_timestamp": { "type": "date" },
      "created_at": { "type": "date" }
    }
  }
}
'

FILES_MAPPING='
{
  "mappings": {
    "properties": {
      "tenant_id": { "type": "keyword" },
      "file_id": { "type": "keyword" },
      "parent_file_id": { "type": "keyword" },
      "original_name": {
        "type": "text",
        "fields": {
          "keyword": { "type": "keyword" }
        }
      },
      "extension": { "type": "keyword" },
      "detected_file_type": { "type": "keyword" },
      "processing_status": { "type": "keyword" },
      "size_bytes": { "type": "long" },
      "sha256_hash": { "type": "keyword" },
      "created_at": { "type": "date" }
    }
  }
}
'

create_or_skip_index "parsed-records" "$PARSED_RECORDS_MAPPING"
create_or_skip_index "files" "$FILES_MAPPING"

echo "OpenSearch index bootstrap complete."

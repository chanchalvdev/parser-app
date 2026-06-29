#!/usr/bin/env bash

set -euo pipefail

API_BASE="${API_BASE_URL:-http://localhost:8088/api/v1}"
API_KEY="${API_KEY:-}"
JOB_TIMEOUT_SECONDS="${JOB_TIMEOUT_SECONDS:-180}"
POLL_INTERVAL_SECONDS="${POLL_INTERVAL_SECONDS:-2}"
DUMMY_FILE_PATH="${DUMMY_FILE_PATH:-/tmp/e2e-dummy-upload.txt}"
DUMMY_FILE_NAME="${DUMMY_FILE_NAME:-local-e2e-test.txt}"
SEARCH_ASSERTIONS="${SEARCH_ASSERTIONS:-dummy-local-test alice@example.com 203.0.113.42}"

if ! command -v curl >/dev/null 2>&1; then
  echo "[e2e] curl is required"
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "[e2e] jq is required"
  exit 1
fi

if [ ! -f "${DUMMY_FILE_PATH}" ]; then
  cat > "${DUMMY_FILE_PATH}" <<'EOF'
dummy-local-test
This is a local smoke-test file.
Contact: alice@example.com
URL: https://example.com/local-check
IP: 203.0.113.42
TOKEN: abc1234-local-test-token
EOF
fi

if [ ! -s "${DUMMY_FILE_PATH}" ]; then
  echo "[e2e] dummy file is empty: ${DUMMY_FILE_PATH}"
  exit 1
fi

FILE_SIZE="$(wc -c < "${DUMMY_FILE_PATH}")"

log() {
  echo "[e2e] $*"
}

api_request() {
  local method="$1"
  local endpoint="$2"
  local body="${3:-}"

  local -a auth_header=()
  if [ -n "${API_KEY}" ]; then
    auth_header=(-H "X-API-Key: ${API_KEY}")
  fi

  if [ "${method}" = "GET" ]; then
    curl -sS -X "${method}" \
      "${auth_header[@]+"${auth_header[@]}"}" \
      "${API_BASE}${endpoint}"
  else
    curl -sS -X "${method}" \
      "${auth_header[@]+"${auth_header[@]}"}" \
      -H "Content-Type: application/json" \
      -d "${body}" \
      "${API_BASE}${endpoint}"
  fi
}

log "Creating upload session for ${DUMMY_FILE_NAME}"
init_payload="$(jq -cn \
  --arg name "${DUMMY_FILE_NAME}" \
  --arg content_type "text/plain" \
  --argjson size "${FILE_SIZE}" \
  '{file_name: $name, content_type: $content_type, size_bytes: $size, password_provided: false}')"

init_resp="$(api_request POST /uploads/initiate "${init_payload}")"
upload_id="$(echo "${init_resp}" | jq -r '.upload_id // empty')"
upload_url="$(echo "${init_resp}" | jq -r '.upload_url // empty')"

if [ -z "${upload_id}" ] || [ -z "${upload_url}" ]; then
  echo "[e2e] failed to initiate upload"
  echo "${init_resp}"
  exit 1
fi

log "Upload initiated: ${upload_id}"

log "Uploading dummy file to presigned URL"
if ! curl -sS -X PUT -T "${DUMMY_FILE_PATH}" "${upload_url}" >/tmp/e2e-upload-put.log; then
  echo "[e2e] upload to object storage failed"
  cat /tmp/e2e-upload-put.log
  exit 1
fi

log "Completing upload"
complete_payload="$(jq -cn --arg upload_id "${upload_id}" '{upload_id: $upload_id}')"
complete_resp="$(api_request POST /uploads/complete "${complete_payload}")"
file_id="$(echo "${complete_resp}" | jq -r '.file_id // empty')"
job_id="$(echo "${complete_resp}" | jq -r '.job_id // empty')"
status="$(echo "${complete_resp}" | jq -r '.status // empty')"

if [ -z "${file_id}" ] || [ -z "${job_id}" ]; then
  echo "[e2e] failed to complete upload"
  echo "${complete_resp}"
  exit 1
fi

log "Upload complete. file_id=${file_id}, job_id=${job_id}, status=${status}"

deadline=$((SECONDS + JOB_TIMEOUT_SECONDS))
final_status=""
while [ "${SECONDS}" -lt "${deadline}" ]; do
  job_resp="$(api_request GET "/jobs/${job_id}")"
  final_status="$(echo "${job_resp}" | jq -r '.status // empty')"
  current_stage="$(echo "${job_resp}" | jq -r '.current_stage // empty')"
  progress="$(echo "${job_resp}" | jq -r '.progress_percent // empty')"

  log "job status: ${final_status:-unknown}, stage: ${current_stage:-unknown}, progress: ${progress:-n/a}%"

  case "${final_status}" in
    completed|failed|password_required|wrong_password|canceled|cancelled)
      break
      ;;
  esac
  sleep "${POLL_INTERVAL_SECONDS}"
done

if [ "${final_status}" != "completed" ]; then
  log "job did not complete successfully: ${final_status:-unknown}"
  echo "${job_resp}"
  exit 1
fi

log "Job completed successfully"

records_resp="$(api_request GET "/files/${file_id}/records?page=1&page_size=10")"
record_total="$(echo "${records_resp}" | jq -r '.total // 0')"
record_count="$(echo "${records_resp}" | jq -r '.records | length')"

log "Parsed records for file ${file_id}: total=${record_total}, returned=${record_count}"

if [ "${record_total}" -le 0 ]; then
  echo "[e2e] expected parsed records but found zero"
  echo "${records_resp}"
  exit 1
fi

for term in ${SEARCH_ASSERTIONS}; do
  encoded_term="$(printf '%s' "${term}" | jq -sRr @uri)"
  results="$(api_request GET "/search?q=${encoded_term}")"
  total_hits="$(echo "${results}" | jq -r '.total // 0')"
  log "search q=${term}: hits=${total_hits}"
  if [ "${total_hits}" -eq 0 ]; then
    echo "[e2e] search assertion failed for term: ${term}"
    echo "${results}"
    exit 1
  fi
done

dashboard_resp="$(api_request GET /dashboard/summary)"
dashboard_total="$(echo "${dashboard_resp}" | jq -r '.total_parsed_records // 0')"

log "Dashboard summary total_parsed_records=${dashboard_total}"
log "Local E2E smoke test passed."

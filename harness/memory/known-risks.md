# Known Risks

## Open Risk

- Search index mapping drift can delay search visibility while DB insert succeeds.
  - **Date identified:** 2026-06-19
  - **Owner:** Search Agent
  - **Mitigation:** Keep DB insert independent; track `search_index_status`; rerun search init as needed.

- Parser output can grow unexpectedly on binary-like files.
  - **Date identified:** 2026-06-19
  - **Owner:** Worker Python Agent
  - **Mitigation:** Apply parser size caps, bounded error logs, and parser-error summaries.

- JSONL malformed input in very large files can produce high parser_error volume.
  - **Date identified:** 2026-06-19
  - **Owner:** Worker Python Agent
  - **Mitigation:** Add rate-limiting/error-summarization with controlled output.

## Accepted Risk

- Initial auth model remains placeholder.
  - **Date identified:** 2026-06-19
  - **Owner:** Security Agent
  - **Mitigation:** Track RBAC roadmap and enforce boundary checks where possible without production auth.

## Resolved Risk

- Presigned upload URL mismatch in browser-based flows.
  - **Date resolved:** 2026-06-19
  - **Resolution:** Added local upload proxy strategy to align public URL and request host.
  - **Owner:** DevOps Agent

- Added from harness self-improve run: test harness command

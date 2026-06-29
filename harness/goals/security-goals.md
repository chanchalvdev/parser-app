# Security Goals

## Security posture
- Never log secrets (passwords, presigned URLs with credentials, raw tokens).
- Enforce local secret abstraction now, with production vault-ready interfaces.
- Validate user-provided inputs before side effects.
- Enforce tenant boundaries in all mutable operations.
- Prevent replay misuse across tenant boundaries.

## Threat model targets
- Archive traversal/path traversal (including zip-slip style).
- Archive expansion abuse (depth/ratio/count/size abuse).
- Parser memory exhaustion from malformed/huge input objects.
- SQL injection and unsafe query construction.
- Status transition regressions that expose data across jobs.

## Review requirements
- All parser-related changes must pass parser-safety review.
- Password/secret flows must pass security review and threat update.
- Settings that affect trust boundaries require explicit threat-model note.

## Acceptance check
- High-risk changes include guardrails + test evidence + residual risk ownership.

# Parser Safety Review Hook

## Trigger
Before merge for parser/extractor changes and JSON-like format handling.

## Checklist
- [ ] Malformed data handling is deterministic and non-fatal where policy allows.
- [ ] Depth/size/expansion limits remain bounded and documented.
- [ ] Error classification is explicit (`parser_error` vs fatal).
- [ ] Structured output remains JSON-safe and serializable.
- [ ] Parser metrics/events capture counts and failures.
- [ ] Regression coverage exists for malformed and edge payloads.

## Blocking criteria
- Unchecked recursion, zip-bomb, or deep expansion risk.
- Parser emits unbounded or non-JSON-safe structured payloads.
- Missing malformed-line recovery when required by mode.
- Security-critical extraction paths changed without explicit safety validation.

## Required evidence
- Parser tests and fixture list.
- Error capture examples and expected continuation/failure behavior.
- Resource impact estimate or bound for worst-case path.

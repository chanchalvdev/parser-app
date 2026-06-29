# Parser Addition Workflow

## Purpose
Add or modify parser behavior safely, with deterministic outputs and bounded failure semantics.

## Phases

1. Review
   - Define parser mode expectations (TXT/CSV/JSON/object/array/JSONL).
   - Confirm extraction limits and malformed-input expectations.
2. Design
   - Define ParsedRecord schema mapping and flattening contract.
   - Define parser error behavior (`parser_error` rows vs fatal failures).
3. Plan
   - Assign parser/extractor/loader touches to Worker Python and Search agents.
   - Define test corpus and fixture expectations.
4. Agree
   - Confirm search and downstream contract compatibility.
5. Execute
   - Implement parser parser mode detection and resilience.
   - Maintain structure preservation where required.
6. Test
   - Add tests for object/array/jsonl and malformed-line recovery.
   - Include archive path/encoding and nested payload cases where relevant.
7. Validation
   - Run one realistic sample through local upload → parse → index → search.
8. Self-improve
   - Update parser risks and decisions in memory files if needed.

## Exit criteria

- Deterministic parsing for supported modes.
- Malformed inputs handled with explicit error policy.
- Searchability and entities contract remain stable.
- Parser safety checklist completed.

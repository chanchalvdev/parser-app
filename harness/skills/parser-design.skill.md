# Parser Design Skill

## When to use
When implementing parser logic, parser modes, malformed input behavior, and extraction rules.

## Procedure
1. Define accepted input modes and mode-detection heuristics.
2. Define normalization and flattening policy.
3. Specify record-level output contract and error handling.
4. Implement tests for normal and malformed inputs.
5. Validate resource constraints and parser time/memory behavior.

## Checklist
- Deterministic output for same input bytes.
- Metadata + `structured_data` retained as required.
- Recoverable malformed input continues safely where policy requires.
- Error classifications are explicit and non-ambiguous.

## Guardrails
- Never parse unbounded nested input without caps.
- Ensure JSONL and streaming-like modes continue past recoverable bad lines.
- Keep parser outputs JSON-serializable and stable.

## Example output
- JSON parser emits one `ParsedRecord` per object for object/array/JSONL modes.
- Malformed JSONL line is captured as parser error and processing continues.
- Entity extraction is triggered from flattened `content_text`.

# Frontend Component Skill

## When to use
When creating pages, components, hooks, and API binding layers.

## Procedure
1. Define local component contract and error/loading states.
2. Implement API hooks and cache strategy.
3. Add URL/query synchronization where practical.
4. Validate accessibility and mobile responsiveness.
5. Add focused tests or smoke checks.

## Checklist
- Client types align with backend contracts.
- Empty/loading/error states are explicit.
- No hard-coded endpoints; use shared client configuration.
- Mutations include success/error feedback and state refresh.

## Guardrails
- Do not block the entire UI for non-critical data.
- Avoid logging sensitive values or upload credentials.
- Keep query behavior deterministic for pagination/sorting.

## Example output
- Search page with filters, pagination, highlighting, and result drawer.
- File/job detail pages showing timeline and status refresh actions.

# Database Migration Skill

## When to use
When changing schema, indexes, constraints, seed data, or migration strategy.

## Procedure
1. Draft up/down migration scripts with safe defaults.
2. Add needed indexes for hot query paths.
3. Validate existing data against new constraints.
4. Add rollback/cleanup strategy and smoke verification SQL.

## Checklist
- Migration ordering is explicit and deterministic.
- Constraints and indexes are justified by query plans.
- Rollback path is practical and documented.
- Seed data remains deterministic and safe for local use.

## Guardrails
- Never destructively migrate without explicit backup/cleanup strategy.
- Ensure tenant and ID fields remain present where expected.
- Preserve existing records unless migration scope explicitly says otherwise.

## Example output
- `migrations/xxxx_add_archive_password_refs.sql` and rollback script with indexes.
- Validation query confirming counts and status consistency pre/post migration.

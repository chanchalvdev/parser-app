# infra/migrations/

## Structure

Migrations use a simple SQL file convention:

- `001_schema.up.sql` — create all core tables and indexes.
- `001_schema.down.sql` — rollback script for core schema.
- `seed.sql` — idempotent reference data + defaults.

Run migrations and seeds via Makefile:

```bash
make migrate
make seed
```

`make migrate` applies `*.up.sql` files in lexical order via the PostgreSQL container.

## Using golang-migrate (optional)

This repo currently uses lightweight container-backed `psql` execution for migrations.
If you prefer `golang-migrate`, you can swap `infra/scripts/migrate.sh` to:

```bash
docker run --rm \
  -v "$(pwd)/infra/migrations:/migrations" \
  --network file-platform_default \
  migrate/migrate \
  -path=/migrations \
  -database "$DATABASE_URL" up
```

When using that approach, keep file names in golang-migrate format like `001_schema.up.sql` and `001_schema.down.sql`.

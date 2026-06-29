#!/usr/bin/env bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
MIGRATIONS_DIR="$PROJECT_ROOT/infra/migrations"
if [ ! -d "$MIGRATIONS_DIR" ] && [ -d "$PROJECT_ROOT/migrations" ]; then
  MIGRATIONS_DIR="$PROJECT_ROOT/migrations"
fi
ENV_FILE="$PROJECT_ROOT/.env"

if [ -f "$ENV_FILE" ]; then
  set -a
  # shellcheck source=/dev/null
  . "$ENV_FILE"
  set +a
fi

if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
  DOCKER_COMPOSE_CMD="docker compose"
elif command -v docker-compose >/dev/null 2>&1; then
  DOCKER_COMPOSE_CMD="docker-compose"
else
  DOCKER_COMPOSE_CMD="docker compose"
fi

if [ -z "${DATABASE_URL:-}" ]; then
  echo "DATABASE_URL is not set. Add DATABASE_URL to $ENV_FILE or export it in your shell."
  exit 1
fi

if [ ! -d "$MIGRATIONS_DIR" ]; then
  echo "Migration directory not found: $MIGRATIONS_DIR" >&2
  exit 1
fi

if ! $DOCKER_COMPOSE_CMD ps --services --filter "status=running" | grep -q '^postgres$'; then
  echo "PostgreSQL is not running. Start services first: make up"
  exit 1
fi

shopt -s nullglob
FILES=("$MIGRATIONS_DIR"/*.up.sql)
if [ ${#FILES[@]} -eq 0 ]; then
  echo "No .up.sql migration files found in $MIGRATIONS_DIR"
  exit 0
fi

for migration in "${FILES[@]}"; do
  echo "Applying migration: $(basename "$migration")"
  $DOCKER_COMPOSE_CMD exec -T postgres psql "$DATABASE_URL" -v ON_ERROR_STOP=1 < "$migration"
done

echo "Database migrations complete."

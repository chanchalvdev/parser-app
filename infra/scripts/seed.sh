#!/usr/bin/env bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SEED_FILE="$PROJECT_ROOT/infra/migrations/seed.sql"
if [ ! -f "$SEED_FILE" ] && [ -f "$PROJECT_ROOT/migrations/seed.sql" ]; then
  SEED_FILE="$PROJECT_ROOT/migrations/seed.sql"
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

if [ ! -f "$SEED_FILE" ]; then
  echo "Seed file not found: $SEED_FILE" >&2
  exit 1
fi

if ! $DOCKER_COMPOSE_CMD ps --services --filter "status=running" | grep -q '^postgres$'; then
  echo "PostgreSQL is not running. Start services first: make up"
  exit 1
fi

$DOCKER_COMPOSE_CMD exec -T postgres psql "$DATABASE_URL" -v ON_ERROR_STOP=1 < "$SEED_FILE"

echo "Seed data loaded."

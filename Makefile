.PHONY: up down logs ps clean restart migrate seed test test-api test-worker test-web test-all lint fmt search-init harness harness-full e2e-upload

DC_CMD := $(shell \
	if docker compose version >/dev/null 2>&1; then \
		echo "docker compose"; \
	elif command -v docker-compose >/dev/null 2>&1; then \
		echo "docker-compose"; \
	else \
		echo "docker compose"; \
	fi \
)

up:
	$(DC_CMD) up --build -d

down:
	$(DC_CMD) down

logs:
	$(DC_CMD) logs -f

ps:
	$(DC_CMD) ps

clean:
	$(DC_CMD) down -v
	rm -rf .pytest_cache
	find . -name "__pycache__" -type d -prune -exec rm -rf {} +

restart:
	$(DC_CMD) restart

migrate:
	./infra/scripts/migrate.sh

seed:
	./infra/scripts/seed.sh

test:
	./infra/scripts/test.sh

# Harness-driven workflow
HARNESS_SCOPE ?= implementation
HARNESS_AGENT ?= architect
HARNESS_MODE ?= light

harness:
	bash harness/scripts/pipeline.sh --scope "$(HARNESS_SCOPE)" --agent "$(HARNESS_AGENT)" --mode "$(HARNESS_MODE)"

harness-full:
	bash harness/scripts/pipeline.sh --scope "$(HARNESS_SCOPE)" --agent "$(HARNESS_AGENT)" --mode full

test-api:
	cd apps/api && go test ./...

test-worker:
	cd apps/worker && python -m pytest -q tests

test-web:
	cd apps/web && npm install --no-audit --no-fund && npm run test

test-all:
	make test-api && make test-worker && make test-web

e2e-upload:
	./scripts/e2e-upload.sh

lint:
	@echo "[lint] API: lint placeholder (go test - run as part of build quality gate)"
	@echo "[lint] Worker: lint placeholder (ruff/flake8 to be added)"
	@echo "[lint] Frontend: lint placeholder (eslint/setup to be added)"

fmt:
	@echo "[fmt] Running Go format for api and worker python code formatting"
	cd apps/api && gofmt -w $$(find . -name '*.go')
	@echo "[fmt] Python formatting placeholder (black/ruff not installed)"
	cd apps/web && npm install --no-audit --no-fund && npm run build

search-init:
	bash ./infra/opensearch/create-indexes.sh

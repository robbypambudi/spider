.PHONY: help install api controller worker frontend test lint typecheck compose-up compose-down migrate seed

help:
	@echo "SPIDER development targets"
	@echo "  make install       Install backend (uv) and frontend (npm) deps"
	@echo "  make compose-up    Start PostgreSQL and Redis"
	@echo "  make migrate       Run Alembic migrations"
	@echo "  make api           Run the FastAPI control plane"
	@echo "  make controller    Run the control-plane reconciler"
	@echo "  make worker        Run a local serving worker (Mock LLM)"
	@echo "  make frontend      Run the Vite dashboard"
	@echo "  make test          Run backend tests"
	@echo "  make lint          Ruff + frontend lint"
	@echo "  make typecheck     mypy + frontend tsc"

install:
	cd backend && uv sync
	cd frontend && npm install

compose-up:
	docker compose up -d postgres redis

compose-down:
	docker compose down

migrate:
	cd backend && uv run alembic upgrade head

api:
	cd backend && uv run uvicorn cmd.api.main:app --reload --host 0.0.0.0 --port 8000

controller:
	cd backend && uv run python -m cmd.controller.main

worker:
	cd backend && uv run python -m cmd.worker.main

frontend:
	cd frontend && npm run dev

test:
	cd backend && uv run pytest -q

lint:
	cd backend && uv run ruff check .
	cd frontend && npm run lint

typecheck:
	cd backend && uv run mypy cmd internal pkg dao
	cd frontend && npm run typecheck

# Contributing to SPIDER

SPIDER is a runtime defense framework for prompt injection detection in
LLM-serving cluster environments. Contributions should preserve that identity.

## Principles

1. The security pipeline inspects every inference request **before** the LLM.
2. A `BLOCK` decision must never call an `LLMProvider`.
3. Thresholds live in policy/configuration, not inside detectors.
4. Handlers do not contain detector logic or database queries.
5. Do not log or persist full prompt text unless explicitly enabled.
6. GPU hardware and Kubernetes are optional. Local development uses Mock LLM.

## Development setup

```bash
# PostgreSQL + Redis
docker compose up -d postgres redis

# Backend
cd backend
uv sync
uv run alembic upgrade head
uv run uvicorn cmd.api.main:app --reload --port 8000

# Frontend
cd frontend
npm install
npm run dev
```

## Tests

```bash
cd backend
uv run pytest
uv run ruff check .
uv run mypy cmd internal pkg dao tests
```

```bash
cd frontend
npm run lint
npm run typecheck
npm run build
```

## Pull requests

* Keep business logic out of `cmd/` entrypoints.
* Add tests for security-path changes, especially enforcement.
* Update docs when pipeline, policy, or serving interfaces change.
* Do not add fake production ML detector implementations.

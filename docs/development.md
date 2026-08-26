# Development

```bash
docker compose up -d postgres redis
cd backend && uv sync && uv run uvicorn cmd.api.main:app --reload --port 8000
cd frontend && npm install && npm run dev
cd backend && uv run pytest && uv run ruff check . && uv run mypy cmd internal pkg dao
```

Schema is created on API startup in development (`create_all` + admin bootstrap). Alembic lives at `backend/dao/migrations` for production deploys:

```bash
cd backend && uv run alembic upgrade head
```

Do not log prompt bodies unless `SPIDER_LOG_PROMPT_CONTENT=true`. Tests use in-memory SQLite and a spy `MockLLMProvider`.

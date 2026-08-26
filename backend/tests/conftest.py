from __future__ import annotations

import os
from cmd.api.main import create_app
from collections.abc import AsyncGenerator

import pytest
from httpx import ASGITransport, AsyncClient
from internal.app import bootstrap_control_plane
from internal.telemetry.metrics import SpiderMetrics
from pkg.config.settings import Settings
from pkg.serving.providers.mock import MockLLMProvider

os.environ.setdefault("DATABASE_URL", "sqlite+aiosqlite:///:memory:")
os.environ.setdefault("SPIDER_ENV", "test")
os.environ.setdefault("SPIDER_JWT_SECRET", "test-secret-please-change-me-32b")
os.environ.setdefault("SPIDER_WORKER_TOKEN", "test-worker-token")


def test_settings() -> Settings:
    return Settings(
        env="test",
        database_url="sqlite+aiosqlite:///:memory:",
        redis_url="redis://localhost:6379/0",
        jwt_secret="test-secret-please-change-me-32b",
        worker_token="test-worker-token",
        default_detector="rule-based",
        default_threshold=0.5,
        fail_mode="closed",
        log_prompt_content=False,
        persist_prompt_content=False,
        bootstrap_admin_email="admin@spider.local",
        bootstrap_admin_password="spider-admin",
    )


@pytest.fixture
async def llm() -> MockLLMProvider:
    return MockLLMProvider()


@pytest.fixture
async def app(llm: MockLLMProvider):
    application = create_app(test_settings(), llm_provider=llm, container=None)
    application.state.container.metrics = SpiderMetrics()
    await bootstrap_control_plane(application.state.container)
    yield application
    await application.state.container.database.dispose()


@pytest.fixture
async def client(app) -> AsyncGenerator[AsyncClient, None]:
    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://test") as http:
        yield http


@pytest.fixture
async def auth_headers(client: AsyncClient) -> dict[str, str]:
    response = await client.post(
        "/api/v1/auth/login",
        json={"email": "admin@spider.local", "password": "spider-admin"},
    )
    assert response.status_code == 200, response.text
    token = response.json()["access_token"]
    return {"Authorization": f"Bearer {token}"}

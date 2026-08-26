from __future__ import annotations

from dataclasses import dataclass

from dao.database import Database
from dao.models.user import User
from dao.repositories.policy import PolicyRepository
from dao.repositories.user import UserRepository
from pkg.config.settings import Settings
from pkg.scheduler.least_loaded import LeastLoadedScheduler
from pkg.security.enforcement import Enforcer
from pkg.security.pipeline import SecurityPipeline, build_default_pipeline
from pkg.serving.providers import PROVIDER_REGISTRY, MockLLMProvider
from pkg.serving.providers.base import LLMProvider
from pkg.serving.router import ServingRouter
from sqlalchemy import select

from internal.service.auth import hash_password
from internal.telemetry.metrics import SpiderMetrics


@dataclass
class AppContainer:
    settings: Settings
    database: Database
    pipeline: SecurityPipeline
    enforcer: Enforcer
    llm_provider: LLMProvider
    serving_router: ServingRouter
    metrics: SpiderMetrics


def build_container(
    settings: Settings,
    *,
    llm_provider: LLMProvider | None = None,
    metrics: SpiderMetrics | None = None,
) -> AppContainer:
    database = Database(settings)
    pipeline = build_default_pipeline(
        detector_name=settings.default_detector,
        threshold=settings.default_threshold,
        chunk_size=settings.chunk_size,
        chunk_overlap=settings.chunk_overlap,
        fail_mode=settings.fail_mode,
    )
    provider = llm_provider
    if provider is None:
        provider_cls = PROVIDER_REGISTRY.get(settings.serving_provider, MockLLMProvider)
        provider = provider_cls()
    scheduler = LeastLoadedScheduler()
    router = ServingRouter(provider=provider, scheduler=scheduler)
    return AppContainer(
        settings=settings,
        database=database,
        pipeline=pipeline,
        enforcer=pipeline.enforcer,
        llm_provider=provider,
        serving_router=router,
        metrics=metrics or SpiderMetrics(),
    )


async def bootstrap_control_plane(container: AppContainer) -> None:
    await container.database.create_schema()
    async with container.database.session_factory() as session:
        users = UserRepository(session)
        existing = await users.get_by_email(container.settings.bootstrap_admin_email)
        if existing is None:
            await users.create(
                email=container.settings.bootstrap_admin_email,
                hashed_password=hash_password(container.settings.bootstrap_admin_password),
                role="ADMIN",
                display_name="SPIDER Admin",
            )
        await PolicyRepository(session).ensure_default(
            name=container.settings.default_security_policy,
            threshold=container.settings.default_threshold,
        )
        await session.commit()


async def database_ready(container: AppContainer) -> bool:
    try:
        async with container.database.session_factory() as session:
            await session.execute(select(User).limit(1))
        return True
    except Exception:
        return False

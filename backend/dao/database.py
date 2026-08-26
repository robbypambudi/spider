from __future__ import annotations

from collections.abc import AsyncGenerator
from typing import Any

from pkg.config.settings import Settings
from sqlalchemy.ext.asyncio import (
    AsyncEngine,
    AsyncSession,
    async_sessionmaker,
    create_async_engine,
)
from sqlalchemy.pool import StaticPool

from dao.models.base import Base


class Database:
    def __init__(self, settings: Settings) -> None:
        engine_kwargs: dict[str, Any] = {"echo": False}
        if settings.database_url.startswith("sqlite"):
            engine_kwargs["connect_args"] = {"check_same_thread": False}
            if ":memory:" in settings.database_url:
                engine_kwargs["poolclass"] = StaticPool
        self.engine: AsyncEngine = create_async_engine(settings.database_url, **engine_kwargs)
        self.session_factory = async_sessionmaker(
            self.engine,
            class_=AsyncSession,
            expire_on_commit=False,
        )

    async def create_schema(self) -> None:
        async with self.engine.begin() as connection:
            await connection.run_sync(Base.metadata.create_all)

    async def dispose(self) -> None:
        await self.engine.dispose()

    async def session(self) -> AsyncGenerator[AsyncSession, None]:
        async with self.session_factory() as session:
            yield session

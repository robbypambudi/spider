from __future__ import annotations

from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from dao.models.policy import Policy


class PolicyRepository:
    def __init__(self, session: AsyncSession) -> None:
        self.session = session

    async def list_policies(self) -> list[Policy]:
        result = await self.session.execute(select(Policy).order_by(Policy.name))
        return list(result.scalars().all())

    async def get_by_name(self, name: str) -> Policy | None:
        result = await self.session.execute(select(Policy).where(Policy.name == name))
        return result.scalar_one_or_none()

    async def ensure_default(self, *, name: str, threshold: float) -> Policy:
        existing = await self.get_by_name(name)
        if existing is not None:
            return existing
        policy = Policy(
            name=name,
            kind="threshold",
            threshold=threshold,
            action_on_detection="block",
            is_default=True,
        )
        self.session.add(policy)
        await self.session.flush()
        return policy

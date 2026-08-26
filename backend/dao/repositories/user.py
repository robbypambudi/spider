from __future__ import annotations

from uuid import UUID

from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from dao.models.user import User


class UserRepository:
    def __init__(self, session: AsyncSession) -> None:
        self.session = session

    async def get_by_email(self, email: str) -> User | None:
        result = await self.session.execute(select(User).where(User.email == email))
        return result.scalar_one_or_none()

    async def get_by_id(self, user_id: UUID) -> User | None:
        result = await self.session.execute(select(User).where(User.id == user_id))
        return result.scalar_one_or_none()

    async def create(
        self,
        *,
        email: str,
        hashed_password: str,
        role: str,
        display_name: str | None = None,
    ) -> User:
        user = User(
            email=email,
            hashed_password=hashed_password,
            role=role,
            display_name=display_name,
        )
        self.session.add(user)
        await self.session.flush()
        return user

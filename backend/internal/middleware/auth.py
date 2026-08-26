from __future__ import annotations

from collections.abc import AsyncGenerator

from dao.models.user import User
from dao.repositories.user import UserRepository
from fastapi import Depends, HTTPException, Request, status
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer
from sqlalchemy.ext.asyncio import AsyncSession

from internal.exceptions import AuthenticationError
from internal.service.auth import AuthService

_bearer = HTTPBearer(auto_error=False)


async def get_db_session(request: Request) -> AsyncGenerator[AsyncSession, None]:
    container = request.app.state.container
    async with container.database.session_factory() as session:
        try:
            yield session
            await session.commit()
        except Exception:
            await session.rollback()
            raise


async def get_current_user(
    request: Request,
    credentials: HTTPAuthorizationCredentials | None = Depends(_bearer),
    session: AsyncSession = Depends(get_db_session),
) -> User:
    if credentials is None:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Not authenticated")
    settings = request.app.state.container.settings
    auth = AuthService(UserRepository(session), settings)
    try:
        return await auth.user_from_token(credentials.credentials)
    except AuthenticationError as exc:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail=str(exc)) from exc


def require_role(*roles: str):  # type: ignore[no-untyped-def]
    async def _inner(user: User = Depends(get_current_user)) -> User:
        if user.role not in roles:
            raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail="Insufficient role")
        return user

    return _inner

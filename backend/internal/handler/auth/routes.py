from __future__ import annotations

from dao.models.user import User
from dao.repositories.user import UserRepository
from fastapi import APIRouter, Depends, Request
from sqlalchemy.ext.asyncio import AsyncSession

from internal.middleware.auth import get_current_user, get_db_session
from internal.payload import LoginRequest, TokenResponse
from internal.service.auth import AuthService

router = APIRouter(tags=["auth"])


@router.post("/auth/login", response_model=TokenResponse)
async def login(
    body: LoginRequest,
    request: Request,
    session: AsyncSession = Depends(get_db_session),
) -> TokenResponse:
    settings = request.app.state.container.settings
    auth = AuthService(UserRepository(session), settings)
    user = await auth.authenticate(body.email, body.password)
    token = auth.issue_token(user)
    return TokenResponse(access_token=token, role=user.role, email=user.email)


@router.get("/auth/me")
async def me(user: User = Depends(get_current_user)) -> dict[str, str]:
    return {"id": str(user.id), "email": user.email, "role": user.role}

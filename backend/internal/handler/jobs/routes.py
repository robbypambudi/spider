from __future__ import annotations

from dao.models.user import User
from fastapi import APIRouter, Depends

from internal.middleware.auth import get_current_user

router = APIRouter(tags=["jobs"])


@router.get("/jobs")
async def list_jobs(user: User = Depends(get_current_user)) -> list[dict[str, str]]:
    _ = user
    return []

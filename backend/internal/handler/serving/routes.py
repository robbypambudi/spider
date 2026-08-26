from __future__ import annotations

from dao.models.user import User
from dao.repositories.worker import WorkerRepository
from fastapi import APIRouter, Depends, Request
from sqlalchemy.ext.asyncio import AsyncSession

from internal.middleware.auth import get_current_user, get_db_session
from internal.service.serving import ServingService

router = APIRouter(tags=["serving"])


@router.get("/serving/nodes")
async def serving_nodes(
    session: AsyncSession = Depends(get_db_session),
    user: User = Depends(get_current_user),
) -> list[dict[str, object]]:
    _ = user
    return await ServingService(WorkerRepository(session)).list_nodes()


@router.get("/serving/models")
async def serving_models(
    request: Request,
    session: AsyncSession = Depends(get_db_session),
    user: User = Depends(get_current_user),
) -> list[dict[str, object]]:
    _ = request
    _ = user
    return await ServingService(WorkerRepository(session)).list_models()

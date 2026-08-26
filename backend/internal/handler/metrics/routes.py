from __future__ import annotations

from dao.models.user import User
from dao.repositories.security import SecurityScanRepository
from fastapi import APIRouter, Depends, Request
from sqlalchemy.ext.asyncio import AsyncSession

from internal.middleware.auth import get_current_user, get_db_session
from internal.service.metrics import MetricsService

router = APIRouter(tags=["metrics"])


@router.get("/metrics/summary")
async def metrics_summary(
    request: Request,
    session: AsyncSession = Depends(get_db_session),
    user: User = Depends(get_current_user),
) -> dict[str, object]:
    _ = request
    _ = user
    return await MetricsService(session, SecurityScanRepository(session)).dashboard_summary()

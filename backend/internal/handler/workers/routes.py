from __future__ import annotations

from dao.models.user import User
from dao.repositories.worker import WorkerRepository
from fastapi import APIRouter, Depends, Header, Request
from pkg.apis.worker.models import WorkerHeartbeat, WorkerResource
from pkg.protocol.worker import WORKER_TOKEN_HEADER
from sqlalchemy.ext.asyncio import AsyncSession

from internal.middleware.auth import get_current_user, get_db_session
from internal.service.worker import WorkerService

router = APIRouter(tags=["workers"])


def _worker_service(request: Request, session: AsyncSession) -> WorkerService:
    return WorkerService(WorkerRepository(session), request.app.state.container.settings)


@router.post("/workers/register", response_model=WorkerResource)
async def register_worker(
    body: WorkerResource,
    request: Request,
    session: AsyncSession = Depends(get_db_session),
    x_spider_worker_token: str | None = Header(default=None, alias=WORKER_TOKEN_HEADER),
) -> WorkerResource:
    service = _worker_service(request, session)
    service.authenticate_token(x_spider_worker_token)
    return await service.register(body)


@router.post("/workers/{worker_id}/heartbeat", response_model=WorkerResource)
async def worker_heartbeat(
    worker_id: str,
    body: WorkerHeartbeat,
    request: Request,
    session: AsyncSession = Depends(get_db_session),
    x_spider_worker_token: str | None = Header(default=None, alias=WORKER_TOKEN_HEADER),
) -> WorkerResource:
    service = _worker_service(request, session)
    service.authenticate_token(x_spider_worker_token)
    body.worker_id = worker_id
    return await service.heartbeat(body)


@router.get("/workers")
async def list_workers(
    request: Request,
    session: AsyncSession = Depends(get_db_session),
    user: User = Depends(get_current_user),
) -> list[WorkerResource]:
    _ = user
    return await _worker_service(request, session).list_workers()


@router.get("/workers/{worker_id}")
async def inspect_worker(
    worker_id: str,
    request: Request,
    session: AsyncSession = Depends(get_db_session),
    user: User = Depends(get_current_user),
) -> WorkerResource:
    _ = user
    return await _worker_service(request, session).inspect(worker_id)

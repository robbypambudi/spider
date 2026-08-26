from __future__ import annotations

from dao.models.user import User
from dao.repositories.inference import InferenceRepository
from dao.repositories.security import SecurityScanRepository
from dao.repositories.worker import WorkerRepository
from fastapi import APIRouter, Depends, Request
from pkg.apis.inference.models import InferenceRequest, ProtectedInferenceResponse
from sqlalchemy.ext.asyncio import AsyncSession

from internal.middleware.auth import get_current_user, get_db_session
from internal.payload import InferenceHttpRequest
from internal.service.inference import InferenceService
from internal.service.security import SecurityService

router = APIRouter(tags=["inference"])


@router.post("/inference", response_model=ProtectedInferenceResponse)
async def infer(
    body: InferenceHttpRequest,
    request: Request,
    session: AsyncSession = Depends(get_db_session),
    user: User = Depends(get_current_user),
) -> ProtectedInferenceResponse:
    container = request.app.state.container
    security = SecurityService(
        container.pipeline,
        SecurityScanRepository(session),
        container.settings,
        container.metrics,
    )
    service = InferenceService(
        security=security,
        enforcer=container.enforcer,
        router=container.serving_router,
        inferences=InferenceRepository(session),
        workers=WorkerRepository(session),
        metrics=container.metrics,
    )
    inference_request = InferenceRequest(
        model=body.model,
        prompt=body.prompt,
        max_tokens=body.max_tokens,
        temperature=body.temperature,
        security_enabled=bool(body.security.get("enabled", True)),
    )
    return await service.infer(inference_request, user_id=user.id)


@router.get("/inference")
async def list_inference(
    session: AsyncSession = Depends(get_db_session),
    user: User = Depends(get_current_user),
    limit: int = 50,
) -> list[dict[str, object]]:
    _ = user
    rows = await InferenceRepository(session).list_recent(limit=limit)
    return [
        {
            "id": str(row.id),
            "request_id": str(row.request_id),
            "model": row.model,
            "status": row.status,
            "decision": row.decision,
            "worker_id": row.worker_id,
            "end_to_end_latency_ms": row.end_to_end_latency_ms,
            "security_overhead_ms": row.security_overhead_ms,
            "inference_latency_ms": row.inference_latency_ms,
            "created_at": row.created_at.isoformat(),
        }
        for row in rows
    ]

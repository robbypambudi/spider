from __future__ import annotations

from dao.models.user import User
from dao.repositories.policy import PolicyRepository
from dao.repositories.security import SecurityScanRepository
from fastapi import APIRouter, Depends, Request
from internal.exceptions import NotFoundError
from internal.middleware.auth import get_current_user, get_db_session
from internal.payload import DetectorView, EvaluateRequest, ScanRequest, ScanResponse
from internal.service.policy import PolicyService
from internal.service.security import SecurityService
from pkg.apis.security.models import SecurityResult
from pkg.security.pipeline import SecurityPipeline
from sqlalchemy.ext.asyncio import AsyncSession

router = APIRouter(tags=["security"])


def _pipeline(request: Request) -> SecurityPipeline:
    return request.app.state.container.pipeline


def _metrics(request: Request):  # type: ignore[no-untyped-def]
    return request.app.state.container.metrics


def _settings(request: Request):  # type: ignore[no-untyped-def]
    return request.app.state.container.settings


def _to_scan_response(result: SecurityResult, scan_id: object | None) -> ScanResponse:
    return ScanResponse(
        scan_id=scan_id,  # type: ignore[arg-type]
        request_id=result.request_id,
        decision=result.decision.value,
        score=result.score,
        chunks_scanned=result.chunks_scanned,
        policy=result.policy,
        threshold=result.metadata.get("threshold"),
        latency_ms=result.total_latency_ms,
        detectors=[
            DetectorView(
                detector=item.detector,
                score=item.score,
                is_injection=item.is_injection,
                latency_ms=item.latency_ms,
                metadata=item.metadata,
            )
            for item in result.detector_results
        ],
        model=result.metadata.get("model"),
    )


@router.post("/security/scan", response_model=ScanResponse)
async def scan(
    body: ScanRequest,
    request: Request,
    session: AsyncSession = Depends(get_db_session),
    user: User = Depends(get_current_user),
) -> ScanResponse:
    service = SecurityService(
        _pipeline(request),
        SecurityScanRepository(session),
        _settings(request),
        _metrics(request),
    )
    result, stored = await service.inspect(
        body.text,
        source=body.source or "api",
        user_id=user.id,
        model=body.model,
        metadata=body.metadata,
    )
    return _to_scan_response(result, stored.id if stored else None)


@router.get("/security/scans")
async def list_scans(
    session: AsyncSession = Depends(get_db_session),
    user: User = Depends(get_current_user),
    limit: int = 50,
    offset: int = 0,
) -> list[dict[str, object]]:
    _ = user
    rows = await SecurityScanRepository(session).list_scans(limit=limit, offset=offset)
    return [
        {
            "id": str(row.id),
            "request_id": str(row.request_id),
            "decision": row.decision,
            "score": row.score,
            "threshold": row.threshold,
            "detector": row.detector,
            "policy": row.policy,
            "chunks_scanned": row.chunks_scanned,
            "chunking_strategy": row.chunking_strategy,
            "latency_ms": row.latency_ms,
            "prompt_hash": row.prompt_hash,
            "prompt_length": row.prompt_length,
            "model_target": row.model_target,
            "created_at": row.created_at.isoformat(),
        }
        for row in rows
    ]


@router.get("/security/scans/{scan_id}")
async def get_scan(
    scan_id: str,
    session: AsyncSession = Depends(get_db_session),
    user: User = Depends(get_current_user),
) -> dict[str, object]:
    _ = user
    repo = SecurityScanRepository(session)
    from uuid import UUID

    try:
        parsed = UUID(scan_id)
    except ValueError as exc:
        raise NotFoundError("Invalid scan id") from exc
    row = await repo.get(parsed)
    if row is None:
        raise NotFoundError("Scan not found")
    chunks = await repo.chunk_results(row.id)
    executions = await repo.detector_executions(row.id)
    return {
        "id": str(row.id),
        "request_id": str(row.request_id),
        "decision": row.decision,
        "score": row.score,
        "threshold": row.threshold,
        "detector": row.detector,
        "detector_version": row.detector_version,
        "policy": row.policy,
        "chunks_scanned": row.chunks_scanned,
        "chunking_strategy": row.chunking_strategy,
        "latency_ms": row.latency_ms,
        "prompt_hash": row.prompt_hash,
        "prompt_length": row.prompt_length,
        "model_target": row.model_target,
        "worker_id": row.worker_id,
        "created_at": row.created_at.isoformat(),
        "chunks": [
            {
                "index": chunk.chunk_index,
                "detector": chunk.detector,
                "score": chunk.score,
                "is_injection": chunk.is_injection,
                "latency_ms": chunk.latency_ms,
            }
            for chunk in chunks
        ],
        "detectors": [
            {
                "detector": item.detector,
                "version": item.detector_version,
                "score": item.score,
                "is_injection": item.is_injection,
                "latency_ms": item.latency_ms,
                "threshold": item.threshold,
            }
            for item in executions
        ],
    }


@router.get("/security/detectors")
async def list_detectors(
    request: Request,
    session: AsyncSession = Depends(get_db_session),
    user: User = Depends(get_current_user),
) -> list[dict[str, str]]:
    _ = user
    service = SecurityService(
        _pipeline(request),
        SecurityScanRepository(session),
        _settings(request),
        _metrics(request),
    )
    return service.list_detectors()


@router.get("/security/policies")
async def list_policies(
    request: Request,
    session: AsyncSession = Depends(get_db_session),
    user: User = Depends(get_current_user),
) -> list[dict[str, object]]:
    _ = user
    service = PolicyService(PolicyRepository(session), _settings(request))
    return await service.list_policies()


@router.post("/security/evaluate")
async def evaluate(
    body: EvaluateRequest,
    request: Request,
    session: AsyncSession = Depends(get_db_session),
    user: User = Depends(get_current_user),
) -> dict[str, object]:
    service = SecurityService(
        _pipeline(request),
        SecurityScanRepository(session),
        _settings(request),
        _metrics(request),
    )
    report = await service.evaluate(
        [sample.model_dump() for sample in body.samples],
        threshold=body.threshold,
        target_fprs=body.target_fpr,
    )
    return report.model_dump()

from __future__ import annotations

import json
from uuid import UUID

from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from dao.models.inference import InferenceEvent, InferenceRequestRecord


class InferenceRepository:
    def __init__(self, session: AsyncSession) -> None:
        self.session = session

    async def create(
        self,
        *,
        request_id: UUID,
        model: str,
        status: str,
        decision: str,
        scan_id: UUID | None,
        user_id: UUID | None,
        worker_id: str | None,
        end_to_end_latency_ms: float,
        security_overhead_ms: float,
        inference_latency_ms: float | None,
        output_preview: str | None,
        metadata: dict[str, object] | None = None,
    ) -> InferenceRequestRecord:
        record = InferenceRequestRecord(
            request_id=request_id,
            user_id=user_id,
            scan_id=scan_id,
            model=model,
            status=status,
            decision=decision,
            worker_id=worker_id,
            end_to_end_latency_ms=end_to_end_latency_ms,
            security_overhead_ms=security_overhead_ms,
            inference_latency_ms=inference_latency_ms,
            output_preview=output_preview,
            metadata_json=json.dumps(metadata or {}, default=str),
        )
        self.session.add(record)
        await self.session.flush()
        return record

    async def add_event(
        self,
        inference_id: UUID,
        event_type: str,
        payload: dict[str, object] | None = None,
    ) -> InferenceEvent:
        event = InferenceEvent(
            inference_id=inference_id,
            event_type=event_type,
            payload_json=json.dumps(payload or {}, default=str),
        )
        self.session.add(event)
        await self.session.flush()
        return event

    async def get_by_request_id(self, request_id: UUID) -> InferenceRequestRecord | None:
        result = await self.session.execute(
            select(InferenceRequestRecord).where(InferenceRequestRecord.request_id == request_id)
        )
        return result.scalar_one_or_none()

    async def list_recent(self, *, limit: int = 50) -> list[InferenceRequestRecord]:
        result = await self.session.execute(
            select(InferenceRequestRecord)
            .order_by(InferenceRequestRecord.created_at.desc())
            .limit(limit)
        )
        return list(result.scalars().all())

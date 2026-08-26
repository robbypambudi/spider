from __future__ import annotations

import json
from typing import Any
from uuid import UUID

from pkg.apis.security.models import DetectionResult, SecurityResult
from sqlalchemy import func, select
from sqlalchemy.ext.asyncio import AsyncSession

from dao.models.security import DetectorExecution, SecurityChunkResult, SecurityScan


class SecurityScanRepository:
    def __init__(self, session: AsyncSession) -> None:
        self.session = session

    async def create_from_result(
        self,
        result: SecurityResult,
        *,
        prompt_hash: str,
        prompt_length: int,
        prompt_text: str | None,
        user_id: UUID | None = None,
        worker_id: str | None = None,
    ) -> SecurityScan:
        metadata = result.metadata
        scan = SecurityScan(
            request_id=result.request_id,
            user_id=user_id,
            decision=result.decision.value,
            score=result.score,
            threshold=metadata.get("threshold"),
            detector=str(metadata.get("detector", "unknown")),
            detector_version=str(metadata.get("detector_version", "unknown")),
            policy=result.policy,
            chunks_scanned=result.chunks_scanned,
            chunking_strategy=str(metadata.get("chunker", "fixed")),
            latency_ms=result.total_latency_ms,
            prompt_hash=prompt_hash,
            prompt_length=prompt_length,
            prompt_text=prompt_text,
            model_target=metadata.get("model"),
            worker_id=worker_id,
            source=metadata.get("source"),
            metadata_json=json.dumps(metadata, default=str),
        )
        self.session.add(scan)
        await self.session.flush()

        for index, detection in enumerate(result.detector_results):
            self.session.add(
                SecurityChunkResult(
                    scan_id=scan.id,
                    chunk_index=index,
                    detector=detection.detector,
                    score=detection.score,
                    is_injection=detection.is_injection,
                    latency_ms=detection.latency_ms,
                )
            )
            self.session.add(
                DetectorExecution(
                    scan_id=scan.id,
                    detector=detection.detector,
                    detector_version=str(detection.metadata.get("version", "unknown")),
                    threshold=detection.threshold,
                    score=detection.score,
                    is_injection=detection.is_injection,
                    latency_ms=detection.latency_ms,
                    metadata_json=json.dumps(detection.metadata, default=str),
                )
            )
        await self.session.flush()
        return scan

    async def get(self, scan_id: UUID) -> SecurityScan | None:
        result = await self.session.execute(select(SecurityScan).where(SecurityScan.id == scan_id))
        return result.scalar_one_or_none()

    async def list_scans(self, *, limit: int = 50, offset: int = 0) -> list[SecurityScan]:
        result = await self.session.execute(
            select(SecurityScan)
            .order_by(SecurityScan.created_at.desc())
            .limit(limit)
            .offset(offset)
        )
        return list(result.scalars().all())

    async def chunk_results(self, scan_id: UUID) -> list[SecurityChunkResult]:
        result = await self.session.execute(
            select(SecurityChunkResult)
            .where(SecurityChunkResult.scan_id == scan_id)
            .order_by(SecurityChunkResult.chunk_index)
        )
        return list(result.scalars().all())

    async def detector_executions(self, scan_id: UUID) -> list[DetectorExecution]:
        result = await self.session.execute(
            select(DetectorExecution).where(DetectorExecution.scan_id == scan_id)
        )
        return list(result.scalars().all())

    async def summary_counts(self) -> dict[str, Any]:
        total = await self.session.scalar(select(func.count(SecurityScan.id))) or 0
        allowed = await self.session.scalar(
            select(func.count(SecurityScan.id)).where(SecurityScan.decision == "ALLOW")
        ) or 0
        blocked = await self.session.scalar(
            select(func.count(SecurityScan.id)).where(SecurityScan.decision == "BLOCK")
        ) or 0
        review = await self.session.scalar(
            select(func.count(SecurityScan.id)).where(SecurityScan.decision == "REVIEW")
        ) or 0
        avg_latency = await self.session.scalar(select(func.avg(SecurityScan.latency_ms))) or 0.0
        return {
            "total_scans": int(total),
            "allowed": int(allowed),
            "blocked": int(blocked),
            "review": int(review),
            "avg_detection_latency_ms": float(avg_latency),
        }

    def detection_results_from_chunks(
        self, chunks: list[SecurityChunkResult]
    ) -> list[DetectionResult]:
        return [
            DetectionResult(
                detector=chunk.detector,
                score=chunk.score,
                is_injection=chunk.is_injection,
                latency_ms=chunk.latency_ms,
            )
            for chunk in chunks
        ]

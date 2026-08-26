from __future__ import annotations

import hashlib
from typing import Any
from uuid import UUID

from dao.models.security import SecurityScan
from dao.repositories.security import SecurityScanRepository
from pkg.apis.security.models import SecurityRequest, SecurityResult
from pkg.config.settings import Settings
from pkg.security.detectors import DETECTOR_REGISTRY
from pkg.security.evaluation import EvaluationReport, evaluate_scores
from pkg.security.pipeline import SecurityPipeline

from internal.telemetry.logging import get_logger
from internal.telemetry.metrics import SpiderMetrics

logger = get_logger("spider-security")


def prompt_hash(text: str) -> str:
    return hashlib.sha256(text.encode("utf-8")).hexdigest()


class SecurityService:
    def __init__(
        self,
        pipeline: SecurityPipeline,
        scans: SecurityScanRepository,
        settings: Settings,
        metrics: SpiderMetrics,
    ) -> None:
        self.pipeline = pipeline
        self.scans = scans
        self.settings = settings
        self.metrics = metrics

    async def inspect(
        self,
        text: str,
        *,
        request_id: UUID | None = None,
        source: str | None = None,
        user_id: UUID | None = None,
        model: str | None = None,
        metadata: dict[str, Any] | None = None,
        persist: bool = True,
        worker_id: str | None = None,
    ) -> tuple[SecurityResult, SecurityScan | None]:
        request = SecurityRequest(
            text=text,
            source=source,
            user_id=user_id,
            model=model,
            metadata=metadata or {},
        )
        if request_id is not None:
            request.request_id = request_id

        result = await self.pipeline.inspect(request)
        digest = prompt_hash(text)
        stored_text = text if self.settings.persist_prompt_content else None

        log_payload: dict[str, Any] = {
            "request_id": str(result.request_id),
            "decision": result.decision.value,
            "score": result.score,
            "detector": result.metadata.get("detector"),
            "threshold": result.metadata.get("threshold"),
            "latency_ms": result.total_latency_ms,
            "prompt_hash": digest,
            "prompt_length": len(text),
            "model": model,
        }
        if self.settings.log_prompt_content:
            log_payload["prompt"] = text
        logger.info("security_scan", **log_payload)

        detector_name = str(result.metadata.get("detector", "unknown"))
        detector_latency = (
            result.detector_results[0].latency_ms if result.detector_results else result.total_latency_ms
        )
        self.metrics.observe_scan(
            decision=result.decision.value,
            pipeline_latency_ms=result.total_latency_ms,
            chunks=result.chunks_scanned,
            detector_name=detector_name,
            detector_score=result.score,
            detector_latency_ms=detector_latency,
        )

        scan: SecurityScan | None = None
        if persist:
            scan = await self.scans.create_from_result(
                result,
                prompt_hash=digest,
                prompt_length=len(text),
                prompt_text=stored_text,
                user_id=user_id,
                worker_id=worker_id,
            )
            logger.info("security_scan_persisted", scan_id=str(scan.id), request_id=str(result.request_id))
        return result, scan

    def list_detectors(self) -> list[dict[str, str]]:
        implemented = [
            {"name": name, "status": "implemented", "warning": "rule-based is development/testing only"}
            if name == "rule-based"
            else {"name": name, "status": "implemented", "warning": ""}
            for name in DETECTOR_REGISTRY
        ]
        prepared = [
            {"name": "prompt-shield", "status": "prepared", "warning": "not implemented"},
            {"name": "flan-t5", "status": "prepared", "warning": "not implemented"},
            {"name": "transformer", "status": "prepared", "warning": "not implemented"},
            {"name": "remote", "status": "prepared", "warning": "not implemented"},
            {"name": "ensemble", "status": "prepared", "warning": "not implemented"},
        ]
        return implemented + prepared

    async def evaluate(
        self,
        samples: list[dict[str, Any]],
        *,
        threshold: float | None = None,
        target_fprs: list[float] | None = None,
    ) -> EvaluationReport:
        scores: list[float] = []
        labels: list[bool] = []
        for sample in samples:
            result, _scan = await self.inspect(
                sample["text"],
                source="evaluate",
                persist=False,
            )
            scores.append(result.score)
            labels.append(bool(sample["is_injection"]))
        used_threshold = (
            threshold if threshold is not None else float(self.pipeline.policy.threshold or 0.5)
        )
        return evaluate_scores(
            labels,
            scores,
            threshold=used_threshold,
            target_fprs=target_fprs or (0.0005, 0.001, 0.005, 0.01),
        )

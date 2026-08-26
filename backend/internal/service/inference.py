from __future__ import annotations

import time
from uuid import UUID, uuid4

from dao.repositories.inference import InferenceRepository
from dao.repositories.worker import WorkerRepository
from pkg.apis.inference.models import InferenceRequest, ProtectedInferenceResponse
from pkg.runtime.status import InferenceStatus
from pkg.security.enforcement import EnforcementAction, Enforcer
from pkg.serving.router import ServingRouter

from internal.service.security import SecurityService
from internal.telemetry.logging import get_logger
from internal.telemetry.metrics import SpiderMetrics

logger = get_logger("spider-inference")


class InferenceService:
    def __init__(
        self,
        security: SecurityService,
        enforcer: Enforcer,
        router: ServingRouter,
        inferences: InferenceRepository,
        workers: WorkerRepository,
        metrics: SpiderMetrics,
    ) -> None:
        self.security = security
        self.enforcer = enforcer
        self.router = router
        self.inferences = inferences
        self.workers = workers
        self.metrics = metrics

    async def infer(
        self,
        request: InferenceRequest,
        *,
        user_id: UUID | None = None,
    ) -> ProtectedInferenceResponse:
        request_id = uuid4()
        started = time.perf_counter()

        security_result, scan = await self.security.inspect(
            request.prompt,
            request_id=request_id,
            source="inference",
            user_id=user_id,
            model=request.model,
            persist=True,
        )
        security_overhead_ms = security_result.total_latency_ms
        action = self.enforcer.resolve(security_result.decision)

        if action is not EnforcementAction.FORWARD:
            status = (
                InferenceStatus.REVIEW.value
                if action is EnforcementAction.HOLD
                else InferenceStatus.BLOCKED.value
            )
            e2e = (time.perf_counter() - started) * 1000.0
            record = await self.inferences.create(
                request_id=request_id,
                model=request.model,
                status=status,
                decision=security_result.decision.value,
                scan_id=scan.id if scan else None,
                user_id=user_id,
                worker_id=None,
                end_to_end_latency_ms=e2e,
                security_overhead_ms=security_overhead_ms,
                inference_latency_ms=None,
                output_preview=None,
            )
            await self.inferences.add_event(
                record.id,
                "blocked" if status == InferenceStatus.BLOCKED.value else "held",
                {"action": action.value},
            )
            self.metrics.observe_inference(
                status=status,
                e2e_latency_ms=e2e,
                security_overhead_ms=security_overhead_ms,
            )
            logger.info(
                "inference_not_forwarded",
                request_id=str(request_id),
                decision=security_result.decision.value,
                action=action.value,
            )
            return ProtectedInferenceResponse(
                request_id=request_id,
                status=status,
                decision=security_result.decision,
                model=request.model,
                output=None,
                security=security_result,
                security_overhead_ms=security_overhead_ms,
                inference_latency_ms=None,
                end_to_end_latency_ms=e2e,
            )

        worker_inventory = await self.workers.list_resources()
        llm_started = time.perf_counter()
        llm_response, selected = await self.router.route(request, workers=worker_inventory)
        inference_latency_ms = (time.perf_counter() - llm_started) * 1000.0
        e2e = (time.perf_counter() - started) * 1000.0
        worker_id = selected.worker_id if selected else None
        preview = (llm_response.output or "")[:512] or None

        record = await self.inferences.create(
            request_id=request_id,
            model=request.model,
            status=InferenceStatus.COMPLETED.value,
            decision=security_result.decision.value,
            scan_id=scan.id if scan else None,
            user_id=user_id,
            worker_id=worker_id,
            end_to_end_latency_ms=e2e,
            security_overhead_ms=security_overhead_ms,
            inference_latency_ms=inference_latency_ms,
            output_preview=preview,
        )
        await self.inferences.add_event(
            record.id,
            "completed",
            {"provider": getattr(self.router.provider, "name", "unknown")},
        )
        self.metrics.observe_inference(
            status=InferenceStatus.COMPLETED.value,
            e2e_latency_ms=e2e,
            security_overhead_ms=security_overhead_ms,
        )
        logger.info(
            "inference_completed",
            request_id=str(request_id),
            decision=security_result.decision.value,
            worker_id=worker_id,
            model=request.model,
            latency_ms=e2e,
        )
        return ProtectedInferenceResponse(
            request_id=request_id,
            status=InferenceStatus.COMPLETED.value,
            decision=security_result.decision,
            model=request.model,
            output=llm_response.output,
            security=security_result,
            security_overhead_ms=security_overhead_ms,
            inference_latency_ms=inference_latency_ms,
            end_to_end_latency_ms=e2e,
            worker_id=worker_id,
        )

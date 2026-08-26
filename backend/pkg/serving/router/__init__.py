from __future__ import annotations

from pkg.apis.inference.models import InferenceRequest, InferenceResponse
from pkg.apis.serving.models import ServingRequest
from pkg.apis.worker.models import WorkerResource
from pkg.scheduler.base import Scheduler
from pkg.serving.providers.base import LLMProvider


class ServingRouter:
    """Route an allowed request to an LLM provider.

    Scheduler selection is used when worker inventory is supplied. Provider
    adapters stay out of HTTP handlers.
    """

    def __init__(
        self,
        provider: LLMProvider,
        scheduler: Scheduler | None = None,
    ) -> None:
        self.provider = provider
        self.scheduler = scheduler

    async def route(
        self,
        request: InferenceRequest,
        *,
        workers: list[WorkerResource] | None = None,
    ) -> tuple[InferenceResponse, WorkerResource | None]:
        selected: WorkerResource | None = None
        if self.scheduler is not None and workers:
            serving = ServingRequest(model=request.model, request=request)
            selected = await self.scheduler.select_worker(serving, workers)
        response = await self.provider.infer(request)
        return response, selected

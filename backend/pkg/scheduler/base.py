from __future__ import annotations

from typing import Protocol

from pkg.apis.serving.models import ServingRequest
from pkg.apis.worker.models import WorkerResource


class Scheduler(Protocol):
    name: str

    async def select_worker(
        self,
        workload: ServingRequest,
        workers: list[WorkerResource],
    ) -> WorkerResource | None: ...

from __future__ import annotations

from pkg.apis.serving.models import ServingRequest
from pkg.apis.worker.models import WorkerResource


class RoundRobinScheduler:
    name = "round-robin"

    async def select_worker(
        self, workload: ServingRequest, workers: list[WorkerResource]
    ) -> WorkerResource | None:
        raise NotImplementedError("RoundRobinScheduler is reserved for cluster experiments.")


class GPUAwareScheduler:
    name = "gpu-aware"

    async def select_worker(
        self, workload: ServingRequest, workers: list[WorkerResource]
    ) -> WorkerResource | None:
        raise NotImplementedError("GPUAwareScheduler requires VRAM reservation accounting.")


class ModelLocalityScheduler:
    name = "model-locality"

    async def select_worker(
        self, workload: ServingRequest, workers: list[WorkerResource]
    ) -> WorkerResource | None:
        raise NotImplementedError("ModelLocalityScheduler requires loaded-model inventory.")


class FairShareScheduler:
    name = "fair-share"

    async def select_worker(
        self, workload: ServingRequest, workers: list[WorkerResource]
    ) -> WorkerResource | None:
        raise NotImplementedError("FairShareScheduler requires tenant usage accounting.")


class LatencyAwareScheduler:
    name = "latency-aware"

    async def select_worker(
        self, workload: ServingRequest, workers: list[WorkerResource]
    ) -> WorkerResource | None:
        raise NotImplementedError("LatencyAwareScheduler requires latency telemetry.")


class ResearchScheduler:
    name = "research"

    async def select_worker(
        self, workload: ServingRequest, workers: list[WorkerResource]
    ) -> WorkerResource | None:
        raise NotImplementedError("ResearchScheduler is a hook for experiment policies.")

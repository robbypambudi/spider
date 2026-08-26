from __future__ import annotations

from datetime import UTC, datetime, timedelta

import pytest
from pkg.apis.inference.models import InferenceRequest
from pkg.apis.serving.models import ServingRequest
from pkg.apis.worker.models import LoadedModel, WorkerResource, WorkerResources
from pkg.reconciler.worker import WorkerReconciler
from pkg.scheduler.least_loaded import LeastLoadedScheduler
from pkg.serving.providers.mock import MockLLMProvider


def _worker(worker_id: str, *, status: str = "ONLINE", running: int = 0) -> WorkerResource:
    return WorkerResource(
        worker_id=worker_id,
        hostname=worker_id,
        status=status,
        resources=WorkerResources(cpu_total=4, memory_total_mb=8192, gpus=[]),
        models=[LoadedModel(name="mock-llm", status="READY")],
        running_requests=running,
    )


@pytest.mark.asyncio
async def test_scheduler_selects_valid_serving_worker() -> None:
    scheduler = LeastLoadedScheduler()
    workers = [
        _worker("offline", status="OFFLINE", running=0),
        _worker("busy", status="ONLINE", running=5),
        _worker("idle", status="ONLINE", running=1),
    ]
    workload = ServingRequest(
        model="mock-llm",
        request=InferenceRequest(model="mock-llm", prompt="hello"),
    )
    selected = await scheduler.select_worker(workload, workers)
    assert selected is not None
    assert selected.worker_id == "idle"


@pytest.mark.asyncio
async def test_scheduler_skips_offline_workers() -> None:
    scheduler = LeastLoadedScheduler()
    selected = await scheduler.select_worker(
        ServingRequest(model="mock-llm", request=InferenceRequest(model="mock-llm", prompt="x")),
        [_worker("a", status="OFFLINE"), _worker("b", status="ERROR")],
    )
    assert selected is None


@pytest.mark.asyncio
async def test_worker_offline_detection() -> None:
    workers = [_worker("stale", status="ONLINE")]
    heartbeats = {"stale": datetime.now(UTC) - timedelta(seconds=120)}
    marked: list[str] = []

    async def list_workers() -> list[WorkerResource]:
        return workers

    async def mark_offline(worker_id: str) -> None:
        marked.append(worker_id)

    async def last_heartbeat(worker_id: str) -> datetime | None:
        return heartbeats.get(worker_id)

    reconciler = WorkerReconciler(
        list_workers=list_workers,
        mark_offline=mark_offline,
        last_heartbeat=last_heartbeat,
        offline_timeout_seconds=30,
    )
    result = await reconciler.reconcile()
    assert result == ["stale"]
    assert marked == ["stale"]


@pytest.mark.asyncio
async def test_mock_llm_provider() -> None:
    provider = MockLLMProvider()
    response = await provider.infer(InferenceRequest(model="mock-llm", prompt="hello"))
    assert response.output is not None
    assert "mock-llm" in response.output
    assert len(provider.calls) == 1

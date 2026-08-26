from __future__ import annotations

import pytest
from httpx import AsyncClient
from pkg.protocol.worker import WORKER_TOKEN_HEADER


def _worker_payload(worker_id: str = "node-cpu-01") -> dict[str, object]:
    return {
        "worker_id": worker_id,
        "hostname": "gpu-lab-01",
        "site": "lab-a",
        "version": "0.1.0",
        "resources": {
            "cpu_total": 8,
            "memory_total_mb": 32768,
            "gpus": [],
        },
        "models": [{"name": "mock-llm", "status": "READY"}],
        "running_requests": 0,
    }


@pytest.mark.asyncio
async def test_worker_registration_and_heartbeat(client: AsyncClient) -> None:
    headers = {WORKER_TOKEN_HEADER: "test-worker-token"}
    registered = await client.post(
        "/api/v1/workers/register",
        json=_worker_payload(),
        headers=headers,
    )
    assert registered.status_code == 200, registered.text
    body = registered.json()
    assert body["worker_id"] == "node-cpu-01"
    assert body["status"] == "ONLINE"

    beat = await client.post(
        "/api/v1/workers/node-cpu-01/heartbeat",
        json={
            "worker_id": "node-cpu-01",
            "status": "ONLINE",
            "running_requests": 1,
            "models": [{"name": "mock-llm", "status": "READY"}],
        },
        headers=headers,
    )
    assert beat.status_code == 200, beat.text
    assert beat.json()["running_requests"] == 1


@pytest.mark.asyncio
async def test_worker_rejects_bad_token(client: AsyncClient) -> None:
    response = await client.post(
        "/api/v1/workers/register",
        json=_worker_payload("evil"),
        headers={WORKER_TOKEN_HEADER: "wrong"},
    )
    assert response.status_code == 401

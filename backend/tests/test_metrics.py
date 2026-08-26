from __future__ import annotations

import pytest
from httpx import AsyncClient


@pytest.mark.asyncio
async def test_prometheus_metrics_expose_security_series(
    client: AsyncClient, auth_headers: dict[str, str]
) -> None:
    await client.post(
        "/api/v1/security/scan",
        json={"text": "Explain TCP congestion control."},
        headers=auth_headers,
    )
    response = await client.get("/metrics")
    assert response.status_code == 200
    body = response.text
    assert "spider_security_scans_total" in body
    assert "spider_security_allowed_total" in body
    assert "spider_security_pipeline_latency_seconds" in body

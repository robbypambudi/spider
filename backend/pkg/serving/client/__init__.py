from __future__ import annotations

import httpx

from pkg.apis.inference.models import InferenceRequest, InferenceResponse


class ServingClient:
    """HTTP client used by workers to talk to a remote serving runtime later."""

    def __init__(self, base_url: str, timeout: float = 30.0) -> None:
        self.base_url = base_url.rstrip("/")
        self.timeout = timeout

    async def infer(self, request: InferenceRequest) -> InferenceResponse:
        async with httpx.AsyncClient(timeout=self.timeout) as client:
            response = await client.post(
                f"{self.base_url}/v1/completions",
                json={"model": request.model, "prompt": request.prompt},
            )
            response.raise_for_status()
            payload = response.json()
        return InferenceResponse.model_validate(payload)

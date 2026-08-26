from __future__ import annotations

from typing import Protocol

from pkg.apis.inference.models import InferenceRequest, InferenceResponse


class LLMProvider(Protocol):
    name: str

    async def infer(self, request: InferenceRequest) -> InferenceResponse: ...

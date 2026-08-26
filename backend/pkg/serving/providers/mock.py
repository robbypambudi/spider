from __future__ import annotations

from pkg.apis.inference.models import InferenceRequest, InferenceResponse


class MockLLMProvider:
    """Deterministic local provider. No GPU required."""

    name = "mock"

    def __init__(self) -> None:
        self.calls: list[InferenceRequest] = []

    async def infer(self, request: InferenceRequest) -> InferenceResponse:
        self.calls.append(request)
        preview = request.prompt.strip().replace("\n", " ")
        if len(preview) > 80:
            preview = preview[:77] + "..."
        output = (
            f"[mock-llm] model={request.model} "
            f"echo={preview!r} tokens={request.max_tokens}"
        )
        return InferenceResponse(
            model=request.model,
            output=output,
            finish_reason="stop",
            usage={"prompt_tokens": max(1, len(request.prompt) // 4), "completion_tokens": 16},
            metadata={"provider": self.name},
        )

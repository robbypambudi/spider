from __future__ import annotations

from pkg.apis.inference.models import InferenceRequest, InferenceResponse


class OpenAICompatibleProvider:
    name = "openai-compatible"

    async def infer(self, request: InferenceRequest) -> InferenceResponse:
        raise NotImplementedError(
            "OpenAICompatibleProvider requires a base URL and API key. See docs/serving.md."
        )


class VLLMProvider:
    name = "vllm"

    async def infer(self, request: InferenceRequest) -> InferenceResponse:
        raise NotImplementedError("VLLMProvider requires a vLLM OpenAI-compatible endpoint.")


class TGIProvider:
    name = "tgi"

    async def infer(self, request: InferenceRequest) -> InferenceResponse:
        raise NotImplementedError("TGIProvider requires a Text Generation Inference endpoint.")


class OllamaProvider:
    name = "ollama"

    async def infer(self, request: InferenceRequest) -> InferenceResponse:
        raise NotImplementedError("OllamaProvider requires a local Ollama endpoint.")


class TritonProvider:
    name = "triton"

    async def infer(self, request: InferenceRequest) -> InferenceResponse:
        raise NotImplementedError("TritonProvider requires a Triton Inference Server contract.")


class RemoteHTTPProvider:
    name = "remote-http"

    async def infer(self, request: InferenceRequest) -> InferenceResponse:
        raise NotImplementedError("RemoteHTTPProvider requires a serving URL. See docs/serving.md.")

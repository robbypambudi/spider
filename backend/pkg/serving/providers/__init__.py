from pkg.serving.providers.base import LLMProvider
from pkg.serving.providers.mock import MockLLMProvider
from pkg.serving.providers.prepared import (
    OllamaProvider,
    OpenAICompatibleProvider,
    RemoteHTTPProvider,
    TGIProvider,
    TritonProvider,
    VLLMProvider,
)

PROVIDER_REGISTRY: dict[str, type] = {
    "mock": MockLLMProvider,
}

__all__ = [
    "LLMProvider",
    "MockLLMProvider",
    "OllamaProvider",
    "OpenAICompatibleProvider",
    "PROVIDER_REGISTRY",
    "RemoteHTTPProvider",
    "TGIProvider",
    "TritonProvider",
    "VLLMProvider",
]

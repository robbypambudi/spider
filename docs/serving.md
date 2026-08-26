# Serving

`LLMProvider` is the only inference backend interface. Handlers never talk to vLLM/Ollama/OpenAI directly.

Implemented: `MockLLMProvider`.

Prepared: `OpenAICompatibleProvider`, `VLLMProvider`, `TGIProvider`, `OllamaProvider`, `TritonProvider`, `RemoteHTTPProvider`.

OpenAI-compatible `POST /v1/chat/completions` is a planned façade over the same pipeline (not in this milestone).

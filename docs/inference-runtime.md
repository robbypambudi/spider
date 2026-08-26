# Inference runtime

`POST /api/v1/inference` always runs the security pipeline first.

```text
validate → SecurityPipeline.inspect → Enforcer
  BLOCK/HOLD → persist + return (LLM not called)
  ALLOW → ServingRouter → LLMProvider → persist + metrics
```

`MockLLMProvider` is the default so local development needs no GPU.

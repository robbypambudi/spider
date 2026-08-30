# Architecture

SPIDER sits between clients and LLM serving. The security pipeline is the product. Cluster components exist to deploy and measure that pipeline under multi-node serving load.

```text
                    Vite UI
                       │
                 Go API (chi)
          ┌────────────┼────────────┐
    Inference API  Security API  Cluster API
          │
   SecurityPipeline
   preprocess → chunk → detect → aggregate → policy → enforce
          │
     ALLOW / BLOCK
          │
    ServingRouter → Scheduler → Serving Node / MockLLM
          │
      Metrics + DAO
```

## Prompt-Shield detector

ML detection uses fine-tuned Flan-T5 classifiers from the
[Prompt-Shield Hugging Face collection](https://huggingface.co/collections/robbypambudi/prompt-shield).
A thin Python sidecar (`backend/cmd/prompt-shield`) loads the model; the Go API calls it via HTTP when
`SPIDER_DEFAULT_DETECTOR=prompt-shield`.

## Layering

* `cmd/` — executable processes only (`api`, `controller`, `worker`, `cli`, `prompt-shield` sidecar).
* `internal/` — application services and HTTP adapters. Handlers do not contain detector logic or raw SQL.
* `pkg/` — reusable domain libraries (pipeline, serving providers, scheduler, reconciler).
* `dao/` — persistence. Repositories are the only database access path.

## Identity

Describe SPIDER as a **runtime defense framework for prompt injection detection**, with LLM-serving cluster integration for evaluation. Do not describe it as a GPU cloud.

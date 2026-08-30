"""Prompt-Shield inference sidecar.

Loads fine-tuned Flan-T5 classifiers from the Prompt-Shield collection:
https://huggingface.co/collections/robbypambudi/prompt-shield

Models:
  - robbypambudi/prompt-shield-flan-t5-small (default, ~61M params)
  - robbypambudi/prompt-shield-flan-t5-base (~248M params)
"""

from __future__ import annotations

import os
import time
from typing import Any

import torch
import uvicorn
from fastapi import FastAPI
from pydantic import BaseModel, Field
from transformers import AutoModelForSequenceClassification, AutoTokenizer

DEFAULT_MODEL = os.getenv(
    "SPIDER_PROMPT_SHIELD_MODEL",
    "robbypambudi/prompt-shield-flan-t5-small",
)

CATALOG = [
    {
        "id": "robbypambudi/prompt-shield-flan-t5-small",
        "name": "Prompt-Shield Flan-T5 Small",
        "params": "60.8M",
    },
    {
        "id": "robbypambudi/prompt-shield-flan-t5-base",
        "name": "Prompt-Shield Flan-T5 Base",
        "params": "0.2B",
    },
]
MAX_LENGTH = int(os.getenv("SPIDER_PROMPT_SHIELD_MAX_LENGTH", "512"))

app = FastAPI(title="Prompt-Shield Detector", version="0.1.0")

_tokenizer: Any = None
_model: Any = None
_loaded_model_id: str = ""


class DetectRequest(BaseModel):
    text: str
    model: str | None = None


class DetectResponse(BaseModel):
    score: float
    is_injection: bool
    latency_ms: float
    model: str
    label_probs: dict[str, float] | None = None
    logits: list[float] | None = None


def load_model(model_id: str) -> None:
    global _tokenizer, _model, _loaded_model_id
    if _loaded_model_id == model_id and _model is not None:
        return
    _tokenizer = AutoTokenizer.from_pretrained(model_id)
    _model = AutoModelForSequenceClassification.from_pretrained(model_id)
    _model.eval()
    _loaded_model_id = model_id


@app.on_event("startup")
async def startup() -> None:
    load_model(DEFAULT_MODEL)


@app.get("/health")
async def health() -> dict[str, str]:
    return {"status": "ok", "model": _loaded_model_id}


@app.get("/models")
async def models() -> dict[str, object]:
    return {
        "active": _loaded_model_id or DEFAULT_MODEL,
        "catalog": CATALOG,
        "collection": "https://huggingface.co/collections/robbypambudi/prompt-shield",
    }


@app.post("/detect", response_model=DetectResponse)
async def detect(body: DetectRequest) -> DetectResponse:
    model_id = body.model or DEFAULT_MODEL
    load_model(model_id)
    started = time.perf_counter()

    enc = _tokenizer(
        body.text,
        return_tensors="pt",
        truncation=True,
        max_length=MAX_LENGTH,
    )
    with torch.no_grad():
        logits = _model(**enc).logits[0]
        probs = torch.softmax(logits, dim=-1)
        id2label = getattr(_model.config, "id2label", {0: "BENIGN", 1: "INJECTION"})
        label_probs = {id2label[i]: float(probs[i].item()) for i in range(len(probs))}
        p_injection = float(label_probs.get("INJECTION", float(probs[1].item())))

    latency_ms = (time.perf_counter() - started) * 1000.0
    return DetectResponse(
        score=p_injection,
        is_injection=p_injection >= 0.5,
        latency_ms=latency_ms,
        model=model_id,
        label_probs=label_probs,
        logits=[float(x) for x in logits.tolist()],
    )


def main() -> None:
    port = int(os.getenv("SPIDER_PROMPT_SHIELD_PORT", "8081"))
    uvicorn.run(app, host="0.0.0.0", port=port)


if __name__ == "__main__":
    main()

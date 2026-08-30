"""Prompt-Shield inference sidecar.

Loads fine-tuned Flan-T5 classifiers from the Prompt-Shield collection:
https://huggingface.co/collections/robbypambudi/prompt-shield

Scoring and chunking follow spider-internal/labs/eval_chunked.py methodology.
"""

from __future__ import annotations

import os
import time
from typing import Any, Literal

import torch
import uvicorn
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field
from transformers import AutoModelForSequenceClassification, AutoTokenizer

DEFAULT_MODEL = os.getenv(
    "SPIDER_PROMPT_SHIELD_MODEL",
    "robbypambudi/prompt-shield-flan-t5-small",
)
DEFAULT_THRESHOLD = float(os.getenv("SPIDER_DEFAULT_THRESHOLD", "0.5"))
HF_TOKEN = os.getenv("HF_TOKEN") or None

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

app = FastAPI(title="Prompt-Shield Detector", version="0.2.0")

_tokenizer: Any = None
_model: Any = None
_loaded_model_id: str = ""


class DetectRequest(BaseModel):
    text: str
    model: str | None = None
    threshold: float | None = None


class DetectResponse(BaseModel):
    score: float
    is_injection: bool
    latency_ms: float
    model: str
    label_probs: dict[str, float] | None = None
    logits: list[float] | None = None


class ChunkItem(BaseModel):
    index: int
    text: str
    start: int
    end: int


class ChunkRequest(BaseModel):
    text: str
    chunk_size: int = Field(default=256, ge=1)
    overlap: int = Field(default=0, ge=0)
    unit: Literal["tokens", "chars"] = "tokens"
    model: str | None = None


class ChunkResponse(BaseModel):
    chunks: list[ChunkItem]
    n_chunks: int
    unit: str
    chunk_size: int
    overlap: int


def chunk_sequence(seq: list, size: int, overlap: int) -> list[list]:
    n = len(seq)
    if n == 0:
        return [seq]
    if size <= 0 or n <= size:
        return [seq]
    step = max(1, size - overlap)
    chunks: list[list] = []
    start = 0
    while start < n:
        chunks.append(seq[start : start + size])
        if start + size >= n:
            break
        start += step
    return chunks


def chunk_spans(n: int, size: int, overlap: int) -> list[tuple[int, int]]:
    windows = chunk_sequence(list(range(n)), size, overlap)
    if n == 0:
        return [(0, 0)]
    return [(w[0], w[-1] + 1) for w in windows]


def _prepare_window_ids(tokenizer, token_ids: list[int]) -> list[int]:
    """Append EOS per window (T5 pools at last </s>)."""
    ids = list(token_ids)
    cls_id = getattr(tokenizer, "cls_token_id", None)
    sep_id = getattr(tokenizer, "sep_token_id", None)
    if cls_id is not None and cls_id != tokenizer.eos_token_id:
        if not ids or ids[0] != cls_id:
            ids = [cls_id] + ids
        if sep_id is not None and (not ids or ids[-1] != sep_id):
            ids.append(sep_id)
        return ids
    eos_id = tokenizer.eos_token_id
    if eos_id is not None and (not ids or ids[-1] != eos_id):
        ids.append(eos_id)
    return ids


def _forward_scores(model, input_ids, attention_mask) -> torch.Tensor:
    with torch.no_grad():
        logits = model(input_ids=input_ids, attention_mask=attention_mask).logits
    return torch.nn.functional.softmax(logits, dim=1)[:, -1]


def score_windows(model, tokenizer, windows: list[list[int]], device: str) -> list[float]:
    if not windows:
        return []
    prepared = [_prepare_window_ids(tokenizer, w) for w in windows]
    max_len = max(len(w) for w in prepared)
    pad_id = tokenizer.pad_token_id
    rows: list[list[int]] = []
    masks: list[list[int]] = []
    for w in prepared:
        pad = max_len - len(w)
        rows.append(w + [pad_id] * pad)
        masks.append([1] * len(w) + [0] * pad)
    input_ids = torch.tensor(rows, dtype=torch.long, device=device)
    attention_mask = torch.tensor(masks, dtype=torch.long, device=device)

    eos_id = tokenizer.eos_token_id
    is_t5 = getattr(model.config, "model_type", "") == "t5"
    if is_t5 and eos_id is not None:
        eos_counts = (input_ids == eos_id).sum(dim=1)
        if int(eos_counts.min()) < 1 or int(eos_counts.max()) != int(eos_counts.min()):
            scores = [
                _forward_scores(model, input_ids[i : i + 1], attention_mask[i : i + 1])
                for i in range(input_ids.size(0))
            ]
            return [float(s.item()) for s in torch.cat(scores)]

    probs = _forward_scores(model, input_ids, attention_mask)
    return [float(p.item()) for p in probs]


def score_text(model, tokenizer, text: str, device: str) -> tuple[float, dict[str, float], list[float]]:
    token_ids = tokenizer(text, add_special_tokens=False, truncation=False)["input_ids"]
    window = _prepare_window_ids(tokenizer, token_ids)
    input_ids = torch.tensor([window], dtype=torch.long, device=device)
    attention_mask = torch.tensor([[1] * len(window)], dtype=torch.long, device=device)

    with torch.no_grad():
        logits = model(input_ids=input_ids, attention_mask=attention_mask).logits[0]
        probs = torch.softmax(logits, dim=-1)

    id2label = getattr(model.config, "id2label", {0: "BENIGN", 1: "INJECTION"})
    label_probs = {id2label[i]: float(probs[i].item()) for i in range(len(probs))}
    p_injection = float(label_probs.get("INJECTION", float(probs[-1].item())))
    return p_injection, label_probs, [float(x) for x in logits.tolist()]


def load_model(model_id: str) -> None:
    global _tokenizer, _model, _loaded_model_id
    if _loaded_model_id == model_id and _model is not None:
        return
    kwargs: dict[str, Any] = {"use_fast": False, "trust_remote_code": True}
    if HF_TOKEN:
        kwargs["token"] = HF_TOKEN
    _tokenizer = AutoTokenizer.from_pretrained(model_id, **kwargs)
    if _tokenizer.pad_token is None:
        _tokenizer.pad_token = _tokenizer.eos_token
    model_kwargs: dict[str, Any] = {"num_labels": 2, "trust_remote_code": True}
    if HF_TOKEN:
        model_kwargs["token"] = HF_TOKEN
    _model = AutoModelForSequenceClassification.from_pretrained(model_id, **model_kwargs)
    _model.eval()
    _loaded_model_id = model_id


def get_device() -> str:
    return "cuda" if torch.cuda.is_available() else "cpu"


@app.on_event("startup")
async def startup() -> None:
    load_model(DEFAULT_MODEL)
    device = get_device()
    if _model is not None:
        _model.to(device)


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
    device = get_device()
    _model.to(device)
    started = time.perf_counter()

    threshold = body.threshold if body.threshold is not None else DEFAULT_THRESHOLD
    p_injection, label_probs, logits = score_text(_model, _tokenizer, body.text, device)

    latency_ms = (time.perf_counter() - started) * 1000.0
    return DetectResponse(
        score=p_injection,
        is_injection=p_injection >= threshold,
        latency_ms=latency_ms,
        model=model_id,
        label_probs=label_probs,
        logits=logits,
    )


@app.post("/chunk", response_model=ChunkResponse)
async def chunk(body: ChunkRequest) -> ChunkResponse:
    if body.overlap >= body.chunk_size:
        raise HTTPException(status_code=400, detail="overlap must be smaller than chunk_size")

    model_id = body.model or DEFAULT_MODEL
    if body.unit == "tokens":
        load_model(model_id)

    text = body.text
    if body.unit == "chars":
        pieces = chunk_sequence(text, body.chunk_size, body.overlap)
        if not pieces:
            pieces = [text]
        spans = chunk_spans(len(text), body.chunk_size, body.overlap)
    else:
        token_ids = _tokenizer(text, add_special_tokens=False, truncation=False)["input_ids"]
        token_windows = chunk_sequence(token_ids, body.chunk_size, body.overlap)
        spans = chunk_spans(len(token_ids), body.chunk_size, body.overlap)
        pieces = [
            _tokenizer.decode(window, skip_special_tokens=True) for window in token_windows
        ]

    items = [
        ChunkItem(index=i, text=piece, start=start, end=end)
        for i, (piece, (start, end)) in enumerate(zip(pieces, spans, strict=True))
    ]
    return ChunkResponse(
        chunks=items,
        n_chunks=len(items),
        unit=body.unit,
        chunk_size=body.chunk_size,
        overlap=body.overlap,
    )


def main() -> None:
    port = int(os.getenv("SPIDER_PROMPT_SHIELD_PORT", "8081"))
    uvicorn.run(app, host="0.0.0.0", port=port)


if __name__ == "__main__":
    main()

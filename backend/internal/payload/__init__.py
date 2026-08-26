from __future__ import annotations

from typing import Any
from uuid import UUID

from pydantic import BaseModel, Field


class LoginRequest(BaseModel):
    email: str
    password: str


class TokenResponse(BaseModel):
    access_token: str
    token_type: str = "bearer"
    role: str
    email: str


class ScanRequest(BaseModel):
    text: str
    source: str | None = None
    model: str | None = None
    metadata: dict[str, Any] = Field(default_factory=dict)


class DetectorView(BaseModel):
    detector: str
    score: float
    is_injection: bool
    latency_ms: float | None = None
    metadata: dict[str, Any] = Field(default_factory=dict)


class ScanResponse(BaseModel):
    scan_id: UUID | None = None
    request_id: UUID
    decision: str
    score: float
    chunks_scanned: int
    policy: str
    threshold: float | None = None
    latency_ms: float
    detectors: list[DetectorView]
    model: str | None = None


class InferenceHttpRequest(BaseModel):
    model: str
    prompt: str
    max_tokens: int | None = 256
    temperature: float | None = 0.0
    security: dict[str, Any] = Field(default_factory=lambda: {"enabled": True})


class LabeledSample(BaseModel):
    text: str
    is_injection: bool


class EvaluateRequest(BaseModel):
    samples: list[LabeledSample]
    threshold: float | None = None
    target_fpr: list[float] = Field(default_factory=lambda: [0.0005, 0.001, 0.005, 0.01])

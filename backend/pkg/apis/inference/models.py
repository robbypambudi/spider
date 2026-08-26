from __future__ import annotations

from typing import Any
from uuid import UUID, uuid4

from pydantic import BaseModel, Field

from pkg.apis.security.models import SecurityDecision, SecurityResult


class InferenceRequest(BaseModel):
    model: str
    prompt: str
    max_tokens: int | None = 256
    temperature: float | None = 0.0
    security_enabled: bool = True
    metadata: dict[str, Any] = Field(default_factory=dict)


class InferenceResponse(BaseModel):
    request_id: UUID = Field(default_factory=uuid4)
    model: str
    output: str | None = None
    finish_reason: str | None = None
    usage: dict[str, int] = Field(default_factory=dict)
    metadata: dict[str, Any] = Field(default_factory=dict)


class ProtectedInferenceResponse(BaseModel):
    request_id: UUID
    status: str
    decision: SecurityDecision
    model: str
    output: str | None = None
    security: SecurityResult
    security_overhead_ms: float
    inference_latency_ms: float | None = None
    end_to_end_latency_ms: float
    worker_id: str | None = None
    metadata: dict[str, Any] = Field(default_factory=dict)

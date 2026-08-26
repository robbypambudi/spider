from __future__ import annotations

from enum import StrEnum
from typing import Any
from uuid import UUID, uuid4

from pydantic import BaseModel, Field


class SecurityDecision(StrEnum):
    ALLOW = "ALLOW"
    BLOCK = "BLOCK"
    REVIEW = "REVIEW"
    ERROR = "ERROR"


class SecurityRequest(BaseModel):
    request_id: UUID = Field(default_factory=uuid4)
    text: str
    source: str | None = None
    user_id: UUID | None = None
    model: str | None = None
    metadata: dict[str, Any] = Field(default_factory=dict)


class DetectionResult(BaseModel):
    detector: str
    score: float
    is_injection: bool
    threshold: float | None = None
    latency_ms: float
    metadata: dict[str, Any] = Field(default_factory=dict)


class AggregatedDetectionResult(BaseModel):
    score: float
    is_injection: bool
    chunks_scanned: int
    detector_results: list[DetectionResult] = Field(default_factory=list)
    metadata: dict[str, Any] = Field(default_factory=dict)


class SecurityResult(BaseModel):
    request_id: UUID
    decision: SecurityDecision
    score: float
    detector_results: list[DetectionResult]
    chunks_scanned: int
    total_latency_ms: float
    policy: str
    metadata: dict[str, Any] = Field(default_factory=dict)

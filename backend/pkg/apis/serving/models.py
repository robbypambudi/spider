from __future__ import annotations

from typing import Any

from pydantic import BaseModel, Field

from pkg.apis.inference.models import InferenceRequest


class ServingRequest(BaseModel):
    """Workload handed to the scheduler after a request is allowed."""

    model: str
    request: InferenceRequest
    site: str | None = None
    metadata: dict[str, Any] = Field(default_factory=dict)


class ServingNodeView(BaseModel):
    worker_id: str
    status: str
    models: list[str] = Field(default_factory=list)
    running_requests: int = 0

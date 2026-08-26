from __future__ import annotations

from typing import Any

from pydantic import BaseModel, Field


class PolicyConfig(BaseModel):
    name: str
    kind: str = "threshold"
    threshold: float = 0.5
    action_on_detection: str = "block"
    metadata: dict[str, Any] = Field(default_factory=dict)

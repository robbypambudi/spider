from __future__ import annotations

from pydantic import BaseModel


class HealthStatus(BaseModel):
    status: str
    service: str
    version: str


class ReadinessStatus(BaseModel):
    status: str
    database: bool
    redis: bool | None = None

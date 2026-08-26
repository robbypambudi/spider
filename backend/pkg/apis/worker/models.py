from __future__ import annotations

from typing import Any

from pydantic import BaseModel, Field


class GPUResource(BaseModel):
    index: int
    vendor: str = "unknown"
    name: str
    memory_total_mb: int = 0
    memory_used_mb: int = 0
    utilization: int = 0


class WorkerResources(BaseModel):
    cpu_total: int
    memory_total_mb: int
    gpus: list[GPUResource] = Field(default_factory=list)


class LoadedModel(BaseModel):
    name: str
    status: str = "READY"


class WorkerResource(BaseModel):
    worker_id: str
    hostname: str
    site: str | None = None
    version: str = "0.1.0"
    status: str = "ONLINE"
    resources: WorkerResources
    models: list[LoadedModel] = Field(default_factory=list)
    running_requests: int = 0
    metadata: dict[str, Any] = Field(default_factory=dict)


class WorkerHeartbeat(BaseModel):
    worker_id: str
    status: str = "ONLINE"
    resources: WorkerResources | None = None
    models: list[LoadedModel] = Field(default_factory=list)
    running_requests: int = 0
    metadata: dict[str, Any] = Field(default_factory=dict)

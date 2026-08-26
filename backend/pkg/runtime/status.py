from __future__ import annotations

from enum import StrEnum


class WorkerStatus(StrEnum):
    REGISTERING = "REGISTERING"
    ONLINE = "ONLINE"
    BUSY = "BUSY"
    DRAINING = "DRAINING"
    OFFLINE = "OFFLINE"
    ERROR = "ERROR"


class InferenceStatus(StrEnum):
    PENDING = "pending"
    SCANNING = "scanning"
    BLOCKED = "blocked"
    REVIEW = "review"
    ROUTING = "routing"
    COMPLETED = "completed"
    FAILED = "failed"


class ModelStatus(StrEnum):
    LOADING = "LOADING"
    READY = "READY"
    UNLOADING = "UNLOADING"
    ERROR = "ERROR"

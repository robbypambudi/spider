from __future__ import annotations

from datetime import datetime
from uuid import UUID

from sqlalchemy import DateTime, ForeignKey, Integer, String, Text
from sqlalchemy.orm import Mapped, mapped_column

from dao.models.base import Base, TimestampMixin, UUIDPrimaryKeyMixin


class Worker(UUIDPrimaryKeyMixin, TimestampMixin, Base):
    __tablename__ = "workers"

    worker_id: Mapped[str] = mapped_column(String(128), unique=True, nullable=False, index=True)
    hostname: Mapped[str] = mapped_column(String(255), nullable=False)
    site: Mapped[str | None] = mapped_column(String(128), nullable=True)
    version: Mapped[str] = mapped_column(String(32), nullable=False, default="0.1.0")
    status: Mapped[str] = mapped_column(String(32), nullable=False, default="REGISTERING")
    cpu_total: Mapped[int] = mapped_column(Integer, nullable=False, default=1)
    memory_total_mb: Mapped[int] = mapped_column(Integer, nullable=False, default=0)
    running_requests: Mapped[int] = mapped_column(Integer, nullable=False, default=0)
    last_heartbeat_at: Mapped[datetime | None] = mapped_column(DateTime(timezone=True), nullable=True)
    models_json: Mapped[str] = mapped_column(Text, nullable=False, default="[]")
    metadata_json: Mapped[str] = mapped_column(Text, nullable=False, default="{}")


class WorkerGPU(UUIDPrimaryKeyMixin, TimestampMixin, Base):
    __tablename__ = "worker_gpus"

    worker_pk: Mapped[UUID] = mapped_column(ForeignKey("workers.id", ondelete="CASCADE"), nullable=False)
    worker_id: Mapped[str] = mapped_column(String(128), nullable=False, index=True)
    gpu_index: Mapped[int] = mapped_column(Integer, nullable=False)
    vendor: Mapped[str] = mapped_column(String(64), nullable=False, default="unknown")
    name: Mapped[str] = mapped_column(String(255), nullable=False)
    memory_total_mb: Mapped[int] = mapped_column(Integer, nullable=False, default=0)
    memory_used_mb: Mapped[int] = mapped_column(Integer, nullable=False, default=0)
    utilization: Mapped[int] = mapped_column(Integer, nullable=False, default=0)


class WorkerHeartbeat(UUIDPrimaryKeyMixin, Base):
    __tablename__ = "worker_heartbeats"

    worker_pk: Mapped[UUID] = mapped_column(ForeignKey("workers.id", ondelete="CASCADE"), nullable=False)
    worker_id: Mapped[str] = mapped_column(String(128), nullable=False, index=True)
    status: Mapped[str] = mapped_column(String(32), nullable=False)
    payload_json: Mapped[str] = mapped_column(Text, nullable=False, default="{}")
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), nullable=False)

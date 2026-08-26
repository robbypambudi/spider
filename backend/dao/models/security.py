from __future__ import annotations

from uuid import UUID

from sqlalchemy import Float, ForeignKey, Integer, String, Text
from sqlalchemy.orm import Mapped, mapped_column

from dao.models.base import Base, TimestampMixin, UUIDPrimaryKeyMixin


class SecurityScan(UUIDPrimaryKeyMixin, TimestampMixin, Base):
    __tablename__ = "security_scans"

    request_id: Mapped[UUID] = mapped_column(nullable=False, index=True)
    user_id: Mapped[UUID | None] = mapped_column(ForeignKey("users.id"), nullable=True)
    decision: Mapped[str] = mapped_column(String(16), nullable=False, index=True)
    score: Mapped[float] = mapped_column(Float, nullable=False)
    threshold: Mapped[float | None] = mapped_column(Float, nullable=True)
    detector: Mapped[str] = mapped_column(String(64), nullable=False)
    detector_version: Mapped[str] = mapped_column(String(64), nullable=False, default="unknown")
    policy: Mapped[str] = mapped_column(String(64), nullable=False)
    chunks_scanned: Mapped[int] = mapped_column(Integer, nullable=False, default=1)
    chunking_strategy: Mapped[str] = mapped_column(String(64), nullable=False, default="fixed")
    latency_ms: Mapped[float] = mapped_column(Float, nullable=False, default=0.0)
    prompt_hash: Mapped[str] = mapped_column(String(64), nullable=False)
    prompt_length: Mapped[int] = mapped_column(Integer, nullable=False)
    prompt_text: Mapped[str | None] = mapped_column(Text, nullable=True)
    model_target: Mapped[str | None] = mapped_column(String(255), nullable=True)
    worker_id: Mapped[str | None] = mapped_column(String(128), nullable=True)
    source: Mapped[str | None] = mapped_column(String(64), nullable=True)
    metadata_json: Mapped[str] = mapped_column(Text, nullable=False, default="{}")


class SecurityChunkResult(UUIDPrimaryKeyMixin, TimestampMixin, Base):
    __tablename__ = "security_chunk_results"

    scan_id: Mapped[UUID] = mapped_column(
        ForeignKey("security_scans.id", ondelete="CASCADE"),
        nullable=False,
        index=True,
    )
    chunk_index: Mapped[int] = mapped_column(Integer, nullable=False)
    detector: Mapped[str] = mapped_column(String(64), nullable=False)
    score: Mapped[float] = mapped_column(Float, nullable=False)
    is_injection: Mapped[bool] = mapped_column(nullable=False, default=False)
    latency_ms: Mapped[float] = mapped_column(Float, nullable=False, default=0.0)


class DetectorExecution(UUIDPrimaryKeyMixin, TimestampMixin, Base):
    __tablename__ = "detector_executions"

    scan_id: Mapped[UUID] = mapped_column(
        ForeignKey("security_scans.id", ondelete="CASCADE"),
        nullable=False,
        index=True,
    )
    detector: Mapped[str] = mapped_column(String(64), nullable=False)
    detector_version: Mapped[str] = mapped_column(String(64), nullable=False, default="unknown")
    threshold: Mapped[float | None] = mapped_column(Float, nullable=True)
    score: Mapped[float] = mapped_column(Float, nullable=False)
    is_injection: Mapped[bool] = mapped_column(nullable=False, default=False)
    latency_ms: Mapped[float] = mapped_column(Float, nullable=False, default=0.0)
    metadata_json: Mapped[str] = mapped_column(Text, nullable=False, default="{}")

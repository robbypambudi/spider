from __future__ import annotations

from uuid import UUID

from sqlalchemy import Float, ForeignKey, String, Text
from sqlalchemy.orm import Mapped, mapped_column

from dao.models.base import Base, TimestampMixin, UUIDPrimaryKeyMixin


class InferenceRequestRecord(UUIDPrimaryKeyMixin, TimestampMixin, Base):
    __tablename__ = "inference_requests"

    request_id: Mapped[UUID] = mapped_column(nullable=False, unique=True, index=True)
    user_id: Mapped[UUID | None] = mapped_column(ForeignKey("users.id"), nullable=True)
    scan_id: Mapped[UUID | None] = mapped_column(ForeignKey("security_scans.id"), nullable=True)
    model: Mapped[str] = mapped_column(String(255), nullable=False)
    status: Mapped[str] = mapped_column(String(32), nullable=False, index=True)
    decision: Mapped[str] = mapped_column(String(16), nullable=False)
    worker_id: Mapped[str | None] = mapped_column(String(128), nullable=True)
    end_to_end_latency_ms: Mapped[float] = mapped_column(Float, nullable=False, default=0.0)
    security_overhead_ms: Mapped[float] = mapped_column(Float, nullable=False, default=0.0)
    inference_latency_ms: Mapped[float | None] = mapped_column(Float, nullable=True)
    output_preview: Mapped[str | None] = mapped_column(String(512), nullable=True)
    metadata_json: Mapped[str] = mapped_column(Text, nullable=False, default="{}")


class InferenceEvent(UUIDPrimaryKeyMixin, TimestampMixin, Base):
    __tablename__ = "inference_events"

    inference_id: Mapped[UUID] = mapped_column(
        ForeignKey("inference_requests.id", ondelete="CASCADE"),
        nullable=False,
        index=True,
    )
    event_type: Mapped[str] = mapped_column(String(64), nullable=False)
    payload_json: Mapped[str] = mapped_column(Text, nullable=False, default="{}")

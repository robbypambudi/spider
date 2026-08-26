from __future__ import annotations

from uuid import UUID

from sqlalchemy import ForeignKey, String
from sqlalchemy.orm import Mapped, mapped_column

from dao.models.base import Base, TimestampMixin, UUIDPrimaryKeyMixin


class ServingModel(UUIDPrimaryKeyMixin, TimestampMixin, Base):
    __tablename__ = "serving_models"

    worker_id: Mapped[str] = mapped_column(String(128), nullable=False, index=True)
    name: Mapped[str] = mapped_column(String(255), nullable=False)
    status: Mapped[str] = mapped_column(String(32), nullable=False, default="READY")


class ServingNode(UUIDPrimaryKeyMixin, TimestampMixin, Base):
    __tablename__ = "serving_nodes"

    worker_pk: Mapped[UUID] = mapped_column(ForeignKey("workers.id", ondelete="CASCADE"), nullable=False)
    worker_id: Mapped[str] = mapped_column(String(128), unique=True, nullable=False)
    status: Mapped[str] = mapped_column(String(32), nullable=False, default="ONLINE")

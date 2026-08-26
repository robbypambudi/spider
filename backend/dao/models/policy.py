from __future__ import annotations

from sqlalchemy import Float, String, Text
from sqlalchemy.orm import Mapped, mapped_column

from dao.models.base import Base, TimestampMixin, UUIDPrimaryKeyMixin


class Policy(UUIDPrimaryKeyMixin, TimestampMixin, Base):
    __tablename__ = "policies"

    name: Mapped[str] = mapped_column(String(128), unique=True, nullable=False)
    kind: Mapped[str] = mapped_column(String(64), nullable=False, default="threshold")
    threshold: Mapped[float] = mapped_column(Float, nullable=False, default=0.5)
    action_on_detection: Mapped[str] = mapped_column(String(32), nullable=False, default="block")
    config_json: Mapped[str] = mapped_column(Text, nullable=False, default="{}")
    is_default: Mapped[bool] = mapped_column(default=False, nullable=False)

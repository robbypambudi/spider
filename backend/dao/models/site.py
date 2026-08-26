from __future__ import annotations

from sqlalchemy import String, Text
from sqlalchemy.orm import Mapped, mapped_column

from dao.models.base import Base, TimestampMixin, UUIDPrimaryKeyMixin


class Site(UUIDPrimaryKeyMixin, TimestampMixin, Base):
    __tablename__ = "sites"

    name: Mapped[str] = mapped_column(String(128), unique=True, nullable=False)
    description: Mapped[str | None] = mapped_column(Text, nullable=True)

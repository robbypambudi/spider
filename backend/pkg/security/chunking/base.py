from __future__ import annotations

from typing import Any, Protocol

from pydantic import BaseModel, Field


class Chunk(BaseModel):
    index: int
    text: str
    start: int | None = None
    end: int | None = None
    metadata: dict[str, Any] = Field(default_factory=dict)


class Chunker(Protocol):
    name: str

    def chunk(self, text: str) -> list[Chunk]: ...

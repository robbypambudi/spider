from __future__ import annotations

from pkg.security.chunking.base import Chunk


class FixedSizeChunker:
    """Character-window chunker with optional overlap. Default research chunker."""

    name = "fixed"

    def __init__(self, size: int = 2048, overlap: int = 128) -> None:
        if size <= 0:
            raise ValueError("chunk size must be positive")
        if overlap < 0 or overlap >= size:
            raise ValueError("overlap must be >= 0 and < size")
        self.size = size
        self.overlap = overlap

    def chunk(self, text: str) -> list[Chunk]:
        if not text:
            return [Chunk(index=0, text="", start=0, end=0, metadata={"strategy": self.name})]

        chunks: list[Chunk] = []
        start = 0
        index = 0
        step = self.size - self.overlap
        length = len(text)
        while start < length:
            end = min(start + self.size, length)
            chunks.append(
                Chunk(
                    index=index,
                    text=text[start:end],
                    start=start,
                    end=end,
                    metadata={"strategy": self.name, "size": self.size, "overlap": self.overlap},
                )
            )
            index += 1
            if end == length:
                break
            start += step
        return chunks

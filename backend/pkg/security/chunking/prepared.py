from __future__ import annotations

from pkg.security.chunking.base import Chunk


class TokenChunker:
    """Placeholder. Implement with a tokenizer when a production detector is wired."""

    name = "token"

    def chunk(self, text: str) -> list[Chunk]:
        raise NotImplementedError(
            "TokenChunker requires a tokenizer binding. See docs/chunking.md."
        )


class SlidingWindowChunker:
    """Placeholder sliding-window chunker for future detector evaluations."""

    name = "sliding-window"

    def chunk(self, text: str) -> list[Chunk]:
        raise NotImplementedError(
            "SlidingWindowChunker is reserved for evaluation experiments. See docs/chunking.md."
        )


class SemanticChunker:
    """Placeholder semantic chunker. Do not fake embeddings."""

    name = "semantic"

    def chunk(self, text: str) -> list[Chunk]:
        raise NotImplementedError(
            "SemanticChunker requires an embedding model interface. See docs/chunking.md."
        )

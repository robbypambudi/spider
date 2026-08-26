from __future__ import annotations

from pkg.security.chunking.fixed import FixedSizeChunker


def test_fixed_size_chunker_single_chunk() -> None:
    chunker = FixedSizeChunker(size=64, overlap=8)
    chunks = chunker.chunk("hello world")
    assert len(chunks) == 1
    assert chunks[0].text == "hello world"
    assert chunks[0].index == 0


def test_fixed_size_chunker_splits_long_text() -> None:
    chunker = FixedSizeChunker(size=8, overlap=2)
    chunks = chunker.chunk("abcdefghijklmnop")
    assert len(chunks) > 1
    assert chunks[0].start == 0
    assert chunks[0].metadata["strategy"] == "fixed"
    reconstructed_span = chunks[-1].end
    assert reconstructed_span == 16

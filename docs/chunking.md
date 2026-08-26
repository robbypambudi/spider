# Chunking

Long prompts are split before detection so an injection buried in a document is still visible.

Implemented: `FixedSizeChunker` (`SPIDER_CHUNK_SIZE`, `SPIDER_CHUNK_OVERLAP`).

Prepared: `TokenChunker`, `SlidingWindowChunker`, `SemanticChunker`.

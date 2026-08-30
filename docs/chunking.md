# Chunking

Long prompts are split before detection so an injection buried in a document is still visible.

Implemented:

- `FixedSizeChunker` — character windows (`SPIDER_CHUNK_SIZE`, `SPIDER_CHUNK_OVERLAP`, `SPIDER_CHUNKER=fixed`)
- `SidecarTokenChunker` — token windows via prompt-shield `/chunk` (`SPIDER_CHUNKER=token`, default size 256, overlap 0, aligned with lab wave-1)

Prepared: `SlidingWindowChunker`, `SemanticChunker`.

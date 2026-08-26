from __future__ import annotations

import asyncio
from collections.abc import Awaitable, Callable


class InMemoryQueue:
    """Simple asyncio queue for local development. Not a distributed job queue."""

    def __init__(self, maxsize: int = 0) -> None:
        self._queue: asyncio.Queue[object] = asyncio.Queue(maxsize=maxsize)

    async def put(self, item: object) -> None:
        await self._queue.put(item)

    async def get(self) -> object:
        return await self._queue.get()

    def qsize(self) -> int:
        return self._queue.qsize()

    async def consume(self, handler: Callable[[object], Awaitable[None]]) -> None:
        item = await self.get()
        await handler(item)

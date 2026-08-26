from __future__ import annotations

from collections.abc import Awaitable, Callable
from datetime import UTC, datetime

from pkg.apis.worker.models import WorkerResource


class WorkerReconciler:
    """Mark workers OFFLINE when heartbeats exceed the configured timeout."""

    name = "worker"

    def __init__(
        self,
        *,
        list_workers: Callable[[], Awaitable[list[WorkerResource]]],
        mark_offline: Callable[[str], Awaitable[None]],
        last_heartbeat: Callable[[str], Awaitable[datetime | None]],
        offline_timeout_seconds: int,
        now: Callable[[], datetime] | None = None,
    ) -> None:
        self._list_workers = list_workers
        self._mark_offline = mark_offline
        self._last_heartbeat = last_heartbeat
        self._timeout = offline_timeout_seconds
        self._now = now or (lambda: datetime.now(UTC))

    async def reconcile(self) -> list[str]:
        marked: list[str] = []
        now = self._now()
        workers = await self._list_workers()
        for worker in workers:
            if worker.status in {"OFFLINE", "ERROR"}:
                continue
            heartbeat = await self._last_heartbeat(worker.worker_id)
            if heartbeat is None:
                await self._mark_offline(worker.worker_id)
                marked.append(worker.worker_id)
                continue
            if heartbeat.tzinfo is None:
                heartbeat = heartbeat.replace(tzinfo=UTC)
            age = (now - heartbeat).total_seconds()
            if age > self._timeout:
                await self._mark_offline(worker.worker_id)
                marked.append(worker.worker_id)
        return marked

from __future__ import annotations

from dao.repositories.worker import WorkerRepository
from pkg.apis.worker.models import WorkerHeartbeat, WorkerResource
from pkg.config.settings import Settings
from pkg.runtime.status import WorkerStatus

from internal.exceptions import NotFoundError, WorkerAuthError


class WorkerService:
    def __init__(self, workers: WorkerRepository, settings: Settings) -> None:
        self.workers = workers
        self.settings = settings

    def authenticate_token(self, token: str | None) -> None:
        if token != self.settings.worker_token:
            raise WorkerAuthError()

    async def register(self, resource: WorkerResource) -> WorkerResource:
        resource.status = WorkerStatus.ONLINE.value
        worker = await self.workers.upsert_registration(resource)
        return await self.workers.as_resource(worker)

    async def heartbeat(self, payload: WorkerHeartbeat) -> WorkerResource:
        worker = await self.workers.record_heartbeat(
            payload.worker_id,
            status=payload.status,
            resources=payload.resources,
            models=payload.models,
            running_requests=payload.running_requests,
            metadata=payload.metadata,
        )
        if worker is None:
            raise NotFoundError(f"Worker {payload.worker_id} is not registered")
        return await self.workers.as_resource(worker)

    async def list_workers(self) -> list[WorkerResource]:
        return await self.workers.list_resources()

    async def inspect(self, worker_id: str) -> WorkerResource:
        worker = await self.workers.get_by_worker_id(worker_id)
        if worker is None:
            raise NotFoundError(f"Worker {worker_id} not found")
        return await self.workers.as_resource(worker)

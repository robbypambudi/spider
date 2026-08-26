from __future__ import annotations

import json
from datetime import UTC, datetime

from pkg.apis.worker.models import (
    GPUResource,
    LoadedModel,
    WorkerResource,
    WorkerResources,
)
from pkg.runtime.status import WorkerStatus
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from dao.models.serving import ServingModel, ServingNode
from dao.models.worker import Worker, WorkerGPU, WorkerHeartbeat


class WorkerRepository:
    def __init__(self, session: AsyncSession) -> None:
        self.session = session

    async def upsert_registration(self, resource: WorkerResource) -> Worker:
        worker = await self.get_by_worker_id(resource.worker_id)
        now = datetime.now(UTC)
        if worker is None:
            worker = Worker(
                worker_id=resource.worker_id,
                hostname=resource.hostname,
                site=resource.site,
                version=resource.version,
                status=WorkerStatus.ONLINE.value,
                cpu_total=resource.resources.cpu_total,
                memory_total_mb=resource.resources.memory_total_mb,
                running_requests=resource.running_requests,
                last_heartbeat_at=now,
                models_json=json.dumps([m.model_dump() for m in resource.models]),
                metadata_json=json.dumps(resource.metadata),
            )
            self.session.add(worker)
            await self.session.flush()
            node = ServingNode(
                worker_pk=worker.id,
                worker_id=worker.worker_id,
                status=WorkerStatus.ONLINE.value,
            )
            self.session.add(node)
        else:
            worker.hostname = resource.hostname
            worker.site = resource.site
            worker.version = resource.version
            worker.status = WorkerStatus.ONLINE.value
            worker.cpu_total = resource.resources.cpu_total
            worker.memory_total_mb = resource.resources.memory_total_mb
            worker.running_requests = resource.running_requests
            worker.last_heartbeat_at = now
            worker.models_json = json.dumps([m.model_dump() for m in resource.models])
            worker.metadata_json = json.dumps(resource.metadata)

        await self._replace_gpus(worker, resource.resources.gpus)
        await self._replace_models(resource.worker_id, resource.models)
        await self.session.flush()
        return worker

    async def record_heartbeat(
        self,
        worker_id: str,
        *,
        status: str,
        resources: WorkerResources | None,
        models: list[LoadedModel],
        running_requests: int,
        metadata: dict[str, object],
    ) -> Worker | None:
        worker = await self.get_by_worker_id(worker_id)
        if worker is None:
            return None
        now = datetime.now(UTC)
        worker.status = status
        worker.running_requests = running_requests
        worker.last_heartbeat_at = now
        if resources is not None:
            worker.cpu_total = resources.cpu_total
            worker.memory_total_mb = resources.memory_total_mb
            await self._replace_gpus(worker, resources.gpus)
        worker.models_json = json.dumps([m.model_dump() for m in models])
        await self._replace_models(worker_id, models)
        self.session.add(
            WorkerHeartbeat(
                worker_pk=worker.id,
                worker_id=worker_id,
                status=status,
                payload_json=json.dumps(metadata, default=str),
                created_at=now,
            )
        )
        node = await self.get_serving_node(worker_id)
        if node is not None:
            node.status = status
        await self.session.flush()
        return worker

    async def get_by_worker_id(self, worker_id: str) -> Worker | None:
        result = await self.session.execute(select(Worker).where(Worker.worker_id == worker_id))
        return result.scalar_one_or_none()

    async def list_workers(self) -> list[Worker]:
        result = await self.session.execute(select(Worker).order_by(Worker.worker_id))
        return list(result.scalars().all())

    async def mark_offline(self, worker_id: str) -> None:
        worker = await self.get_by_worker_id(worker_id)
        if worker is None:
            return
        worker.status = WorkerStatus.OFFLINE.value
        node = await self.get_serving_node(worker_id)
        if node is not None:
            node.status = WorkerStatus.OFFLINE.value
        await self.session.flush()

    async def last_heartbeat_at(self, worker_id: str) -> datetime | None:
        worker = await self.get_by_worker_id(worker_id)
        return worker.last_heartbeat_at if worker else None

    async def get_serving_node(self, worker_id: str) -> ServingNode | None:
        result = await self.session.execute(
            select(ServingNode).where(ServingNode.worker_id == worker_id)
        )
        return result.scalar_one_or_none()

    async def list_serving_nodes(self) -> list[ServingNode]:
        result = await self.session.execute(select(ServingNode))
        return list(result.scalars().all())

    async def list_models(self) -> list[ServingModel]:
        result = await self.session.execute(select(ServingModel))
        return list(result.scalars().all())

    async def gpus_for(self, worker_id: str) -> list[WorkerGPU]:
        result = await self.session.execute(
            select(WorkerGPU).where(WorkerGPU.worker_id == worker_id).order_by(WorkerGPU.gpu_index)
        )
        return list(result.scalars().all())

    async def as_resource(self, worker: Worker) -> WorkerResource:
        gpus = await self.gpus_for(worker.worker_id)
        models_raw = json.loads(worker.models_json or "[]")
        return WorkerResource(
            worker_id=worker.worker_id,
            hostname=worker.hostname,
            site=worker.site,
            version=worker.version,
            status=worker.status,
            resources=WorkerResources(
                cpu_total=worker.cpu_total,
                memory_total_mb=worker.memory_total_mb,
                gpus=[
                    GPUResource(
                        index=gpu.gpu_index,
                        vendor=gpu.vendor,
                        name=gpu.name,
                        memory_total_mb=gpu.memory_total_mb,
                        memory_used_mb=gpu.memory_used_mb,
                        utilization=gpu.utilization,
                    )
                    for gpu in gpus
                ],
            ),
            models=[LoadedModel.model_validate(item) for item in models_raw],
            running_requests=worker.running_requests,
        )

    async def list_resources(self) -> list[WorkerResource]:
        workers = await self.list_workers()
        return [await self.as_resource(worker) for worker in workers]

    async def _replace_gpus(self, worker: Worker, gpus: list[GPUResource]) -> None:
        existing = await self.gpus_for(worker.worker_id)
        for row in existing:
            await self.session.delete(row)
        await self.session.flush()
        for gpu in gpus:
            self.session.add(
                WorkerGPU(
                    worker_pk=worker.id,
                    worker_id=worker.worker_id,
                    gpu_index=gpu.index,
                    vendor=gpu.vendor,
                    name=gpu.name,
                    memory_total_mb=gpu.memory_total_mb,
                    memory_used_mb=gpu.memory_used_mb,
                    utilization=gpu.utilization,
                )
            )

    async def _replace_models(self, worker_id: str, models: list[LoadedModel]) -> None:
        result = await self.session.execute(
            select(ServingModel).where(ServingModel.worker_id == worker_id)
        )
        for row in result.scalars().all():
            await self.session.delete(row)
        await self.session.flush()
        for model in models:
            self.session.add(
                ServingModel(worker_id=worker_id, name=model.name, status=model.status)
            )

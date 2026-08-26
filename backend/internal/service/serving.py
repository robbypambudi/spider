from __future__ import annotations

from dao.repositories.worker import WorkerRepository


class ServingService:
    def __init__(self, workers: WorkerRepository) -> None:
        self.workers = workers

    async def list_nodes(self) -> list[dict[str, object]]:
        nodes = await self.workers.list_serving_nodes()
        return [
            {"worker_id": node.worker_id, "status": node.status, "id": str(node.id)} for node in nodes
        ]

    async def list_models(self) -> list[dict[str, object]]:
        models = await self.workers.list_models()
        return [
            {"worker_id": model.worker_id, "name": model.name, "status": model.status}
            for model in models
        ]

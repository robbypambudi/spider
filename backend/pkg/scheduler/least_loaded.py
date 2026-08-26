from __future__ import annotations

from pkg.apis.serving.models import ServingRequest
from pkg.apis.worker.models import WorkerResource

_SERVING_STATUSES = {"ONLINE", "BUSY"}


class LeastLoadedScheduler:
    """Pick a serving-capable worker with the lowest running request count.

    Considers status, GPU availability, VRAM pressure, and model locality when
    reported. CPU-only workers remain eligible when no GPU constraint is set.
    """

    name = "least-loaded"

    async def select_worker(
        self,
        workload: ServingRequest,
        workers: list[WorkerResource],
    ) -> WorkerResource | None:
        eligible = [worker for worker in workers if self._is_eligible(worker, workload)]
        if not eligible:
            return None

        def sort_key(worker: WorkerResource) -> tuple[int, int, float, int]:
            locality_penalty = 0
            if workload.model and worker.models:
                ready = {item.name for item in worker.models if item.status == "READY"}
                if workload.model not in ready:
                    locality_penalty = 1
            gpu_util = 0.0
            vram_used = 0
            if worker.resources.gpus:
                gpu_util = sum(gpu.utilization for gpu in worker.resources.gpus) / len(
                    worker.resources.gpus
                )
                vram_used = sum(gpu.memory_used_mb for gpu in worker.resources.gpus)
            return (locality_penalty, worker.running_requests, gpu_util, vram_used)

        return min(eligible, key=sort_key)

    def _is_eligible(self, worker: WorkerResource, workload: ServingRequest) -> bool:
        if worker.status not in _SERVING_STATUSES:
            return False
        if workload.site and worker.site and workload.site != worker.site:
            return False
        return True

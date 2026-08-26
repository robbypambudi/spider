from __future__ import annotations

from dao.models.security import SecurityScan
from dao.models.worker import Worker, WorkerGPU
from dao.repositories.security import SecurityScanRepository
from pkg.runtime.status import WorkerStatus
from sqlalchemy import func, select
from sqlalchemy.ext.asyncio import AsyncSession


def _percentile(values: list[float], percentile: float) -> float:
    if not values:
        return 0.0
    ordered = sorted(values)
    index = min(len(ordered) - 1, max(0, int(round((percentile / 100.0) * (len(ordered) - 1)))))
    return float(ordered[index])


class MetricsService:
    def __init__(self, session: AsyncSession, scans: SecurityScanRepository) -> None:
        self.session = session
        self.scans = scans

    async def dashboard_summary(self) -> dict[str, object]:
        counts = await self.scans.summary_counts()
        total = counts["total_scans"]
        blocked = counts["blocked"]
        detection_rate = (blocked / total) if total else 0.0

        latencies_result = await self.session.execute(select(SecurityScan.latency_ms))
        latencies = [float(value) for value in latencies_result.scalars().all()]

        workers_total = await self.session.scalar(select(func.count(Worker.id))) or 0
        workers_online = await self.session.scalar(
            select(func.count(Worker.id)).where(Worker.status == WorkerStatus.ONLINE.value)
        ) or 0
        gpus_total = await self.session.scalar(select(func.count(WorkerGPU.id))) or 0

        overhead_result = await self.session.execute(
            select(func.avg(SecurityScan.latency_ms))
        )
        avg_overhead = float(overhead_result.scalar() or 0.0)

        return {
            "total_scans": total,
            "allowed": counts["allowed"],
            "blocked": blocked,
            "review": counts["review"],
            "detection_rate": detection_rate,
            "avg_detection_latency_ms": counts["avg_detection_latency_ms"],
            "p95_detection_latency_ms": _percentile(latencies, 95),
            "avg_security_overhead_ms": avg_overhead,
            "active_serving_nodes": int(workers_online),
            "total_gpus": int(gpus_total),
            "workers_total": int(workers_total),
        }

from __future__ import annotations

import asyncio

from dao.database import Database
from dao.repositories.worker import WorkerRepository
from internal.telemetry.logging import configure_logging, get_logger
from pkg.config.settings import get_settings
from pkg.reconciler.prepared import ServingNodeReconciler
from pkg.reconciler.worker import WorkerReconciler

logger = get_logger("spider-controller")


async def run_controller() -> None:
    settings = get_settings()
    configure_logging(log_prompt_content=settings.log_prompt_content)
    database = Database(settings)
    await database.create_schema()
    serving_reconciler = ServingNodeReconciler()
    logger.info(
        "controller_started",
        heartbeat_timeout=settings.worker_offline_timeout,
    )
    try:
        while True:
            async with database.session_factory() as session:
                repo = WorkerRepository(session)
                reconciler = WorkerReconciler(
                    list_workers=repo.list_resources,
                    mark_offline=repo.mark_offline,
                    last_heartbeat=repo.last_heartbeat_at,
                    offline_timeout_seconds=settings.worker_offline_timeout,
                )
                marked = await reconciler.reconcile()
                await serving_reconciler.reconcile()
                await session.commit()
                if marked:
                    logger.info("workers_marked_offline", worker_ids=marked)
            await asyncio.sleep(5)
    finally:
        await database.dispose()


def main() -> None:
    asyncio.run(run_controller())


if __name__ == "__main__":
    main()

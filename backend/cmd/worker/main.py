from __future__ import annotations

import asyncio

import httpx
import typer
from internal.telemetry.logging import configure_logging, get_logger
from internal.version import __version__
from pkg.apis.worker.models import LoadedModel, WorkerHeartbeat, WorkerResource
from pkg.config.settings import get_settings
from pkg.protocol.worker import HEARTBEAT_PATH, REGISTER_PATH, WORKER_TOKEN_HEADER
from pkg.workerctl.discovery import cuda_driver_info, discover_resources
from pkg.workerctl.identity import load_identity, platform_summary

cli_app = typer.Typer(help="SPIDER serving worker agent.")
logger = get_logger("spider-worker")


async def _run_worker() -> None:
    settings = get_settings()
    configure_logging(log_prompt_content=settings.log_prompt_content)
    identity = load_identity(version=__version__)
    resources = discover_resources()
    models = [LoadedModel(name=settings.default_model, status="READY")]
    payload = WorkerResource(
        worker_id=identity.worker_id,
        hostname=identity.hostname,
        site=identity.site,
        version=identity.version,
        status="ONLINE",
        resources=resources,
        models=models,
        metadata={**platform_summary(), **cuda_driver_info()},
    )
    headers = {WORKER_TOKEN_HEADER: settings.worker_token}
    base = settings.api_base_url.rstrip("/")
    async with httpx.AsyncClient(timeout=10.0) as client:
        response = await client.post(f"{base}{REGISTER_PATH}", json=payload.model_dump(), headers=headers)
        response.raise_for_status()
        logger.info("worker_registered", worker_id=identity.worker_id, gpus=len(resources.gpus))
        while True:
            current = discover_resources()
            heartbeat = WorkerHeartbeat(
                worker_id=identity.worker_id,
                status="ONLINE",
                resources=current,
                models=models,
                running_requests=0,
                metadata=platform_summary(),
            )
            path = HEARTBEAT_PATH.format(worker_id=identity.worker_id)
            beat = await client.post(f"{base}{path}", json=heartbeat.model_dump(), headers=headers)
            beat.raise_for_status()
            logger.info("worker_heartbeat", worker_id=identity.worker_id)
            await asyncio.sleep(settings.worker_heartbeat_interval)


@cli_app.command("start")
def start() -> None:
    """Start the worker: identity → discovery → register → heartbeat."""
    asyncio.run(_run_worker())


def main() -> None:
    cli_app()


if __name__ == "__main__":
    main()

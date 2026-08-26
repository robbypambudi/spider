from __future__ import annotations

from collections.abc import AsyncGenerator
from contextlib import asynccontextmanager

from fastapi import FastAPI, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse, PlainTextResponse, Response
from internal.app import AppContainer, bootstrap_control_plane, build_container, database_ready
from internal.exceptions import SpiderError
from internal.handler import (
    auth_router,
    inference_router,
    jobs_router,
    metrics_router,
    security_router,
    serving_router,
    workers_router,
)
from internal.middleware.logging import RequestContextMiddleware
from internal.telemetry.logging import configure_logging, get_logger
from internal.version import SERVICE_NAME, __version__
from pkg.config.settings import Settings, get_settings
from pkg.monitor.health import HealthStatus, ReadinessStatus
from pkg.serving.providers.base import LLMProvider

logger = get_logger("spider-api")


def create_app(
    settings: Settings | None = None,
    *,
    llm_provider: LLMProvider | None = None,
    container: AppContainer | None = None,
) -> FastAPI:
    settings = settings or get_settings()
    configure_logging(log_prompt_content=settings.log_prompt_content)
    app_container = container or build_container(settings, llm_provider=llm_provider)

    @asynccontextmanager
    async def lifespan(_app: FastAPI) -> AsyncGenerator[None, None]:
        await bootstrap_control_plane(app_container)
        logger.info("api_started", service=SERVICE_NAME, version=__version__, env=settings.env)
        yield
        await app_container.database.dispose()

    app = FastAPI(
        title="SPIDER",
        description=(
            "Runtime defense framework for prompt injection detection in LLM-serving "
            "cluster environments."
        ),
        version=__version__,
        lifespan=lifespan,
    )
    app.state.container = app_container

    app.add_middleware(RequestContextMiddleware)
    app.add_middleware(
        CORSMiddleware,
        allow_origins=settings.cors_origin_list,
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )

    @app.exception_handler(SpiderError)
    async def spider_error_handler(_request: Request, exc: SpiderError) -> JSONResponse:
        return JSONResponse(
            status_code=exc.status_code,
            content={"error": exc.code, "message": exc.message},
        )

    app.include_router(auth_router, prefix="/api/v1")
    app.include_router(security_router, prefix="/api/v1")
    app.include_router(inference_router, prefix="/api/v1")
    app.include_router(workers_router, prefix="/api/v1")
    app.include_router(serving_router, prefix="/api/v1")
    app.include_router(metrics_router, prefix="/api/v1")
    app.include_router(jobs_router, prefix="/api/v1")

    @app.get("/health", response_model=HealthStatus, tags=["health"])
    async def health() -> HealthStatus:
        return HealthStatus(status="ok", service=SERVICE_NAME, version=__version__)

    @app.get("/ready", response_model=ReadinessStatus, tags=["health"])
    async def ready() -> ReadinessStatus:
        db_ok = await database_ready(app_container)
        status = "ok" if db_ok else "degraded"
        return ReadinessStatus(status=status, database=db_ok, redis=None)

    @app.get("/metrics")
    async def prometheus_metrics() -> Response:
        payload, content_type = app_container.metrics.render()
        return PlainTextResponse(payload.decode("utf-8"), media_type=content_type)

    return app


app = create_app()


def main() -> None:
    import uvicorn

    settings = get_settings()
    uvicorn.run(
        "cmd.api.main:app",
        host=settings.api_host,
        port=settings.api_port,
        reload=settings.is_development,
    )


if __name__ == "__main__":
    main()

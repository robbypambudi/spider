from internal.handler.auth.routes import router as auth_router
from internal.handler.inference.routes import router as inference_router
from internal.handler.jobs.routes import router as jobs_router
from internal.handler.metrics.routes import router as metrics_router
from internal.handler.security.routes import router as security_router
from internal.handler.serving.routes import router as serving_router
from internal.handler.workers.routes import router as workers_router

__all__ = [
    "auth_router",
    "inference_router",
    "jobs_router",
    "metrics_router",
    "security_router",
    "serving_router",
    "workers_router",
]

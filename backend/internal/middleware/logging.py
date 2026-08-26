from __future__ import annotations

from collections.abc import Awaitable, Callable

from fastapi import Request, Response
from starlette.middleware.base import BaseHTTPMiddleware

from internal.telemetry.logging import bind_request_context, clear_request_context, get_logger

logger = get_logger("spider-http")


class RequestContextMiddleware(BaseHTTPMiddleware):
    async def dispatch(
        self,
        request: Request,
        call_next: Callable[[Request], Awaitable[Response]],
    ) -> Response:
        request_id = request.headers.get("x-request-id")
        bind_request_context(service="spider-api", path=str(request.url.path), request_id=request_id)
        try:
            response = await call_next(request)
        except Exception:
            logger.exception("unhandled_request_error", method=request.method, path=str(request.url.path))
            raise
        finally:
            clear_request_context()
        return response

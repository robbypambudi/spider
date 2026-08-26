from __future__ import annotations

from typing import Any

import structlog

_configured = False


def configure_logging(*, log_prompt_content: bool = False) -> None:
    global _configured
    if _configured:
        return

    structlog.configure(
        processors=[
            structlog.contextvars.merge_contextvars,
            structlog.processors.add_log_level,
            structlog.processors.TimeStamper(fmt="iso", utc=True, key="timestamp"),
            structlog.processors.StackInfoRenderer(),
            structlog.processors.format_exc_info,
            structlog.processors.JSONRenderer(),
        ],
        wrapper_class=structlog.make_filtering_bound_logger(0),
        context_class=dict,
        logger_factory=structlog.PrintLoggerFactory(),
        cache_logger_on_first_use=True,
    )
    structlog.contextvars.bind_contextvars(log_prompt_content=log_prompt_content)
    _configured = True


def get_logger(service: str = "spider-api") -> structlog.stdlib.BoundLogger:
    return structlog.get_logger(service=service)


def bind_request_context(**kwargs: Any) -> None:
    """Bind structured fields. Never pass raw prompt text here unless enabled."""
    structlog.contextvars.bind_contextvars(**kwargs)


def clear_request_context() -> None:
    structlog.contextvars.clear_contextvars()

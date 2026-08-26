from __future__ import annotations


class SpiderError(Exception):
    """Base SPIDER error."""

    def __init__(self, message: str, *, code: str = "spider_error", status_code: int = 500) -> None:
        super().__init__(message)
        self.message = message
        self.code = code
        self.status_code = status_code


class AuthenticationError(SpiderError):
    def __init__(self, message: str = "Authentication required") -> None:
        super().__init__(message, code="unauthenticated", status_code=401)


class AuthorizationError(SpiderError):
    def __init__(self, message: str = "Insufficient permissions") -> None:
        super().__init__(message, code="forbidden", status_code=403)


class NotFoundError(SpiderError):
    def __init__(self, message: str = "Resource not found") -> None:
        super().__init__(message, code="not_found", status_code=404)


class ValidationFailedError(SpiderError):
    def __init__(self, message: str) -> None:
        super().__init__(message, code="validation_error", status_code=422)


class SecurityPipelineError(SpiderError):
    def __init__(self, message: str) -> None:
        super().__init__(message, code="security_pipeline_error", status_code=500)


class ServingError(SpiderError):
    def __init__(self, message: str) -> None:
        super().__init__(message, code="serving_error", status_code=502)


class WorkerAuthError(SpiderError):
    def __init__(self, message: str = "Invalid worker token") -> None:
        super().__init__(message, code="worker_unauthorized", status_code=401)

"""Policy catalog lives under GET /api/v1/security/policies."""

from internal.handler.security.routes import router

__all__ = ["router"]

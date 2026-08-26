from __future__ import annotations

from collections.abc import Callable

from dao.models.user import User

from internal.exceptions import AuthorizationError

RoleChecker = Callable[[User], None]


def require_roles(*roles: str) -> Callable[[User], None]:
    allowed = set(roles)

    def _check(user: User) -> None:
        if user.role not in allowed:
            raise AuthorizationError(f"Role {user.role} cannot perform this action")

    return _check

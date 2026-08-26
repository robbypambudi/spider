from __future__ import annotations

import hashlib
import hmac
import secrets
from datetime import UTC, datetime, timedelta
from uuid import UUID

import jwt
from dao.models.user import User
from dao.repositories.user import UserRepository
from pkg.config.settings import Settings

from internal.exceptions import AuthenticationError, AuthorizationError

ROLES = ("ADMIN", "RESEARCHER", "STUDENT", "SERVICE")


def hash_password(password: str) -> str:
    salt = secrets.token_hex(16)
    derived = hashlib.pbkdf2_hmac("sha256", password.encode("utf-8"), salt.encode("utf-8"), 120_000)
    return f"pbkdf2_sha256$120000${salt}${derived.hex()}"


def verify_password(password: str, stored: str) -> bool:
    try:
        scheme, iterations, salt, digest = stored.split("$")
    except ValueError:
        return False
    if scheme != "pbkdf2_sha256":
        return False
    derived = hashlib.pbkdf2_hmac(
        "sha256",
        password.encode("utf-8"),
        salt.encode("utf-8"),
        int(iterations),
    )
    return hmac.compare_digest(derived.hex(), digest)


class AuthService:
    def __init__(self, users: UserRepository, settings: Settings) -> None:
        self.users = users
        self.settings = settings

    async def authenticate(self, email: str, password: str) -> User:
        user = await self.users.get_by_email(email)
        if user is None or not user.is_active or not verify_password(password, user.hashed_password):
            raise AuthenticationError("Invalid email or password")
        return user

    def issue_token(self, user: User) -> str:
        now = datetime.now(UTC)
        payload = {
            "sub": str(user.id),
            "email": user.email,
            "role": user.role,
            "iat": int(now.timestamp()),
            "exp": int((now + timedelta(minutes=self.settings.jwt_expire_minutes)).timestamp()),
        }
        return jwt.encode(payload, self.settings.jwt_secret, algorithm=self.settings.jwt_algorithm)

    def decode_token(self, token: str) -> dict[str, str]:
        try:
            payload = jwt.decode(
                token,
                self.settings.jwt_secret,
                algorithms=[self.settings.jwt_algorithm],
            )
        except jwt.PyJWTError as exc:
            raise AuthenticationError("Invalid or expired token") from exc
        return {
            "sub": str(payload.get("sub", "")),
            "email": str(payload.get("email", "")),
            "role": str(payload.get("role", "")),
        }

    async def user_from_token(self, token: str) -> User:
        claims = self.decode_token(token)
        try:
            user_id = UUID(claims["sub"])
        except ValueError as exc:
            raise AuthenticationError("Invalid token subject") from exc
        user = await self.users.get_by_id(user_id)
        if user is None or not user.is_active:
            raise AuthenticationError("User not found")
        return user

    def require_roles(self, user: User, allowed: tuple[str, ...]) -> None:
        if user.role not in allowed:
            raise AuthorizationError(f"Role {user.role} is not permitted")

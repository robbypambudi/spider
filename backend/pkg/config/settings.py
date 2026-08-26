from __future__ import annotations

from functools import lru_cache
from typing import Literal

from pydantic import Field, field_validator
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """Runtime configuration. Thresholds and fail-mode live here, not in detectors."""

    model_config = SettingsConfigDict(
        env_file=(".env", "../.env"),
        env_file_encoding="utf-8",
        extra="ignore",
        populate_by_name=True,
    )

    env: str = Field(default="development", validation_alias="SPIDER_ENV")

    database_url: str = Field(
        default="postgresql+asyncpg://spider:spider@localhost:5432/spider",
        validation_alias="DATABASE_URL",
    )
    redis_url: str = Field(default="redis://localhost:6379/0", validation_alias="REDIS_URL")

    api_host: str = Field(default="0.0.0.0", validation_alias="SPIDER_API_HOST")
    api_port: int = Field(default=8000, validation_alias="SPIDER_API_PORT")

    jwt_secret: str = Field(default="change-me", validation_alias="SPIDER_JWT_SECRET")
    jwt_algorithm: str = Field(default="HS256", validation_alias="SPIDER_JWT_ALGORITHM")
    jwt_expire_minutes: int = Field(default=1440, validation_alias="SPIDER_JWT_EXPIRE_MINUTES")

    worker_token: str = Field(
        default="development-worker-token",
        validation_alias="SPIDER_WORKER_TOKEN",
    )

    cors_origins: str = Field(
        default="http://localhost:5173",
        validation_alias="SPIDER_CORS_ORIGINS",
    )

    default_detector: str = Field(default="rule-based", validation_alias="SPIDER_DEFAULT_DETECTOR")
    default_security_policy: str = Field(
        default="threshold",
        validation_alias="SPIDER_DEFAULT_SECURITY_POLICY",
    )
    default_threshold: float = Field(default=0.5, validation_alias="SPIDER_DEFAULT_THRESHOLD")

    fail_mode: Literal["open", "closed"] = Field(
        default="closed",
        validation_alias="SPIDER_FAIL_MODE",
    )

    log_prompt_content: bool = Field(default=False, validation_alias="SPIDER_LOG_PROMPT_CONTENT")
    persist_prompt_content: bool = Field(
        default=False,
        validation_alias="SPIDER_PERSIST_PROMPT_CONTENT",
    )

    worker_heartbeat_interval: int = Field(
        default=10,
        validation_alias="SPIDER_WORKER_HEARTBEAT_INTERVAL",
    )
    worker_offline_timeout: int = Field(
        default=30,
        validation_alias="SPIDER_WORKER_OFFLINE_TIMEOUT",
    )

    chunker: str = Field(default="fixed", validation_alias="SPIDER_CHUNKER")
    chunk_size: int = Field(default=2048, validation_alias="SPIDER_CHUNK_SIZE")
    chunk_overlap: int = Field(default=128, validation_alias="SPIDER_CHUNK_OVERLAP")

    default_model: str = Field(default="mock-llm", validation_alias="SPIDER_DEFAULT_MODEL")
    serving_provider: str = Field(default="mock", validation_alias="SPIDER_SERVING_PROVIDER")

    bootstrap_admin_email: str = Field(
        default="admin@spider.local",
        validation_alias="SPIDER_BOOTSTRAP_ADMIN_EMAIL",
    )
    bootstrap_admin_password: str = Field(
        default="spider-admin",
        validation_alias="SPIDER_BOOTSTRAP_ADMIN_PASSWORD",
    )

    api_base_url: str = Field(default="http://localhost:8000", validation_alias="SPIDER_API_BASE_URL")

    @field_validator("default_threshold")
    @classmethod
    def _threshold_range(cls, value: float) -> float:
        if not 0.0 <= value <= 1.0:
            msg = "SPIDER_DEFAULT_THRESHOLD must be between 0.0 and 1.0"
            raise ValueError(msg)
        return value

    @property
    def cors_origin_list(self) -> list[str]:
        return [origin.strip() for origin in self.cors_origins.split(",") if origin.strip()]

    @property
    def is_development(self) -> bool:
        return self.env.lower() in {"development", "dev", "test"}

    @property
    def log_prompts(self) -> bool:
        return self.log_prompt_content


@lru_cache(maxsize=1)
def get_settings() -> Settings:
    return Settings()


def reset_settings_cache() -> None:
    get_settings.cache_clear()

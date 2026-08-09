"""Validated process settings."""

from __future__ import annotations

from typing import Literal

from pydantic import Field, HttpUrl, SecretStr
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """Complete immutable settings graph, constructed once at startup."""

    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        extra="forbid",
        frozen=True,
    )

    app_env: Literal["local", "test", "staging", "production"] = "local"
    log_level: Literal["DEBUG", "INFO", "WARNING", "ERROR"] = "INFO"
    log_json: bool = False
    http_host: str = "127.0.0.1"
    http_port: int = Field(default=8080, ge=1, le=65535)
    http_concurrency: int = Field(default=100, ge=1, le=10_000)
    http_keep_alive_seconds: int = Field(default=5, ge=1, le=300)
    request_timeout_seconds: float = Field(default=5.0, gt=0, le=120)
    shutdown_grace_seconds: int = Field(default=15, ge=1, le=300)
    outbound_base_url: HttpUrl
    outbound_api_token: SecretStr
    outbound_timeout_seconds: float = Field(default=2.0, gt=0, le=30)
    outbound_max_attempts: int = Field(default=3, ge=1, le=5)
    outbound_concurrency: int = Field(default=8, ge=1, le=100)
    worker_concurrency: int = Field(default=2, ge=1, le=16)
    worker_queue_size: int = Field(default=32, ge=1, le=1_000)


def load_settings() -> Settings:
    """Load required settings from the configured settings sources."""
    # BaseSettings supplies required fields from environment sources at runtime;
    # mypy sees only the generated model constructor and cannot represent that path.
    return Settings()  # type: ignore[call-arg]

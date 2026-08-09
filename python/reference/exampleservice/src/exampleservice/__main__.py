"""Thin process entrypoint."""

from __future__ import annotations

import uvicorn

from exampleservice.app import create_app
from exampleservice.config import load_settings
from exampleservice.telemetry.logging import configure_logging


def main() -> None:
    """Validate settings, configure logging, and start Uvicorn."""
    settings = load_settings()
    configure_logging(settings)
    uvicorn.run(
        create_app(settings),
        host=settings.http_host,
        port=settings.http_port,
        limit_concurrency=settings.http_concurrency,
        timeout_keep_alive=settings.http_keep_alive_seconds,
        timeout_graceful_shutdown=settings.shutdown_grace_seconds,
    )


if __name__ == "__main__":
    main()

"""One-owner service logging configuration."""

from __future__ import annotations

import json
import logging
import logging.config
from datetime import UTC, datetime

from exampleservice.config import Settings


class JsonFormatter(logging.Formatter):
    """Serialize a stable, redacted service log envelope."""

    def format(self, record: logging.LogRecord) -> str:
        """Render one JSON object per line."""
        payload: dict[str, object] = {
            "timestamp": datetime.now(UTC).isoformat(),
            "level": record.levelname.lower(),
            "service": "exampleservice",
            "logger": record.name,
            "message": record.getMessage(),
        }
        if record.exc_info is not None:
            payload["exception"] = self.formatException(record.exc_info)
        return json.dumps(payload, separators=(",", ":"), ensure_ascii=False)


def configure_logging(settings: Settings) -> None:
    """Configure the process logging graph once at composition."""
    formatter = "json" if settings.log_json else "text"
    logging.config.dictConfig(
        {
            "version": 1,
            "disable_existing_loggers": False,
            "handlers": {
                "console": {
                    "class": "logging.StreamHandler",
                    "formatter": formatter,
                }
            },
            "formatters": {
                "json": {"()": "exampleservice.telemetry.logging.JsonFormatter"},
                "text": {"format": "%(levelname)s %(name)s %(message)s"},
            },
            "root": {"handlers": ["console"], "level": settings.log_level},
        }
    )

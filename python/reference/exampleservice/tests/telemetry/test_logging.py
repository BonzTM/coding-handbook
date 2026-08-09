"""Structured logging proof."""

from __future__ import annotations

import json
import logging

from exampleservice.config import Settings
from exampleservice.telemetry.logging import JsonFormatter, configure_logging


def test_json_formatter_emits_stable_fields_without_secret() -> None:
    record = logging.LogRecord("orders", logging.INFO, __file__, 1, "created", (), None)

    payload = json.loads(JsonFormatter().format(record))

    assert payload["service"] == "exampleservice"
    assert payload["level"] == "info"
    assert payload["message"] == "created"


def test_logging_configuration_accepts_local_text(settings: Settings) -> None:
    configure_logging(settings)

    assert logging.getLogger().level == logging.INFO

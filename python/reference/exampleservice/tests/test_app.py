"""Assembled lifespan ownership proof."""

from __future__ import annotations

import pytest

from exampleservice.app import create_app
from exampleservice.config import Settings
from exampleservice.telemetry.health import Readiness


@pytest.mark.asyncio
async def test_lifespan_becomes_ready_then_drains(settings: Settings) -> None:
    app = create_app(settings)

    async with app.router.lifespan_context(app):
        readiness: Readiness = app.state.readiness
        assert readiness.is_ready()

    assert not readiness.is_ready()

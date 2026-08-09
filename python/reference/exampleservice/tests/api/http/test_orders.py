"""HTTP validation, problem, correlation, and serialization proof."""

from __future__ import annotations

import asyncio
import json
from datetime import UTC, datetime
from pathlib import Path

import httpx
import pytest
from fastapi import FastAPI, status

from exampleservice.api.http.dependencies import get_metrics, get_order_service, get_readiness
from exampleservice.app import create_app
from exampleservice.config import Settings
from exampleservice.core.models import Money, Sku
from exampleservice.core.service import OrderService
from exampleservice.db.memory import MemoryOrderRepository
from exampleservice.telemetry.health import Readiness
from exampleservice.telemetry.metrics import Metrics

_NOW = datetime(2026, 1, 2, 3, 4, 5, tzinfo=UTC)
_GOLDEN = Path(__file__).parents[2] / "testdata" / "order.json"
_REQUEST_ID_HEX_LENGTH = 32


class FixedClock:
    """Fixed domain clock."""

    def now(self) -> datetime:
        return _NOW


class FakeCatalog:
    """Offline catalog port."""

    async def price_for(self, _sku: Sku) -> Money:
        return Money(minor_units=1250, currency="USD")


def build_test_app(settings: Settings) -> tuple[FastAPI, Metrics, Readiness]:
    app = create_app(settings)
    metrics = Metrics()
    readiness = Readiness()
    readiness.set(True)
    service = OrderService(MemoryOrderRepository(), FakeCatalog(), FixedClock())
    app.dependency_overrides[get_order_service] = lambda: service
    app.dependency_overrides[get_metrics] = lambda: metrics
    app.dependency_overrides[get_readiness] = lambda: readiness
    return app, metrics, readiness


@pytest.mark.asyncio
async def test_create_matches_golden_and_echoes_request_id(settings: Settings) -> None:
    app, _metrics, _readiness = build_test_app(settings)
    transport = httpx.ASGITransport(app=app)
    async with httpx.AsyncClient(transport=transport, base_url="http://test") as client:
        response = await client.post(
            "/orders",
            headers={"x-request-id": "request-123"},
            json={"orderId": "order-1", "sku": "sku-1"},
        )
    expected = json.loads(await asyncio.to_thread(_GOLDEN.read_text, encoding="utf-8"))

    assert response.status_code == status.HTTP_201_CREATED
    assert response.headers["x-request-id"] == "request-123"
    assert response.headers["x-content-type-options"] == "nosniff"
    assert response.headers["referrer-policy"] == "no-referrer"
    assert response.json() == expected


@pytest.mark.asyncio
async def test_unknown_field_returns_stable_problem(settings: Settings) -> None:
    app, _metrics, _readiness = build_test_app(settings)
    transport = httpx.ASGITransport(app=app)
    async with httpx.AsyncClient(transport=transport, base_url="http://test") as client:
        response = await client.post(
            "/orders",
            headers={"x-request-id": "validation-1"},
            json={"orderId": "order-1", "sku": "sku-1", "unexpected": True},
        )

    assert response.status_code == status.HTTP_422_UNPROCESSABLE_CONTENT
    assert response.headers["content-type"] == "application/problem+json"
    assert response.json() == {
        "type": "https://example.invalid/problems/validation",
        "title": "Request validation failed",
        "status": 422,
        "requestId": "validation-1",
    }


@pytest.mark.asyncio
async def test_missing_order_maps_to_problem_and_replaces_bad_request_id(
    settings: Settings,
) -> None:
    app, _metrics, _readiness = build_test_app(settings)
    transport = httpx.ASGITransport(app=app)
    async with httpx.AsyncClient(transport=transport, base_url="http://test") as client:
        response = await client.get("/orders/missing", headers={"x-request-id": "not valid!"})

    request_id = response.headers["x-request-id"]
    assert response.status_code == status.HTTP_404_NOT_FOUND
    assert len(request_id) == _REQUEST_ID_HEX_LENGTH
    assert response.json()["requestId"] == request_id


@pytest.mark.asyncio
async def test_health_and_metrics_are_distinct_operational_surfaces(settings: Settings) -> None:
    app, _metrics, readiness = build_test_app(settings)
    transport = httpx.ASGITransport(app=app)
    async with httpx.AsyncClient(transport=transport, base_url="http://test") as client:
        live = await client.get("/livez")
        ready = await client.get("/readyz")
        metrics = await client.get("/metrics")
        readiness.set(False)
        draining = await client.get("/readyz")

    assert live.json() == {"status": "live"}
    assert ready.status_code == status.HTTP_200_OK
    assert draining.status_code == status.HTTP_503_SERVICE_UNAVAILABLE
    assert "exampleservice_orders_total" in metrics.text

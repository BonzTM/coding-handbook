"""Typed access to lifespan-owned application state."""

from __future__ import annotations

from typing import Protocol, cast

from fastapi import Request

from exampleservice.core.service import OrderService


class ReadinessPort(Protocol):
    """Readiness behavior required by probes."""

    def is_ready(self) -> bool: ...


class MetricsPort(Protocol):
    """Metric behavior required by HTTP endpoints."""

    def record_order(self, operation: str, outcome: str) -> None: ...

    def render(self) -> bytes: ...


def get_order_service(request: Request) -> OrderService:
    """Resolve the core use case wired by lifespan."""
    return cast("OrderService", request.app.state.order_service)


def get_readiness(request: Request) -> ReadinessPort:
    """Resolve the health state wired by lifespan."""
    return cast("ReadinessPort", request.app.state.readiness)


def get_metrics(request: Request) -> MetricsPort:
    """Resolve the metric facade wired by lifespan."""
    return cast("MetricsPort", request.app.state.metrics)

"""Thin order and operational HTTP adapters."""

from __future__ import annotations

import asyncio
from typing import Annotated

from fastapi import APIRouter, Depends, Response, status
from prometheus_client import CONTENT_TYPE_LATEST

from exampleservice.api.http.dependencies import (
    MetricsPort,
    ReadinessPort,
    get_metrics,
    get_order_service,
    get_readiness,
)
from exampleservice.api.http.dto import CreateOrderRequest, OrderResponse, to_order_response
from exampleservice.core.models import OrderId, Sku
from exampleservice.core.service import OrderService

OrderServiceDependency = Annotated[OrderService, Depends(get_order_service)]
ReadinessDependency = Annotated[ReadinessPort, Depends(get_readiness)]
MetricsDependency = Annotated[MetricsPort, Depends(get_metrics)]


def build_router(request_timeout_seconds: float) -> APIRouter:
    """Build routes with one explicit request deadline."""
    router = APIRouter()

    @router.post("/orders", response_model=OrderResponse, status_code=status.HTTP_201_CREATED)
    async def create_order(
        command: CreateOrderRequest,
        service: OrderServiceDependency,
        metrics: MetricsDependency,
    ) -> OrderResponse:
        async with asyncio.timeout(request_timeout_seconds):
            order = await service.create(OrderId(command.order_id), Sku(command.sku))
        metrics.record_order("create", "success")
        return to_order_response(order)

    @router.get("/orders/{order_id}", response_model=OrderResponse)
    async def get_order(
        order_id: str,
        service: OrderServiceDependency,
        metrics: MetricsDependency,
    ) -> OrderResponse:
        async with asyncio.timeout(request_timeout_seconds):
            order = await service.get(OrderId(order_id))
        metrics.record_order("get", "success")
        return to_order_response(order)

    @router.get("/livez")
    async def livez() -> dict[str, str]:
        return {"status": "live"}

    @router.get("/readyz")
    async def readyz(readiness: ReadinessDependency) -> Response:
        ready = readiness.is_ready()
        code = status.HTTP_200_OK if ready else status.HTTP_503_SERVICE_UNAVAILABLE
        return Response(content="ready\n" if ready else "not ready\n", status_code=code)

    @router.get("/metrics")
    async def metrics_endpoint(metrics: MetricsDependency) -> Response:
        return Response(content=metrics.render(), media_type=CONTENT_TYPE_LATEST)

    return router

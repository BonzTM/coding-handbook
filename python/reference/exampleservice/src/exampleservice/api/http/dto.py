"""Explicit HTTP wire contracts."""

from __future__ import annotations

from datetime import datetime

from pydantic import BaseModel, ConfigDict, Field

from exampleservice.core.models import Order


class CreateOrderRequest(BaseModel):
    """Strict inbound order command."""

    model_config = ConfigDict(extra="forbid", strict=True)

    order_id: str = Field(alias="orderId", min_length=1, max_length=64)
    sku: str = Field(min_length=1, max_length=64, pattern=r"^[A-Za-z0-9._-]+$")


class OrderResponse(BaseModel):
    """Stable outbound order representation."""

    model_config = ConfigDict(extra="forbid")

    order_id: str = Field(alias="orderId")
    sku: str
    minor_units: int = Field(alias="minorUnits")
    currency: str
    created_at: datetime = Field(alias="createdAt")


def to_order_response(order: Order) -> OrderResponse:
    """Map one domain snapshot into the public DTO."""
    return OrderResponse(
        orderId=order.order_id,
        sku=order.sku,
        minorUnits=order.price.minor_units,
        currency=order.price.currency,
        createdAt=order.created_at,
    )

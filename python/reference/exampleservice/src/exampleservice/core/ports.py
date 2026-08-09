"""Consumer-owned boundary Protocols."""

from __future__ import annotations

from datetime import datetime
from typing import Protocol

from exampleservice.core.models import Money, Order, OrderId, Sku


class Clock(Protocol):
    """Wall clock used for domain timestamps."""

    def now(self) -> datetime: ...


class OrderRepository(Protocol):
    """Persistence required by the order use case."""

    async def add(self, order: Order) -> None: ...

    async def get(self, order_id: OrderId) -> Order | None: ...


class CatalogPort(Protocol):
    """Catalog behavior required by order creation."""

    async def price_for(self, sku: Sku) -> Money: ...

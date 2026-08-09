"""Process-local repository for the deliberately persistence-free exemplar."""

from __future__ import annotations

import asyncio

from exampleservice.core.models import Order, OrderId
from exampleservice.core.service import OrderAlreadyExistsError


class MemoryOrderRepository:
    """Concurrency-safe in-memory OrderRepository implementation."""

    def __init__(self) -> None:
        self._orders: dict[OrderId, Order] = {}
        self._lock = asyncio.Lock()

    async def add(self, order: Order) -> None:
        """Add an order exactly once within this process."""
        async with self._lock:
            if order.order_id in self._orders:
                raise OrderAlreadyExistsError
            self._orders[order.order_id] = order

    async def get(self, order_id: OrderId) -> Order | None:
        """Return one immutable snapshot when present."""
        async with self._lock:
            return self._orders.get(order_id)

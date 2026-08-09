"""Order use cases and typed failures."""

from __future__ import annotations

from exampleservice.core.models import Order, OrderId, Sku
from exampleservice.core.ports import CatalogPort, Clock, OrderRepository


class OrderError(Exception):
    """Base class for expected order failures."""


class OrderAlreadyExistsError(OrderError):
    """Raised when an order identifier is already present."""


class OrderNotFoundError(OrderError):
    """Raised when an order identifier is absent."""


class CatalogUnavailableError(OrderError):
    """Raised when the required catalog dependency cannot answer."""


class OrderService:
    """Create and retrieve immutable orders."""

    def __init__(
        self,
        repository: OrderRepository,
        catalog: CatalogPort,
        clock: Clock,
    ) -> None:
        self._repository = repository
        self._catalog = catalog
        self._clock = clock

    async def create(self, order_id: OrderId, sku: Sku) -> Order:
        """Price and persist one new order."""
        existing = await self._repository.get(order_id)
        if existing is not None:
            raise OrderAlreadyExistsError
        price = await self._catalog.price_for(sku)
        order = Order(order_id=order_id, sku=sku, price=price, created_at=self._clock.now())
        await self._repository.add(order)
        return order

    async def get(self, order_id: OrderId) -> Order:
        """Return an order or a typed not-found failure."""
        order = await self._repository.get(order_id)
        if order is None:
            raise OrderNotFoundError
        return order

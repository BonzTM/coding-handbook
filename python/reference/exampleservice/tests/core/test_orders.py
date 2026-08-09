"""Framework-free order behavior proof."""

from __future__ import annotations

from dataclasses import dataclass
from datetime import UTC, datetime

import pytest

from exampleservice.core.models import Money, OrderId, Sku
from exampleservice.core.service import (
    OrderAlreadyExistsError,
    OrderNotFoundError,
    OrderService,
)
from exampleservice.db.memory import MemoryOrderRepository

_NOW = datetime(2026, 1, 2, 3, 4, 5, tzinfo=UTC)


@dataclass(frozen=True, slots=True)
class FakeClock:
    """Fixed UTC wall clock."""

    value: datetime = _NOW

    def now(self) -> datetime:
        return self.value


class FakeCatalog:
    """Deterministic catalog port."""

    async def price_for(self, _sku: Sku) -> Money:
        return Money(minor_units=1250, currency="USD")


@pytest.mark.asyncio
async def test_create_round_trips_plain_domain_order() -> None:
    service = OrderService(MemoryOrderRepository(), FakeCatalog(), FakeClock())

    created = await service.create(OrderId("order-1"), Sku("sku-1"))
    loaded = await service.get(OrderId("order-1"))

    assert loaded == created
    assert created.created_at == _NOW
    assert created.price == Money(minor_units=1250, currency="USD")


@pytest.mark.asyncio
async def test_create_rejects_duplicate_identifier() -> None:
    service = OrderService(MemoryOrderRepository(), FakeCatalog(), FakeClock())
    await service.create(OrderId("order-1"), Sku("sku-1"))

    with pytest.raises(OrderAlreadyExistsError):
        await service.create(OrderId("order-1"), Sku("sku-2"))


@pytest.mark.asyncio
async def test_get_reports_typed_not_found() -> None:
    service = OrderService(MemoryOrderRepository(), FakeCatalog(), FakeClock())

    with pytest.raises(OrderNotFoundError):
        await service.get(OrderId("missing"))


@pytest.mark.parametrize(
    ("minor_units", "currency"),
    [(-1, "USD"), (1, "usd"), (1, "US")],
)
def test_money_rejects_invalid_values(minor_units: int, currency: str) -> None:
    with pytest.raises(ValueError, match=r"minor_units|currency"):
        Money(minor_units=minor_units, currency=currency)

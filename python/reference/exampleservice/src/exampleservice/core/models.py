"""Orders domain values."""

from __future__ import annotations

from dataclasses import dataclass
from datetime import UTC, datetime
from typing import NewType

OrderId = NewType("OrderId", str)
Sku = NewType("Sku", str)
_CURRENCY_CODE_LENGTH = 3


@dataclass(frozen=True, slots=True)
class Money:
    """Non-negative amount in integer minor units."""

    minor_units: int
    currency: str

    def __post_init__(self) -> None:
        if self.minor_units < 0:
            raise ValueError("minor_units must be non-negative")
        if len(self.currency) != _CURRENCY_CODE_LENGTH or not self.currency.isupper():
            raise ValueError("currency must be a three-letter uppercase code")


@dataclass(frozen=True, slots=True)
class Order:
    """Immutable order snapshot."""

    order_id: OrderId
    sku: Sku
    price: Money
    created_at: datetime

    def __post_init__(self) -> None:
        if not self.order_id:
            raise ValueError("order_id must not be empty")
        if not self.sku:
            raise ValueError("sku must not be empty")
        if self.created_at.tzinfo is None or self.created_at.utcoffset() is None:
            raise ValueError("created_at must be timezone-aware")
        if self.created_at.utcoffset() != UTC.utcoffset(self.created_at):
            raise ValueError("created_at must be UTC")

"""Outbound catalog retry, timeout, and parsing proof."""

from __future__ import annotations

from dataclasses import dataclass, field

import httpx
import pytest

from exampleservice.clients.catalog import CatalogClient, MonotonicClock
from exampleservice.core.models import Sku
from exampleservice.core.service import CatalogUnavailableError

_EXPECTED_PRICE = 2500


@dataclass(slots=True)
class FakeClock:
    """Manually advanced monotonic clock."""

    value: float = 0.0

    def now(self) -> float:
        return self.value


@dataclass(slots=True)
class FakeSleeper:
    """Sleeper that advances fake time without waiting."""

    clock: FakeClock
    delays: list[float] = field(default_factory=list)

    async def sleep(self, delay_seconds: float) -> None:
        self.delays.append(delay_seconds)
        self.clock.value += delay_seconds


class FixedRandom:
    """Select the maximum jitter for exact proof."""

    def uniform(self, _lower: float, upper: float) -> float:
        return upper


@dataclass(slots=True)
class ExpiringClock:
    """Expire the logical total budget on its second read."""

    calls: int = 0

    def now(self) -> float:
        self.calls += 1
        return 0.0 if self.calls == 1 else 2.0


def build_catalog(
    client: httpx.AsyncClient,
    *,
    attempts: int,
    clock: MonotonicClock,
    sleeper: FakeSleeper,
) -> CatalogClient:
    return CatalogClient(
        client,
        max_attempts=attempts,
        total_timeout_seconds=1.0,
        concurrency=2,
        clock=clock,
        sleeper=sleeper,
        random_source=FixedRandom(),
    )


@pytest.mark.asyncio
async def test_retry_uses_bounded_full_jitter_without_real_sleep() -> None:
    responses = iter([503, 200])

    def handler(_request: httpx.Request) -> httpx.Response:
        status = next(responses)
        return httpx.Response(status, json={"minorUnits": 2500, "currency": "USD"})

    clock = FakeClock()
    sleeper = FakeSleeper(clock)
    async with httpx.AsyncClient(
        base_url="https://catalog.example.invalid",
        transport=httpx.MockTransport(handler),
        timeout=httpx.Timeout(1.0),
    ) as client:
        price = await build_catalog(client, attempts=2, clock=clock, sleeper=sleeper).price_for(
            Sku("sku-1")
        )

    assert price.minor_units == _EXPECTED_PRICE
    assert sleeper.delays == [0.05]


@pytest.mark.asyncio
async def test_permanent_status_is_not_retried() -> None:
    calls = 0

    def handler(_request: httpx.Request) -> httpx.Response:
        nonlocal calls
        calls += 1
        return httpx.Response(400)

    clock = FakeClock()
    sleeper = FakeSleeper(clock)
    async with httpx.AsyncClient(
        base_url="https://catalog.example.invalid",
        transport=httpx.MockTransport(handler),
        timeout=httpx.Timeout(1.0),
    ) as client:
        with pytest.raises(CatalogUnavailableError):
            await build_catalog(client, attempts=3, clock=clock, sleeper=sleeper).price_for(
                Sku("sku-1")
            )

    assert calls == 1
    assert sleeper.delays == []


@pytest.mark.asyncio
async def test_transport_timeout_exhausts_attempt_bound() -> None:
    def handler(request: httpx.Request) -> httpx.Response:
        raise httpx.ReadTimeout("timed out", request=request)

    clock = FakeClock()
    sleeper = FakeSleeper(clock)
    async with httpx.AsyncClient(
        base_url="https://catalog.example.invalid",
        transport=httpx.MockTransport(handler),
        timeout=httpx.Timeout(1.0),
    ) as client:
        with pytest.raises(CatalogUnavailableError):
            await build_catalog(client, attempts=2, clock=clock, sleeper=sleeper).price_for(
                Sku("sku-1")
            )

    assert sleeper.delays == [0.05]


@pytest.mark.asyncio
async def test_retry_stops_when_fake_clock_exhausts_total_timeout() -> None:
    def handler(_request: httpx.Request) -> httpx.Response:
        return httpx.Response(503)

    sleeper = FakeSleeper(FakeClock())
    async with httpx.AsyncClient(
        base_url="https://catalog.example.invalid",
        transport=httpx.MockTransport(handler),
        timeout=httpx.Timeout(1.0),
    ) as client:
        catalog = build_catalog(client, attempts=3, clock=ExpiringClock(), sleeper=sleeper)
        with pytest.raises(CatalogUnavailableError):
            await catalog.price_for(Sku("sku-1"))

    assert sleeper.delays == []

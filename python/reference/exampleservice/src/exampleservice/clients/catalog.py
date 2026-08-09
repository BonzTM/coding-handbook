"""Bounded HTTP catalog adapter with deterministic retry seams."""

from __future__ import annotations

import asyncio
import secrets
import time
from typing import Protocol

import httpx
from pydantic import BaseModel, ConfigDict, Field, ValidationError

from exampleservice.core.models import Money, Sku
from exampleservice.core.service import CatalogUnavailableError

_RETRYABLE_STATUS = frozenset({429, 502, 503, 504})
_BACKOFF_BASE_SECONDS = 0.05
_BACKOFF_CAP_SECONDS = 1.0


class MonotonicClock(Protocol):
    """Elapsed-time source for total retry budgets."""

    def now(self) -> float: ...


class Sleeper(Protocol):
    """Cancellable delay seam."""

    async def sleep(self, delay_seconds: float) -> None: ...


class RandomSource(Protocol):
    """Jitter source."""

    def uniform(self, lower: float, upper: float) -> float: ...


class SystemMonotonicClock:
    """Production elapsed-time source."""

    def now(self) -> float:
        """Read the monotonic clock."""
        return time.monotonic()


class AsyncioSleeper:
    """Production cancellable sleeper."""

    async def sleep(self, delay_seconds: float) -> None:
        """Sleep without blocking the event loop."""
        await asyncio.sleep(delay_seconds)


class SecureRandomSource:
    """Process-local full-jitter source."""

    def __init__(self) -> None:
        self._random = secrets.SystemRandom()

    def uniform(self, lower: float, upper: float) -> float:
        """Return one random delay inside the inclusive bounds."""
        return self._random.uniform(lower, upper)


class CatalogPrice(BaseModel):
    """Tolerant DTO for the independently evolving catalog response."""

    model_config = ConfigDict(extra="ignore", strict=True)

    minor_units: int = Field(alias="minorUnits", ge=0)
    currency: str = Field(min_length=3, max_length=3, pattern=r"^[A-Z]{3}$")


class CatalogClient:
    """HTTPX-backed CatalogPort implementation."""

    def __init__(
        self,
        client: httpx.AsyncClient,
        *,
        max_attempts: int,
        total_timeout_seconds: float,
        concurrency: int,
        clock: MonotonicClock,
        sleeper: Sleeper,
        random_source: RandomSource,
    ) -> None:
        if max_attempts < 1 or concurrency < 1 or total_timeout_seconds <= 0:
            raise ValueError("catalog bounds must be positive")
        self._client = client
        self._max_attempts = max_attempts
        self._total_timeout_seconds = total_timeout_seconds
        self._clock = clock
        self._sleeper = sleeper
        self._random = random_source
        self._semaphore = asyncio.Semaphore(concurrency)

    async def price_for(self, sku: Sku) -> Money:
        """Fetch one price within a total deadline and concurrency bound."""
        async with self._semaphore:
            try:
                async with asyncio.timeout(self._total_timeout_seconds):
                    return await self._attempts(sku)
            except TimeoutError as error:
                raise CatalogUnavailableError from error

    async def _attempts(self, sku: Sku) -> Money:
        deadline = self._clock.now() + self._total_timeout_seconds
        for attempt in range(self._max_attempts):
            try:
                response = await self._client.get(f"/v1/catalog/{sku}")
                if response.status_code not in _RETRYABLE_STATUS:
                    return self._parse(response)
            except httpx.TransportError as error:
                if attempt + 1 == self._max_attempts:
                    raise CatalogUnavailableError from error
            if attempt + 1 < self._max_attempts:
                await self._backoff(attempt, deadline)
        raise CatalogUnavailableError

    def _parse(self, response: httpx.Response) -> Money:
        try:
            response.raise_for_status()
            price = CatalogPrice.model_validate_json(response.content)
        except (httpx.HTTPStatusError, ValidationError) as error:
            raise CatalogUnavailableError from error
        return Money(minor_units=price.minor_units, currency=price.currency)

    async def _backoff(self, attempt: int, deadline: float) -> None:
        maximum = min(_BACKOFF_CAP_SECONDS, _BACKOFF_BASE_SECONDS * (2**attempt))
        remaining = deadline - self._clock.now()
        if remaining <= 0:
            raise CatalogUnavailableError
        await self._sleeper.sleep(min(self._random.uniform(0.0, maximum), remaining))

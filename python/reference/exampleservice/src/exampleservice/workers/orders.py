"""Bounded order verification worker."""

from __future__ import annotations

import asyncio
from dataclasses import dataclass
from typing import Protocol

from exampleservice.core.models import OrderId


class OrderReader(Protocol):
    """Use-case surface consumed by the worker."""

    async def get(self, order_id: OrderId) -> object: ...


@dataclass(frozen=True, slots=True)
class _Stop:
    """Private queue sentinel."""


_STOP = _Stop()
_MAX_ITEMS_PER_CONSUMER = 1_000_000


class OrderWorker:
    """Fixed-size worker set over a bounded queue."""

    def __init__(self, reader: OrderReader, *, concurrency: int, queue_size: int) -> None:
        if concurrency < 1 or queue_size < 1:
            raise ValueError("worker bounds must be positive")
        self._reader = reader
        self._concurrency = concurrency
        self._queue: asyncio.Queue[OrderId | _Stop] = asyncio.Queue(maxsize=queue_size)
        self.started = asyncio.Event()

    async def submit(self, order_id: OrderId, *, timeout_seconds: float) -> None:
        """Enqueue one item within the caller's admission budget."""
        async with asyncio.timeout(timeout_seconds):
            await self._queue.put(order_id)

    async def run(self) -> None:
        """Own a fixed TaskGroup until every consumer receives a stop marker."""
        async with asyncio.TaskGroup() as task_group:
            for worker_number in range(self._concurrency):
                task_group.create_task(self._consume(), name=f"order-worker-{worker_number}")
            self.started.set()

    async def stop(self) -> None:
        """Request an ordered drain without abandoning queued work."""
        for _ in range(self._concurrency):
            await self._queue.put(_STOP)

    async def _consume(self) -> None:
        for _ in range(_MAX_ITEMS_PER_CONSUMER):
            item = await self._queue.get()
            try:
                if isinstance(item, _Stop):
                    return
                await self._reader.get(item)
            finally:
                self._queue.task_done()
        raise RuntimeError("worker lifetime item bound exhausted")

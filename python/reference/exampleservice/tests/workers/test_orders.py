"""Structured worker ownership and shutdown proof."""

from __future__ import annotations

import asyncio

import pytest

from exampleservice.core.models import OrderId
from exampleservice.workers.orders import OrderWorker


class FakeReader:
    """Observable worker dependency."""

    def __init__(self) -> None:
        self.seen: list[OrderId] = []
        self.called = asyncio.Event()

    async def get(self, order_id: OrderId) -> object:
        self.seen.append(order_id)
        self.called.set()
        return object()


@pytest.mark.asyncio
async def test_worker_drains_and_stops_owned_task_group() -> None:
    reader = FakeReader()
    worker = OrderWorker(reader, concurrency=2, queue_size=2)

    async with asyncio.TaskGroup() as tasks:
        tasks.create_task(worker.run())
        await worker.started.wait()
        await worker.submit(OrderId("order-1"), timeout_seconds=1.0)
        await reader.called.wait()
        await worker.stop()

    assert reader.seen == [OrderId("order-1")]

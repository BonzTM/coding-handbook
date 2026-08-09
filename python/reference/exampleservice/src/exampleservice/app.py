"""FastAPI composition and lifespan ownership."""

from __future__ import annotations

import asyncio
from collections.abc import AsyncIterator, Callable
from contextlib import AbstractAsyncContextManager, asynccontextmanager
from datetime import UTC, datetime

import httpx
from fastapi import FastAPI

from exampleservice.api.http.errors import install_exception_handlers
from exampleservice.api.http.middleware import RequestIdMiddleware
from exampleservice.api.http.routers import build_router
from exampleservice.clients.catalog import (
    AsyncioSleeper,
    CatalogClient,
    SecureRandomSource,
    SystemMonotonicClock,
)
from exampleservice.config import Settings
from exampleservice.core.service import OrderService
from exampleservice.db.memory import MemoryOrderRepository
from exampleservice.telemetry.health import Readiness
from exampleservice.telemetry.metrics import Metrics
from exampleservice.telemetry.tracing import Tracing
from exampleservice.workers.orders import OrderWorker

Lifespan = Callable[[FastAPI], AbstractAsyncContextManager[None]]


class SystemClock:
    """Production UTC wall clock."""

    def now(self) -> datetime:
        """Return one aware UTC instant."""
        return datetime.now(UTC)


def _http_client(settings: Settings) -> httpx.AsyncClient:
    timeout = httpx.Timeout(settings.outbound_timeout_seconds)
    limits = httpx.Limits(max_connections=settings.outbound_concurrency)
    token = settings.outbound_api_token.get_secret_value()
    return httpx.AsyncClient(
        base_url=str(settings.outbound_base_url),
        headers={"Authorization": f"Bearer {token}"},
        timeout=timeout,
        limits=limits,
    )


def _order_service(settings: Settings, client: httpx.AsyncClient) -> OrderService:
    catalog = CatalogClient(
        client,
        max_attempts=settings.outbound_max_attempts,
        total_timeout_seconds=settings.outbound_timeout_seconds,
        concurrency=settings.outbound_concurrency,
        clock=SystemMonotonicClock(),
        sleeper=AsyncioSleeper(),
        random_source=SecureRandomSource(),
    )
    return OrderService(MemoryOrderRepository(), catalog, SystemClock())


def _wire_state(
    app: FastAPI,
    service: OrderService,
    readiness: Readiness,
    metrics: Metrics,
    settings: Settings,
) -> OrderWorker:
    app.state.order_service = service
    app.state.readiness = readiness
    app.state.metrics = metrics
    return OrderWorker(
        service,
        concurrency=settings.worker_concurrency,
        queue_size=settings.worker_queue_size,
    )


def _lifespan(settings: Settings, tracing: Tracing) -> Lifespan:
    @asynccontextmanager
    async def lifespan(app: FastAPI) -> AsyncIterator[None]:
        readiness = Readiness()
        metrics = Metrics()
        try:
            async with _http_client(settings) as client:
                tracing.instrument_client(client)
                worker = _wire_state(
                    app, _order_service(settings, client), readiness, metrics, settings
                )
                async with asyncio.TaskGroup() as tasks:
                    tasks.create_task(worker.run(), name="order-worker-supervisor")
                    await worker.started.wait()
                    readiness.set(True)
                    try:
                        yield
                    finally:
                        readiness.set(False)
                        async with asyncio.timeout(settings.shutdown_grace_seconds):
                            await worker.stop()
        finally:
            tracing.shutdown()

    return lifespan


def create_app(settings: Settings) -> FastAPI:
    """Construct an unstarted application with no import-time I/O."""
    tracing = Tracing()
    app = FastAPI(title="exampleservice", version="0.1.0", lifespan=_lifespan(settings, tracing))
    app.state.settings = settings
    app.add_middleware(RequestIdMiddleware)
    install_exception_handlers(app)
    app.include_router(build_router(settings.request_timeout_seconds))
    tracing.instrument_app(app)
    return app

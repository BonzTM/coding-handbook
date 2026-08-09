"""Offline-safe OpenTelemetry ownership."""

from __future__ import annotations

import httpx
from fastapi import FastAPI
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from opentelemetry.instrumentation.httpx import HTTPXClientInstrumentor
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.sampling import ALWAYS_OFF


class Tracing:
    """Own an offline-safe provider and adapter instrumentation."""

    def __init__(self) -> None:
        self._provider = TracerProvider(
            sampler=ALWAYS_OFF,
            resource=Resource.create({"service.name": "exampleservice"}),
        )

    def instrument_app(self, app: FastAPI) -> None:
        """Attach server spans to one application instance."""
        FastAPIInstrumentor.instrument_app(app, tracer_provider=self._provider)

    def instrument_client(self, client: httpx.AsyncClient) -> None:
        """Attach client spans to one lifespan-owned client."""
        HTTPXClientInstrumentor.instrument_client(client, tracer_provider=self._provider)

    def shutdown(self) -> None:
        """Flush and stop telemetry after resource users exit."""
        self._provider.shutdown()

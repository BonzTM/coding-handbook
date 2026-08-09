"""Private-registry Prometheus metrics."""

from __future__ import annotations

from prometheus_client import CollectorRegistry, Counter, generate_latest


class Metrics:
    """Low-cardinality application metric facade."""

    def __init__(self) -> None:
        self._registry = CollectorRegistry()
        self._orders = Counter(
            "exampleservice_orders_total",
            "Order operations by bounded outcome",
            labelnames=("operation", "outcome"),
            registry=self._registry,
        )

    def record_order(self, operation: str, outcome: str) -> None:
        """Record one bounded order outcome."""
        self._orders.labels(operation=operation, outcome=outcome).inc()

    def render(self) -> bytes:
        """Render this service's private registry."""
        return generate_latest(self._registry)

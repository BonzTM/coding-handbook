"""Concurrency-safe event-loop readiness state."""

from __future__ import annotations


class Readiness:
    """Small readiness state owned by application lifespan."""

    def __init__(self) -> None:
        self._ready = False

    def set(self, value: bool) -> None:
        """Change readiness during startup or drain."""
        self._ready = value

    def is_ready(self) -> bool:
        """Report whether the instance can accept work."""
        return self._ready

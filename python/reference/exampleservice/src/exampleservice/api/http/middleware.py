"""Bounded request correlation middleware."""

from __future__ import annotations

import contextvars
import re
import secrets

from starlette.datastructures import Headers, MutableHeaders
from starlette.types import ASGIApp, Message, Receive, Scope, Send

_REQUEST_ID_PATTERN = re.compile(r"^[A-Za-z0-9._-]{1,64}$")
_request_id: contextvars.ContextVar[str] = contextvars.ContextVar("request_id", default="unknown")


def current_request_id() -> str:
    """Return the current bounded correlation identifier."""
    return _request_id.get()


def _select_request_id(scope: Scope) -> str:
    supplied = Headers(scope=scope).get("x-request-id", "")
    if _REQUEST_ID_PATTERN.fullmatch(supplied) is not None:
        return supplied
    return secrets.token_hex(16)


class RequestIdMiddleware:
    """Validate, scope, and echo a request identifier."""

    def __init__(self, app: ASGIApp) -> None:
        self._app = app

    async def __call__(self, scope: Scope, receive: Receive, send: Send) -> None:
        if scope["type"] != "http":
            await self._app(scope, receive, send)
            return
        request_id = _select_request_id(scope)
        token = _request_id.set(request_id)

        async def send_with_id(message: Message) -> None:
            if message["type"] == "http.response.start":
                headers = MutableHeaders(scope=message)
                headers.append("x-request-id", request_id)
                headers["x-content-type-options"] = "nosniff"
                headers["referrer-policy"] = "no-referrer"
            await send(message)

        try:
            await self._app(scope, receive, send_with_id)
        finally:
            _request_id.reset(token)

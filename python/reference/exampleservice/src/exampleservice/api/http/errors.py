"""Single RFC 9457-style exception boundary."""

from __future__ import annotations

import logging

from fastapi import FastAPI, Request
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse

from exampleservice.api.http.middleware import current_request_id
from exampleservice.core.service import (
    CatalogUnavailableError,
    OrderAlreadyExistsError,
    OrderNotFoundError,
)

_LOGGER = logging.getLogger(__name__)
_PROBLEM_MEDIA_TYPE = "application/problem+json"


def _problem(*, problem_type: str, title: str, status: int) -> JSONResponse:
    return JSONResponse(
        status_code=status,
        media_type=_PROBLEM_MEDIA_TYPE,
        content={
            "type": f"https://example.invalid/problems/{problem_type}",
            "title": title,
            "status": status,
            "requestId": current_request_id(),
        },
    )


def install_exception_handlers(app: FastAPI) -> None:
    """Register every HTTP error mapping in one place."""

    @app.exception_handler(OrderNotFoundError)
    async def not_found(_request: Request, _error: OrderNotFoundError) -> JSONResponse:
        return _problem(problem_type="order-not-found", title="Order not found", status=404)

    @app.exception_handler(OrderAlreadyExistsError)
    async def conflict(_request: Request, _error: OrderAlreadyExistsError) -> JSONResponse:
        return _problem(problem_type="order-exists", title="Order already exists", status=409)

    @app.exception_handler(CatalogUnavailableError)
    async def unavailable(_request: Request, _error: CatalogUnavailableError) -> JSONResponse:
        return _problem(problem_type="catalog-unavailable", title="Catalog unavailable", status=503)

    @app.exception_handler(RequestValidationError)
    async def validation(_request: Request, _error: RequestValidationError) -> JSONResponse:
        return _problem(problem_type="validation", title="Request validation failed", status=422)

    @app.exception_handler(TimeoutError)
    async def timeout(_request: Request, _error: TimeoutError) -> JSONResponse:
        return _problem(problem_type="timeout", title="Request deadline exceeded", status=504)

    @app.exception_handler(Exception)
    async def unexpected(_request: Request, error: Exception) -> JSONResponse:
        _LOGGER.exception("unhandled request failure", exc_info=error)
        return _problem(problem_type="internal", title="Internal server error", status=500)

"""Suite-wide immutable configuration fixture."""

from __future__ import annotations

import pytest
from pydantic import HttpUrl, SecretStr

from exampleservice.config import Settings


@pytest.fixture
def settings() -> Settings:
    """Return complete test settings without reading a dotenv file."""
    return Settings(
        app_env="test",
        outbound_base_url=HttpUrl("https://catalog.example.invalid"),
        outbound_api_token=SecretStr("fixed-test-token"),
    )

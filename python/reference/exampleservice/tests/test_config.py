"""Configuration boundary proof."""

from __future__ import annotations

from pathlib import Path

import pytest
from pydantic import ValidationError

from exampleservice.config import Settings, load_settings


def test_settings_reject_missing_required_values(
    monkeypatch: pytest.MonkeyPatch,
    tmp_path: Path,
) -> None:
    monkeypatch.delenv("OUTBOUND_BASE_URL", raising=False)
    monkeypatch.delenv("OUTBOUND_API_TOKEN", raising=False)
    monkeypatch.chdir(tmp_path)

    with pytest.raises(ValidationError) as captured:
        load_settings()

    missing = {str(error["loc"][0]) for error in captured.value.errors()}
    assert missing == {"outbound_base_url", "outbound_api_token"}


def test_secret_representation_is_redacted(settings: Settings) -> None:
    rendered = repr(settings)

    assert "fixed-test-token" not in rendered
    assert "**********" in rendered

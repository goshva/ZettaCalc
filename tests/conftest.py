import json
from pathlib import Path

import pytest

from src.domain.catalog import Catalog, Sku

FIXTURES = Path(__file__).parent / "fixtures"
NOW = "2026-11-03T00:00:00+00:00"


@pytest.fixture
def catalog() -> Catalog:
    data = json.loads((FIXTURES / "catalog.json").read_text(encoding="utf-8"))
    return Catalog([Sku(**item) for item in data["items"]], data["snapshot"])


def load_json(path: Path):
    return json.loads(path.read_text(encoding="utf-8"))

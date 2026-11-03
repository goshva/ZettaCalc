import pytest

from src.domain.catalog import UnknownSkuError


def test_known_sku(catalog):
    assert catalog.get("SKU-GROUT-5").price_rub == 1150.0


def test_unknown_sku_is_never_invented(catalog):
    with pytest.raises(UnknownSkuError):
        catalog.get("SKU-MADE-UP")

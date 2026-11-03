import pytest

from src.domain.catalog import UnknownSkuError
from src.domain.estimate import build_estimate


def test_total_and_range(catalog):
    est = build_estimate([("SKU-GLUE-C2-25", 2), ("SKU-GROUT-5", 1)], catalog, ["a", "b"])
    assert est.total_rub == 2430.0
    assert est.uncertainty_pct == 6.0
    assert est.low_rub < est.total_rub < est.high_rub


def test_uncertainty_is_capped(catalog):
    est = build_estimate([("SKU-GROUT-5", 1)], catalog, ["x"] * 50)
    assert est.uncertainty_pct == 25.0


def test_unknown_sku_rejected(catalog):
    with pytest.raises(UnknownSkuError):
        build_estimate([("NOPE", 1)], catalog, [])

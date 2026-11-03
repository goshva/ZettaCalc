"""Tool layer: тонкие обёртки над детерминированным ядром (read-only)."""
from __future__ import annotations

from typing import Any

from src.domain.catalog import Catalog, UnknownSkuError
from src.domain.geometry import Contour
from src.harness.runtime import Tool


def make_tools(catalog: Catalog) -> dict[str, Tool]:
    def get_catalog_item(sku: str) -> dict[str, Any]:
        try:
            item = catalog.get(sku)
        except UnknownSkuError:
            return {"found": False, "sku": sku}
        return {"found": True, "sku": item.sku, "name": item.name, "price_rub": item.price_rub}

    def validate_geometry(contour_mm: list[list[float]]) -> dict[str, Any]:
        if len(contour_mm) < 3:
            return {"valid": False, "reason": "меньше 3 точек"}
        area = Contour([(p[0], p[1]) for p in contour_mm]).area_m2()
        return {"valid": area > 0, "area_m2": round(area, 2)}

    return {
        "get_catalog_item": Tool(get_catalog_item),
        "validate_geometry": Tool(validate_geometry),
    }

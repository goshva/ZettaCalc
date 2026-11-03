"""Catalog & Pricing (PIM): SKU, посадочные размеры, цены.

Правило: цены и SKU — только из PIM. Никаких значений «из головы».
"""
from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True)
class Sku:
    sku: str
    name: str
    category: str
    width_mm: int
    depth_mm: int
    height_mm: int
    seat_width_mm: int | None = None
    seat_depth_mm: int | None = None
    price_rub: float = 0.0


def filter_by_fit(items: list[Sku], seat_width_mm: int, tolerance_mm: int = 5) -> list[Sku]:
    """Сценарий замены: отфильтровать SKU по посадочному размеру."""
    return [
        s for s in items
        if s.seat_width_mm is not None
        and abs(s.seat_width_mm - seat_width_mm) <= tolerance_mm
    ]

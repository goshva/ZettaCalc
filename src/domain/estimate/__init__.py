"""Estimate: предварительный расчёт. Цены только из PIM, диапазон явный."""
from __future__ import annotations

from dataclasses import dataclass, field

from src.domain.catalog import Catalog


@dataclass(frozen=True)
class EstimateLine:
    sku: str
    qty: float
    unit_price_rub: float

    @property
    def total_rub(self) -> float:
        return round(self.qty * self.unit_price_rub, 2)


UNCERTAINTY_PER_ASSUMPTION_PCT = 3.0
UNCERTAINTY_CAP_PCT = 25.0


@dataclass
class Estimate:
    lines: list[EstimateLine]
    assumptions: list[str] = field(default_factory=list)
    uncertainty_pct: float = 0.0  # растёт с числом неподтверждённых допущений

    @property
    def total_rub(self) -> float:
        return round(sum(line.total_rub for line in self.lines), 2)

    @property
    def low_rub(self) -> float:
        return round(self.total_rub * (1 - self.uncertainty_pct / 100), 2)

    @property
    def high_rub(self) -> float:
        return round(self.total_rub * (1 + self.uncertainty_pct / 100), 2)


def build_estimate(
    items: list[tuple[str, float]], catalog: Catalog, assumptions: list[str]
) -> Estimate:
    """items: (sku, количество). Неизвестный SKU поднимает UnknownSkuError."""
    lines = [EstimateLine(sku, qty, catalog.get(sku).price_rub) for sku, qty in items]
    pct = min(len(assumptions) * UNCERTAINTY_PER_ASSUMPTION_PCT, UNCERTAINTY_CAP_PCT)
    return Estimate(lines, list(assumptions), pct)

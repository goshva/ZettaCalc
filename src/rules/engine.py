"""Rule Engine: подбор техкарт по условиям. Детерминированный, без LLM.

Обязательные операции техкарты нельзя убрать ради цены. Если условия не
выполнены или неизвестны, карта не применяется молча: результат содержит
причину и флаг needs_review.
"""
from __future__ import annotations

import json
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any

TECHCARDS_DIR = Path(__file__).parent / "techcards"


@dataclass(frozen=True)
class Surface:
    kind: str
    material: str
    area_m2: float
    unevenness_mm_per_m: float | None = None  # None = клиент не знает


@dataclass
class Selection:
    techcard: dict[str, Any] | None
    needs_review: bool
    reasons: list[str] = field(default_factory=list)
    assumptions: list[str] = field(default_factory=list)


def load_techcards(directory: Path = TECHCARDS_DIR) -> list[dict[str, Any]]:
    cards = [json.loads(p.read_text(encoding="utf-8")) for p in sorted(directory.glob("*.json"))]
    return [c for c in cards if "conditions" in c]


def select_techcard(surface: Surface, cards: list[dict[str, Any]]) -> Selection:
    candidates = [
        c
        for c in cards
        if c["conditions"].get("surface") == surface.kind
        and c["conditions"].get("material") == surface.material
    ]
    if not candidates:
        return Selection(None, True, [f"нет техкарты для {surface.kind}/{surface.material}"])

    card = max(candidates, key=lambda c: c["version"])
    cond = card["conditions"]
    reasons: list[str] = []
    assumptions: list[str] = []

    if surface.area_m2 < cond.get("min_area_m2", 0):
        reasons.append(f"площадь {surface.area_m2} м² меньше минимума {cond['min_area_m2']} м²")

    limit = cond.get("max_unevenness_mm_per_m")
    if limit is not None:
        if surface.unevenness_mm_per_m is None:
            assumptions.append(
                f"{card['code']}: ровность основания неизвестна, требует проверки на объекте"
            )
        elif surface.unevenness_mm_per_m > limit:
            reasons.append(
                f"неровность {surface.unevenness_mm_per_m} мм/м выше допустимых {limit} мм/м"
            )

    if reasons:
        return Selection(None, True, reasons, assumptions)
    return Selection(card, False, [], assumptions)

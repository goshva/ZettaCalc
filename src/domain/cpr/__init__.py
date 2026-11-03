"""CPR Generator: сборка ЦПР из входа (rule engine + PIM).

Правило: обязательные операции нельзя молча убирать ради более низкой цены.
"""
from __future__ import annotations

from dataclasses import dataclass, field
from typing import Any

from src.domain.catalog import Catalog
from src.domain.estimate import build_estimate
from src.domain.geometry import Contour
from src.rules.engine import Surface, select_techcard


@dataclass
class Cpr:
    cpr_id: str
    version: int = 1
    status: str = "draft"
    assumptions: list[str] = field(default_factory=list)
    checklist: dict[str, bool] = field(default_factory=dict)

    def can_approve(self) -> bool:
        return all(self.checklist.values()) if self.checklist else False


DISCLAIMER = (
    "Предварительный расчёт. Допущения требуют подтверждения; "
    "итог может измениться после проверки объекта."
)


@dataclass(frozen=True)
class RoomInput:
    room_id: str
    type: str
    contour_mm: list[tuple[float, float]]
    ceiling_height_m: float
    floor_material: str
    floor_unevenness_mm_per_m: float | None  # None = клиент не знает
    completeness_checklist: dict[str, bool]


def build_cpr(
    *,
    cpr_seq: int,
    year: int,
    project_id: str,
    room: RoomInput,
    catalog: Catalog,
    cards: list[dict[str, Any]],
    now: str,
    created_by: str = "cpr-generator",
) -> dict[str, Any]:
    """Детерминированная сборка ЦПР (dict по schemas/json/cpr.schema.json).

    Никаких вызовов LLM. Один и тот же вход даёт побайтно тот же выход
    (время передаётся снаружи), поэтому pass^k проверяется на равенстве.
    """
    area = round(Contour(room.contour_mm).area_m2(), 2)
    selection = select_techcard(
        Surface("floor", room.floor_material, area, room.floor_unevenness_mm_per_m), cards
    )
    assumptions = list(selection.assumptions)
    status = "draft"
    works: list[dict[str, Any]] = []
    items: list[tuple[str, float]] = []
    materials: list[dict[str, Any]] = []

    if selection.techcard is None:
        status = "pending_review"
        assumptions.extend(selection.reasons)
    else:
        card = selection.techcard
        works = [
            {"code": op["code"], "name": op["name"], "unit": op["unit"], "qty": area}
            for op in card["operations"]
        ]
        for key, sku in card["materials"].items():
            qty = round(card["consumption"][key] * area, 2)
            items.append((sku, qty))
            materials.append({"sku": sku, "qty": qty, "techcard": card["code"]})

    if not all(room.completeness_checklist.values()):
        status = "pending_review"

    estimate = build_estimate(items, catalog, assumptions)
    return {
        "cpr_id": f"CPR-{year:04d}-{cpr_seq:06d}",
        "version": 1,
        "status": status,
        "created_at": now,
        "updated_at": now,
        "source": {
            "reference_id": "n/a",
            "implementation_variant": "balanced",
            "brief_id": f"{project_id}-brief",
            "geometry_id": f"{room.room_id}-geometry",
            "catalog_snapshot": catalog.snapshot,
        },
        "client": {
            "project_id": project_id,
            "apartment_area_m2": area,
            "rooms_count": 1,
            "assumptions": assumptions,
        },
        "rooms": [
            {
                "room_id": room.room_id,
                "type": room.type,
                "area_m2": area,
                "ceiling_height_m": room.ceiling_height_m,
                "geometry": {"contour_mm": [list(p) for p in room.contour_mm]},
                "completeness_checklist": dict(room.completeness_checklist),
                "works": works,
            }
        ],
        "specification": {"materials": materials, "furniture": [], "lighting": []},
        "estimate": {
            "materials_rub": estimate.total_rub,
            "works_rub": 0,
            "total_rub": estimate.total_rub,
            "currency": "RUB",
            "disclaimer": DISCLAIMER,
            "range_rub": [estimate.low_rub, estimate.high_rub],
        },
        "deliverables": {"drawings": [], "schedules": [], "attachments": []},
        "tender": None,
        "audit": {
            "created_by": created_by,
            "approved_by": None,
            "approval_at": None,
            "changelog": [{"version": 1, "at": now, "note": "авто-сборка"}],
        },
    }

"""CPR Generator: сборка ЦПР из входа (rule engine + PIM).

Правило: обязательные операции нельзя молча убирать ради более низкой цены.
"""
from __future__ import annotations

from dataclasses import dataclass, field


@dataclass
class Cpr:
    cpr_id: str
    version: int = 1
    status: str = "draft"
    assumptions: list[str] = field(default_factory=list)
    checklist: dict[str, bool] = field(default_factory=dict)

    def can_approve(self) -> bool:
        return all(self.checklist.values()) if self.checklist else False

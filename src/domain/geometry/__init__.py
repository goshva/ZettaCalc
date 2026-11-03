"""Geometry Service: контур в мм, сцена (1 unit = 1 m)."""
from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True)
class Contour:
    points_mm: list[tuple[float, float]]

    def area_m2(self) -> float:
        # Shoelace
        pts = self.points_mm
        s = 0.0
        for i in range(len(pts)):
            x1, y1 = pts[i]
            x2, y2 = pts[(i + 1) % len(pts)]
            s += x1 * y2 - x2 * y1
        return abs(s) / 2.0 / 1_000_000.0

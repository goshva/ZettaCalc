#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
extract_a_sections.py — извлекает секции главной страницы из «Прототип А.html»
и подменяет inline data:image на /api/images/file/<fname> по md5 уже
сохранённых в API-хранилище файлов.

Позиции секций снимаются по <section ... id="..."> / <footer> / <dialog>,
поэтому файл воспроизводим при повторных запусках.

Запуск:  python vue-prototypes/extract_a_sections.py
"""
import base64
import hashlib
import re
import sys
from pathlib import Path

sys.stdout.reconfigure(encoding="utf-8", errors="replace")

ROOT = Path("C:/soft/ZettaCalc")
SRC = ROOT / "Прототип А.html"
IMG = ROOT / "vue-prototypes" / "api" / "static" / "images"
OUT = ROOT / "vue-prototypes" / "api" / "data" / "a_page"

# ── 1. индекс: md5 байтов картинки -> имя файла в хранилище ────────────────
by_hash: dict[str, str] = {}
for p in sorted(IMG.glob("*.jpg")):
    by_hash.setdefault(hashlib.md5(p.read_bytes()).hexdigest(), p.name)
print(f"хранилище: {len(by_hash)} файлов")

html = SRC.read_text(encoding="utf-8", errors="ignore")

RE_IMG = re.compile(r"data:image/(\w+);base64,([A-Za-z0-9+/=]+)")
stats = {"ok": 0, "miss": 0}


def swap(text: str) -> str:
    def rep(m):
        try:
            raw = base64.b64decode(m.group(2))
        except Exception:
            stats["miss"] += 1
            return m.group(0)
        name = by_hash.get(hashlib.md5(raw).hexdigest())
        if name:
            stats["ok"] += 1
            return f"/api/images/file/{name}"
        stats["miss"] += 1
        return m.group(0)
    return RE_IMG.sub(rep, text)


# ── 2. границы блоков ──────────────────────────────────────────────────────
# (старт-маркер, имя файла); конец = следующий маркер
BLOCKS = [
    ('<section class="section" id="supply"', "supply"),
    ('<section class="center section" id="center"', "center"),
    ('<section class="budget section" id="budget"', "budget"),
    ('<section class="repair section" id="repair"', "repair"),
    ('<section id="project"', "project"),
    ('<section id="companySection"', "company"),
    ('<section class="section" id="faq"', "faq"),
    ('<section class="section final"', "final"),
    ("<footer", "footer"),
    ('<dialog class="modal" id="meetingDialog"', "dialog_meeting"),
    ('<dialog class="modal video-modal" id="videoDialog"', "dialog_video"),
    ('<dialog class="modal" id="savedDialog"', "dialog_saved"),
]

pos = [(html.find(marker), name, marker) for marker, name in BLOCKS]
for start, name, marker in pos:
    if start < 0:
        print(f"  НЕ НАЙДЕН: {name} ({marker[:40]})")

pos = sorted([p for p in pos if p[0] >= 0])
OUT.mkdir(parents=True, exist_ok=True)

for i, (start, name, marker) in enumerate(pos):
    end = pos[i + 1][0] if i + 1 < len(pos) else html.find("<script", start)
    frag = swap(html[start:end])
    # нормализуем переводы строк для читаемости
    (OUT / f"section_{name}.html").write_text(frag.strip(), encoding="utf-8")
    print(f"  {name:16} {len(frag):8} симв.")

print(f"\nкартинок заменено: {stats['ok']}, не найдено: {stats['miss']}")

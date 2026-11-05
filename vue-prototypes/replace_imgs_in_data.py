# -*- coding: utf-8 -*-
"""
replace_imgs_in_data.py — заменяет data:image в извлечённых JSON/HTML на
пути /api/images/file/<fname> по content-hash (md5) уже сохранённых файлов.

Запуск:  python replace_imgs_in_data.py
"""
import base64
import hashlib
import json
import re
import sys
from pathlib import Path

sys.stdout.reconfigure(encoding="utf-8", errors="replace")

DATA = Path("C:/soft/ZettaCalc/vue-prototypes/api/data")
IMG = Path("C:/soft/ZettaCalc/vue-prototypes/api/static/images")

# ── 1. Индекс: md5(файла) -> имя ───────────────────────────────────────────
print("Хеширование картинок на диске...")
by_hash: dict[str, str] = {}
for p in sorted(IMG.glob("*.jpg")):
    by_hash[hashlib.md5(p.read_bytes()).hexdigest()] = p.name
print(f"  {len(by_hash)} файлов")

RE_IMG = re.compile(r"data:image/(\w+);base64,([A-Za-z0-9+/=]+)")

stats = {"replaced": 0, "missing": 0}

def replace_in_text(text: str) -> str:
    def rep(m):
        b64 = m.group(2)
        try:
            raw = base64.b64decode(b64)
        except Exception:
            stats["missing"] += 1
            return m.group(0)
        fname = by_hash.get(hashlib.md5(raw).hexdigest())
        if fname:
            stats["replaced"] += 1
            return f"/api/images/file/{fname}"
        stats["missing"] += 1
        return m.group(0)
    return RE_IMG.sub(rep, text)

# ── 2. Обработка файлов ────────────────────────────────────────────────────
targets = [
    DATA / "proto_a_hotspotData60.json",
    DATA / "proto_a_extraAnalogues64.json",
    *sorted((DATA / "apps52").glob("*.html")),
    *sorted((DATA / "apps52").glob("*/*.json")),
]

for f in targets:
    if not f.exists():
        print(f"  пропуск (нет): {f.name}")
        continue
    before = stats["replaced"]
    txt = f.read_text(encoding="utf-8")
    new = replace_in_text(txt)
    if new != txt:
        f.write_text(new, encoding="utf-8")
    n = stats["replaced"] - before
    print(f"  {f.relative_to(DATA)}: заменено {n}")

print(f"\nИтого: заменено {stats['replaced']}, не найдено {stats['missing']}")

# ── 3. Структура датасетов ─────────────────────────────────────────────────
print("\n=== СТРУКТУРА ДАННЫХ ===")

def describe(obj, prefix="  ", depth=0):
    if depth > 2:
        return
    if isinstance(obj, dict):
        for k, v in list(obj.items())[:15]:
            if isinstance(v, list):
                extra = ""
                if v and isinstance(v[0], dict):
                    extra = " [0]: " + ",".join(list(v[0].keys())[:12])
                print(f"{prefix}{k}: list[{len(v)}]{extra}")
            elif isinstance(v, dict):
                print(f"{prefix}{k}: dict({len(v)})")
                describe(v, prefix + "  ", depth + 1)
            else:
                print(f"{prefix}{k}: {json.dumps(v, ensure_ascii=False)[:70]}")

for rel in ["apps52/interiors/dataset.json", "apps52/catalog/data.json",
            "apps52/center/photos.json", "apps52/center/interiors.json",
            "apps52/interiors/adaptdata.json"]:
    f = DATA / rel
    if not f.exists():
        continue
    print(f"\n--- {rel} ---")
    obj = json.loads(f.read_text(encoding="utf-8"))
    describe(obj)

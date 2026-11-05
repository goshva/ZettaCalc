# -*- coding: utf-8 -*-
"""
Извлекает JSON-блоки прототипа А и заменяет data:image на пути.
Использует positional mapping: порядок data:image в JSON совпадает с порядком
в исходном HTML.

Уже есть 669 картинок на диске. Нужно только сопоставить их по порядку.
"""
import base64
import json
import re
import sys
from pathlib import Path

sys.stdout = sys.stdout

ROOT = Path("C:/soft/ZettaCalc")
SRC = ROOT / "Прототип А.html"
DATA_DIR = ROOT / "vue-prototypes/api/data"
DATA_DIR.mkdir(parents=True, exist_ok=True)

print("Чтение HTML...")
html = SRC.read_text(encoding="utf-8")
print(f"  Размер: {len(html):,} chars")

# ── Порядок data:image в HTML ──────────────────────────────────────────────
# Картинки из extract_data.py сохранены в порядке:
#   a_interiors_img_0001..0262 (из interiors.html)
#   a_catalog_img_0001..0102
#   a_center_img_0001..0044
#   a_repair_img_0001..0006
#   a_about_img_0001..0003
#   a_main_img_0001..0070 (вне JSON)
#   b_img_0001..0143 (прототип B)
#   + 39 именованных

# Строим глобальный порядок data:image из HTML
all_imgs = re.findall(r"data:image/\w+;base64,([A-Za-z0-9+/=]+)", html)
print(f"  data:image всего: {len(all_imgs)}")

# Распределяем по блокам:
# Сначала идут картинки в apps52 приложениях
# Потом hotspotData60, extraAnalogues64
# Потом "вне JSON"

# apps52
raw_apps = re.search(r'<script[^>]*id="apps52"[^>]*>(.*?)</script>', html, re.S)
apps = json.loads(raw_apps.group(1))

# Считаем картинки в каждом приложении
app_img_counts = {}
app_img_ranges = {}  # (start, end) для каждого приложения
pos = 0
for app_name in ["interiors", "catalog", "center", "repair", "about"]:
    h = apps[app_name].get("html", "")
    count = len(re.findall(r"data:image/\w+;base64,", h))
    app_img_counts[app_name] = count
    if count > 0:
        app_img_ranges[app_name] = (pos, pos + count)
    pos += count

print(f"  apps52 картинки: {app_img_counts}")

# hotspotData60
hm = re.search(r'<script[^>]*\bid="hotspotData60"[^>]*>(.*?)</script>', html, re.S)
hotspot_img_count = len(re.findall(r"data:image/\w+;base64,", hm.group(1))) if hm else 0
print(f"  hotspotData60 images: {hotspot_img_count}")

# extraAnalogues64
em = re.search(r'<script[^>]*\bid="extraAnalogues64"[^>]*>(.*?)</script>', html, re.S)
extra_img_count = len(re.findall(r"data:image/\w+;base64,", em.group(1))) if em else 0
print(f"  extraAnalogues64 images: {extra_img_count}")

# Вне JSON = всё остальное
outside_count = len(all_imgs) - pos - hotspot_img_count - extra_img_count
print(f"  вне JSON: {outside_count}")

# ── Mapping: index_in_block -> filename ───────────────────────────────────
def build_mapping(block_name: str, start_idx: int, count: int) -> dict:
    """Возвращает {b64_prefix: fname} для картинок блока."""
    mapping = {}
    for i in range(count):
        global_idx = start_idx + i
        if global_idx < len(all_imgs):
            b64 = all_imgs[global_idx]
            # Определяем имя файла
            if block_name in ("interiors", "catalog", "center", "repair", "about"):
                fname = f"a_{block_name}_img_{i+1:04d}.jpg"
            elif block_name == "main_outside":
                fname = f"a_main_img_{i+1:04d}.jpg"
            else:
                fname = f"{block_name}_img_{i+1:04d}.jpg"
            if (Path("C:/soft/ZettaCalc/vue-prototypes/api/static/images") / fname).exists():
                mapping[b64[:64]] = fname
    return mapping

# Строим mapping для каждого блока
mapping = {}
# interiors: starts at 0
mapping.update(build_mapping("interiors", 0, app_img_counts.get("interiors", 0)))
# catalog: starts after interiors
catalog_start = app_img_counts.get("interiors", 0)
mapping.update(build_mapping("catalog", catalog_start, app_img_counts.get("catalog", 0)))
# center: starts after catalog
center_start = catalog_start + app_img_counts.get("catalog", 0)
mapping.update(build_mapping("center", center_start, app_img_counts.get("center", 0)))
# repair
repair_start = center_start + app_img_counts.get("center", 0)
mapping.update(build_mapping("repair", repair_start, app_img_counts.get("repair", 0)))
# about
about_start = repair_start + app_img_counts.get("repair", 0)
mapping.update(build_mapping("about", about_start, app_img_counts.get("about", 0)))
# outside JSON
outside_global_start = about_start + app_img_counts.get("about", 0)
mapping.update(build_mapping("main_outside", outside_global_start, outside_count))

# hotspotData60
hotspot_global_start = outside_global_start + outside_count
mapping.update(build_mapping("hotspotData60", hotspot_global_start, hotspot_img_count))
# extraAnalogues64
extra_global_start = hotspot_global_start + hotspot_img_count
mapping.update(build_mapping("extraAnalogues64", extra_global_start, extra_img_count))

print(f"  Mapping size: {len(mapping)}")

# ── Заменa data:image в JSON на пути ──────────────────────────────────────
def replace_imgs(text: str) -> str:
    def rep(m):
        b64 = m.group(1)
        fname = mapping.get(b64[:64])
        if fname:
            return f"/api/images/{fname}"
        return f"/api/images/placeholder.jpg"
    return re.sub(r"data:image/\w+;base64,([A-Za-z0-9+/=]+)", rep, text)

# ── Извлечение JSON-блоков ────────────────────────────────────────────────
BLOCK_IDS = ["hotspotData60", "extraAnalogues64", "apps52",
             "adaptdata", "dataset", "data", "photos", "interiors"]

for bid in BLOCK_IDS:
    # plain-кавычки
    m = re.search(r'<script[^>]*\bid="%s"[^>]*>(.*?)</script>' % re.escape(bid), html, re.S)
    if m:
        content = m.group(1)
    else:
        # экранированные
        m = re.search(r'<script[^>]*\bid=\\"%s\\"[^>]*>(.*?)</script>' % re.escape(bid), html, re.S)
        if not m:
            print(f"  SKIPPED: {bid}")
            continue
        content = m.group(1).replace('\\"', '"')

    try:
        obj = json.loads(content)
        # Заменяем data:image
        dumped = json.dumps(obj, ensure_ascii=False)
        dumped = replace_imgs(dumped)
        obj = json.loads(dumped)
        out = DATA_DIR / f"proto_a_{bid}.json"
        out.write_text(json.dumps(obj, ensure_ascii=False, indent=2), encoding="utf-8")
        print(f"  SAVED: {bid} -> {out.name}")
    except json.JSONDecodeError as e:
        print(f"  JSON ERROR: {bid}: {e}")

# apps52 — сохраняем html/css приложений
apps_dir = DATA_DIR / "apps52"
apps_dir.mkdir(exist_ok=True)
for app_name, app in apps.items():
    h = app.get("html", "")
    c = app.get("css", "")
    if h:
        (apps_dir / f"{app_name}.html").write_text(replace_imgs(h), encoding="utf-8")
    if c:
        (apps_dir / f"{app_name}.css").write_text(c, encoding="utf-8")
    print(f"  APP: {app_name} html={len(h):,} css={len(c):,}")

print("\nDONE")

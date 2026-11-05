# -*- coding: utf-8 -*-
"""Извлекает ВСЕ JSON-блоки из прототипа А, включая вложенные (из HTML приложений)."""
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

# ── Топ-уровень ────────────────────────────────────────────────────────────
print("\n=== Top-level JSON ===")

def extract_json_block(html, bid):
    """Извлекает JSON по id, обрабатывая экранированные кавычки."""
    # plain
    m = re.search(r'<script[^>]*\bid="%s"[^>]*>(.*?)</script>' % re.escape(bid), html, re.S)
    if m:
        return m.group(1)
    # escaped
    m = re.search(r'<script[^>]*\bid=\\"%s\\"[^>]*>(.*?)</script>' % re.escape(bid), html, re.S)
    if m:
        return m.group(1).replace('\\"', '"')
    return None

top_ids = ["hotspotData60", "extraAnalogues64", "apps52", "adaptdata", "dataset", "data", "photos", "interiors"]
for bid in top_ids:
    content = extract_json_block(html, bid)
    if content:
        try:
            obj = json.loads(content)
            out = DATA_DIR / f"proto_a_{bid}.json"
            out.write_text(json.dumps(obj, ensure_ascii=False, indent=2), encoding="utf-8")
            kind = type(obj).__name__
            info = ""
            if isinstance(obj, dict):
                info = f"keys={list(obj.keys())[:10]}"
            elif isinstance(obj, list):
                info = f"list[{len(obj)}]"
                if obj and isinstance(obj[0], dict):
                    info += f" [0]keys={list(obj[0].keys())[:10]}"
            print(f"  {bid}: {kind} {info}")
        except json.JSONDecodeError as e:
            print(f"  {bid}: JSON ERROR {e}")

# ── apps52 приложения ─────────────────────────────────────────────────────
print("\n=== apps52 ===")
raw_apps = extract_json_block(html, "apps52")
if raw_apps:
    apps = json.loads(raw_apps)
    apps_dir = DATA_DIR / "apps52"
    apps_dir.mkdir(exist_ok=True)
    
    for app_name, app in apps.items():
        h = app.get("html", "")
        c = app.get("css", "")
        if h:
            (apps_dir / f"{app_name}.html").write_text(h, encoding="utf-8")
        if c:
            (apps_dir / f"{app_name}.css").write_text(c, encoding="utf-8")
        print(f"  {app_name}: html={len(h):,} css={len(c):,}")
        
        # Вложенные JSON внутри HTML приложения
        nested = {}
        for m in re.finditer(
            r'<script[^>]*\bid=["\']([^\s"\']+)["\'][^>]*>(.*?)</script>',
            h, re.S | re.I
        ):
            nid = m.group(1)
            ncontent = m.group(2).strip()
            if not ncontent.startswith(('{', '[')):
                continue
            try:
                obj = json.loads(ncontent)
                nested[nid] = obj
            except json.JSONDecodeError:
                nested[nid] = {"_raw": ncontent[:500]}
        
        if nested:
            nested_dir = apps_dir / app_name
            nested_dir.mkdir(exist_ok=True)
            for nid, obj in nested.items():
                (nested_dir / f"{nid}.json").write_text(
                    json.dumps(obj, ensure_ascii=False, indent=2), encoding="utf-8"
                )
                kind = type(obj).__name__
                info = ""
                if isinstance(obj, dict):
                    info = f"keys={list(obj.keys())[:10]}"
                elif isinstance(obj, list):
                    info = f"list[{len(obj)}]"
                    if obj and isinstance(obj[0], dict):
                        info += f" [0]keys={list(obj[0].keys())[:10]}"
                print(f"    nested {nid}: {kind} {info}")

print("\nDONE")

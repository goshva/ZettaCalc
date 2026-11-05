# -*- coding: utf-8 -*-
"""
extract_main_page.py — извлекает разметку ГЛАВНОЙ страницы прототипа А из
<body> (приложение main имеет html=0, т.к. живёт прямо в документе),
разбивает на секции и сохраняет для переноса в Vue.
"""
import json
import re
import sys
from pathlib import Path

sys.stdout.reconfigure(encoding="utf-8", errors="replace")

ROOT = Path("C:/soft/ZettaCalc")
SRC = ROOT / "Прототип А.html"
DATA = ROOT / "vue-prototypes/api/data"
OUT = DATA / "main_page"
OUT.mkdir(parents=True, exist_ok=True)

html = SRC.read_text(encoding="utf-8")

# ── center/photos.json и center/interiors.json (списки) ────────────────────
for rel in ["apps52/center/photos.json", "apps52/center/interiors.json"]:
    f = DATA / rel
    obj = json.loads(f.read_text(encoding="utf-8"))
    print(f"--- {rel}: list[{len(obj)}] ---")
    for i, item in enumerate(obj[:4]):
        if isinstance(item, dict):
            print(f"  [{i}] " + json.dumps({k: (v[:60] + '…' if isinstance(v, str) and len(v) > 60 else v) for k, v in item.items()}, ensure_ascii=False)[:300])
    if obj and isinstance(obj[0], dict) and len(obj[0].keys()) > 0:
        print(f"  keys: {list(obj[0].keys())}")
    print()

# ── Тело документа: от <body> до первого <script> ─────────────────────────
body_m = re.search(r"<body[^>]*>(.*)", html, re.S)
body = body_m.group(1)
# обрезаем по первому крупному script (JSON-блоки/роутер)
cut = body.find("<script")
body_main = body[:cut] if cut > 0 else body
print(f"body до первого script: {len(body_main):,} chars")

# ── Секции: <section id="..."> ... </section> ─────────────────────────────
sections = []
for m in re.finditer(r"<(section|header|footer|nav|dialog|div)\b([^>]*\bid=\"[\w-]+\"[^>]*)>", body_main):
    tag = m.group(1)
    attrs = m.group(2)
    idm = re.search(r'id="([\w-]+)"', attrs)
    if not idm:
        continue
    sid = idm.group(1)
    # ищем закрывающий тег того же уровня (грубо — следующий <section|header|footer ... id=)
    sections.append((tag, sid, m.start()))

print(f"Элементов с id в body: {len(sections)}")
for tag, sid, pos in sections:
    print(f"  <{tag} id=\"{sid}\"> @ {pos:,}")

# ── Режем body на куски по этим позициям ──────────────────────────────────
pieces = {}
for i, (tag, sid, pos) in enumerate(sections):
    end = sections[i + 1][2] if i + 1 < len(sections) else len(body_main)
    pieces[sid] = body_main[pos:end]

# ── Заменяем data:image на пути ────────────────────────────────────────────
import base64, hashlib
IMG = ROOT / "vue-prototypes/api/static/images"
by_hash = {hashlib.md5(p.read_bytes()).hexdigest(): p.name for p in IMG.glob("*.jpg")}
RE_IMG = re.compile(r"data:image/(\w+);base64,([A-Za-z0-9+/=]+)")
rep_n = [0, 0]

def replace_imgs(text):
    def rep(m):
        try:
            raw = base64.b64decode(m.group(2))
        except Exception:
            rep_n[1] += 1
            return m.group(0)
        fname = by_hash.get(hashlib.md5(raw).hexdigest())
        if fname:
            rep_n[0] += 1
            return f"/api/images/file/{fname}"
        rep_n[1] += 1
        return m.group(0)
    return RE_IMG.sub(rep, text)

manifest = []
for sid, chunk in pieces.items():
    clean = replace_imgs(chunk)
    (OUT / f"{sid}.html").write_text(clean, encoding="utf-8")
    # заголовок секции
    h = re.search(r"<h([123])[^>]*>(.*?)</h\1>", clean, re.S)
    title = re.sub(r"<[^>]+>", " ", h.group(2)).strip()[:70] if h else ""
    manifest.append({"id": sid, "chars": len(clean), "title": title})

(OUT / "_manifest.json").write_text(json.dumps(manifest, ensure_ascii=False, indent=2), encoding="utf-8")
print(f"\nСохранено секций: {len(pieces)} → {OUT}")
print(f"Картинки: заменено {rep_n[0]}, не найдено {rep_n[1]}")
print("\nМанифест:")
for m_ in manifest:
    print(f"  {m_['id']:28} {m_['chars']:>8,}  {m_['title']}")

# -*- coding: utf-8 -*-
"""
extract_a_main.py — извлекает ГЛАВНУЮ прототипа А:
  · shell (shellbar + shell-header + note) — 3.7 КБ
  · <main id="main"> … </main> — 4.6 МБ (hero с фото + 5 секций)
  · футер и диалоги после main
data:image → /api/images/file/<имя>.jpg по md5.
"""
import base64
import hashlib
import re
import sys
from pathlib import Path

sys.stdout.reconfigure(encoding="utf-8", errors="replace")

ROOT = Path("C:/soft/ZettaCalc")
SRC = ROOT / "Прототип А.html"
OUT = ROOT / "vue-prototypes/api/data/a_page"
OUT.mkdir(parents=True, exist_ok=True)

html = SRC.read_text(encoding="utf-8")
body = re.search(r"<body[^>]*>(.*)", html, re.S).group(1)
bm = body[: body.find("<script")]
print(f"body_main: {len(bm):,}")

IMG = ROOT / "vue-prototypes/api/static/images"
by_hash = {hashlib.md5(p.read_bytes()).hexdigest(): p.name for p in IMG.glob("*.jpg")}
RE_IMG = re.compile(r"data:image/(\w+);base64,([A-Za-z0-9+/=]+)")
stats = [0, 0]
missing = []

def replace_imgs(text):
    def rep(m):
        try:
            raw = base64.b64decode(m.group(2))
        except Exception:
            stats[1] += 1
            return m.group(0)
        h = hashlib.md5(raw).hexdigest()
        fname = by_hash.get(h)
        if fname:
            stats[0] += 1
            return f"/api/images/file/{fname}"
        stats[1] += 1
        missing.append(h[:12])
        return m.group(0)
    return RE_IMG.sub(rep, text)

bm = replace_imgs(bm)
print(f"картинки: заменено {stats[0]}, не найдено {stats[1]}")
if missing:
    print("первые недостающие:", missing[:8])

# ── режем ──────────────────────────────────────────────────────────────────
i_main = bm.find("<main")
shell = bm[:i_main]
rest = bm[i_main:]
i_end = rest.find("</main>")
if i_end < 0:
    # </main> нет — режем по первому <footer> или берём до конца
    i_foot = rest.find("<footer")
    main_html = rest if i_foot < 0 else rest[:i_foot]
    tail = "" if i_foot < 0 else rest[i_foot:]
else:
    main_html = rest[: i_end + len("</main>")]
    tail = rest[i_end + len("</main>"):]

(OUT / "shell.html").write_text(shell, encoding="utf-8")
(OUT / "main.html").write_text(main_html, encoding="utf-8")
if tail.strip():
    (OUT / "tail.html").write_text(tail, encoding="utf-8")

print(f"shell: {len(shell):,}")
print(f"main:  {len(main_html):,}")
print(f"tail:  {len(tail):,}")

# сколько картинок в main
print("img в main:", main_html.count("/api/images/file/"))
print("img в shell:", shell.count("/api/images/file/"))

# секции main
for m in re.finditer(r'<section\b[^>]*class="([^"]*)"[^>]*>', main_html):
    print("  section:", m.group(1))

# -*- coding: utf-8 -*-
"""
extract_a_full.py — полная разметка главной А: склеивает всё содержимое <body>
без <script>-блоков (разметка идёт и до, и после роутера: hero, секции,
футер, диалоги). data:image → /api/images/file/<имя>.jpg.
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
body = body[: body.rfind("</body>")]
print(f"body: {len(body):,}")

# вырезаем все <script…>…</script>
parts = []
pos = 0
n_scripts = 0
for m in re.finditer(r"<script\b[^>]*>.*?</script>", body, re.S):
    parts.append(body[pos:m.start()])
    pos = m.end()
    n_scripts += 1
parts.append(body[pos:])
markup = "".join(parts)
print(f"вырезано script-блоков: {n_scripts}")
print(f"чистая разметка: {len(markup):,}")

# картинки
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

markup = replace_imgs(markup)
print(f"картинки: заменено {stats[0]}, не найдено {stats[1]}")
if missing:
    print("недостающие хэши:", missing[:10])

(OUT / "full.html").write_text(markup, encoding="utf-8")
print(f"→ {OUT / 'full.html'}")

# ── режем на куски для компонентов Vue ────────────────────────────────────
def cut_between(text, start_marker, end_marker, from_pos=0):
    i = text.find(start_marker, from_pos)
    if i < 0:
        return None, from_pos
    j = text.find(end_marker, i)
    if j < 0:
        return text[i:], len(text)
    return text[i:j], j

pieces = {}
# 1. shellbar
i_shellbar = markup.find('<div class="shellbar">')
i_shell_hdr = markup.find('<header class="shell-header">')
i_note = markup.find('<div class="prototype-note51">')
i_surface = markup.find('<div id="surface52">')
i_inner_hdr = markup.find("<header><div class=\"wrap nav\">")
i_main = markup.find("<main id=\"main\">")

if i_main < 0:
    i_main = markup.find("<main")
i_end_main = markup.find("</main>")
if i_end_main >= 0:
    i_end_main += len("</main>")
i_footer = markup.find("<footer")

print(f"\nshellbar@{i_shellbar} shell-header@{i_shell_hdr} note@{i_note} surface@{i_surface} inner-header@{i_inner_hdr} main@{i_main} main-end@{i_end_main} footer@{i_footer}")

segs = [
    ("shellbar",  i_shellbar, i_shell_hdr),
    ("shell_header", i_shell_hdr, i_note),
    ("proto_note", i_note, i_surface),
    ("surface_head", i_surface, i_inner_hdr),   # skip + proto-bar
    ("inner_header", i_inner_hdr, i_main),
    ("main", i_main, i_end_main),
    ("after_main", i_end_main, i_footer if i_footer > 0 else len(markup)),
    ("footer", i_footer, markup.rfind("</footer>") + len("</footer>") if i_footer > 0 else i_footer),
]
for name, a, b in segs:
    if a is None or a < 0 or b is None or b <= a:
        print(f"  {name:14} ПРОПУСК (a={a}, b={b})")
        continue
    chunk = markup[a:b]
    (OUT / f"{name}.html").write_text(chunk, encoding="utf-8")
    print(f"  {name:14} {len(chunk):>9,} chars  imgs={chunk.count('/api/images/file/')}")

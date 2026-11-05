# -*- coding: utf-8 -*-
"""
extract_b_page.py — извлекает полную статичную разметку прототипа В
(shell + 17 секций + футер + диалоги) и сохраняет кусками для переноса в Vue.
"""
import base64
import hashlib
import json
import re
import sys
from pathlib import Path

sys.stdout.reconfigure(encoding="utf-8", errors="replace")

ROOT = Path("C:/soft/ZettaCalc")
SRC = ROOT / "Прототип В.html"
OUT = ROOT / "vue-prototypes/api/data/b_page"
OUT.mkdir(parents=True, exist_ok=True)

html = SRC.read_text(encoding="utf-8")
body = re.search(r"<body[^>]*>(.*?)<script", html, re.S).group(1)
print(f"body: {len(body):,} chars")

# ── картинки: IMG-словарь в JS + data:image в разметке ─────────────────────
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
        fname = by_hash.get(hashlib.md5(raw).hexdigest())
        if fname:
            stats[0] += 1
            return f"/api/images/file/{fname}"
        stats[1] += 1
        missing.append(hashlib.md5(raw).hexdigest()[:12])
        return m.group(0)
    return RE_IMG.sub(rep, text)

body = replace_imgs(body)
print(f"картинки: заменено {stats[0]}, не найдено {stats[1]}")

# ── режем по секциям/шапке/футеру ──────────────────────────────────────────
TAG_RE = re.compile(r"<(section|header|footer|nav|aside|div|dialog)\b([^>]*)>")

def attr(attrs, name):
    m = re.search(r'\b%s\s*=\s*"([^"]*)"' % name, attrs)
    return m.group(1) if m else ""

# блоки верхнего уровня: режем по <section>/<header>/<footer> с учётом вложенности section
marks = []
for m in re.finditer(r"<(section|header|footer)\b([^>]*)>", body):
    marks.append((m.start(), m.group(1), m.group(2)))

# всё до первого mark — shell
if marks:
    shell = body[: marks[0][0]]
    (OUT / "00_shell.html").write_text(shell, encoding="utf-8")
    print(f"00_shell.html: {len(shell):,}")

pieces = []
for i, (pos, tag, attrs) in enumerate(marks):
    end = marks[i + 1][0] if i + 1 < len(marks) else len(body)
    chunk = body[pos:end]
    sid = attr(attrs, "id") or f"{tag}{i}"
    fname = f"{i+1:02d}_{tag}_{sid}.html"
    (OUT / fname).write_text(chunk, encoding="utf-8")
    hs = re.findall(r"<h([123])[^>]*>(.*?)</h\1>", chunk, re.S)
    title = re.sub(r"<[^>]+>", " ", hs[0][1]).strip()[:60] if hs else ""
    nimg = chunk.count("/api/images/file/")
    pieces.append({"file": fname, "tag": tag, "id": attr(attrs, "id"),
                   "class": attr(attrs, "class"), "chars": len(chunk),
                   "imgs": nimg, "h": title})

(OUT / "_manifest.json").write_text(json.dumps(pieces, ensure_ascii=False, indent=2), encoding="utf-8")
print(f"\n{'файл':34} {'размер':>9} {'img':>4}  заголовок")
for p in pieces:
    print(f"{p['file']:34} {p['chars']:>9,} {p['imgs']:>4}  {p['h']}")
if missing:
    print("\nне найденные хэши:", missing[:10])

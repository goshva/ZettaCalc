# -*- coding: utf-8 -*-
"""Извлекает секции главной А: pairing по <section>/<header>/<footer>/<dialog>."""
import json
import re
import sys
import base64
import hashlib
from pathlib import Path

sys.stdout.reconfigure(encoding="utf-8", errors="replace")

ROOT = Path("C:/soft/ZettaCalc")
SRC = ROOT / "Прототип А.html"
DATA = ROOT / "vue-prototypes/api/data"
OUT = DATA / "main_page"
OUT.mkdir(parents=True, exist_ok=True)

html = SRC.read_text(encoding="utf-8")
body_m = re.search(r"<body[^>]*>(.*)", html, re.S)
body = body_m.group(1)
cut = body.find("<script")
body_main = body[:cut] if cut > 0 else body
print(f"body_main: {len(body_main):,} chars")

# ── pairing для block-тегов ───────────────────────────────────────────────
BLOCK_TAGS = ["section", "header", "footer", "dialog", "nav"]
TAG_RE = re.compile(r"<(/?)(section|header|footer|dialog|nav)\b([^>]*)>", re.I)

found = []          # (tag, attrs, start, end)
open_stack = []     # (tag, attrs, start)

for m in TAG_RE.finditer(body_main):
    closing, tag, attrs = m.group(1), m.group(2).lower(), m.group(3)
    if not closing:
        open_stack.append((tag, attrs, m.start()))
    else:
        # закрываем ближайший соответствующий
        for i in range(len(open_stack) - 1, -1, -1):
            if open_stack[i][0] == tag:
                t, a, s = open_stack.pop(i)
                found.append((t, a, s, m.end()))
                del open_stack[i + 1:]  # незакрытые внутренности теряем
                break

found.sort(key=lambda x: x[2])
print(f"Спаренных block-элементов: {len(found)}")

# ── картинки ───────────────────────────────────────────────────────────────
IMG = ROOT / "vue-prototypes/api/static/images"
by_hash = {hashlib.md5(p.read_bytes()).hexdigest(): p.name for p in IMG.glob("*.jpg")}
RE_IMG = re.compile(r"data:image/(\w+);base64,([A-Za-z0-9+/=]+)")
stats = [0, 0]

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
        return m.group(0)
    return RE_IMG.sub(rep, text)

def attr(attrs, name):
    m = re.search(r'\b%s\s*=\s*"([^"]*)"' % name, attrs)
    return m.group(1) if m else ""

manifest = []
for i, (tag, attrs, start, end) in enumerate(found):
    sid = attr(attrs, "id")
    scls = attr(attrs, "class")
    label = sid or (scls.split()[0] if scls else tag)
    chunk = replace_imgs(body_main[start:end])
    fname = f"{i:02d}_{tag}_{label}.html"
    (OUT / fname).write_text(chunk, encoding="utf-8")
    hs = re.findall(r"<h([123])[^>]*>(.*?)</h\1>", chunk, re.S)
    title = re.sub(r"<[^>]+>", " ", hs[0][1]).strip()[:70] if hs else ""
    nimg = chunk.count("/api/images/file/")
    manifest.append({"file": fname, "tag": tag, "id": sid, "class": scls,
                     "chars": len(chunk), "imgs": nimg, "h": title})

(OUT / "_manifest.json").write_text(json.dumps(manifest, ensure_ascii=False, indent=2), encoding="utf-8")
print(f"Картинки: заменено {stats[0]}, не найдено {stats[1]}\n")
print(f"{'файл':42} {'размер':>10} {'img':>4}  заголовок")
for m_ in manifest:
    print(f"{m_['file']:42} {m_['chars']:>10,} {m_['imgs']:>4}  {m_['h'][:58]}")

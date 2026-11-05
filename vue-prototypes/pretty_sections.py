"""Разбивает минифицированные HTML-секции на строки для удобного чтения."""
import re
from pathlib import Path

SRC = Path(__file__).parent / "api" / "data" / "a_page"
DST = SRC / "pretty"
DST.mkdir(exist_ok=True)

for f in sorted(SRC.glob("section_*.html")) + [SRC / "main.html", SRC / "shell.html"]:
    html = f.read_text(encoding="utf-8")
    # Перенос строки перед каждым открывающим/закрывающим тегом
    out = re.sub(r"><", ">\n<", html)
    (DST / f.name).write_text(out, encoding="utf-8")
    print(f.name, len(out.splitlines()), "lines")

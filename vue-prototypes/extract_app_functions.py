"""Вырезает именованные функции из оригинального app.js прототипа А.

Нужно, чтобы переносить разметку экранов «Интерьеров» в Vue по частям:
каждая функция пишется отдельным файлом в out_dir.
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

SRC = Path(__file__).parent.parent / "unpacked" / "Прототип А" / "js" / "app.js"
OUT = Path(__file__).parent / "_extracted"

NAMES = sys.argv[1:] or [
    "crumb", "band", "card", "hub", "listing", "results", "category",
    "detail", "gallery", "estimate", "products", "variants", "apartment",
    "favs", "uploadPage", "listing", "updateResults", "lineRows",
]

# В оригинале часть экранов объявлена дважды (каталог и интерьеры),
# поэтому сохраняем все вхождения с индексом: card.js, card.2.js и т.д.
DECL = r"(?:^|[\n;])(?:function\s+{n}\s*\(|(?:const|let|var)?\s*{n}\s*=\s*(?:async\s*)?function\s*\()"


def extract_all(src: str, name: str) -> list[str]:
    found: list[str] = []
    for m in re.finditer(DECL.format(n=re.escape(name)), src, re.M):
        brace = src.index("{", m.start())
        depth = 0
        for i in range(brace, len(src)):
            ch = src[i]
            if ch == "{":
                depth += 1
            elif ch == "}":
                depth -= 1
                if depth == 0:
                    found.append(src[m.start():i + 1])
                    break
    return found


def main() -> None:
    src = SRC.read_text(encoding="utf-8")
    OUT.mkdir(exist_ok=True)
    for name in NAMES:
        bodies = extract_all(src, name)
        if not bodies:
            print(f"не найдено: {name}")
            continue
        for n, body in enumerate(bodies, start=1):
            suffix = "" if n == 1 else f".{n}"
            (OUT / f"{name}{suffix}.js").write_text(body, encoding="utf-8")
        print(f"{name}: версий {len(bodies)}, размеры {[len(b) for b in bodies]}")


if __name__ == "__main__":
    main()
